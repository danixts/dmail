package dmail

import (
	"errors"
	"net/smtp"
	"strings"
)

type loginAuth struct {
	username string
	password string
}

func newLoginAuth(username, password string) smtp.Auth {
	return &loginAuth{username: username, password: password}
}

func (a *loginAuth) Start(*smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", nil, nil
}

func (a *loginAuth) Next(prompt []byte, more bool) ([]byte, error) {
	if !more {
		return nil, nil
	}
	question := strings.ToLower(strings.TrimSpace(string(prompt)))
	if strings.HasPrefix(question, "user") {
		return []byte(a.username), nil
	}
	if strings.HasPrefix(question, "pass") {
		return []byte(a.password), nil
	}
	return nil, errors.New("dmail: unexpected SMTP prompt: " + string(prompt))
}
