package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Harness struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Available bool   `json:"available"`
	Error     string `json:"error,omitempty"`
}

type binary struct {
	path string
	args []string
}

func executable(path string) (string, bool) {
	resolved, err := exec.LookPath(path)
	if err != nil {
		return "", false
	}
	// Never execute Windows npm shell shims, or inspect/evaluate their contents.
	if runtime.GOOS == "windows" && !strings.EqualFold(filepath.Ext(resolved), ".exe") {
		return "", false
	}
	resolved, err = filepath.Abs(resolved)
	return resolved, err == nil
}

func discoverHarnesses() ([]Harness, map[string]binary) {
	harnesses := []Harness{{ID: "pi", Name: "Pi"}, {ID: "opencode", Name: "OpenCode"}, {ID: "codex", Name: "Codex"}}
	binaries := map[string]binary{}
	npmRoots := []string{}
	if appdata := os.Getenv("APPDATA"); appdata != "" {
		npmRoots = append(npmRoots, filepath.Join(appdata, "npm", "node_modules"))
	}
	for _, shim := range []string{"pi", "opencode"} {
		if path, err := exec.LookPath(shim); err == nil {
			npmRoots = append(npmRoots, filepath.Join(filepath.Dir(path), "node_modules"))
		}
	}
	for i := range harnesses {
		h := &harnesses[i]
		override := os.Getenv("KLM_" + strings.ToUpper(h.ID) + "_BIN")
		candidates := []string{}
		if override != "" {
			candidates = append(candidates, override)
		} else {
			if h.ID == "pi" {
				for _, root := range npmRoots {
					candidates = append(candidates, filepath.Join(root, "@earendil-works", "pi-coding-agent", "dist", "cli.js"))
				}
			} else if h.ID == "opencode" {
				for _, root := range npmRoots {
					candidates = append(candidates, filepath.Join(root, "opencode-ai", "bin", "opencode.exe"))
				}
			} else if local := os.Getenv("LOCALAPPDATA"); local != "" {
				candidates = append(candidates, filepath.Join(local, "Programs", "OpenAI", "Codex", "bin", "codex.exe"))
			}
			if runtime.GOOS == "windows" {
				candidates = append(candidates, h.ID+".exe")
			} else {
				candidates = append(candidates, h.ID)
			}
		}
		for _, candidate := range candidates {
			if h.ID == "pi" && strings.EqualFold(filepath.Ext(candidate), ".js") {
				info, err := os.Stat(candidate)
				if err != nil || !info.Mode().IsRegular() {
					continue
				}
				node, ok := executable("node")
				if !ok {
					continue
				}
				path, err := filepath.Abs(candidate)
				if err != nil {
					continue
				}
				binaries[h.ID] = binary{path: node, args: []string{path}}
				h.Available = true
				break
			}
			if path, ok := executable(candidate); ok {
				binaries[h.ID] = binary{path: path}
				h.Available = true
				break
			}
		}
		if !h.Available {
			h.Error = "Executable not found. Install the harness or configure KLM_" + strings.ToUpper(h.ID) + "_BIN and restart the engine."
			if h.ID == "pi" {
				h.Error += " Pi's cli.js also requires Node.js on PATH."
			}
		}
	}
	return harnesses, binaries
}

func commandArgs(s Session, native nativeSession) []string {
	var args []string
	switch s.Harness {
	case "pi":
		args = []string{"--mode", "json", "--print", "--no-extensions", "--session", native.Path}
		if s.Model != "" {
			args = append(args, "--model", s.Model)
		}
	case "opencode":
		args = []string{"run", "--format", "json", "--thinking"}
		if native.ID != "" {
			args = append(args, "--session", native.ID)
		}
		if s.Model != "" {
			args = append(args, "--model", s.Model)
		}
	case "codex":
		args = []string{"exec", "--json", "--color", "never", "--sandbox", "read-only", "--skip-git-repo-check"}
		if s.Model != "" {
			args = append(args, "--model", s.Model)
		}
		if native.ID != "" {
			args = append(args, "resume", native.ID)
		}
		args = append(args, "-")
	}
	return args
}

// stdout and stderr have independent os/exec copy goroutines. Stderr is bounded
// and intentionally never returned or logged, since CLI diagnostics can contain credentials.
type cappedBuffer struct {
	data      []byte
	truncated bool
}

func (b *cappedBuffer) Write(p []byte) (int, error) {
	n := len(p)
	remaining := (16 << 10) - len(b.data)
	if len(p) > remaining {
		p = p[:remaining]
		b.truncated = true
	}
	b.data = append(b.data, p...)
	return n, nil
}

func (b *cappedBuffer) String() string { return string(b.data) }

type jsonLines struct {
	pending []byte
	total   int
	err     error
	consume func([]byte) error
	cancel  func()
}

func (l *jsonLines) Write(p []byte) (int, error) {
	if l.err != nil {
		return 0, l.err
	}
	n := len(p)
	l.total += n
	if l.total > 32<<20 {
		return 0, l.reject("Harness output exceeded the 32 MiB per-turn limit.")
	}
	for len(p) > 0 {
		end := bytes.IndexByte(p, '\n')
		part := p
		if end >= 0 {
			part = p[:end]
		}
		if len(l.pending)+len(part) > 2<<20 {
			return 0, l.reject("Harness JSON line exceeded the 2 MiB limit.")
		}
		l.pending = append(l.pending, part...)
		if end < 0 {
			break
		}
		if err := l.flush(); err != nil {
			return 0, err
		}
		p = p[end+1:]
	}
	return n, nil
}

func (l *jsonLines) reject(text string) error {
	l.err = errors.New(text)
	l.cancel()
	return l.err
}

func (l *jsonLines) flush() error {
	if l.err != nil {
		return l.err
	}
	line := bytes.TrimSpace(bytes.TrimPrefix(l.pending, []byte{0xef, 0xbb, 0xbf}))
	if len(line) > 0 {
		if err := l.consume(line); err != nil {
			return l.reject(err.Error())
		}
	}
	l.pending = l.pending[:0]
	return nil
}

const linkedPromptPrefix = "KLM linked-agent tools are available: linked_discover, linked_read, linked_ask, linked_answer. When the user refers to work in the linked main agent or side agent conversation, retrieve it or consult that agent before continuing. For linked_ask, action=continue means the result is ready: use the answer to respond to the user or continue their task now. Only action=yield means it is still pending: finish the turn so KLM can resume you automatically. A KLM continuation already contains the result and requires no further wait or user follow-up. Do not poll or repeat a pending question.\n\n"

func (a *app) execute(t *turn, s Session, native nativeSession, b binary, cwd string, payload submission) {
	defer a.wg.Done()
	defer t.cancel()
	p := &adapter{app: a, turn: t, id: s.ID, harness: s.Harness,
		keys: map[string]string{}, toolNames: map[string]string{}, commands: map[string]string{}}
	var err error
	// Resolve defaults from the harness catalog, not a previous native turn's override.
	p.model, p.effort = s.Model, s.Effort
	if catalog, catalogErr := a.catalog(t.ctx, s, cwd, false); catalogErr == nil {
		if p.model == "" {
			p.model = catalog.DefaultModel
		}
		if p.effort == "" {
			if option := catalog.find(p.model); option != nil {
				p.effort = option.DefaultEffort
			}
			if p.model == catalog.DefaultModel && catalog.DefaultEffort != "" {
				p.effort = catalog.DefaultEffort
			}
		}
	}
	bridge, bridgeErr := p.startLinkedBridge()
	if bridgeErr != nil {
		err = bridgeErr
	} else {
		p.bridge = bridge
		payload.Text = linkedPromptPrefix + payload.Text
		switch s.Harness {
		case "opencode":
			err = p.runOpenCode(b, cwd, payload)
		case "pi":
			err = p.runPi(b, cwd, payload.piText())
		case "codex":
			err = p.runCodex(b, cwd, payload)
		default:
			err = errors.New("Unsupported harness.")
		}
		bridge.close()
	}
	a.mu.Lock()
	finalErr := a.commitLocked(func(d *diskState) {
		s := d.session(s.ID)
		s.Status, s.UpdatedAt = "idle", now()
		s.Permissions = nil
		s.Questions = nil
		var last Event
		switch {
		case t.ctx.Err() != nil:
			last = event("status", "Execution stopped.")
			last.Status = "cancelled"
		case err != nil:
			s.Status = "error"
			last = event("error", err.Error())
		case p.failed:
			s.Status = "error"
		case !p.completed:
			s.Status = "error"
			last = event("error", "Harness exited without reporting a completed turn.")
		}
		if last.ID != "" {
			last.ConsultationID = t.consultationID
			s.Events = append(s.Events, last)
		}
		// A killed process cannot send terminal updates for its in-flight items.
		for i := range s.Events {
			e := &s.Events[i]
			if e.Status == "running" || e.Status == "pending" {
				if s.Status == "error" {
					e.Status = "error"
				} else {
					e.Status = "cancelled"
				}
			}
		}
		if t.consultationID != "" {
			c := d.consultation(t.consultationID)
			if t.delivery {
				if c.Delivery != "cancelled" {
					c.Delivery = "delivered"
					if t.ctx.Err() != nil || s.Status == "error" {
						c.Delivery = "failed"
					}
				}
			} else if !consultationTerminal(c.Status) {
				c.Status = "failed"
				c.Error = "Agent finished without returning a consultation answer."
				if err != nil {
					c.Error = err.Error()
				}
				if t.ctx.Err() != nil {
					c.Status = "cancelled"
					c.Error = "Consultation stopped."
				}
			}
			syncConsultation(d, c)
		}
		// A failed/stopped requester must not cause an unattended continuation.
		if t.ctx.Err() != nil || s.Status == "error" {
			for i := range d.Consultations {
				c := &d.Consultations[i]
				if c.From == s.ID && (c.Delivery == "waiting" || c.Delivery == "pending") {
					c.Delivery = "cancelled"
					if !consultationTerminal(c.Status) {
						c.Status = "cancelled"
						c.Error = "Requesting turn stopped."
					}
					syncConsultation(d, c)
				}
			}
		}
	})
	if finalErr != nil {
		// No in-memory success is published after a failed disk write. Restart
		// recovers the last durable running snapshot as interrupted.
		p.failed = true
	}
	delete(a.runs, s.ID)
	for _, c := range a.state.Consultations {
		if c.From == s.ID && c.Status == "cancelled" {
			if recipient := a.runs[c.To]; recipient != nil && recipient.consultationID == c.ID {
				recipient.cancel()
			}
		}
	}
	t.approvals = nil
	t.questions = nil
	close(t.done)
	a.scheduleLinkedLocked()
	a.mu.Unlock()
}

type adapter struct {
	bridge    *linkedBridge
	app       *app
	turn      *turn
	id        string
	harness   string
	keys      map[string]string
	toolNames map[string]string
	commands  map[string]string
	piMessage int
	failed    bool
	completed bool
	model     string
	effort    string
}

func object(value any) map[string]any {
	m, _ := value.(map[string]any)
	return m
}

func str(m map[string]any, key string) string {
	s, _ := m[key].(string)
	return s
}

func truth(m map[string]any, key string) bool {
	b, _ := m[key].(bool)
	return b
}

func humanError(value any) string {
	if text, ok := value.(string); ok {
		if len(text) > 4096 {
			return strings.ToValidUTF8(text[:4096], "") + " [truncated]"
		}
		return text
	}
	if m := object(value); m != nil {
		for _, key := range []string{"message", "errorMessage", "error", "data"} {
			if text := humanError(m[key]); text != "" {
				return text
			}
		}
		return str(m, "name")
	}
	return ""
}

func contentText(value any) string {
	if text, ok := value.(string); ok {
		return text
	}
	if m := object(value); m != nil {
		if text := str(m, "text"); text != "" {
			return text
		}
		return contentText(m["content"])
	}
	if parts, ok := value.([]any); ok {
		texts := []string{}
		for _, part := range parts {
			if text := contentText(part); text != "" {
				texts = append(texts, text)
			}
		}
		return strings.Join(texts, "\n")
	}
	return ""
}

// Stable keys upsert harness items in first-observed order. Only Pi deltas
// append text; accumulated tool results and completed messages replace text.
func (p *adapter) put(key, kind, title, text, status string, appendText bool, raw map[string]any) error {
	id := p.keys[key]
	if key == "" || id == "" {
		id = newID()
		if key != "" {
			p.keys[key] = id
		}
	}
	p.app.mu.Lock()
	defer p.app.mu.Unlock()
	return p.app.commitLocked(func(d *diskState) {
		s := d.session(p.id)
		var e *Event
		for i := len(s.Events) - 1; i >= 0; i-- {
			if s.Events[i].ID == id {
				e = &s.Events[i]
				break
			}
		}
		if e == nil {
			s.Events = append(s.Events, Event{ID: id, Type: kind, ConsultationID: p.turn.consultationID, CreatedAt: now(), Data: map[string]any{"harness": p.harness}})
			e = &s.Events[len(s.Events)-1]
		}
		if title != "" {
			e.Title = title
		}
		if status != "" {
			e.Status = status
		}
		if appendText {
			e.Text += text
		} else {
			e.Text = text
		}
		for k, v := range raw {
			e.Data[k] = v
		}
		s.UpdatedAt = now()
	})
}

func (p *adapter) nativeID(id string) error {
	if id == "" {
		return nil
	}
	if len(id) > 256 || strings.HasPrefix(id, "-") || strings.IndexFunc(id, func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_')
	}) != -1 {
		return errors.New("Harness returned an invalid native session identifier.")
	}
	p.app.mu.Lock()
	defer p.app.mu.Unlock()
	if p.app.state.Native[p.id].ID == id {
		return nil
	}
	return p.app.commitLocked(func(d *diskState) {
		n := d.Native[p.id]
		n.ID = id
		d.Native[p.id] = n
	})
}

func (p *adapter) failure(raw map[string]any) error {
	p.failed = true
	text := humanError(raw)
	if text == "" {
		text = "Harness reported an error."
	}
	// Retain the event type but not arbitrary diagnostic payloads, which can
	// contain authentication headers or other credentials.
	return p.put("", "error", "", text, "error", false, map[string]any{"type": raw["type"]})
}

func (p *adapter) consume(line []byte) error {
	var raw map[string]any
	if err := json.Unmarshal(line, &raw); err != nil || raw == nil || str(raw, "type") == "" {
		return errors.New("Harness returned invalid JSON event output.")
	}
	switch p.harness {
	case "pi":
		return p.pi(raw)
	case "opencode":
		return p.opencode(raw)
	case "codex":
		return p.codex(raw)
	}
	return errors.New("Unknown harness protocol.")
}

func (p *adapter) pi(raw map[string]any) error {
	typ := str(raw, "type")
	switch typ {
	case "session":
		return p.nativeID(str(raw, "id"))
	case "message_start":
		if str(object(raw["message"]), "role") == "assistant" {
			p.piMessage++
		}
		return nil
	case "message_update":
		delta := object(raw["assistantMessageEvent"])
		kind := str(delta, "type")
		if kind != "text_delta" && kind != "thinking_delta" {
			return nil
		}
		target := "assistant"
		if kind == "thinking_delta" {
			target = "reasoning"
		}
		index, _ := delta["contentIndex"].(float64)
		if number, ok := delta["contentIndex"].(json.Number); ok {
			index, _ = number.Float64()
		}
		key := fmt.Sprintf("message/%d/%d", p.piMessage, int(index))
		// Keep delta metadata, not the repeatedly growing partial message copy.
		return p.put(key, target, "", str(delta, "delta"), "running", true,
			map[string]any{"type": typ, "contentIndex": index})
	case "message_end":
		message := object(raw["message"])
		if str(message, "role") != "assistant" {
			return nil
		}
		status := "completed"
		stop := str(message, "stopReason")
		if stop == "error" || stop == "aborted" {
			status = "error"
		}
		parts, _ := message["content"].([]any)
		for index, value := range parts {
			part := object(value)
			kind, text := "assistant", str(part, "text")
			if str(part, "type") == "thinking" {
				kind, text = "reasoning", str(part, "thinking")
			} else if str(part, "type") != "text" {
				continue
			}
			if err := p.put(fmt.Sprintf("message/%d/%d", p.piMessage, index), kind, "", text, status, false,
				map[string]any{"type": typ, "contentIndex": index, "stopReason": stop, "model": message["model"], "usage": message["usage"]}); err != nil {
				return err
			}
		}
		if status == "error" {
			return p.failure(map[string]any{"type": typ, "message": message["errorMessage"]})
		}
		return nil
	case "tool_execution_start", "tool_execution_update", "tool_execution_end":
		key := "tool/" + str(raw, "toolCallId")
		name := str(raw, "toolName")
		if name != "" {
			p.toolNames[key] = name
		} else {
			name = p.toolNames[key]
		}
		if args := object(raw["args"]); args != nil {
			p.commands[key] = str(args, "command")
		}
		kind, text, status := "tool", "", "running"
		if name == "bash" {
			kind = "command"
		} else if strings.HasPrefix(name, "mcp") {
			kind = "mcp"
		}
		if typ == "tool_execution_update" {
			text = contentText(raw["partialResult"])
		} else if typ == "tool_execution_end" {
			text, status = contentText(raw["result"]), "completed"
			if truth(raw, "isError") {
				status = "error"
			}
		}
		if command := p.commands[key]; command != "" {
			text = command + "\n" + text
		}
		return p.put(key, kind, name, strings.TrimSuffix(text, "\n"), status, false, raw)
	case "agent_settled":
		p.completed = true
		return p.put("", "status", "", "Agent settled.", "completed", false, raw)
	case "agent_end":
		// agent_end includes complete messages already represented above.
		return nil
	case "turn_start", "turn_end":
		return nil
	case "error":
		return p.failure(raw)
	default:
		return p.put("", "status", typ, typ, "", false, raw)
	}
}

func (p *adapter) opencode(raw map[string]any) error {
	if err := p.nativeID(str(raw, "sessionID")); err != nil {
		return err
	}
	typ := str(raw, "type")
	part := object(raw["part"])
	key := typ + "/" + str(part, "id")
	switch typ {
	case "text", "reasoning":
		kind := "assistant"
		if typ == "reasoning" {
			kind = "reasoning"
			if strings.TrimSpace(str(part, "text")) == "" {
				return nil
			}
		}
		return p.put(key, kind, "", str(part, "text"), "completed", false,
			map[string]any{"type": typ, "partID": part["id"]})
	case "tool_use":
		state := object(part["state"])
		name := str(part, "tool")
		kind, text := "tool", contentText(state["output"])
		status := str(state, "status")
		if status == "error" {
			text = humanError(state["error"])
		}
		if name == "bash" {
			kind = "command"
			if command := str(object(state["input"]), "command"); command != "" {
				text = command + "\n" + text
			}
		} else if strings.HasPrefix(name, "mcp_") || strings.HasPrefix(name, "mcp.") {
			kind = "mcp"
		}
		return p.put(key, kind, name, strings.TrimSuffix(text, "\n"), status, false, raw)
	case "step_start", "step_finish":
		status := "started"
		p.completed = false
		if typ == "step_finish" {
			status = "completed"
			p.completed = str(part, "reason") != "tool-calls"
		}
		return p.put("", "status", typ, typ, status, false, raw)
	case "error":
		return p.failure(raw)
	default:
		return p.put("", "status", typ, typ, "", false, raw)
	}
}

func (p *adapter) codex(raw map[string]any) error {
	typ := str(raw, "type")
	switch typ {
	case "thread.started":
		return p.nativeID(str(raw, "thread_id"))
	case "turn.failed", "error":
		return p.failure(raw)
	case "turn.completed":
		p.completed = true
		return p.put("", "status", "", "Turn completed.", "completed", false, raw)
	case "item.started", "item.updated", "item.completed":
		item := object(raw["item"])
		kind, title, text := "tool", str(item, "type"), str(item, "text")
		status := str(item, "status")
		if status == "" || status == "in_progress" {
			status = "running"
			if typ == "item.completed" {
				status = "completed"
			}
		} else if status == "failed" {
			status = "error"
		}
		switch title {
		case "agent_message":
			kind, title = "assistant", ""
		case "reasoning":
			kind, title = "reasoning", ""
		case "command_execution":
			kind, title = "command", "Command"
			text = str(item, "command")
			if output := str(item, "aggregated_output"); output != "" {
				text += "\n" + output
			}
			if code, ok := item["exit_code"].(float64); ok && code != 0 {
				status = "error"
			}
		case "mcp_tool_call":
			kind, title = "mcp", str(item, "server")+" / "+str(item, "tool")
			text = contentText(item["result"])
			if message := humanError(item["error"]); message != "" {
				text, status = message, "error"
			}
		case "error":
			// Item errors are nonfatal diagnostics; only turn.failed/error fails the turn.
			kind, status, title = "status", "warning", "Harness warning"
			if text == "" {
				text = humanError(item)
			}
		}
		return p.put("item/"+str(item, "id"), kind, title, text, status, false, raw)
	default:
		return p.put("", "status", typ, typ, "", false, raw)
	}
}
