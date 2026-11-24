package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func runnerProcessBirth(pid int) (string, bool, error) {
	identity, err := linuxProcIdentity(pid)
	if os.IsNotExist(err) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	boot, err := os.ReadFile("/proc/sys/kernel/random/boot_id")
	return string(boot) + "/" + identity.key, true, err
}

func newRunnerRuntimeProcessGroup(dir string) (*runtimeProcessGroup, runnerProcessDomain, error) {
	domain := runnerProcessDomain{ID: newID(), Kind: "linux_supervisor_receipt", EnginePID: os.Getpid()}
	var err error
	domain.EngineBirth, _, err = runnerProcessBirth(domain.EnginePID)
	if err != nil {
		return nil, domain, err
	}
	folder := filepath.Join(dir, "runner-ownership")
	if err := os.MkdirAll(folder, 0700); err != nil {
		return nil, domain, err
	}
	file, err := os.OpenFile(filepath.Join(folder, domain.ID+".json"), os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	if err != nil {
		return nil, domain, err
	}
	defer file.Close()
	group, err := newRuntimeProcessGroup()
	if err != nil {
		return nil, domain, err
	}
	_, err = group.call(linuxPacket{Op: "recovery-receipt", Token: domain.ID}, []*os.File{file})
	if err != nil {
		group.Close()
		return nil, domain, err
	}
	return group, domain, nil
}

func reconcileRunnerProcessDomain(ctx context.Context, dir string, domain runnerProcessDomain) error {
	if domain.Kind != "linux_supervisor_receipt" || !validAgentSourceUID(domain.ID) {
		return errors.New("Retained Linux process-domain identity is invalid.")
	}
	birth, alive, err := runnerProcessBirth(domain.EnginePID)
	if err != nil {
		return err
	}
	if alive && birth == domain.EngineBirth {
		return errors.New("The original engine still owns the process supervisor.")
	}
	boot, err := os.ReadFile("/proc/sys/kernel/random/boot_id")
	if err != nil {
		return err
	}
	if strings.SplitN(domain.EngineBirth, "/", 2)[0] != string(boot) {
		return nil
	}
	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()
	for {
		data, err := os.ReadFile(filepath.Join(dir, "runner-ownership", domain.ID+".json"))
		var receipt struct {
			ID      string `json:"id"`
			Drained bool   `json:"drained"`
		}
		if err == nil && len(data) < 1024 && json.Unmarshal(data, &receipt) == nil && receipt.ID == domain.ID && receipt.Drained {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline.C:
			return errors.New("The original Linux subreaper has not retained its confirmed domain-drainage receipt.")
		case <-time.After(20 * time.Millisecond):
		}
	}
}
