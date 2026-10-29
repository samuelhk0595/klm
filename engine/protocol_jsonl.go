package main

import (
	"bytes"
	"context"
	"errors"
)

type protocolLineInput struct {
	ctx     context.Context
	dir     string
	harness string
}

func (l *jsonLines) writeProjected(data []byte) (int, error) {
	n := len(data)
	for len(data) > 0 {
		end := bytes.IndexByte(data, '\n')
		part := data
		if end >= 0 {
			part = data[:end]
		}
		if l.stage == nil {
			l.stage = &protocolFrame{ctx: l.projection.ctx, dir: l.projection.dir}
		}
		if _, err := l.stage.Write(part); err != nil {
			return 0, l.reject(protocolFrameError(l.projection.harness, "", l.stage.size, err).Error())
		}
		if end < 0 {
			break
		}
		if err := l.flushProjected(); err != nil {
			return 0, err
		}
		data = data[end+1:]
	}
	return n, nil
}

func (l *jsonLines) flushProjected() error {
	f := l.stage
	if f == nil {
		return nil
	}
	l.stage = nil
	var data []byte
	var err error
	if f.size > 0 && (f.file != nil || len(bytes.TrimSpace(f.memory)) > 0) {
		data, err = f.project(l.projection.harness, 4<<20)
	}
	err = errors.Join(err, f.Close())
	if err != nil {
		return l.reject(err.Error())
	}
	if len(data) > 0 {
		if err := l.consume(data); err != nil {
			return l.reject(err.Error())
		}
	}
	return nil
}

func (l *jsonLines) closeStaging() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.stage != nil {
		l.err = errors.Join(l.err, l.stage.Close())
		l.stage = nil
	}
}
