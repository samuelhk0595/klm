package main

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

func (a *app) runnerJobContent(w http.ResponseWriter, r *http.Request) {
	owner, err := a.runnerOwner(r)
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	var body []byte
	err = a.store.readers.QueryRowContext(r.Context(), `SELECT j.body FROM runner_jobs j JOIN runner_executions e ON e.id=j.execution_id WHERE j.id=? AND e.integration_id=?`, r.PathValue("jobId"), owner).Scan(&body)
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	var job runnerJob
	if err := json.Unmarshal(body, &job); err != nil {
		sqlHTTPError(w, err)
		return
	}
	if job.Digest == "" {
		fail(w, 409, "Full output is not finalized; inspect the current job preview.")
		return
	}
	if digest := r.URL.Query().Get("digest"); digest != "" && digest != job.Digest {
		fail(w, 409, "Output digest changed.")
		return
	}
	offset := int64(0)
	if value := r.URL.Query().Get("offset"); value != "" {
		offset, err = strconv.ParseInt(value, 10, 64)
	}
	if err != nil || offset < 0 || offset > job.Bytes {
		fail(w, 400, "Invalid output offset.")
		return
	}
	file, err := os.Open(filepath.Join(a.dir, "runner-output", job.ID+".log"))
	if err != nil {
		sqlHTTPError(w, err)
		return
	}
	defer file.Close()
	chunk := make([]byte, min(int64(64<<10), job.Bytes-offset))
	n, err := file.ReadAt(chunk, offset)
	if err != nil && n != len(chunk) {
		sqlHTTPError(w, err)
		return
	}
	respond(w, 200, map[string]any{"id": job.ID, "execution_id": job.ExecutionID, "digest": job.Digest, "bytes": job.Bytes, "offset": strconv.FormatInt(offset, 10), "next_offset": strconv.FormatInt(offset+int64(n), 10), "eof": offset+int64(n) == job.Bytes, "data_base64": base64.StdEncoding.EncodeToString(chunk[:n])})
}
