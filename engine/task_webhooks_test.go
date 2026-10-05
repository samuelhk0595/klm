package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"testing"
)

func TestTaskWebhookAuthentication(t *testing.T) {
	body, secret := []byte(`{"event":"inventory.updated","quantity":3}`), "test-only-secret"
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	signature := hex.EncodeToString(mac.Sum(nil))
	for _, prefix := range []string{"", "sha256="} {
		headers := http.Header{"X-Custom-Signature": {prefix + signature}}
		if !webhookAuthenticated(headers, body, secret, "X-Custom-Signature") {
			t.Fatal("custom signature header should authenticate")
		}
		if webhookAuthenticated(headers, []byte(`{"quantity":999}`), secret, "X-Custom-Signature") {
			t.Fatal("tampered body authenticated")
		}
	}
	if !webhookAuthenticated(http.Header{"Authorization": {"Bearer " + secret}}, body, secret, "X-Webhook-Signature") {
		t.Fatal("bearer token should authenticate without a provider header")
	}
	for _, headers := range []http.Header{{}, {"Authorization": {"Bearer wrong"}}, {"X-Custom-Signature": {"sha256=not-hex"}}} {
		if webhookAuthenticated(headers, body, secret, "X-Custom-Signature") {
			t.Fatal("invalid credentials authenticated")
		}
	}
}

func TestTaskWebhookPayloadAndMetadata(t *testing.T) {
	for _, input := range []TaskInput{
		{ContentType: "application/json", Body: `{"sensor":"temperature","value":24}`},
		{ContentType: "application/x-www-form-urlencoded", Body: "event=created&name=sample"},
		{ContentType: "text/plain", Body: "arbitrary task input"},
		{ContentType: "text/plain", Body: ""},
	} {
		if err := validateTaskInput(input); err != nil {
			t.Fatal(err)
		}
	}
	for _, input := range []TaskInput{
		{ContentType: "application/json", Body: `{bad`},
		{ContentType: "text/plain", Body: string([]byte{0xff})},
		{ContentType: "text/plain", Body: string(make([]byte, taskPayloadLimit+1))},
	} {
		if validateTaskInput(input) == nil {
			t.Fatal("invalid input accepted")
		}
	}
	metadata := webhookMetadata(http.Header{
		"Authorization": {"Bearer private"}, "Cookie": {"private"}, "X-Custom-Signature": {"private"},
		"X-Secret-Event": {"private"}, "Idempotency-Key": {"delivery-1"}, "X-Sender-Event": {"opened"},
	}, "X-Custom-Signature", "Idempotency-Key")
	if len(metadata) != 2 || metadata["Idempotency-Key"] != "delivery-1" || metadata["X-Sender-Event"] != "opened" {
		t.Fatalf("unexpected captured metadata: %#v", metadata)
	}
}
