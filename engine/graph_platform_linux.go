//go:build linux

package main

import (
	"errors"
	"runtime"
	"sync"
)

var graphSupervisionProbe struct {
	once sync.Once
	err  error
}

// Exercise the trusted helper handshake, including subreaper and pidfd setup,
// before advertising execution. Every activation still creates its own domain.
func graphContainmentAvailable() error {
	graphSupervisionProbe.once.Do(func() {
		if runtime.GOARCH != "amd64" {
			graphSupervisionProbe.err = errors.New("Graphs require Linux amd64.")
			return
		}
		g, err := newRuntimeProcessGroup()
		if err == nil {
			err = g.Close()
		}
		graphSupervisionProbe.err = err
	})
	return graphSupervisionProbe.err
}
