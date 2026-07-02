package dmail

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
)

// Envelope is the raw SMTP envelope produced from an [Email]: sender,
// recipients (To+Cc+Bcc) and the fully rendered message bytes.
type Envelope struct {
	From       string
	Recipients []string
	Message    []byte
}

// Transport delivers a rendered [Envelope]. The default implementation
// speaks SMTP directly; [WithTransport] can override it, e.g. for tests.
// Implementations should stop as soon as ctx is done.
type Transport interface {
	Deliver(ctx context.Context, envelope Envelope) error
}

type smtpTransport struct {
	config Config
}

func newSMTPTransport(config Config) smtpTransport {
	return smtpTransport{config: config}
}

func (t smtpTransport) Deliver(ctx context.Context, envelope Envelope) error {
	conn, err := t.dial(ctx)
	if err != nil {
		return err
	}

	// Unblocks any pending read/write on conn as soon as ctx is done,
	// since net/smtp has no context-aware API of its own.
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer stop()

	client, err := smtp.NewClient(conn, t.config.Host)
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("dmail: smtp client: %w", err)
	}
	defer func() { _ = client.Close() }()

	if !t.config.isImplicitTLS() {
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

func (t smtpTransport) dial(ctx context.Context) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: t.config.Timeout}

	if t.config.isImplicitTLS() {
		tlsDialer := &tls.Dialer{NetDialer: dialer, Config: t.tlsConfig()}
		conn, err := tlsDialer.DialContext(ctx, "tcp", t.config.address())
		if err != nil {
			return nil, fmt.Errorf("dmail: tls dial: %w", err)
		}
		return conn, nil
	}

	conn, err := dialer.DialContext(ctx, "tcp", t.config.address())
	if err != nil {
		return nil, fmt.Errorf("dmail: dial: %w", err)
	}
	return conn, nil
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
