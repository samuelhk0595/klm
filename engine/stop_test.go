package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStopCancelsEvenWhenStorageFails(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	a := &app{
		storageErr: errors.New("storage unavailable"),
		state:      diskState{Sessions: []Session{{ID: "session"}}},
		runs:       map[string]*turn{"session": {ctx: ctx, cancel: cancel}},
	}
	r := httptest.NewRequest(http.MethodPost, "/api/sessions/session/stop", nil)
	r.SetPathValue("id", "session")
	w := httptest.NewRecorder()
	a.stop(w, r)
	if ctx.Err() == nil {
		t.Fatal("Stop left the turn executing because a state write failed")
	}
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("storage failure must still be reported: HTTP %d", w.Code)
	}
}
