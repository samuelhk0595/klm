//go:build !windows

package main

import (
	"os"
	"runtime"
)

const platformNativeDirectoryPicker = false

func linuxDirectoryPickerBackend() string {
	if runtime.GOOS != "linux" || os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "" {
		return ""
	}
	for _, name := range []string{"zenity", "kdialog"} {
		if _, ok := executable(name); ok {
			return name
		}
	}
	return ""
}

func nativeDirectoryPickerAvailable() bool { return linuxDirectoryPickerBackend() != "" }

func platformDirectoryRoots() ([]filesystemDirectory, error) {
	return []filesystemDirectory{{Name: "/", Path: "/"}}, nil
}
