package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
	_ "time/tzdata"
)

type EngineTimeSettings struct {
	Timezone  string `json:"timezone"`
	Revision  uint64 `json:"revision"`
	UpdatedAt string `json:"updatedAt"`
}

var validEngineTimezones sync.Map

func validEngineTimezone(zone string) bool {
	if len(zone) > 100 {
		return false
	}
	if _, ok := validEngineTimezones.Load(zone); ok {
		return true
	}
	// Local is process-relative, not a persisted IANA identity clients can use.
	if zone != "UTC" && (!strings.Contains(zone, "/") || strings.HasPrefix(zone, "/") || strings.Contains(zone, "..")) {
		return false
	}
	_, err := time.LoadLocation(zone)
	if err == nil {
		validEngineTimezones.Store(zone, true)
	}
	return err == nil
}

func detectEngineTimezone() string {
	if runtime.GOOS == "windows" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		// WinRT returns the IANA identifier rather than a Windows display name.
		// Use the system PowerShell, without profile scripts or user-controlled input.
		shell := filepath.Join(os.Getenv("SystemRoot"), "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
		if !filepath.IsAbs(shell) {
			return ""
		}
		cmd := exec.CommandContext(ctx, shell, "-NoLogo", "-NoProfile", "-NonInteractive", "-Command", "[Windows.Globalization.Calendar,Windows.Globalization,ContentType=WindowsRuntime]::new().GetTimeZone()")
		configureProcess(cmd)
		out, err := cmd.Output()
		zone := strings.TrimSpace(string(out))
		if err == nil && validEngineTimezone(zone) {
			return zone
		}
		return ""
	}
	// Unix installations may expose an IANA name through TZ or /etc/localtime.
	if zone := strings.TrimPrefix(os.Getenv("TZ"), ":"); validEngineTimezone(zone) {
		return zone
	}
	if target, err := filepath.EvalSymlinks("/etc/localtime"); err == nil {
		if _, zone, ok := strings.Cut(filepath.ToSlash(target), "/zoneinfo/"); ok && validEngineTimezone(zone) {
			return zone
		}
	}
	if b, err := os.ReadFile("/etc/timezone"); err == nil {
		if zone := strings.TrimSpace(string(b)); validEngineTimezone(zone) {
			return zone
		}
	}
	return ""
}

func validateEngineTimeSettings(s *EngineTimeSettings) error {
	if s == nil {
		return nil
	} // Pre-migration state.
	if s.Revision == 0 || !validGraphTime(s.UpdatedAt) || s.Timezone != "" && !validEngineTimezone(s.Timezone) {
		return errors.New("Invalid persisted engine timezone.")
	}
	return nil
}

func initializeEngineTime(dir string, d *diskState) error {
	if d.EngineTime != nil {
		return validateEngineTimeSettings(d.EngineTime)
	}
	// Even failed detection is recorded once: require Settings instead of changing
	// timezone silently on a later restart or OS configuration change.
	d.EngineTime = &EngineTimeSettings{Timezone: detectEngineTimezone(), Revision: 1, UpdatedAt: now()}
	return saveState(dir, d)
}

func (a *app) getEngineTime(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	settings := a.state.EngineTime
	a.mu.Unlock()
	respond(w, 200, settings)
}

func (a *app) updateEngineTime(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Timezone string `json:"timezone"`
		Revision uint64 `json:"revision"`
	}
	if !decode(w, r, &body) {
		return
	}
	if !validEngineTimezone(body.Timezone) {
		fail(w, 400, "Select a valid IANA time zone.")
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.state.EngineTime == nil || a.state.EngineTime.Revision != body.Revision {
		fail(w, 409, "Engine time zone changed. Reload and retry.")
		return
	}
	settings := EngineTimeSettings{Timezone: body.Timezone, Revision: body.Revision + 1, UpdatedAt: now()}
	if err := a.commitTransactionLocked(func(d *diskState) error {
		// Only the global setting changes. Existing instants and any future run
		// snapshots must remain untouched; this is not a catch-up trigger.
		d.EngineTime = &settings
		return nil
	}); err != nil {
		fail(w, 500, err.Error())
		return
	}
	respond(w, 200, settings)
}
