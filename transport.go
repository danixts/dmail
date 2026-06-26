package dmail

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
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
	client, err := t.dial()
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()

	if !t.config.usesImplicitTLS() {
		if err := client.StartTLS(t.tlsConfig()); err != nil {
			return fmt.Errorf("dmail: starttls: %w", err)
		}
	}
	if err := client.Auth(t.auth(client)); err != nil {
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

func (t smtpTransport) dial() (*smtp.Client, error) {
	if !t.config.usesImplicitTLS() {
		client, err := smtp.Dial(t.config.address())
		if err != nil {
			return nil, fmt.Errorf("dmail: dial: %w", err)
		}
		return client, nil
	}

	conn, err := tls.Dial("tcp", t.config.address(), t.tlsConfig())
	if err != nil {
		return nil, fmt.Errorf("dmail: tls dial: %w", err)
	}
	client, err := smtp.NewClient(conn, t.config.Host)
	if err != nil {
		return nil, fmt.Errorf("dmail: smtp client: %w", err)
	}
	return client, nil
}

func (t smtpTransport) tlsConfig() *tls.Config {
	if t.config.TLSConfig != nil {
		return t.config.TLSConfig
	}
	return &tls.Config{ServerName: t.config.Host}
}

func (t smtpTransport) auth(client *smtp.Client) smtp.Auth {
	if ok, mechanisms := client.Extension("AUTH"); ok && strings.Contains(mechanisms, "PLAIN") {
		return smtp.PlainAuth("", t.config.Username, t.config.Password, t.config.Host)
	}
	return newLoginAuth(t.config.Username, t.config.Password)
}
