package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
	"unicode/utf8"
)

const taskPayloadLimit = 1 << 20

type TaskInput struct {
	ContentType string            `json:"contentType"`
	Body        string            `json:"body"`
	Headers     map[string]string `json:"headers,omitempty"`
}

type WebhookBinding struct {
	OperationID     string `json:"operationId"`
	ID              string `json:"id"`
	TaskID          string `json:"taskId"`
	ProjectID       string `json:"projectId"`
	Revision        uint64 `json:"revision,omitempty"`
	PublicOrigin    string `json:"publicOrigin"`
	SignatureHeader string `json:"signatureHeader,omitempty"`
	DeliveryHeader  string `json:"deliveryHeader,omitempty"`
	ProtectedSecret string `json:"protectedSecret"`
	CreatedAt       string `json:"createdAt"`
	LastReceiptAt   string `json:"lastReceiptAt,omitempty"`
	LastDisposition string `json:"lastDisposition,omitempty"`
	// Legacy storage fields are retained solely to replay old journals intact.
	LegacyRepositoryID int64  `json:"repositoryId,omitempty"`
	LegacyRepository   string `json:"repository,omitempty"`
	LegacyPingAt       string `json:"lastPingAt,omitempty"`
}

type WebhookReceipt struct {
	BindingID     string `json:"bindingId"`
	DeliveryID    string `json:"deliveryId"`
	Digest        string `json:"digest"`
	DigestVersion int    `json:"digestVersion,omitempty"`
	LogicalKey    string `json:"logicalKey,omitempty"` // historical receipts only
	Disposition   string `json:"disposition"`
	RunID         string `json:"runId,omitempty"`
	ReceivedAt    string `json:"receivedAt"`
}

func (d *diskState) webhookBinding(taskID string) *WebhookBinding {
	for i := range d.WebhookBindings {
		if d.WebhookBindings[i].TaskID == taskID {
			return &d.WebhookBindings[i]
		}
	}
	return nil
}

func webhookBindingView(b *WebhookBinding) any {
	if b == nil {
		return nil
	}
	path := "/hooks/tasks/" + b.ID
	return map[string]any{"id": b.ID, "taskId": b.TaskID, "revision": b.Revision, "path": path, "publicOrigin": b.PublicOrigin, "signatureHeader": b.SignatureHeader, "deliveryHeader": b.DeliveryHeader, "lastReceiptAt": b.LastReceiptAt, "lastDisposition": b.LastDisposition}
}

func validWebhookHeader(name string) bool {
	if name == "" || len(name) > 100 {
		return false
	}
	for _, c := range name {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-') {
			return false
		}
	}
	return true
}

func validateTaskInput(input TaskInput) error {
	if len(input.Body) > taskPayloadLimit || !utf8.ValidString(input.Body) || strings.ContainsRune(input.Body, 0) {
		return errors.New("Payload must be UTF-8 text up to 1 MiB.")
	}
	media, _, err := mime.ParseMediaType(input.ContentType)
	if err != nil || len(input.ContentType) > 256 {
		return errors.New("Enter a valid content type.")
	}
	if (media == "application/json" || strings.HasSuffix(media, "+json")) && !json.Valid([]byte(input.Body)) {
		return errors.New("Payload is not valid JSON.")
	}
	total := 0
	for name, value := range input.Headers {
		total += len(name) + len(value)
		if !validWebhookHeader(name) || len(value) > 2048 || !utf8.ValidString(value) || strings.ContainsAny(value, "\r\n\x00") {
			return errors.New("Invalid webhook metadata.")
		}
	}
	if total > 8192 {
		return errors.New("Webhook metadata exceeds 8 KiB.")
	}
	return nil
}

func (a *app) getWebhookBinding(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()
	t := a.state.task(r.PathValue("taskId"))
	if t == nil || t.ProjectID != r.PathValue("id") {
		fail(w, 404, "Task not found.")
		return
	}
	respond(w, 200, map[string]any{"binding": webhookBindingView(a.state.webhookBinding(t.ID))})
}

func (a *app) saveWebhookBinding(w http.ResponseWriter, r *http.Request) {
	var args struct {
		PublicOrigin    string `json:"publicOrigin"`
		SignatureHeader string `json:"signatureHeader"`
		DeliveryHeader  string `json:"deliveryHeader"`
		OperationID     string `json:"operationId"`
		Revision        uint64 `json:"revision"`
	}
	if !decode(w, r, &args) {
		return
	}
	args.PublicOrigin = strings.TrimRight(strings.TrimSpace(args.PublicOrigin), "/")
	args.SignatureHeader = http.CanonicalHeaderKey(strings.TrimSpace(args.SignatureHeader))
	args.DeliveryHeader = http.CanonicalHeaderKey(strings.TrimSpace(args.DeliveryHeader))
	u, err := url.Parse(args.PublicOrigin)
	if err != nil || args.PublicOrigin != "" && (u.Scheme != "https" || u.Host == "" || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "") {
		fail(w, 400, "Public origin must be an HTTPS origin without a path, or left empty for local setup.")
		return
	}
	if !validLinkedText(args.OperationID, 128) || !validWebhookHeader(args.SignatureHeader) || !validWebhookHeader(args.DeliveryHeader) || strings.EqualFold(args.SignatureHeader, args.DeliveryHeader) || strings.EqualFold(args.SignatureHeader, "Authorization") || strings.EqualFold(args.DeliveryHeader, "Authorization") || strings.EqualFold(args.SignatureHeader, "Cookie") || strings.EqualFold(args.DeliveryHeader, "Cookie") {
		fail(w, 400, "Provide distinct signature and delivery header names and an operation ID.")
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	t := a.state.task(r.PathValue("taskId"))
	if t == nil || t.ProjectID != r.PathValue("id") || t.DeletedAt != "" || t.Trigger != "webhook" {
		fail(w, 404, "Webhook Task not found.")
		return
	}
	if b := a.state.webhookBinding(t.ID); b != nil {
		if b.OperationID == args.OperationID {
			if b.PublicOrigin != args.PublicOrigin || b.SignatureHeader != args.SignatureHeader || b.DeliveryHeader != args.DeliveryHeader {
				fail(w, 409, "Operation already belongs to different webhook settings.")
				return
			}
			respond(w, 200, map[string]any{"binding": webhookBindingView(b)})
			return
		}
		if b.Revision != args.Revision {
			fail(w, 409, "Webhook changed. Reload its settings before saving.")
			return
		}
		taskID := t.ID
		err := a.commitLocked(func(d *diskState) {
			next := d.webhookBinding(taskID)
			next.PublicOrigin, next.SignatureHeader, next.DeliveryHeader = args.PublicOrigin, args.SignatureHeader, args.DeliveryHeader
			next.OperationID = args.OperationID
			next.Revision++
		})
		if err != nil {
			fail(w, 503, err.Error())
			return
		}
		respond(w, 200, map[string]any{"binding": webhookBindingView(a.state.webhookBinding(taskID))})
		return
	}
	secret := newID() + newID()
	protected, err := protectTaskSecret(secret)
	if err != nil {
		fail(w, 503, err.Error())
		return
	}
	b := WebhookBinding{OperationID: args.OperationID, ID: newID(), TaskID: t.ID, ProjectID: t.ProjectID, Revision: 1, PublicOrigin: args.PublicOrigin, SignatureHeader: args.SignatureHeader, DeliveryHeader: args.DeliveryHeader, ProtectedSecret: protected, CreatedAt: now()}
	if err := a.commitLocked(func(d *diskState) { d.WebhookBindings = append(d.WebhookBindings, b) }); err != nil {
		fail(w, 503, err.Error())
		return
	}
	respond(w, 201, map[string]any{"binding": webhookBindingView(&b), "secret": secret})
}

func (a *app) rotateWebhookSecret(w http.ResponseWriter, r *http.Request) {
	secret := newID() + newID()
	blob, err := protectTaskSecret(secret)
	if err != nil {
		fail(w, 503, err.Error())
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	b := a.state.webhookBinding(r.PathValue("taskId"))
	t := a.state.task(r.PathValue("taskId"))
	if b == nil || b.ProjectID != r.PathValue("id") || t == nil || t.DeletedAt != "" {
		fail(w, 404, "Webhook not found.")
		return
	}
	taskID := b.TaskID
	if err := a.commitLocked(func(d *diskState) { next := d.webhookBinding(taskID); next.ProtectedSecret = blob; next.Revision++ }); err != nil {
		fail(w, 503, err.Error())
		return
	}
	respond(w, 200, map[string]any{"secret": secret, "binding": webhookBindingView(a.state.webhookBinding(taskID))})
}

func webhookAuthenticated(headers http.Header, raw []byte, secret, signatureHeader string) bool {
	if auth := headers.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") && hmac.Equal([]byte(strings.TrimPrefix(auth, "Bearer ")), []byte(secret)) {
		return true
	}
	sig := headers.Get(signatureHeader)
	decoded, err := hex.DecodeString(strings.TrimPrefix(sig, "sha256="))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(raw)
	return err == nil && hmac.Equal(decoded, mac.Sum(nil))
}

// Only event/delivery metadata is passed to the model, never credentials,
// cookies, request URLs or arbitrary headers. Provider names have no semantics.
func webhookMetadata(headers http.Header, signatureHeader, deliveryHeader string) map[string]string {
	result := map[string]string{}
	for name, values := range headers {
		lower := strings.ToLower(name)
		if strings.EqualFold(name, signatureHeader) || strings.Contains(lower, "authorization") || strings.Contains(lower, "cookie") || strings.Contains(lower, "secret") || strings.Contains(lower, "token") || strings.Contains(lower, "signature") {
			continue
		}
		if strings.EqualFold(name, deliveryHeader) || strings.HasSuffix(lower, "-event") || strings.HasSuffix(lower, "-event-type") || strings.HasSuffix(lower, "-delivery") || strings.HasSuffix(lower, "-request-id") {
			result[http.CanonicalHeaderKey(name)] = strings.Join(values, ", ")
		}
	}
	return result
}

func (a *app) receiveTaskWebhook(w http.ResponseWriter, r *http.Request) {
	_ = http.NewResponseController(w).SetWriteDeadline(time.Now().Add(9 * time.Second))
	a.mu.Lock()
	var binding WebhookBinding
	for _, b := range a.state.WebhookBindings {
		if b.ID == r.PathValue("bindingId") {
			binding = b
			break
		}
	}
	a.mu.Unlock()
	if binding.ID == "" {
		fail(w, 404, "Unknown webhook.")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, taskPayloadLimit)
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		var large *http.MaxBytesError
		if errors.As(err, &large) {
			fail(w, 413, "Payload too large.")
		} else {
			fail(w, 400, "Invalid body.")
		}
		return
	}
	secret, err := unprotectTaskSecret(binding.ProtectedSecret)
	if err != nil {
		fail(w, 503, "Webhook unavailable.")
		return
	}
	if !webhookAuthenticated(r.Header, raw, secret, binding.SignatureHeader) {
		fail(w, 401, "Invalid webhook credentials.")
		return
	}
	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "text/plain"
	}
	input := TaskInput{ContentType: contentType, Body: string(raw), Headers: webhookMetadata(r.Header, binding.SignatureHeader, binding.DeliveryHeader)}
	if err := validateTaskInput(input); err != nil {
		fail(w, 400, err.Error())
		return
	}
	delivery := r.Header.Get(binding.DeliveryHeader)
	if delivery == "" {
		delivery = newID()
	}
	if !validLinkedText(delivery, 200) {
		fail(w, 400, "Invalid delivery identity.")
		return
	}
	encoded, _ := json.Marshal(input)
	digest := sha256.Sum256(encoded)
	receipt := WebhookReceipt{BindingID: binding.ID, DeliveryID: delivery, Digest: hex.EncodeToString(digest[:]), DigestVersion: 1, Disposition: "accepted", ReceivedAt: now()}
	a.mu.Lock()
	defer a.mu.Unlock()
	current := a.state.webhookBinding(binding.TaskID)
	if current == nil || current.Revision != binding.Revision || current.ProtectedSecret != binding.ProtectedSecret {
		fail(w, 409, "Webhook changed. Retry the request.")
		return
	}
	for _, previous := range a.state.WebhookReceipts {
		if previous.BindingID == binding.ID && previous.DeliveryID == delivery {
			expected := receipt.Digest
			if previous.DigestVersion == 0 {
				legacyDigest := sha256.Sum256(raw)
				expected = hex.EncodeToString(legacyDigest[:])
			}
			if previous.Digest != expected {
				fail(w, 409, "Conflicting delivery.")
				return
			}
			respond(w, 200, map[string]string{"runId": previous.RunID})
			return
		}
	}
	err = a.commitTransactionLocked(func(d *diskState) error {
		t := d.task(binding.TaskID)
		p := d.project(binding.ProjectID)
		if t == nil || t.DeletedAt != "" || !t.Enabled || t.Trigger != "webhook" || p == nil || p.Removed {
			receipt.Disposition = "disabled"
		} else {
			run, err := a.admitTaskRun(d, *t, input, "webhook", binding.ID+"/"+delivery)
			if err != nil {
				return err
			}
			receipt.RunID = run.ID
		}
		d.WebhookReceipts = append(d.WebhookReceipts, receipt)
		b := d.webhookBinding(binding.TaskID)
		b.LastReceiptAt, b.LastDisposition = receipt.ReceivedAt, receipt.Disposition
		return nil
	})
	if err != nil {
		fail(w, 503, "Admission unavailable.")
		return
	}
	a.scheduleTasksLocked()
	if receipt.Disposition == "disabled" {
		w.WriteHeader(204)
		return
	}
	respond(w, 202, map[string]string{"runId": receipt.RunID})
}

func (a *app) runTaskNow(w http.ResponseWriter, r *http.Request) {
	var args struct {
		Body        string `json:"body"`
		ContentType string `json:"contentType"`
		OperationID string `json:"operationId"`
	}
	if !decode(w, r, &args) {
		return
	}
	if !validLinkedText(args.OperationID, 128) {
		fail(w, 400, "Operation ID required.")
		return
	}
	if args.ContentType == "" {
		args.ContentType = "text/plain"
	}
	input := TaskInput{Body: args.Body, ContentType: args.ContentType}
	if err := validateTaskInput(input); err != nil {
		fail(w, 400, err.Error())
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	t := a.state.task(r.PathValue("taskId"))
	if t == nil || t.ProjectID != r.PathValue("id") || t.DeletedAt != "" {
		fail(w, 404, "Task not found.")
		return
	}
	for _, run := range a.state.TaskRuns {
		if run.TaskID == t.ID && run.OperationID == args.OperationID {
			if run.Origin != "manual" || !reflect.DeepEqual(run.Input, &input) {
				fail(w, 409, "Operation already belongs to another run input.")
				return
			}
			respond(w, 200, map[string]string{"runId": run.ID, "sessionId": run.SessionID})
			return
		}
	}
	var accepted TaskRun
	err := a.commitTransactionLocked(func(d *diskState) error {
		var err error
		accepted, err = a.admitTaskRun(d, *d.task(t.ID), input, "manual", args.OperationID)
		return err
	})
	if err != nil {
		fail(w, 409, err.Error())
		return
	}
	a.scheduleTasksLocked()
	respond(w, 202, map[string]string{"runId": accepted.ID, "sessionId": accepted.SessionID})
}
