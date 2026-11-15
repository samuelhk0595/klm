package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRunnerLinuxControlDiscovery(t *testing.T) {
	a, _ := runnerRegressionApp(t, "coordinator")
	a.controlReady = true
	// Keep the Unix socket path below the platform's fixed pathname limit.
	dir, err := os.MkdirTemp("", "klm-control-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	shutdown := make(chan struct{})
	listener, err := serveEngineControl(dir, a, shutdown)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	command := func(value string) string {
		t.Helper()
		conn, err := net.DialTimeout("unix", filepath.Join(dir, "control.sock"), time.Second)
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()
		_ = conn.SetDeadline(time.Now().Add(2 * time.Second))
		if _, err := fmt.Fprintln(conn, value); err != nil {
			t.Fatal(err)
		}
		body, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			t.Fatalf("%s control failed: %v", value, err)
		}
		return body
	}
	var discovery struct {
		Ready    bool   `json:"ready"`
		EngineID string `json:"engine_id"`
		Protocol string `json:"runner_protocol"`
	}
	if err := json.Unmarshal([]byte(command("discover")), &discovery); err != nil {
		t.Fatal(err)
	}
	if !discovery.Ready || discovery.EngineID == "" || discovery.Protocol != "1" {
		t.Fatalf("invalid discovery: %+v", discovery)
	}
	if body := command("ready"); body != "ok\n" {
		t.Fatalf("normal readiness failed after discovery: %q", body)
	}
	if body := command("stop"); body != "ok\n" {
		t.Fatalf("normal stop failed after discovery: %q", body)
	}
	select {
	case <-shutdown:
	default:
		t.Fatal("control Stop did not signal shutdown")
	}
}
