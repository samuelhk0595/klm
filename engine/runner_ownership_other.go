//go:build !windows && !linux

package main

import (
	"context"
	"errors"
)

func newRunnerRuntimeProcessGroup(dir string) (*runtimeProcessGroup, runnerProcessDomain, error) {
	return nil, runnerProcessDomain{}, errors.New("Durable runner process domains are unavailable on this platform.")
}
func reconcileRunnerProcessDomain(ctx context.Context, dir string, domain runnerProcessDomain) error {
	return errors.New("Durable runner process reconciliation is unavailable on this platform.")
}
