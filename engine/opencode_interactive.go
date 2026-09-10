package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

const openCodeFrameLimit = 2 << 20
const openCodeResponseLimit = 32 << 20

type openCodeHTTP struct {
	client    *http.Client
	base      string
	directory string
	password  string
}

func (h *openCodeHTTP) request(ctx context.Context, method, path string, body any) (*http.Response, error) {
	var data []byte
	var err error
	if body != nil {
		data, err = json.Marshal(body)
		if err != nil {
			return nil, errors.New("Could not encode the OpenCode request.")
		}
		if len(data) > openCodeResponseLimit {
			return nil, errors.New("OpenCode request exceeded the 32 MiB limit.")
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, h.base+path+"?directory="+url.QueryEscape(h.directory), bytes.NewReader(data))
	if err != nil {
		return nil, errors.New("Could not construct the OpenCode request.")
	}
	req.SetBasicAuth("opencode", h.password)
	req.Header.Set("Accept", "application/json")
	if path == "/event" {
		req.Header.Set("Accept", "text/event-stream")
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	response, err := h.client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		// Neither diagnostics nor HTTP bodies are safe to expose as transport errors.
		return nil, errors.New("Could not communicate with the owned OpenCode server.")
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		response.Body.Close()
		return nil, fmt.Errorf("OpenCode %s request failed (HTTP %d).", method, response.StatusCode)
	}
	return response, nil
}

func (h *openCodeHTTP) json(ctx context.Context, method, path string, body, result any) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	response, err := h.request(ctx, method, path, body)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	data, err := io.ReadAll(io.LimitReader(response.Body, openCodeResponseLimit+1))
	if err != nil {
		return errors.New("Could not read the OpenCode response.")
	}
	if len(data) > openCodeResponseLimit {
		return errors.New("OpenCode response exceeded the 32 MiB limit.")
	}
	if result != nil {
		if err := json.Unmarshal(data, result); err != nil {
			return errors.New("OpenCode returned an invalid JSON response.")
		}
	}
	return nil
}

// SSE frames, including multi-line data fields, have the same bound as CLI JSON lines.
func openCodeEvents(ctx context.Context, body io.Reader, events chan<- map[string]any, failed chan<- error) {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 4096), openCodeFrameLimit+1)
	var data []byte
	frameBytes := 0
	err := errors.New("OpenCode event stream disconnected before the turn completed.")
	for scanner.Scan() {
		line := scanner.Bytes()
		frameBytes += len(line) + 1
		if frameBytes > openCodeFrameLimit {
			err = errors.New("OpenCode event frame exceeded the 2 MiB limit.")
			break
		}
		if len(line) == 0 {
			if len(data) > 0 {
				var event map[string]any
				if json.Unmarshal(data, &event) != nil || str(event, "type") == "" {
					err = errors.New("OpenCode returned an invalid event frame.")
					break
				}
				select {
				case events <- event:
				case <-ctx.Done():
					return
				}
			}
			data = data[:0]
			frameBytes = 0
		} else if bytes.HasPrefix(line, []byte("data:")) {
			value := bytes.TrimPrefix(line[5:], []byte(" "))
			if len(data) > 0 {
				data = append(data, '\n')
			}
			data = append(data, value...)
		}
	}
	if scanner.Err() != nil {
		err = errors.New("OpenCode event stream could not be read or exceeded the 2 MiB frame limit.")
	}
	select {
	case failed <- err:
	case <-ctx.Done():
	}
}

func (p *adapter) runOpenCode(b binary, cwd, text string) error {
	p.app.mu.Lock()
	s := *p.app.state.session(p.id)
	native := p.app.state.Native[p.id]
	p.app.mu.Unlock()
	t := p.turn
	ctx, cancel := context.WithCancel(t.ctx)
	defer cancel()
	p.completed = false

	var secret [32]byte
	if _, err := rand.Read(secret[:]); err != nil {
		return errors.New("Could not generate OpenCode server credentials.")
	}
	h := &openCodeHTTP{directory: cwd, password: hex.EncodeToString(secret[:])}
	transport := &http.Transport{
		DialContext:           (&net.Dialer{Timeout: 5 * time.Second}).DialContext,
		ResponseHeaderTimeout: 30 * time.Second,
		IdleConnTimeout:       30 * time.Second,
	}
	h.client = &http.Client{Transport: transport, CheckRedirect: func(*http.Request, []*http.Request) error {
		return errors.New("OpenCode server redirects are not allowed.")
	}}
	defer transport.CloseIdleConnections()

	args := append(append([]string{}, b.args...), "serve", "--hostname", "127.0.0.1", "--port", "0", "--mdns=false")
	cmd := exec.Command(b.path, args...)
	cmd.Dir = cwd
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if !strings.EqualFold(key, "OPENCODE_SERVER_PASSWORD") && !strings.EqualFold(key, "OPENCODE_SERVER_USERNAME") &&
			!strings.EqualFold(key, "OPENCODE_ENABLE_QUESTION_TOOL") {
			cmd.Env = append(cmd.Env, entry)
		}
	}
	cmd.Env = append(cmd.Env, "OPENCODE_SERVER_USERNAME=opencode", "OPENCODE_SERVER_PASSWORD="+h.password, "OPENCODE_ENABLE_QUESTION_TOOL=true")
	configureProcess(cmd)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return errors.New("Could not open OpenCode server output.")
	}
	defer stdout.Close()
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return errors.New("Could not open OpenCode server diagnostics.")
	}
	defer stderr.Close()
	announced := make(chan string, 1)
	outputFailed := make(chan error, 1)
	go func() {
		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 4096), openCodeFrameLimit+1)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if address, ok := strings.CutPrefix(line, "opencode server listening on "); ok {
				select {
				case announced <- address:
				default:
				}
			}
		}
		if scanner.Err() != nil {
			outputFailed <- errors.New("Could not read OpenCode server output within the 2 MiB line limit.")
		}
	}()
	go func() { _, _ = io.Copy(io.Discard, stderr) }()
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return errors.New("Could not start the OpenCode server. Check its installation and configuration.")
	}
	processDone := make(chan struct{})
	go func() {
		_ = cmd.Wait()
		close(processDone)
	}()
	sessionID := ""
	promptAttempted := false
	defer func() {
		// Cancel local subscriptions/replies first, then abort only our native turn.
		cancel()
		if sessionID != "" && promptAttempted && !p.completed {
			abortCtx, abortCancel := context.WithTimeout(context.Background(), 2*time.Second)
			_ = h.json(abortCtx, http.MethodPost, "/session/"+url.PathEscape(sessionID)+"/abort", nil, nil)
			abortCancel()
		}
		select {
		case <-processDone:
			return
		default:
			_ = killTree(cmd.Process)
		}
		select {
		case <-processDone:
		case <-time.After(5 * time.Second):
		}
	}()

	startup := time.NewTimer(30 * time.Second)
	defer startup.Stop()
	select {
	case address := <-announced:
		u, parseErr := url.Parse(address)
		if parseErr != nil || u.Scheme != "http" || u.Hostname() != "127.0.0.1" || u.User != nil ||
			u.Path != "" || u.RawQuery != "" || u.Fragment != "" || u.Opaque != "" {
			return errors.New("OpenCode announced an invalid loopback server address.")
		}
		port, parseErr := strconv.Atoi(u.Port())
		if parseErr != nil || port < 1 || port > 65535 || u.Host != net.JoinHostPort("127.0.0.1", u.Port()) {
			return errors.New("OpenCode announced an invalid loopback server port.")
		}
		h.base = u.String()
	case err := <-outputFailed:
		return err
	case <-processDone:
		return errors.New("OpenCode server exited before startup completed.")
	case <-startup.C:
		return errors.New("OpenCode server startup timed out.")
	case <-ctx.Done():
		return ctx.Err()
	}

	var bridgeStatus map[string]any
	if err := h.json(ctx, http.MethodPost, "/mcp", map[string]any{"name": "klm_linked", "config": map[string]any{"type": "remote", "url": p.bridge.url, "headers": map[string]string{"Authorization": "Bearer " + p.bridge.token}, "oauth": false, "enabled": true, "timeout": 10000}}, &bridgeStatus); err != nil {
		return errors.New("Could not configure the OpenCode linked-agent bridge.")
	}
	if str(object(bridgeStatus["klm_linked"]), "status") != "connected" {
		return errors.New("OpenCode did not connect to the linked-agent bridge.")
	}
	if err := p.bridge.waitReady(ctx); err != nil {
		return err
	}
	var session map[string]any
	if native.ID != "" {
		if err := h.json(ctx, http.MethodGet, "/session/"+url.PathEscape(native.ID), nil, &session); err != nil {
			return fmt.Errorf("Could not resume the stored OpenCode session: %w", err)
		}
		if str(session, "id") != native.ID {
			return errors.New("OpenCode returned a different native session than requested.")
		}
	} else if err := h.json(ctx, http.MethodPost, "/session", map[string]any{"title": s.Title}, &session); err != nil {
		return err
	}
	wanted, err := filepath.Abs(cwd)
	if err != nil {
		return errors.New("Could not resolve the OpenCode working directory.")
	}
	actual := str(session, "directory")
	if actual == "" || !filepath.IsAbs(actual) {
		return errors.New("OpenCode returned a session without an absolute working directory.")
	}
	wanted, actual = filepath.Clean(wanted), filepath.Clean(actual)
	match := wanted == actual
	if runtime.GOOS == "windows" {
		match = strings.EqualFold(wanted, actual)
	}
	if !match {
		return errors.New("Stored OpenCode session belongs to a different working directory.")
	}
	sessionID = str(session, "id")
	if sessionID == "" {
		return errors.New("OpenCode returned a session without an identifier.")
	}
	if err := p.nativeID(sessionID); err != nil {
		return err
	}
	sessionPath := "/session/" + url.PathEscape(sessionID)
	var history []map[string]any
	if err := h.json(ctx, http.MethodGet, sessionPath+"/message", nil, &history); err != nil {
		return err
	}
	baseline := make(map[string]bool, len(history))
	usage := &openCodeUsage{messages: map[string]map[string]any{}, steps: map[string]map[string]map[string]any{}, windows: map[string]*int64{}}
	var providers struct {
		All []struct {
			ID     string `json:"id"`
			Models map[string]struct {
				Limit struct {
					Context float64 `json:"context"`
				} `json:"limit"`
			} `json:"models"`
		} `json:"all"`
	}
	// Model limits are optional telemetry; a failed lookup must not block execution.
	metricsCtx, metricsCancel := context.WithTimeout(ctx, 3*time.Second)
	if h.json(metricsCtx, http.MethodGet, "/provider", nil, &providers) == nil {
		for _, provider := range providers.All {
			for id, model := range provider.Models {
				usage.windows[provider.ID+"/"+id] = tokenCount(model.Limit.Context)
			}
		}
	}
	metricsCancel()
	for _, message := range history {
		info := object(message["info"])
		if str(info, "sessionID") != sessionID {
			return errors.New("OpenCode history contained a different native session.")
		}
		usage.message(info)
		parts, _ := message["parts"].([]any)
		for _, value := range parts {
			part := object(value)
			if str(part, "sessionID") == sessionID && str(part, "messageID") == str(info, "id") {
				usage.part(part)
			}
		}
		if id := str(info, "id"); id != "" {
			baseline[id] = true
		}
	}
	if err := p.setUsage(usage.snapshot()); err != nil {
		return err
	}
	history = nil

	stream, err := h.request(ctx, http.MethodGet, "/event", nil)
	if err != nil {
		return err
	}
	defer stream.Body.Close()
	if !strings.HasPrefix(strings.ToLower(stream.Header.Get("Content-Type")), "text/event-stream") {
		return errors.New("OpenCode did not provide an event stream.")
	}
	events := make(chan map[string]any, 64)
	streamFailed := make(chan error, 1)
	go openCodeEvents(ctx, stream.Body, events, streamFailed)

	permissions := map[string]bool{}
	questions := map[string]*atomic.Bool{}
	defer func() {
		cancel()
		for id, live := range questions {
			live.Store(false)
			_ = p.dismissQuestion(id)
		}
	}()
	requestQuestion := func(properties map[string]any) error {
		if str(properties, "sessionID") != sessionID {
			return nil
		}
		id := str(properties, "id")
		if strings.TrimSpace(id) == "" {
			return errors.New("OpenCode returned a question without an identifier.")
		}
		if questions[id] != nil {
			return nil
		}
		encoded, err := json.Marshal(properties["questions"])
		if err != nil || len(encoded) > openCodeFrameLimit {
			return errors.New("OpenCode question request exceeded the 2 MiB limit.")
		}
		var nativeQuestions []struct {
			Question string           `json:"question"`
			Header   string           `json:"header"`
			Options  []QuestionOption `json:"options"`
			Multiple bool             `json:"multiple"`
			Custom   *bool            `json:"custom"`
		}
		if json.Unmarshal(encoded, &nativeQuestions) != nil {
			return errors.New("OpenCode returned an invalid question request.")
		}
		req := QuestionRequest{SourceID: id}
		for index, question := range nativeQuestions {
			req.Items = append(req.Items, QuestionItem{
				ID: strconv.Itoa(index), Header: question.Header, Text: question.Question,
				Options: question.Options, Multiple: question.Multiple,
				Custom: question.Custom == nil || *question.Custom,
			})
		}
		live := &atomic.Bool{}
		live.Store(true)
		questions[id] = live
		return p.requestQuestion(req, func(answers [][]string, cancelled bool) error {
			if t.ctx.Err() != nil || ctx.Err() != nil || !live.CompareAndSwap(true, false) {
				return errors.New("The OpenCode question is no longer active.")
			}
			path := "/question/" + url.PathEscape(id)
			var body any
			if cancelled {
				path += "/reject"
			} else {
				path += "/reply"
				body = map[string]any{"answers": answers}
			}
			var accepted bool
			if err := h.json(ctx, http.MethodPost, path, body, &accepted); err != nil {
				return err
			}
			if !accepted {
				return errors.New("OpenCode did not accept the question response.")
			}
			return nil
		})
	}
	requestPermission := func(properties map[string]any) error {
		if str(properties, "sessionID") != sessionID {
			return nil
		}
		encoded, err := json.Marshal(properties)
		if err != nil || len(encoded) > openCodeFrameLimit {
			return errors.New("OpenCode permission request exceeded the 2 MiB limit.")
		}
		id, kind := str(properties, "id"), str(properties, "permission")
		if id == "" || kind == "" {
			return errors.New("OpenCode returned an invalid permission request.")
		}
		if permissions[id] {
			return nil
		}
		if p.internalOpenCodeTool(kind) {
			if err := h.json(ctx, http.MethodPost, "/permission/"+url.PathEscape(id)+"/reply", map[string]any{"reply": "once"}, nil); err != nil {
				return err
			}
			permissions[id] = true
			return nil
		}
		patterns := []string{}
		values, ok := properties["patterns"].([]any)
		if !ok {
			return errors.New("OpenCode permission request has invalid patterns.")
		}
		for _, value := range values {
			pattern, ok := value.(string)
			if !ok {
				return errors.New("OpenCode permission request has an invalid pattern.")
			}
			patterns = append(patterns, pattern)
		}
		sort.Strings(patterns)
		metadata := map[string]any{}
		for key, value := range object(properties["metadata"]) {
			switch key {
			case "sessionID", "messageID", "partID", "callID", "toolCallID", "requestID", "permissionID":
				continue
			}
			metadata[key] = value
		}
		var scope map[string]any
		if len(patterns) > 0 {
			// Exact semantic equality only; never use native `always` patterns or IDs.
			scope = map[string]any{"permission": kind, "patterns": patterns, "metadata": metadata, "directory": wanted}
		}
		description := strings.Join(patterns, "\n")
		if detail := str(metadata, "description"); detail != "" {
			description = detail + "\n" + description
		}
		req := Permission{
			ID: newID(), Harness: p.harness, Kind: kind, Title: "Allow " + kind,
			Description: strings.TrimSpace(description), Patterns: patterns,
			Details: metadata, AllowLabel: "Allow", CreatedAt: now(), SourceID: id,
		}
		if scope == nil {
			req.Decisions = []string{"once", "reject"}
		}
		if err := p.requestPermission(req, scope, func(allow bool) error {
			if err := t.ctx.Err(); err != nil {
				return err
			}
			if err := ctx.Err(); err != nil {
				return err
			}
			reply := "reject"
			if allow {
				reply = "once"
			}
			return h.json(ctx, http.MethodPost, "/permission/"+url.PathEscape(id)+"/reply", map[string]any{"reply": reply}, nil)
		}); err != nil {
			return err
		}
		permissions[id] = true
		return nil
	}
	var pending []map[string]any
	if err := h.json(ctx, http.MethodGet, "/permission", nil, &pending); err != nil {
		return err
	}
	for _, req := range pending {
		if err := requestPermission(req); err != nil {
			return err
		}
	}
	pending = nil
	if err := h.json(ctx, http.MethodGet, "/question", nil, &pending); err != nil {
		return err
	}
	for _, req := range pending {
		if err := requestQuestion(req); err != nil {
			return err
		}
	}

	infos := map[string]map[string]any{}
	parts := map[string]map[string]any{}
	putPart := func(part map[string]any) error {
		messageID, partID := str(part, "messageID"), str(part, "id")
		if baseline[messageID] || str(infos[messageID], "role") != "assistant" {
			return nil
		}
		if messageID == "" || partID == "" || str(part, "sessionID") != sessionID {
			return errors.New("OpenCode returned an invalid message part.")
		}
		encoded, err := json.Marshal(part)
		if err != nil || len(encoded) > openCodeFrameLimit {
			return errors.New("OpenCode message part exceeded the 2 MiB limit.")
		}
		kind, title, body, status := str(part, "type"), "", str(part, "text"), "running"
		raw := map[string]any{"type": kind, "partID": partID, "messageID": messageID}
		switch kind {
		case "text", "reasoning":
			if kind == "text" {
				kind = "assistant"
			} else if strings.TrimSpace(body) == "" {
				return nil
			}
			if object(part["time"])["end"] != nil || object(infos[messageID]["time"])["completed"] != nil {
				status = "completed"
			}
			// Reasoning metadata can contain encrypted provider state; retain text only.
		case "tool":
			state := object(part["state"])
			title, body, status = str(part, "tool"), contentText(state["output"]), str(state, "status")
			if status == "" {
				status = "pending"
			}
			if status == "error" {
				body = humanError(state["error"])
			}
			if title == "bash" {
				kind = "command"
				if command := str(object(state["input"]), "command"); command != "" {
					body = strings.TrimSuffix(command+"\n"+body, "\n")
				}
			} else if strings.HasPrefix(title, "mcp_") || strings.HasPrefix(title, "mcp.") || strings.Contains(title, "__") {
				kind = "mcp"
			}
			// Opaque tool names stay intact; do not guess that every custom tool is MCP.
			raw["tool"], raw["callID"], raw["state"] = title, part["callID"], state
		default:
			return nil
		}
		return p.put("opencode/"+messageID+"/"+partID, kind, title, body, status, false, raw)
	}

	prompt := map[string]any{"parts": []any{map[string]any{"type": "text", "text": text}}}
	if p.model != "" {
		provider, model, ok := strings.Cut(p.model, "/")
		if !ok || provider == "" || model == "" {
			return errors.New("OpenCode model must use provider/model format.")
		}
		prompt["model"] = map[string]any{"providerID": provider, "modelID": model}
	}
	if p.effort != "" {
		prompt["variant"] = p.effort
	}
	select {
	case err := <-streamFailed:
		return err
	case <-ctx.Done():
		return ctx.Err()
	case <-processDone:
		return errors.New("OpenCode server exited before prompt submission.")
	default:
	}
	promptAttempted = true
	if err := h.json(ctx, http.MethodPost, sessionPath+"/prompt_async", prompt, nil); err != nil {
		return err
	}
	// Events queued during prompt_async are processed only after submission succeeds.
	busy, ownedMessage := false, false
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-processDone:
			return errors.New("OpenCode server exited before the turn completed.")
		case err := <-outputFailed:
			return err
		case err := <-streamFailed:
			return err
		case event := <-events:
			typ, properties := str(event, "type"), object(event["properties"])
			scopedID := str(properties, "sessionID")
			if typ == "message.updated" {
				scopedID = str(object(properties["info"]), "sessionID")
			} else if typ == "message.part.updated" {
				scopedID = str(object(properties["part"]), "sessionID")
			} else if typ == "message.part.delta" && scopedID == "" {
				part := parts[str(properties, "messageID")+"/"+str(properties, "partID")]
				scopedID = str(part, "sessionID")
			}
			if scopedID != sessionID {
				continue
			}
			switch typ {
			case "session.compacted":
				for _, info := range usage.messages {
					if created := tokenCount(object(info["time"])["created"]); created != nil && *created > usage.compactedAt {
						usage.compactedAt = *created
					}
				}
				if err := p.setUsage(usage.snapshot()); err != nil {
					return err
				}
			case "permission.asked":
				if err := requestPermission(properties); err != nil {
					return err
				}
			case "permission.replied":
				id := str(properties, "requestID")
				if id == "" {
					return errors.New("OpenCode returned an invalid permission resolution.")
				}
				permissions[id] = true
				if err := p.dismissPermission(id); err != nil {
					return err
				}
			case "question.asked":
				if err := requestQuestion(properties); err != nil {
					return err
				}
			case "question.replied", "question.rejected":
				id := str(properties, "requestID")
				if strings.TrimSpace(id) == "" {
					return errors.New("OpenCode returned an invalid question resolution.")
				}
				if questions[id] == nil {
					questions[id] = &atomic.Bool{}
				}
				questions[id].Store(false)
				if err := p.dismissQuestion(id); err != nil {
					return err
				}
			case "session.error":
				return p.failure(map[string]any{"type": typ, "error": properties["error"]})
			case "message.updated":
				info := object(properties["info"])
				id := str(info, "id")
				if id == "" || baseline[id] {
					continue
				}
				ownedMessage = true
				infos[id] = info
				usage.message(info)
				if str(info, "role") == "assistant" {
					if err := p.resolvedSelection(str(info, "providerID")+"/"+str(info, "modelID"), str(info, "variant")); err != nil {
						return err
					}
				}
				if err := p.setUsage(usage.snapshot()); err != nil {
					return err
				}
				if str(info, "role") == "assistant" && info["error"] != nil {
					return p.failure(map[string]any{"type": typ, "error": info["error"]})
				}
				for _, part := range parts {
					if str(part, "messageID") == id {
						if err := putPart(part); err != nil {
							return err
						}
					}
				}
			case "message.part.updated":
				part := object(properties["part"])
				messageID, partID := str(part, "messageID"), str(part, "id")
				if baseline[messageID] {
					continue
				}
				if messageID == "" || partID == "" {
					return errors.New("OpenCode returned a message part without identifiers.")
				}
				parts[messageID+"/"+partID] = part
				usage.part(part)
				if str(part, "type") == "step-finish" {
					if err := p.setUsage(usage.snapshot()); err != nil {
						return err
					}
				}
				if err := putPart(part); err != nil {
					return err
				}
			case "message.part.delta":
				if str(properties, "field") != "text" {
					continue
				}
				part := parts[str(properties, "messageID")+"/"+str(properties, "partID")]
				if part == nil || (str(part, "type") != "text" && str(part, "type") != "reasoning") {
					// A missing initial part is recovered from the final snapshot.
					continue
				}
				delta := str(properties, "delta")
				if len(str(part, "text"))+len(delta) > openCodeFrameLimit {
					return errors.New("OpenCode text part exceeded the 2 MiB limit.")
				}
				part["text"] = str(part, "text") + delta
				if err := putPart(part); err != nil {
					return err
				}
			case "session.status", "session.idle":
				status := str(object(properties["status"]), "type")
				if status == "busy" || status == "retry" {
					busy = true
					continue
				}
				if (typ != "session.idle" && status != "idle") || (!busy && !ownedMessage) {
					continue
				}
				if err := h.json(ctx, http.MethodGet, sessionPath+"/message", nil, &history); err != nil {
					return err
				}
				terminalAssistant, incomplete := false, false
				latestCreated, latestID := float64(-1), ""
				for _, message := range history {
					info := object(message["info"])
					id := str(info, "id")
					if id == "" || baseline[id] || str(info, "role") != "assistant" {
						continue
					}
					if str(info, "sessionID") != sessionID {
						return errors.New("OpenCode history contained a different native session.")
					}
					infos[id] = info
					usage.message(info)
					if info["error"] != nil {
						return p.failure(map[string]any{"type": "message.updated", "error": info["error"]})
					}
					completed := object(info["time"])["completed"] != nil
					if !completed {
						incomplete = true
					}
					created, _ := object(info["time"])["created"].(float64)
					if created > latestCreated || (created == latestCreated && id > latestID) {
						latestCreated, latestID = created, id
						finish := str(info, "finish")
						terminalAssistant = completed && finish != "" && finish != "tool-calls" && finish != "unknown"
					}
					completeParts, ok := message["parts"].([]any)
					if !ok {
						return errors.New("OpenCode history did not contain message parts.")
					}
					for _, value := range completeParts {
						part := object(value)
						if str(part, "messageID") != id {
							return errors.New("OpenCode history contained an invalid message part.")
						}
						usage.part(part)
						if str(part, "type") == "tool" {
							state := str(object(part["state"]), "status")
							if state != "completed" && state != "error" {
								incomplete = true
							}
						}
						if err := putPart(part); err != nil {
							return err
						}
					}
				}
				if err := p.setUsage(usage.snapshot()); err != nil {
					return err
				}
				if incomplete || !terminalAssistant {
					return errors.New("OpenCode became idle without a completed assistant response.")
				}
				select {
				case err := <-streamFailed:
					return err
				case <-ctx.Done():
					return ctx.Err()
				case <-processDone:
					return errors.New("OpenCode server exited during final reconciliation.")
				default:
				}
				if err := p.put("opencode/completed", "status", "", "Turn completed.", "completed", false, map[string]any{"type": typ}); err != nil {
					return err
				}
				p.completed = true
				return nil
			}
		}
	}
}
