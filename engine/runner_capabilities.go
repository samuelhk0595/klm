package main

import (
	"net/http"
	"runtime"
)

func (a *app) runnerCapabilities(w http.ResponseWriter, r *http.Request) {
	var engineID string
	if err := a.store.readers.QueryRowContext(r.Context(), `SELECT value FROM store_meta WHERE key='database_id'`).Scan(&engineID); err != nil {
		sqlHTTPError(w, err)
		return
	}
	profiles := []map[string]any{}
	for _, harness := range a.harnessSnapshot() {
		ready := harness.Available && harness.Compatible && managedChatPlatform()
		diagnostics := []string{}
		if !harness.Available {
			diagnostics = append(diagnostics, "Harness executable is unavailable.")
		}
		if harness.Available && !harness.Compatible {
			diagnostics = append(diagnostics, "Install the pinned managed harness version before external execution.")
		}
		if !managedChatPlatform() {
			diagnostics = append(diagnostics, "Owned managed process supervision is unavailable on this platform.")
		}
		profiles = append(profiles, map[string]any{"id": harness.ID, "installed_version": harness.InstalledVersion, "required_version": harness.RequiredVersion, "ready": ready, "mechanisms": map[string]bool{"operational_manifest": ready, "owned_shell_jobs": ready, "next_step_input": ready, "native_questions": ready, "native_permissions": ready, "mcp_elicitation": ready, "native_continuation": ready}, "runtime_validation": map[string]any{"platform": "windows_amd64", "mechanism_probe_recorded": runtime.GOOS == "windows" && runtime.GOARCH == "amd64"}, "diagnostics": diagnostics})
	}
	respond(w, 200, map[string]any{"engine_id": engineID, "protocol": "1", "harnesses": profiles, "shell_wait_max_ms": 5000, "job_wait_max_ms": 5000, "question_kinds": []string{"question", "permission", "elicitation"}})
}
