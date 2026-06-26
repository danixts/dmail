package dmail

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
)

type Envelope struct {
	From       string
	Recipients []string
	Message    []byte
}

type Transport interface {
	Deliver(envelope Envelope) error
}

type smtpTransport struct {
	config Config
}

func newSMTPTransport(config Config) smtpTransport {
	return smtpTransport{config: config}
}

func (t smtpTransport) Deliver(envelope Envelope) error {
	client, err := smtp.Dial(t.config.address())
	if err != nil {
		return fmt.Errorf("dmail: dial: %w", err)
	}
	defer client.Close()

	if err := client.StartTLS(&tls.Config{ServerName: t.config.Host}); err != nil {
		return fmt.Errorf("dmail: starttls: %w", err)
	}
	if err := client.Auth(newLoginAuth(t.config.User, t.config.Pass)); err != nil {
		return fmt.Errorf("dmail: auth: %w", err)
	}
	if err := client.Mail(envelope.From); err != nil {
		return fmt.Errorf("dmail: mail from: %w", err)
	}
	for _, recipient := range envelope.Recipients {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("dmail: rcpt %s: %w", recipient, err)
		}
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("dmail: data: %w", err)
	}
	if _, err := writer.Write(envelope.Message); err != nil {
		return fmt.Errorf("dmail: write: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("dmail: close: %w", err)
	}
	return client.Quit()
}
