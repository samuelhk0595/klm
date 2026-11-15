package main

import "context"

func (a *app) protectRunnerReply(body []byte) ([]byte, error) {
	var existing bool
	err := a.store.readers.QueryRowContext(context.Background(), `SELECT EXISTS(SELECT 1 FROM data_mcp_credentials) OR EXISTS(SELECT 1 FROM runner_requests WHERE reply IS NOT NULL)`).Scan(&existing)
	if err != nil {
		return nil, err
	}
	return protectMCPCredential(body, mcpCredentialStorage{Directory: a.dir, Existing: existing})
}
