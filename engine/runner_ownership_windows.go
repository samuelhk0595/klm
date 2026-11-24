package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var openRunnerJob = syscall.NewLazyDLL("kernel32.dll").NewProc("OpenJobObjectW")

func runnerProcessBirth(pid int) (string, bool, error) {
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION|windows.SYNCHRONIZE, false, uint32(pid))
	if errors.Is(err, windows.ERROR_INVALID_PARAMETER) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	defer windows.CloseHandle(handle)
	state, err := windows.WaitForSingleObject(handle, 0)
	if err != nil {
		return "", false, err
	}
	if state == windows.WAIT_OBJECT_0 {
		return "", false, nil
	}
	var created, exited, kernel, user windows.Filetime
	if err := windows.GetProcessTimes(handle, &created, &exited, &kernel, &user); err != nil {
		return "", false, err
	}
	return fmt.Sprintf("%x:%x", created.HighDateTime, created.LowDateTime), true, nil
}

func newRunnerRuntimeProcessGroup(dir string) (*runtimeProcessGroup, runnerProcessDomain, error) {
	domain := runnerProcessDomain{ID: newID(), Kind: "windows_job", EnginePID: os.Getpid()}
	var err error
	domain.EngineBirth, _, err = runnerProcessBirth(domain.EnginePID)
	if err != nil {
		return nil, domain, err
	}
	domain.Name = "Local\\KLMRunner-" + domain.ID
	name, err := syscall.UTF16PtrFromString(domain.Name)
	if err != nil {
		return nil, domain, err
	}
	group, err := newWindowsRuntimeProcessGroup(uintptr(unsafe.Pointer(name)))
	return group, domain, err
}

func reconcileRunnerProcessDomain(ctx context.Context, dir string, domain runnerProcessDomain) error {
	if domain.Kind != "windows_job" || domain.Name != "Local\\KLMRunner-"+domain.ID || domain.EngineBirth == "" {
		return errors.New("Retained Windows ownership identity is invalid.")
	}
	if !domain.Contained || domain.PendingStarts != 0 {
		return errors.New("A native process start was not confirmed inside its retained Windows job.")
	}
	birth, alive, err := runnerProcessBirth(domain.EnginePID)
	if err != nil {
		return err
	}
	if alive && birth == domain.EngineBirth {
		return errors.New("The original engine process is still alive; its domain was not taken over.")
	}
	name, err := syscall.UTF16PtrFromString(domain.Name)
	if err != nil {
		return err
	}
	// The unpredictable engine-owned job identity covers every descendant. Root
	// PID absence alone never establishes drainage or selects a kill target.
	handle, _, callErr := openRunnerJob.Call(0x0004|0x0008, 0, uintptr(unsafe.Pointer(name)))
	if handle == 0 {
		if errors.Is(callErr, syscall.ERROR_FILE_NOT_FOUND) {
			return nil
		}
		return fmt.Errorf("Cannot reopen the retained process domain: %w", callErr)
	}
	group := &runtimeProcessGroup{job: syscall.Handle(handle)}
	if err := group.Close(); err != nil {
		return err
	}
	if !group.drained {
		return errors.New("Retained Windows process domain did not confirm drainage.")
	}
	return ctx.Err()
}
