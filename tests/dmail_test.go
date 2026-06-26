package dmail_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/danixts/dmail"
)

type captureTransport struct {
	envelope dmail.Envelope
}

func (c *captureTransport) Deliver(envelope dmail.Envelope) error {
	c.envelope = envelope
	return nil
}

func newClient(t *testing.T, transport dmail.Transport) *dmail.Client {
	t.Helper()
	client, err := dmail.New(dmail.Config{
		Host:     "smtp.example.com",
		Username: "user",
		Password: "pass",
		From:     "no-reply@example.com",
		FromName: "Test Sender",
	}, dmail.WithTransport(transport))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return client
}

func TestSendBuildsEnvelope(t *testing.T) {
	capture := &captureTransport{}
	client := newClient(t, capture)

	err := client.Send(dmail.Email{
		To:      []string{"a@example.com"},
		Cc:      []string{"b@example.com"},
		ReplyTo: "reply@example.com",
		Subject: "Hello",
		Text:    "Hello",
		HTML:    "<b>Hello</b>",
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}

	if got := len(capture.envelope.Recipients); got != 2 {
		t.Fatalf("recipients = %d, want 2", got)
	}
	if capture.envelope.From != "no-reply@example.com" {
		t.Errorf("envelope From = %q", capture.envelope.From)
	}

	message := string(capture.envelope.Message)
	for _, want := range []string{
		"From: Test Sender <no-reply@example.com>",
		"To: a@example.com",
		"Cc: b@example.com",
		"Reply-To: reply@example.com",
		"Subject: Hello",
		"multipart/alternative",
		"text/plain",
		"text/html",
	} {
		if !strings.Contains(message, want) {
			t.Errorf("message missing %q\n---\n%s", want, message)
		}
	}
}

func TestSendWithAttachment(t *testing.T) {
	capture := &captureTransport{}
	client := newClient(t, capture)

	err := client.Send(dmail.Email{
		To:      []string{"a@example.com"},
		Subject: "Files",
		Text:    "see attachment",
		Attachments: []dmail.Attachment{
			{Filename: "report.csv", Content: []byte("a,b,c")},
			{Filename: "blob.bin", Content: make([]byte, 200)},
		},
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}

	message := string(capture.envelope.Message)
	for _, want := range []string{
		"multipart/mixed",
		`Content-Disposition: attachment; filename="report.csv"`,
		"Content-Transfer-Encoding: base64",
		"text/csv",
	} {
		if !strings.Contains(message, want) {
			t.Errorf("message missing %q\n---\n%s", want, message)
		}
	}
}

func TestSendSingleBody(t *testing.T) {
	capture := &captureTransport{}
	client := newClient(t, capture)

	if err := client.Send(dmail.Email{To: []string{"a@x"}, Subject: "h", HTML: "<i>x</i>"}); err != nil {
		t.Fatalf("Send: %v", err)
	}
	message := string(capture.envelope.Message)
	if !strings.Contains(message, "text/html") || strings.Contains(message, "multipart") {
		t.Errorf("html-only message wrong:\n%s", message)
	}
}

func TestSendValidatesEmail(t *testing.T) {
	client := newClient(t, &captureTransport{})

	if err := client.Send(dmail.Email{Subject: "x", Text: "x"}); err == nil {
		t.Error("expected error when To is empty")
	}
	if err := client.Send(dmail.Email{To: []string{"a@example.com"}}); err == nil {
		t.Error("expected error when Text and HTML are empty")
	}
}

func TestConfigValidation(t *testing.T) {
	cases := map[string]dmail.Config{
		"missing host":     {Username: "u", Password: "p", From: "a@b.com"},
		"missing username": {Host: "h", Password: "p", From: "a@b.com"},
		"missing from":     {Host: "h", Username: "u", Password: "p"},
	}
	for name, config := range cases {
		if _, err := dmail.New(config); err == nil {
			t.Errorf("%s: expected validation error", name)
		}
	}
}

func TestNewFromEnv(t *testing.T) {
	t.Setenv("SMTP_HOST", "smtp.test")
	t.Setenv("SMTP_PORT", "2525")
	t.Setenv("SMTP_USERNAME", "u")
	t.Setenv("SMTP_PASSWORD", "p")
	t.Setenv("SMTP_FROM", "from@test")
	t.Setenv("SMTP_FROM_NAME", "From")

	capture := &captureTransport{}
	client, err := dmail.NewFromEnv(dmail.WithTransport(capture), dmail.WithLogger(nil))
	if err != nil {
		t.Fatalf("NewFromEnv: %v", err)
	}
	if err := client.Send(dmail.Email{To: []string{"a@x"}, Subject: "s", Text: "t"}); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if !strings.Contains(string(capture.envelope.Message), "From: From <from@test>") {
		t.Errorf("From header not built from env:\n%s", capture.envelope.Message)
	}
}

func TestSenderWithoutName(t *testing.T) {
	capture := &captureTransport{}
	client, err := dmail.New(dmail.Config{
		Host:     "h",
		Username: "u",
		Password: "p",
		From:     "plain@example.com",
	}, dmail.WithTransport(capture))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := client.Send(dmail.Email{To: []string{"a@x"}, Subject: "s", Text: "t"}); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if !strings.Contains(string(capture.envelope.Message), "From: plain@example.com") {
		t.Errorf("plain From wrong:\n%s", capture.envelope.Message)
	}
}

func TestAttachFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "note.txt")
	if err := os.WriteFile(path, []byte("hi"), 0o600); err != nil {
		t.Fatal(err)
	}

	att, err := dmail.AttachFile(path)
	if err != nil {
		t.Fatalf("AttachFile: %v", err)
	}
	if att.Filename != "note.txt" || string(att.Content) != "hi" {
		t.Errorf("unexpected attachment: %+v", att)
	}

	if _, err := dmail.AttachFile(filepath.Join(dir, "missing.txt")); err == nil {
		t.Error("expected error for missing file")
	}
}

func TestTemplateRendering(t *testing.T) {
	tmpl, err := dmail.NewTemplate(dmail.TemplateConfig{
		Subject: "Invoice {{.Number}}",
		HTML:    "<h1>Hello {{.Name}}</h1>",
		Text:    "Hello {{.Name}}",
	})
	if err != nil {
		t.Fatalf("NewTemplate: %v", err)
	}

	capture := &captureTransport{}
	client := newClient(t, capture)
	err = client.Send(dmail.Email{
		To:       []string{"client@example.com"},
		Template: tmpl,
		Data:     map[string]any{"Name": "Ariel", "Number": "F-001"},
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}

	message := string(capture.envelope.Message)
	for _, want := range []string{"Subject: Invoice F-001", "Hello Ariel"} {
		if !strings.Contains(message, want) {
			t.Errorf("rendered message missing %q\n---\n%s", want, message)
		}
	}
}

func TestTemplateEscapesHTML(t *testing.T) {
	tmpl := dmail.MustTemplate(dmail.TemplateConfig{HTML: "<p>{{.Name}}</p>"})
	rendered, err := tmpl.Render(map[string]any{"Name": "<script>alert(1)</script>"})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if strings.Contains(rendered.HTML, "<script>") {
		t.Errorf("html/template did not escape input: %s", rendered.HTML)
	}
}

func TestTemplateParseError(t *testing.T) {
	if _, err := dmail.NewTemplate(dmail.TemplateConfig{HTML: "{{.Name"}); err == nil {
		t.Error("expected parse error for malformed template")
	}
}
