package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

type turn struct {
	consultationID string
	delivery       bool
	ctx            context.Context
	cancel         context.CancelFunc
	done           chan struct{}
	approvals      map[string]*pendingApproval
	questions      map[string]*pendingQuestion
}

type app struct {
	mu         sync.Mutex
	dir        string
	state      diskState
	storageErr error
	harnesses  []Harness
	binaries   map[string]binary
	runs       map[string]*turn
	listeners  map[string]map[chan struct{}]bool
	picker     chan struct{}
	ctx        context.Context
	closing    bool
	wg         sync.WaitGroup
	catalogMu  sync.Mutex
	catalogs   map[string]catalogCache
	quotaMu    sync.Mutex
	quotas     map[string]quotaCache
}

func main() {
	if err := run(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func run() error {
	addr := flag.String("addr", "127.0.0.1:7331", "loopback HTTP listen address")
	dataDir := flag.String("data-dir", "", "state directory (default: OS user config directory/klm/engine)")
	flag.Parse()
	host, _, err := net.SplitHostPort(*addr)
	if err != nil {
		return errors.New("-addr must be a loopback IP and port, for example 127.0.0.1:7331")
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return errors.New("-addr must use a literal loopback IP; wildcard and network binds are forbidden")
	}
	if *dataDir == "" {
		config, err := os.UserConfigDir()
		if err != nil {
			return err
		}
		*dataDir = filepath.Join(config, "klm", "engine")
	}
	dir, err := filepath.Abs(*dataDir)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(dir, "sessions"), 0700); err != nil {
		return errors.New("cannot create engine data directory")
	}
	lock, err := lockDataDir(filepath.Join(dir, "engine.lock"))
	if err != nil {
		return err
	}
	defer lock.Close()
	state, err := loadState(dir)
	if err != nil {
		return err
	}
	interrupted := false
	interruptedConsultations := map[string]string{}
	for i := range state.Consultations {
		c := &state.Consultations[i]
		if c.Status == "answering" {
			interruptedConsultations[c.To] = c.ID
		}
		if c.Delivery == "delivering" {
			interruptedConsultations[c.From] = c.ID
		}
		if !consultationTerminal(c.Status) || c.Delivery == "pending" || c.Delivery == "delivering" || c.Delivery == "waiting" {
			if !consultationTerminal(c.Status) {
				c.Status = "interrupted"
				c.Error = "Consultation interrupted by engine restart."
			}
			c.Delivery = "interrupted"
			syncConsultation(&state, c)
			interrupted = true
		}
	}
	for i := range state.Sessions {
		s := &state.Sessions[i]
		if len(s.Questions) > 0 {
			s.Questions = nil
			s.UpdatedAt = now()
			interrupted = true
		}
		if len(s.Permissions) > 0 {
			s.Permissions = nil
			s.UpdatedAt = now()
			interrupted = true
		}
		if s.Status == "running" {
			s.Status = "error"
			s.UpdatedAt = now()
			entry := event("error", "Execution interrupted by engine restart.")
			entry.ConsultationID = interruptedConsultations[s.ID]
			s.Events = append(s.Events, entry)
			interrupted = true
		}
	}
	if interrupted {
		if err := saveState(dir, &state); err != nil {
			return errors.New("cannot persist interrupted sessions")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	harnesses, binaries := discoverHarnesses()
	a := &app{dir: dir, state: state, harnesses: harnesses, binaries: binaries,
		runs: map[string]*turn{}, listeners: map[string]map[chan struct{}]bool{}, picker: make(chan struct{}, 1), ctx: ctx}
	go a.expireConsultations()
	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		return fmt.Errorf("cannot listen on %s: %w", *addr, err)
	}
	server := &http.Server{Handler: a.routes(), ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout: 15 * time.Second, WriteTimeout: 20 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16 << 10,
		BaseContext: func(net.Listener) context.Context { return ctx }}
	stopped := make(chan os.Signal, 1)
	signal.Notify(stopped, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stopped)
	serveErr := make(chan error, 1)
	go func() { serveErr <- server.Serve(listener) }()
	log.Printf("KLM engine listening on http://%s", listener.Addr())
	select {
	case <-stopped:
	case err = <-serveErr:
	}
	a.mu.Lock()
	a.closing = true
	cancel()
	a.mu.Unlock()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer shutdownCancel()
	_ = server.Shutdown(shutdownCtx)
	_ = server.Close()
	a.wg.Wait()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
