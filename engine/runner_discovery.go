package main

import (
	"encoding/json"
	"io"
)

func (a *app) writeRunnerDiscovery(w io.Writer) {
	a.mu.Lock()
	ready := a.controlReady && !a.closing && a.storageErr == nil && a.runner != nil
	a.mu.Unlock()
	var engineID string
	if a.store != nil {
		if err := a.store.readers.QueryRow(`SELECT value FROM store_meta WHERE key='database_id'`).Scan(&engineID); err != nil {
			ready = false
		}
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"engine_id": engineID, "api_port": apiPort, "runner_protocol": "1", "ready": ready})
}
