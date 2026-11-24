package main

import (
	"fmt"
	"os"
	"syscall"
)

func runnerDirectoryIdentity(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return "", fmt.Errorf("Workspace directory identity is unavailable.")
	}
	return fmt.Sprintf("%x:%x", stat.Dev, stat.Ino), nil
}
