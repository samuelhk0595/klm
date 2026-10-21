package main

import (
	"context"
	"testing"
)

func TestOpenCodeRuntimeTurnBoundary(t *testing.T) {
	for _, perTurn := range []bool{false, true} {
		name := "reusable"
		if perTurn {
			name = "managed"
		}
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			a := &app{ctx: ctx, runtimes: map[string]*sessionRuntime{}}
			defer a.shutdownSessionRuntimes()
			first, err := a.beginSessionRuntime("session")
			if err != nil {
				t.Fatal(err)
			}
			group, err := newRuntimeProcessGroup()
			if err != nil {
				t.Fatal(err)
			}
			first.group = group
			// Exercise runtime ownership without starting a harness. Its process
			// lifecycle is covered separately by the owned-process tests.
			if err := first.retainOpenCode(nil, "captured-profile", perTurn); err != nil {
				t.Fatal(err)
			}
			first.finishTurn()
			if !first.active() || a.runtimes["session"] != first {
				t.Fatal("successful turn lost its idle runtime")
			}
			second, err := a.beginSessionRuntime("session")
			if err != nil {
				t.Fatal(err)
			}
			if perTurn {
				if second == first || first.active() || !group.closed {
					t.Fatal("new managed turn reused its previous process container")
				}
			} else if second != first || !second.active() {
				t.Fatal("reusable runtime was unnecessarily replaced")
			}
			if err := second.shutdown(); err != nil {
				t.Fatal(err)
			}
			if second.active() || a.runtimes["session"] != nil {
				t.Fatal("explicit shutdown retained the session runtime")
			}
		})
	}
}
