package dmail

import (
	"errors"
	"os"
)

const (
	defaultHost = "smtp.azurecomm.net"
	defaultPort = "587"
)

type Config struct {
	Host       string
	Port       string
	User       string
	Pass       string
	SenderAddr string
	SenderName string
}

func ConfigFromEnv() Config {
	return Config{
		Host:       os.Getenv("ACS_SMTP_HOST"),
		Port:       os.Getenv("ACS_SMTP_PORT"),
		User:       os.Getenv("ACS_SMTP_USER"),
		Pass:       os.Getenv("ACS_SMTP_PASS"),
		SenderAddr: os.Getenv("ACS_SENDER"),
		SenderName: os.Getenv("ACS_SENDER_NAME"),
	}
}

func (c Config) withDefaults() Config {
	if c.Host == "" {
		c.Host = defaultHost
	}
	if c.Port == "" {
		c.Port = defaultPort
	}
	return c
}

func (c Config) validate() error {
	if c.User == "" || c.Pass == "" {
		return errors.New("dmail: User and Pass are required")
	}
	if c.SenderAddr == "" {
		return errors.New("dmail: SenderAddr is required")
	}
	return nil
}

func (c Config) address() string {
	return c.Host + ":" + c.Port
}
