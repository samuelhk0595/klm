//go:build !dev

package main

func projectGitIdentityVerifierForBuild() projectGitIdentityVerifier {
	return verifyProjectGitToken
}
