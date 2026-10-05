//go:build !windows

package main

import "errors"

func protectTaskSecret(string) (string, error) {
	return "", errors.New("Task webhooks require Windows secret protection.")
}
func unprotectTaskSecret(string) (string, error) {
	return "", errors.New("Task webhooks require Windows secret protection.")
}
