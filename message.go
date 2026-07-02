package dmail

import (
	"errors"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"
)

// Errors returned by [Email.validate] when a required field is missing.
var (
	ErrNoRecipients = errors.New("dmail: at least one recipient is required in To")
	ErrEmptyBody    = errors.New("dmail: email requires Text or HTML")
)

// Email describes a single message to send, including its recipients,
// content, optional attachments and optional [Template].
type Email struct {
	To          []string
	Cc          []string
	Bcc         []string
	ReplyTo     string
	Subject     string
	Text        string
	HTML        string
	Attachments []Attachment
	Template    *Template
	Data        any
}

func (e *Email) applyTemplate() error {
	if e.Template == nil {
		return nil
	}
	rendered, err := e.Template.Render(e.Data)
	if err != nil {
		return err
	}
	if rendered.Subject != "" {
		e.Subject = rendered.Subject
	}
	if rendered.HTML != "" {
		e.HTML = rendered.HTML
	}
	if rendered.Text != "" {
		e.Text = rendered.Text
	}
	return nil
}

// Attachment is a file attached to an [Email]. ContentType is detected from
// Filename's extension when left empty.
type Attachment struct {
	Filename    string
	Content     []byte
	ContentType string
}

// AttachFile reads the file at path and returns it as an [Attachment].
func AttachFile(path string) (Attachment, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Attachment{}, fmt.Errorf("dmail: read attachment %s: %w", path, err)
	}
	return Attachment{Filename: filepath.Base(path), Content: content}, nil
}

func (a Attachment) resolvedContentType() string {
	if a.ContentType != "" {
		return a.ContentType
	}
	if detected := mime.TypeByExtension(filepath.Ext(a.Filename)); detected != "" {
		return detected
	}
	return "application/octet-stream"
}

func (e Email) recipients() []string {
	recipients := make([]string, 0, len(e.To)+len(e.Cc)+len(e.Bcc))
	recipients = append(recipients, e.To...)
	recipients = append(recipients, e.Cc...)
	recipients = append(recipients, e.Bcc...)
	return recipients
}

func (e Email) validate() error {
	if len(e.To) == 0 {
		return ErrNoRecipients
	}
	if strings.TrimSpace(e.Text) == "" && strings.TrimSpace(e.HTML) == "" {
		return ErrEmptyBody
	}
	return nil
}
