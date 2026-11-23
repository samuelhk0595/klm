package main

import "runtime"

func defaultTerminalShell() string {
	if runtime.GOOS == "linux" {
		return "sh"
	}
	return "powershell"
}
