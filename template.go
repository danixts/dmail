package dmail

import (
	"bytes"
	"fmt"
	htmltemplate "html/template"
	texttemplate "text/template"
)

type TemplateConfig struct {
	Subject string
	HTML    string
	Text    string
}

type Template struct {
	subject *texttemplate.Template
	html    *htmltemplate.Template
	text    *texttemplate.Template
}

func NewTemplate(config TemplateConfig) (*Template, error) {
	template := &Template{}

	if config.Subject != "" {
		parsed, err := texttemplate.New("subject").Parse(config.Subject)
		if err != nil {
			return nil, fmt.Errorf("dmail: parse subject template: %w", err)
		}
		template.subject = parsed
	}
	if config.HTML != "" {
		parsed, err := htmltemplate.New("html").Parse(config.HTML)
		if err != nil {
			return nil, fmt.Errorf("dmail: parse html template: %w", err)
		}
		template.html = parsed
	}
	if config.Text != "" {
		parsed, err := texttemplate.New("text").Parse(config.Text)
		if err != nil {
			return nil, fmt.Errorf("dmail: parse text template: %w", err)
		}
		template.text = parsed
	}
	return template, nil
}

func MustTemplate(config TemplateConfig) *Template {
	template, err := NewTemplate(config)
	if err != nil {
		panic(err)
	}
	return template
}

type RenderedTemplate struct {
	Subject string
	HTML    string
	Text    string
}

func (t *Template) Render(data any) (RenderedTemplate, error) {
	var rendered RenderedTemplate

	if t.subject != nil {
		value, err := executeText(t.subject, data)
		if err != nil {
			return rendered, fmt.Errorf("dmail: render subject: %w", err)
		}
		rendered.Subject = value
	}
	if t.html != nil {
		var buffer bytes.Buffer
		if err := t.html.Execute(&buffer, data); err != nil {
			return rendered, fmt.Errorf("dmail: render html: %w", err)
		}
		rendered.HTML = buffer.String()
	}
	if t.text != nil {
		value, err := executeText(t.text, data)
		if err != nil {
			return rendered, fmt.Errorf("dmail: render text: %w", err)
		}
		rendered.Text = value
	}
	return rendered, nil
}

func executeText(template *texttemplate.Template, data any) (string, error) {
	var buffer bytes.Buffer
	if err := template.Execute(&buffer, data); err != nil {
		return "", err
	}
	return buffer.String(), nil
}
