package main

import (
	"context"
	"errors"
	"os/exec"
)

// A durable kernel-domain or supervisor receipt, never a PID-only kill grant.
type runnerProcessDomain struct {
	ID            string `json:"id"`
	Kind          string `json:"kind"`
	Name          string `json:"name,omitempty"`
	EnginePID     int    `json:"engine_pid"`
	EngineBirth   string `json:"engine_birth"`
	Contained     bool   `json:"contained"`
	PendingStarts int    `json:"pending_starts"`
}

func (s *runnerService) confirmProcessDomain(executionID, turnID, domainID string) error {
	a := s.app
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.writeSQLLocked(a.ctx, func(st *sqliteSessionTx) error {
		execution, err := st.runnerExecution(executionID)
		if err != nil {
			return err
		}
		if execution.CancelRequested || execution.State != "running" || execution.TurnID != turnID {
			return errors.New("Process admission is sealed.")
		}
		if err := st.runnerGroupOpen(execution.IntegrationID, execution.ProjectID, execution.InstanceKey); err != nil {
			return err
		}
		for i := range execution.ProcessDomains {
			if execution.ProcessDomains[i].ID == domainID {
				execution.ProcessDomains[i].Contained = true
				execution.ProcessDomains[i].PendingStarts--
				return st.putRunnerExecution(execution)
			}
		}
		return errors.New("Retained process-domain admission was not found.")
	})
}

func (s *runnerService) retainProcessDomain(executionID, turnID string, domain runnerProcessDomain) error {
	a := s.app
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.writeSQLLocked(a.ctx, func(st *sqliteSessionTx) error {
		execution, err := st.runnerExecution(executionID)
		if err != nil {
			return err
		}
		if execution.CancelRequested || execution.State != "running" || execution.TurnID != turnID {
			return errors.New("Native process admission belongs to a sealed turn.")
		}
		if err := st.runnerGroupOpen(execution.IntegrationID, execution.ProjectID, execution.InstanceKey); err != nil {
			return err
		}
		for i := range execution.ProcessDomains {
			if execution.ProcessDomains[i].ID == domain.ID {
				execution.ProcessDomains[i].PendingStarts++
				return st.putRunnerExecution(execution)
			}
		}
		domain.PendingStarts = 1
		execution.ProcessDomains = append(execution.ProcessDomains, domain)
		return st.putRunnerExecution(execution)
	})
}

func prepareRunnerOwnedProcess(cmd *exec.Cmd, dir string) (*ownedProcess, *runnerProcessDomain, error) {
	group, domain, err := newRunnerRuntimeProcessGroup(dir)
	if err != nil {
		return nil, nil, err
	}
	owner, err := prepareSessionOwnedProcess(cmd, group)
	if err != nil {
		group.Close()
		return nil, nil, err
	}
	owner.exclusive = true
	return owner, &domain, nil
}

func (a *app) reconcileRunnerOwnership(ctx context.Context, execution *runnerExecution) error {
	if execution.ProcessesDrained {
		return nil
	}
	if len(execution.ProcessDomains) == 0 {
		// No native turn has ever been claimed for an idle prepared worker.
		if execution.TurnID == "" {
			execution.ProcessesDrained = true
			return nil
		}
		execution.RecoveryBlocked = "native_process_domain_missing"
		return nil
	}
	for _, domain := range execution.ProcessDomains {
		if err := reconcileRunnerProcessDomain(ctx, a.dir, domain); err != nil {
			execution.RecoveryBlocked = "native_process_domain_unconfirmed"
			execution.Error = err.Error()
			return nil
		}
	}
	execution.ProcessesDrained = true
	return nil
}
