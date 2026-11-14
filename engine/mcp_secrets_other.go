//go:build !windows && !linux

package main

import "errors"

// Do not silently persist plaintext credentials on an unsupported platform.
func protectMCPCredential([]byte, ...mcpCredentialStorage) ([]byte, error) {
	return nil, errors.New("Protected MCP credential storage is currently supported on Windows only.")
}
func unprotectMCPCredential([]byte, ...mcpCredentialStorage) ([]byte, error) {
	return nil, errors.New("Protected MCP credential storage is currently supported on Windows only.")
}
