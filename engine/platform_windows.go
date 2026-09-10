package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"
	"unsafe"
)

var moveFileEx = syscall.NewLazyDLL("kernel32.dll").NewProc("MoveFileExW")

func replaceFile(from, to string) error {
	f, err := syscall.UTF16PtrFromString(from)
	if err != nil {
		return err
	}
	t, err := syscall.UTF16PtrFromString(to)
	if err != nil {
		return err
	}
	// REPLACE_EXISTING | WRITE_THROUGH, on the same volume as the temp file.
	ok, _, callErr := moveFileEx.Call(uintptr(unsafe.Pointer(f)), uintptr(unsafe.Pointer(t)), 0x1|0x8)
	if ok == 0 {
		return callErr
	}
	return nil
}

func lockDataDir(path string) (*os.File, error) {
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	// The kernel releases this exclusive handle on crashes. A leftover file is not a stale lock.
	h, err := syscall.CreateFile(p, syscall.GENERIC_READ|syscall.GENERIC_WRITE, 0, nil,
		syscall.OPEN_ALWAYS, syscall.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		return nil, errors.New("data directory is locked by another engine or cannot be opened")
	}
	f := os.NewFile(uintptr(h), path)
	if err := f.Truncate(0); err != nil {
		f.Close()
		return nil, err
	}
	if _, err := fmt.Fprintf(f, "%d\n", os.Getpid()); err != nil {
		f.Close()
		return nil, err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return nil, err
	}
	return f, nil
}

func configureProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}

func killTree(p *os.Process) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "taskkill.exe", "/PID", strconv.Itoa(p.Pid), "/T", "/F")
	configureProcess(cmd)
	if err := cmd.Run(); err != nil {
		return p.Kill()
	}
	return nil
}

const directoryPickerScript = `
$ErrorActionPreference = 'Stop'
[Console]::OutputEncoding = [System.Text.UTF8Encoding]::new($false)
Add-Type -AssemblyName System.Windows.Forms
$dialog = New-Object System.Windows.Forms.FolderBrowserDialog
$dialog.Description = 'Select project folder'
$dialog.ShowNewFolderButton = $false
try {
  if ($dialog.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK) {
    [Console]::Write($dialog.SelectedPath)
  }
} finally { $dialog.Dispose() }
`

func pickerCommand(ctx context.Context) (*exec.Cmd, error) {
	cmd := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-NonInteractive", "-STA", "-Command", directoryPickerScript)
	configureProcess(cmd)
	cmd.Cancel = func() error { return killTree(cmd.Process) }
	cmd.WaitDelay = 5 * time.Second
	return cmd, nil
}
