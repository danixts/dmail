package dmail

import (
	"crypto/tls"
	"errors"
	"os"
	"time"
)

const (
	defaultPort    = "587"
	defaultTimeout = 30 * time.Second
)

// Errors returned by [Config.validate] when a required field is missing.
var (
	ErrMissingHost        = errors.New("dmail: Host is required")
	ErrMissingCredentials = errors.New("dmail: Username and Password are required")
	ErrMissingFrom        = errors.New("dmail: From is required")
)

// Config holds the SMTP connection settings and sender identity used by [Client].
type Config struct {
	Host      string
	Port      string
	Username  string
	Password  string
	From      string
	FromName  string
	Timeout   time.Duration
	TLSConfig *tls.Config
}

// ConfigFromEnv builds a Config from the SMTP_HOST, SMTP_PORT, SMTP_USERNAME,
// SMTP_PASSWORD, SMTP_FROM and SMTP_FROM_NAME environment variables.
func ConfigFromEnv() Config {
	return Config{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     os.Getenv("SMTP_PORT"),
		Username: os.Getenv("SMTP_USERNAME"),
		Password: os.Getenv("SMTP_PASSWORD"),
		From:     os.Getenv("SMTP_FROM"),
		FromName: os.Getenv("SMTP_FROM_NAME"),
	}
}

func (c Config) withDefaults() Config {
	if c.Port == "" {
		c.Port = defaultPort
	}
	if c.Timeout == 0 {
		c.Timeout = defaultTimeout
	}
	return c
}

func (c Config) validate() error {
	if c.Host == "" {
		return ErrMissingHost
	}
	if c.Username == "" || c.Password == "" {
		return ErrMissingCredentials
	}
	if c.From == "" {
		return ErrMissingFrom
	}
	return nil
}

func (c Config) address() string {
	return c.Host + ":" + c.Port
}

func (c Config) isImplicitTLS() bool {
	return c.Port == "465"
}
