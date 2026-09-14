use axum::{
    body::Body,
    http::{header, StatusCode, Uri},
    response::{IntoResponse, Response},
    routing::get,
    Router,
};
use include_dir::{include_dir, Dir};
use std::sync::{
    atomic::{AtomicBool, Ordering},
    Arc, Mutex,
};
use tokio::sync::oneshot;

// The same Vite output as Tauri's asset protocol, embedded at compile time.
// Serving never reads the checkout, user data or arbitrary filesystem paths.
static ASSETS: Dir<'_> = include_dir!("$CARGO_MANIFEST_DIR/../dist");
#[cfg(dev)]
pub const FOCUS_URL: &str = "http://localhost:17332";
#[cfg(not(dev))]
pub const FOCUS_URL: &str = "http://localhost:7332";

const WEB_PORT: u16 = if cfg!(dev) { 17332 } else { 7332 };

pub struct WebServer {
    running: Arc<AtomicBool>,
    error: Arc<Mutex<Option<String>>>,
    shutdown: Mutex<Option<oneshot::Sender<()>>>,
    task: Mutex<Option<tauri::async_runtime::JoinHandle<()>>>,
}

impl WebServer {
    pub fn start() -> Self {
        let mut server = Self {
            running: Arc::new(AtomicBool::new(false)),
            error: Arc::new(Mutex::new(None)),
            shutdown: Mutex::new(None),
            task: Mutex::new(None),
        };
        let listener = match std::net::TcpListener::bind(("0.0.0.0", WEB_PORT)) {
            Ok(listener) => listener,
            Err(error) => {
                *server.error.lock().unwrap() = Some(format!(
                    "Cannot serve Focus on port {WEB_PORT}: {error}. Close the program using this port and restart KLM Desktop."
                ));
                return server;
            }
        };
        if let Err(error) = listener.set_nonblocking(true) {
            *server.error.lock().unwrap() = Some(format!("Cannot start Focus: {error}"));
            return server;
        }
        let (tx, rx) = oneshot::channel();
        *server.shutdown.get_mut().unwrap() = Some(tx);
        let running = server.running.clone();
        let error = server.error.clone();
        *server.task.get_mut().unwrap() = Some(tauri::async_runtime::spawn(async move {
            let result = async {
                let listener = tokio::net::TcpListener::from_std(listener)?;
                running.store(true, Ordering::Release);
                axum::serve(listener, Router::new().fallback_service(get(asset)))
                    .with_graceful_shutdown(async { let _ = rx.await; })
                    .await
            }.await;
            running.store(false, Ordering::Release);
            if let Err(cause) = result {
                *error.lock().unwrap() = Some(format!("Focus server stopped: {cause}. Restart KLM Desktop."));
            }
        }));
        server
    }

    pub fn error(&self) -> Option<String> {
        self.error.lock().unwrap().clone()
    }

    pub fn ready(&self) -> Result<(), String> {
        if let Some(error) = self.error() { return Err(error); }
        if !self.running.load(Ordering::Acquire) {
            return Err("Focus server is not ready. Retry in a moment.".into());
        }
        Ok(())
    }

    pub fn stop(&self) {
        self.running.store(false, Ordering::Release);
        if let Some(tx) = self.shutdown.lock().unwrap().take() { let _ = tx.send(()); }
        if let Some(mut task) = self.task.lock().unwrap().take() {
            // Bound slow HTTP peers while still requesting graceful server shutdown.
            tauri::async_runtime::block_on(async {
                if tokio::time::timeout(std::time::Duration::from_secs(2), &mut task).await.is_err() {
                    task.abort();
                }
            });
        }
    }
}

async fn asset(uri: Uri) -> Response {
    let decoded = match percent_encoding::percent_decode_str(uri.path()).decode_utf8() {
        Ok(path) => path,
        Err(_) => return StatusCode::BAD_REQUEST.into_response(),
    };
    if decoded.contains(['\\', '\0']) || decoded.split('/').any(|part| part == ".." || part == ".") {
        return StatusCode::BAD_REQUEST.into_response();
    }
    let path = decoded.trim_start_matches('/');
    let file = ASSETS.get_file(if path.is_empty() { "index.html" } else { path })
        .or_else(|| {
            // SPA routes fall back to index; missing chunks/images remain real 404s.
            if std::path::Path::new(path).extension().is_none() { ASSETS.get_file("index.html") } else { None }
        });
    let Some(file) = file else { return StatusCode::NOT_FOUND.into_response(); };
    let mime = mime_guess::from_path(file.path()).first_or_octet_stream();
    Response::builder()
        .header(header::CONTENT_TYPE, mime.as_ref())
        .header(header::CACHE_CONTROL, "no-cache")
        .header("X-Content-Type-Options", "nosniff")
        .body(Body::from(file.contents()))
        .unwrap()
}
