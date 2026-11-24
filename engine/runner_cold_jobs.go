package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
)

func (a *app) reconcileRunnerJobs(ctx context.Context, execution *runnerExecution) error {
	rows, err := a.store.readers.QueryContext(ctx, `SELECT body FROM runner_jobs WHERE execution_id=? AND state IN ('starting','running','unknown')`, execution.ID)
	if err != nil {
		return err
	}
	jobs := []runnerJob{}
	for rows.Next() {
		var body []byte
		var job runnerJob
		if err := rows.Scan(&body); err != nil {
			rows.Close()
			return err
		}
		if err := json.Unmarshal(body, &job); err != nil {
			rows.Close()
			return err
		}
		jobs = append(jobs, job)
	}
	if err := rows.Close(); err != nil {
		return err
	}
	for _, job := range jobs {
		if job.Domain == nil {
			execution.RecoveryBlocked = "owned_job_domain_missing"
			continue
		}
		if err := reconcileRunnerProcessDomain(ctx, a.dir, *job.Domain); err != nil {
			execution.RecoveryBlocked = "owned_job_domain_unconfirmed"
			execution.Error = err.Error()
			continue
		}
		job.State = "interrupted"
		job.Error = "The original owned process domain is drained. Exit outcome is unknown; retained output and workspace artifacts were preserved, and the command was not replayed."
		job.ExitCode = nil
		job.ContentID = job.ID
		file, err := os.Open(filepath.Join(a.dir, "runner-output", job.ID+".log"))
		if err == nil {
			hash := sha256.New()
			job.Bytes, err = io.Copy(hash, file)
			file.Close()
			if err != nil {
				return err
			}
			job.Digest = hex.EncodeToString(hash.Sum(nil))
		} else if !os.IsNotExist(err) {
			return err
		}
		if _, err := a.store.sessionTransaction(ctx, func(st *sqliteSessionTx) error {
			_, err := st.tx.ExecContext(st.ctx, `UPDATE runner_jobs SET state=?,body=? WHERE id=? AND state IN ('starting','running','unknown')`, job.State, runnerRawJSON(job), job.ID)
			if err != nil {
				return err
			}
			return st.runnerEvent(*execution, "job", job)
		}); err != nil {
			return err
		}
	}
	return nil
}
