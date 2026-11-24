//go:build !windows && !linux

package main

import "errors"

func runnerDirectoryIdentity(path string) (string, error) {
	return "", errors.New("Workspace directory identity is unavailable on this platform.")
}
