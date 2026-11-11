package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Resolve before a harness replaces XDG_CONFIG_HOME. Never copy credentials or
// modify HOME; gh still owns token and host precedence.
func ghConfigurationEnvironment(env []string) []string {
	dir := os.Getenv("GH_CONFIG_DIR")
	if dir == "" {
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			dir = filepath.Join(xdg, "gh")
		} else if runtime.GOOS == "windows" && os.Getenv("APPDATA") != "" {
			dir = filepath.Join(os.Getenv("APPDATA"), "GitHub CLI")
		} else if home, err := os.UserHomeDir(); err == nil {
			dir = filepath.Join(home, ".config", "gh")
		}
	}
	if dir == "" {
		return env
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return env
	}
	result := make([]string, 0, len(env)+1)
	for _, entry := range env {
		key, _, _ := strings.Cut(entry, "=")
		if !strings.EqualFold(key, "GH_CONFIG_DIR") {
			result = append(result, entry)
		}
	}
	return append(result, "GH_CONFIG_DIR="+abs)
}
