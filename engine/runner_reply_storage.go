package main

import "context"

func (a *app) protectRunnerReply(body []byte) ([]byte, error) {
	var existing bool
	err := a.store.readers.QueryRowContext(context.Background(), existingProtectedCredentialsSQL).Scan(&existing)
	if err != nil {
		return nil, err
	}
	return protectMCPCredential(body, mcpCredentialStorage{Directory: a.dir, Existing: existing})
}
