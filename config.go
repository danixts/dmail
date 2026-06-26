package dmail

import (
	"crypto/tls"
	"errors"
	"os"
)

const defaultPort = "587"

type Config struct {
	Host      string
	Port      string
	Username  string
	Password  string
	From      string
	FromName  string
	TLSConfig *tls.Config
}

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
	return c
}

func (c Config) validate() error {
	if c.Host == "" {
		return errors.New("dmail: Host is required")
	}
	if c.Username == "" || c.Password == "" {
		return errors.New("dmail: Username and Password are required")
	}
	if c.From == "" {
		return errors.New("dmail: From is required")
	}
	return nil
}

func (c Config) address() string {
	return c.Host + ":" + c.Port
}

func (c Config) usesImplicitTLS() bool {
	return c.Port == "465"
}
