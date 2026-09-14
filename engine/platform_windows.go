package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"sync"
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

var (
	createJobObject     = syscall.NewLazyDLL("kernel32.dll").NewProc("CreateJobObjectW")
	setJobInformation   = syscall.NewLazyDLL("kernel32.dll").NewProc("SetInformationJobObject")
	queryJobInformation = syscall.NewLazyDLL("kernel32.dll").NewProc("QueryInformationJobObject")
	assignProcessToJob  = syscall.NewLazyDLL("kernel32.dll").NewProc("AssignProcessToJobObject")
	terminateJobObject  = syscall.NewLazyDLL("kernel32.dll").NewProc("TerminateJobObject")
	resumeOwnedProcess  = syscall.NewLazyDLL("ntdll.dll").NewProc("NtResumeProcess")
)

type ownedProcess struct {
	mu      sync.Mutex
	cmd     *exec.Cmd
	job     syscall.Handle
	stopped bool
	drained bool
}

// Launch suspended: assigning a job after an ordinary Start leaves a child-spawn
// race. No target user code runs until assignment has succeeded. No breakaway
// flags are granted; engine crashes close the non-inherited kill-on-close job.
func prepareOwnedProcess(cmd *exec.Cmd) (*ownedProcess, error) {
	configureProcess(cmd)
	cmd.SysProcAttr.CreationFlags |= 0x4 // CREATE_SUSPENDED
	job, _, err := createJobObject.Call(0, 0)
	if job == 0 {
		return nil, errors.New("Cannot create the owned harness job.")
	}
	// JOBOBJECT_EXTENDED_LIMIT_INFORMATION, pointer-size aware on Windows.
	limits := struct {
		PerProcessUserTimeLimit int64
		PerJobUserTimeLimit     int64
		LimitFlags              uint32
		MinimumWorkingSetSize   uintptr
		MaximumWorkingSetSize   uintptr
		ActiveProcessLimit      uint32
		Affinity                uintptr
		PriorityClass           uint32
		SchedulingClass         uint32
		IOCounters              [6]uint64
		ProcessMemoryLimit      uintptr
		JobMemoryLimit          uintptr
		PeakProcessMemoryUsed   uintptr
		PeakJobMemoryUsed       uintptr
	}{LimitFlags: 0x2000} // JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
	ok, _, err := setJobInformation.Call(job, 9, uintptr(unsafe.Pointer(&limits)), unsafe.Sizeof(limits))
	if ok == 0 {
		syscall.CloseHandle(syscall.Handle(job))
		return nil, fmt.Errorf("Cannot configure the owned harness job: %w", err)
	}
	return &ownedProcess{cmd: cmd, job: syscall.Handle(job)}, nil
}

func (o *ownedProcess) Attach() error {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.stopped {
		_ = o.cmd.Process.Kill()
		return errors.New("Owned harness was cancelled before job assignment.")
	}
	// This PID comes directly from the still-suspended process we just created.
	h, err := syscall.OpenProcess(0x0100|0x0800|0x0001, false, uint32(o.cmd.Process.Pid))
	if err != nil {
		_ = o.cmd.Process.Kill()
		return errors.New("Cannot open the suspended owned harness.")
	}
	defer syscall.CloseHandle(h)
	if ok, _, _ := assignProcessToJob.Call(uintptr(o.job), uintptr(h)); ok == 0 {
		_ = o.cmd.Process.Kill()
		return errors.New("Cannot contain the harness in its owned job; execution was refused.")
	}
	if status, _, _ := resumeOwnedProcess.Call(uintptr(h)); status != 0 {
		_ = o.cmd.Process.Kill()
		return errors.New("Cannot resume the contained harness.")
	}
	return nil
}

func (o *ownedProcess) Stop() error {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.stopped {
		return nil
	}
	o.stopped = true
	if o.job != 0 {
		if ok, _, err := terminateJobObject.Call(uintptr(o.job), 1); ok == 0 {
			if o.cmd.Process != nil {
				_ = o.cmd.Process.Kill()
			}
			return fmt.Errorf("Cannot terminate the owned harness job: %w", err)
		}
	}
	// Also covers cancellation between Start and Attach, while still suspended.
	if o.cmd.Process != nil {
		_ = o.cmd.Process.Kill()
	}
	return nil
}

func (o *ownedProcess) Close() error {
	stopErr := o.Stop()
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.job == 0 {
		return stopErr
	}
	defer func() { syscall.CloseHandle(o.job); o.job = 0 }()
	deadline := time.Now().Add(5 * time.Second)
	for {
		accounting := struct {
			Times                    [4]int64
			TotalPageFaultCount      uint32
			TotalProcesses           uint32
			ActiveProcesses          uint32
			TotalTerminatedProcesses uint32
		}{}
		ok, _, _ := queryJobInformation.Call(uintptr(o.job), 1, uintptr(unsafe.Pointer(&accounting)), unsafe.Sizeof(accounting), 0)
		if ok == 0 {
			return errors.New("Cannot confirm owned harness job drainage.")
		}
		if accounting.ActiveProcesses == 0 {
			o.drained = true
			return stopErr
		}
		if time.Now().After(deadline) {
			return errors.New("Owned harness job drainage is unconfirmed.")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (o *ownedProcess) Drained() bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.drained
}

const directoryPickerScript = `
$ErrorActionPreference = 'Stop'
[Console]::OutputEncoding = [System.Text.UTF8Encoding]::new($false)
Add-Type -AssemblyName System.Windows.Forms
$dialog = New-Object System.Windows.Forms.FolderBrowserDialog
$dialog.Description = 'Select project folder'
$dialog.ShowNewFolderButton = $false
$dialog.SelectedPath = [Environment]::GetFolderPath('UserProfile')
# A visible topmost owner keeps the dialog above the desktop/browser even though
# the engine itself has no window. Dispose it together with the picker.
$owner = New-Object System.Windows.Forms.Form
$owner.Text = 'Select project folder'
$owner.Size = New-Object System.Drawing.Size(1, 1)
$owner.StartPosition = [System.Windows.Forms.FormStartPosition]::CenterScreen
$owner.ShowInTaskbar = $false
$owner.FormBorderStyle = [System.Windows.Forms.FormBorderStyle]::None
$owner.TopMost = $true
try {
  $owner.Show()
  $owner.Activate()
  if ($dialog.ShowDialog($owner) -eq [System.Windows.Forms.DialogResult]::OK) {
    [Console]::Write($dialog.SelectedPath)
  }
} finally { $dialog.Dispose(); $owner.Dispose() }
`

func pickerCommand(ctx context.Context) (*exec.Cmd, error) {
	cmd := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-NonInteractive", "-STA", "-Command", directoryPickerScript)
	configureProcess(cmd)
	cmd.Cancel = func() error { return killTree(cmd.Process) }
	cmd.WaitDelay = 5 * time.Second
	return cmd, nil
}
