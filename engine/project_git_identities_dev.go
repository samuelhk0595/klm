//go:build dev

package main

import (
	"context"
	"os"
)

func projectGitIdentityVerifierForBuild() projectGitIdentityVerifier {
	if os.Getenv("KLM_DEV_SYNTHETIC_GIT_IDENTITIES") == "1" {
		return syntheticProjectGitIdentityVerifier
	}
	return verifyProjectGitToken
}

// Only these exact synthetic candidates are accepted, with no external requests.
func syntheticProjectGitIdentityVerifier(_ context.Context, token string) (string, error) {
	switch token {
	case "klm-synthetic-github-alice":
		return "synthetic-alice", nil
	case "klm-synthetic-github-bob":
		return "synthetic-bob", nil
	default:
		return "", sqliteReject(400, "GitHub could not verify this token.")
	}
}
