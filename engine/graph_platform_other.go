//go:build !linux

package main

import (
	"errors"
	"runtime"
)

func graphContainmentAvailable() error {
	if runtime.GOOS == "windows" {
		return nil
	}
	return errors.New("Graph process supervision requires Windows or Linux amd64.")
}
