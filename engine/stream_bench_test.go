package main

import (
	"context"
	"strings"
	"testing"
)

// A burst arriving while the engine has unrelated conversation history. Compare
// the former per-fragment persistence path with the production batching path.
// Run explicitly; this does not start a harness or touch the user's engine data.
func BenchmarkStreamBurst(b *testing.B) {
	for _, mode := range []string{"per_fragment", "batched_delta", "batched_snapshot"} {
		b.Run(mode, func(b *testing.B) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			a := &app{dir: b.TempDir(), state: diskState{
				Version: 2, Native: map[string]nativeSession{},
				Projects: []Project{{ID: "project", Folders: []string{}}},
				Sessions: []Session{
					{ID: "active", ProjectID: "project", Status: "running", Events: []Event{}},
					{ID: "history", ProjectID: "project", Status: "idle", Events: []Event{event("assistant", strings.Repeat("history ", 1<<20))}},
				},
			}}
			const fragments = 100
			const delta = "word "
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				p := &adapter{app: a, id: "active", harness: "pi", keys: map[string]string{}, turn: &turn{ctx: ctx, cancel: cancel}}
				if mode != "per_fragment" {
					p.stream = newStreamBatch(p)
				}
				for i := 0; i < fragments; i++ {
					text, appendText := delta, true
					if mode == "batched_snapshot" {
						text, appendText = strings.Repeat(delta, i+1), false
					}
					if err := p.put("message", "assistant", "", text, "running", appendText, nil); err != nil {
						b.Fatal(err)
					}
				}
				// A synchronous boundary must persist all text first and retain order.
				if err := p.put("tool", "tool", "Read", "result", "completed", false, nil); err != nil {
					b.Fatal(err)
				}
				events := a.state.session("active").Events
				if len(events) != (n+1)*2 || events[n*2].Text != strings.Repeat(delta, fragments) || events[n*2+1].Title != "Read" {
					b.Fatal("stream lost text or changed event ordering")
				}
				// Final native snapshots replace accumulated text, never append it.
				if err := p.put("message", "assistant", "", "Final answer", "completed", false, nil); err != nil {
					b.Fatal(err)
				}
				if p.stream != nil {
					if err := p.stream.close(); err != nil {
						b.Fatal(err)
					}
				}
				final := a.state.session("active").Events[n*2]
				if final.Text != "Final answer" || final.Status != "completed" {
					b.Fatal("stream overwrote the completed message")
				}
			}
			b.ReportMetric(float64(a.state.GraphRevision)/float64(b.N), "commits/burst")
		})
	}
}
