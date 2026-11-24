package main

import (
	"bytes"
	"encoding/base64"
	"io"
)

// Preserve streaming output while withholding a possible credential suffix until
// the next chunk. Native CLI errors/debug changes cannot echo the selected token.
type projectGitRedactor struct {
	target  io.Writer
	secrets [][]byte
	pending []byte
	keep    int
}

func newProjectGitRedactor(target io.Writer, token string) *projectGitRedactor {
	if target == nil {
		target = io.Discard
	}
	secrets := [][]byte{
		[]byte(token),
		[]byte(base64.StdEncoding.EncodeToString([]byte("x-access-token:" + token))),
	}
	keep := 0
	for _, secret := range secrets {
		if len(secret)-1 > keep {
			keep = len(secret) - 1
		}
	}
	return &projectGitRedactor{target: target, secrets: secrets, pending: []byte{}, keep: keep}
}

func (r *projectGitRedactor) Write(data []byte) (int, error) {
	r.pending = append(r.pending, data...)
	for {
		first, length := -1, 0
		for _, secret := range r.secrets {
			if len(secret) == 0 {
				continue
			}
			index := bytes.Index(r.pending, secret)
			if index >= 0 && (first < 0 || index < first) {
				first, length = index, len(secret)
			}
		}
		if first < 0 {
			break
		}
		if _, err := r.target.Write(r.pending[:first]); err != nil {
			return 0, err
		}
		if _, err := io.WriteString(r.target, "[redacted]"); err != nil {
			return 0, err
		}
		r.pending = r.pending[first+length:]
	}
	if count := len(r.pending) - r.keep; count > 0 {
		if _, err := r.target.Write(r.pending[:count]); err != nil {
			return 0, err
		}
		r.pending = append(r.pending[:0], r.pending[count:]...)
	}
	return len(data), nil
}

func (r *projectGitRedactor) flush() error {
	_, err := r.target.Write(r.pending)
	r.pending = r.pending[:0]
	return err
}
