package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os/exec"
	"sync"
	"time"
)

type interactiveProcess struct {
	Frames <-chan map[string]any
	Done   <-chan struct{}
	stdin  io.WriteCloser
	cancel context.CancelFunc
	mu     sync.Mutex
	err    error
}

func startInteractive(t *turn, b binary, args []string, cwd string) (*interactiveProcess, error) {
	ctx, cancel := context.WithCancel(t.ctx)
	cmd := exec.CommandContext(ctx, b.path, append(append([]string{}, b.args...), args...)...)
	cmd.Dir = cwd
	configureProcess(cmd)
	cmd.Cancel = func() error { return killTree(cmd.Process) }
	cmd.WaitDelay = 5 * time.Second
	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return nil, errors.New("Could not open harness input.")
	}
	frames := make(chan map[string]any, 64)
	done := make(chan struct{})
	p := &interactiveProcess{Frames: frames, Done: done, stdin: stdin, cancel: cancel}
	lines := &jsonLines{cancel: cancel, consume: func(line []byte) error {
		var frame map[string]any
		d := json.NewDecoder(bytes.NewReader(line))
		d.UseNumber()
		if err := d.Decode(&frame); err != nil || frame == nil {
			return errors.New("Harness returned invalid protocol JSON.")
		}
		select {
		case frames <- frame:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}}
	cmd.Stdout = lines
	cmd.Stderr = &cappedBuffer{}
	if err := cmd.Start(); err != nil {
		cancel()
		stdin.Close()
		return nil, errors.New("Could not start the harness. Check its installation.")
	}
	go func() {
		err := cmd.Wait()
		if ctx.Err() == nil {
			_ = lines.flush()
		}
		if lines.err != nil {
			p.err = lines.err
		} else if err != nil {
			p.err = errors.New("Harness process exited unsuccessfully. Check its authentication and configuration.")
		}
		close(frames)
		close(done)
	}()
	return p, nil
}

func (p *interactiveProcess) Send(value any) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	select {
	case <-p.Done:
		return errors.New("Harness connection has closed.")
	default:
	}
	return json.NewEncoder(p.stdin).Encode(value)
}

func (p *interactiveProcess) Err() error { <-p.Done; return p.err }

func (p *interactiveProcess) Close() error {
	p.cancel()
	_ = p.stdin.Close()
	<-p.Done
	return p.err
}
