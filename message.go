package dmail

import (
	"errors"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"
)

type Email struct {
	To          []string
	Cc          []string
	Bcc         []string
	ReplyTo     string
	Subject     string
	Text        string
	HTML        string
	Attachments []Attachment
}

type Attachment struct {
	Filename    string
	Content     []byte
	ContentType string
}

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
		return errors.New("dmail: at least one recipient is required in To")
	}
	if strings.TrimSpace(e.Text) == "" && strings.TrimSpace(e.HTML) == "" {
		return errors.New("dmail: email requires Text or HTML")
	}
	return nil
}
