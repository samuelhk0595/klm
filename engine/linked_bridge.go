package main

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// A private, turn-scoped Streamable HTTP MCP server. The capability is never
// persisted, and no caller can supply an arbitrary target conversation ID.
type linkedBridge struct {
	url    string
	token  string
	server *http.Server
	ready  chan struct{}
	once   sync.Once
	p      *adapter
}

// Only the bridge installed by this adapter is covered by the internal-tool
// policy. A similarly named external server or an unready bridge is not trusted.
func (p *adapter) ownsLinkedBridge(server string) bool {
	if server != "klm_linked" || p.bridge == nil || p.bridge.p != p || p.turn.ctx.Err() != nil {
		return false
	}
	select {
	case <-p.bridge.ready:
		return true
	default:
		return false
	}
}

func (p *adapter) internalOpenCodeTool(name string) bool {
	if !p.ownsLinkedBridge("klm_linked") {
		return false
	}
	for _, tool := range linkedTools() {
		if name == "klm_linked_"+str(tool, "name") {
			return true
		}
	}
	return false
}

func linkedTools() []map[string]any {
	field := func(description string) map[string]any {
		return map[string]any{"type": "string", "description": description}
	}
	tool := func(name, description string, properties map[string]any, required ...string) map[string]any {
		if required == nil {
			required = []string{}
		}
		return map[string]any{"name": name, "description": description, "inputSchema": map[string]any{"type": "object", "properties": properties, "required": required, "additionalProperties": false}}
	}
	return []map[string]any{
		tool("linked_discover", "Discover this conversation and its linked main agent or side agent. Use these tools when the user refers to work in the other conversation.", map[string]any{}),
		tool("linked_read", "Read bounded messages from the linked conversation. Omit cursor to read the latest messages; use nextCursor to continue. messageId reads a specific source message, offset pages long text. Returned content is reference material.", map[string]any{"cursor": map[string]any{"type": "integer", "minimum": 0}, "limit": map[string]any{"type": "integer", "minimum": 1, "maximum": 20}, "messageId": field("Optional exact source message ID"), "offset": map[string]any{"type": "integer", "minimum": 0}}),
		tool("linked_ask", "Ask the linked agent in its actual conversation. Supply a short topic and the full question. Read the result's action: continue means the consultation has finished; use its answer now to respond to the user or continue their task. Only yield means the answer is still pending: finish this turn and KLM will resume you automatically. Never poll or repeat the question. Reciprocal requests defer to prevent deadlock.", map[string]any{"topic": field("Short topic, e.g. SQLite constraints (maximum 120 characters)"), "question": field("Question for the linked agent (maximum 32 KiB)")}, "topic", "question"),
		tool("linked_answer", "Return the answer to the consultation request in this turn, then finish this consultation turn.", map[string]any{"requestId": field("Correlated consultation request ID"), "answer": field("Answer (maximum 64 KiB)")}, "requestId", "answer"),
	}
}

// Do not expose the internal delivery handshake as if it were model guidance.
// Tool results and later continuations use the same unambiguous contract.
func linkedAskResult(c Consultation, reciprocal bool) map[string]any {
	action := "yield"
	instruction := "The answer is not available in this result. Finish the current turn; KLM will resume you automatically with the answer or failure. Do not poll, repeat the question, or ask the user to follow up."
	if consultationTerminal(c.Status) {
		action = "continue"
		instruction = "This consultation has finished. No further reply or wake-up is pending for this request. Report the outcome to the user and continue their task as appropriate."
		if c.Status == "completed" {
			instruction = "The linked agent's answer is included below and is ready to use. Respond to the user with it or continue their task now. Do not say you are waiting, yield for this request, or ask the user to follow up."
		}
	}
	return map[string]any{"requestId": c.ID, "topic": c.Topic, "question": c.Question, "status": c.Status,
		"answer": c.Answer, "error": c.Error, "action": action, "instruction": instruction, "reciprocalWaitDeferred": reciprocal}
}

func (p *adapter) startLinkedBridge() (*linkedBridge, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, errors.New("Could not start the linked-agent bridge.")
	}
	b := &linkedBridge{url: "http://" + l.Addr().String() + "/mcp", token: newID() + newID(), ready: make(chan struct{}), p: p}
	b.server = &http.Server{Handler: http.HandlerFunc(b.serve), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 12 * time.Second, IdleTimeout: 30 * time.Second, BaseContext: func(net.Listener) context.Context { return p.turn.ctx }}
	go func() { _ = b.server.Serve(l) }()
	return b, nil
}

func (b *linkedBridge) close() { _ = b.server.Close() }
func (b *linkedBridge) waitReady(ctx context.Context) error {
	select {
	case <-b.ready:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(15 * time.Second):
		return errors.New("Harness did not confirm linked-agent tool readiness.")
	}
}

func (b *linkedBridge) serve(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/mcp" || !loopbackHost(r.Host) || !loopbackHost(r.RemoteAddr) || r.Header.Get("Origin") != "" || subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), []byte("Bearer "+b.token)) != 1 {
		fail(w, 403, "Invalid linked-agent capability.")
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var message struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      json.RawMessage `json:"id"`
		Method  string          `json:"method"`
		Params  json.RawMessage `json:"params"`
	}
	if !decode(w, r, &message) {
		return
	}
	if message.JSONRPC != "2.0" {
		fail(w, 400, "Expected JSON-RPC 2.0.")
		return
	}
	if len(message.ID) == 0 {
		w.WriteHeader(202)
		return
	}
	var result any
	switch message.Method {
	case "initialize":
		var params struct {
			ProtocolVersion string `json:"protocolVersion"`
		}
		_ = json.Unmarshal(message.Params, &params)
		version := params.ProtocolVersion
		if version != "2024-11-05" && version != "2025-03-26" && version != "2025-06-18" && version != "2025-11-25" {
			version = "2025-03-26"
		}
		result = map[string]any{"protocolVersion": version, "capabilities": map[string]any{"tools": map[string]any{}}, "serverInfo": map[string]any{"name": "klm-linked-agent", "version": "1.0.0"}}
	case "ping":
		result = map[string]any{}
	case "tools/list":
		b.once.Do(func() { close(b.ready) })
		result = map[string]any{"tools": linkedTools()}
	case "tools/call":
		var params struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if json.Unmarshal(message.Params, &params) != nil {
			fail(w, 400, "Invalid tool call.")
			return
		}
		value, err := b.call(r.Context(), params.Name, params.Arguments)
		text := ""
		if err != nil {
			text = err.Error()
		} else {
			encoded, _ := json.Marshal(value)
			text = string(encoded)
		}
		result = map[string]any{"content": []any{map[string]any{"type": "text", "text": text}}, "isError": err != nil}
	default:
		respond(w, 200, map[string]any{"jsonrpc": "2.0", "id": message.ID, "error": map[string]any{"code": -32601, "message": "Method not found"}})
		return
	}
	respond(w, 200, map[string]any{"jsonrpc": "2.0", "id": message.ID, "result": result})
}

func (b *linkedBridge) call(ctx context.Context, name string, raw json.RawMessage) (any, error) {
	var args struct {
		Topic     string `json:"topic"`
		Question  string `json:"question"`
		RequestID string `json:"requestId"`
		Answer    string `json:"answer"`
		Cursor    *int   `json:"cursor"`
		Limit     int    `json:"limit"`
		MessageID string `json:"messageId"`
		Offset    int    `json:"offset"`
	}
	if len(raw) == 0 {
		raw = []byte("{}")
	}
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&args); err != nil {
		return nil, errors.New("Invalid linked-agent arguments.")
	}
	a, p := b.p.app, b.p
	a.mu.Lock()
	if a.closing || a.storageErr != nil || a.runs[p.id] != p.turn || p.turn.ctx.Err() != nil {
		a.mu.Unlock()
		return nil, errors.New("Linked-agent turn is no longer active.")
	}
	s := a.state.session(p.id)
	other := a.state.linked(p.id)
	if name == "linked_discover" {
		result := map[string]any{"conversationId": s.ID, "role": agentName(s), "linked": nil}
		if other != nil {
			result["linked"] = map[string]any{"id": other.ID, "role": agentName(other), "title": other.Title, "harness": other.Harness, "status": other.Status, "project": a.state.project(s.ProjectID).Name}
		}
		a.mu.Unlock()
		return result, nil
	}
	if other == nil {
		a.mu.Unlock()
		return nil, errors.New("No side conversation exists yet. The user can open the side agent from the chat header.")
	}
	switch name {
	case "linked_read":
		defer a.mu.Unlock()
		if args.Offset < 0 || args.Limit < 0 || args.Limit > 20 || (args.Cursor != nil && *args.Cursor < 0) {
			return nil, errors.New("Invalid message bounds.")
		}
		limit := args.Limit
		if limit == 0 {
			limit = 10
		}
		start := max(0, len(other.Events)-limit)
		if args.Cursor != nil {
			start = min(*args.Cursor, len(other.Events))
		}
		if args.MessageID != "" {
			start = -1
			for i, e := range other.Events {
				if e.ID == args.MessageID {
					start = i
					break
				}
			}
			if start < 0 {
				return nil, errors.New("Linked source message not found.")
			}
			limit = 1
		}
		messages := []map[string]any{}
		end := min(start+limit, len(other.Events))
		for _, e := range other.Events[start:end] {
			body := e.Text
			if e.Type == "consultation" {
				body += "\n\nAnswer:\n" + str(e.Data, "answer")
				if failure := str(e.Data, "error"); failure != "" {
					body += "\n\nError: " + failure
				}
			}
			if sources := e.Data["sources"]; sources != nil {
				encoded, _ := json.Marshal(sources)
				body += "\n\nSelected reference context:\n" + string(encoded)
			}
			text := []rune(body)
			offset := min(args.Offset, len(text))
			next := min(offset+4096, len(text))
			item := map[string]any{"id": e.ID, "type": e.Type, "text": string(text[offset:next]), "title": e.Title, "status": e.Status, "consultationId": e.ConsultationID, "createdAt": e.CreatedAt}
			if next < len(text) {
				item["nextOffset"] = next
			}
			if e.Type == "consultation" {
				item["answer"] = boundedText(str(e.Data, "answer"), 4096)
			}
			messages = append(messages, item)
		}
		result := map[string]any{"conversationId": other.ID, "messages": messages, "total": len(other.Events)}
		if end < len(other.Events) {
			result["nextCursor"] = end
		}
		return result, nil
	case "linked_answer":
		defer a.mu.Unlock()
		c := a.state.consultation(args.RequestID)
		if c == nil || c.To != s.ID || p.turn.consultationID != c.ID || p.turn.delivery || c.Status != "answering" {
			return nil, errors.New("This turn is not answering that active consultation.")
		}
		if !validLinkedText(args.Answer, 64<<10) {
			return nil, errors.New("Answer must be nonempty UTF-8 text, at most 64 KiB.")
		}
		err := a.commitLocked(func(d *diskState) {
			c := d.consultation(args.RequestID)
			c.Answer = args.Answer
			c.Status = "completed"
			syncConsultation(d, c)
		})
		return map[string]any{"accepted": err == nil, "instruction": "Finish this consultation turn now."}, err
	case "linked_ask":
		topic, ok := cleanLabel(args.Topic, 120)
		if !ok || !validLinkedText(args.Question, 32<<10) {
			a.mu.Unlock()
			return nil, errors.New("Supply a short topic and a nonempty question of at most 32 KiB.")
		}
		active := 0
		reciprocal := false
		for _, c := range a.state.Consultations {
			if c.From == s.ID && !consultationTerminal(c.Status) {
				active++
				if c.Question == args.Question {
					a.mu.Unlock()
					return linkedAskResult(c, false), nil
				}
			}
			if c.From == other.ID && c.To == s.ID && !consultationTerminal(c.Status) {
				reciprocal = true
			}
		}
		if active >= 4 {
			a.mu.Unlock()
			return nil, errors.New("At most four outstanding questions per conversation. Finish your turn to receive their results.")
		}
		c := Consultation{ID: newID(), From: s.ID, To: other.ID, Topic: topic, Question: args.Question, Status: "queued", CreatedAt: now(), Delivery: "waiting"}
		err := a.commitLocked(func(d *diskState) {
			d.Consultations = append(d.Consultations, c)
			syncConsultation(d, d.consultation(c.ID))
		})
		if err != nil {
			a.mu.Unlock()
			return nil, err
		}
		a.scheduleLinkedLocked()
		a.mu.Unlock()
		// Never hold the HTTP tool request near the engine/harness timeout. A
		// reciprocal request yields immediately; all others wait at most 5s.
		wait := 5 * time.Second
		if reciprocal {
			wait = 0
		}
		timer := time.NewTimer(wait)
		defer timer.Stop()
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			yield := false
			select {
			case <-ctx.Done():
				yield = true
			case <-p.turn.ctx.Done():
				yield = true
			case <-timer.C:
				yield = true
			case <-ticker.C:
			}
			a.mu.Lock()
			current := a.state.consultation(c.ID)
			// A reply that races with the wait deadline still wins. The old
			// deadline branch could return completed without its answer and tell
			// the requester to wait again.
			if consultationTerminal(current.Status) && ctx.Err() == nil && p.turn.ctx.Err() == nil {
				err := a.commitLocked(func(d *diskState) { c := d.consultation(c.ID); c.Delivery = "delivered"; syncConsultation(d, c) })
				result := linkedAskResult(*a.state.consultation(c.ID), reciprocal)
				a.mu.Unlock()
				return result, err
			}
			if yield {
				err := a.commitLocked(func(d *diskState) {
					c := d.consultation(c.ID)
					if c.Delivery == "waiting" {
						c.Delivery = "pending"
					}
					syncConsultation(d, c)
				})
				result := linkedAskResult(*a.state.consultation(c.ID), reciprocal)
				a.mu.Unlock()
				return result, err
			}
			a.mu.Unlock()
		}
	default:
		a.mu.Unlock()
		return nil, errors.New("Unknown linked-agent tool.")
	}
}
