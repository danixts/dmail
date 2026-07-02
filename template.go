package dmail

import (
	"bytes"
	"fmt"
	htmltemplate "html/template"
	texttemplate "text/template"
)

// TemplateConfig holds the text/template sources for a [Template]. Each
// field is optional; empty ones are left unrendered.
type TemplateConfig struct {
	Subject string
	HTML    string
	Text    string
}

// Template renders an [Email]'s subject, HTML body and text body from
// parsed templates. HTML is rendered with html/template for auto-escaping;
// Subject and Text use text/template.
type Template struct {
	subject *texttemplate.Template
	html    *htmltemplate.Template
	text    *texttemplate.Template
}

// NewTemplate parses config into a [Template].
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

// MustTemplate is like [NewTemplate] but panics if config fails to parse.
func MustTemplate(config TemplateConfig) *Template {
	template, err := NewTemplate(config)
	if err != nil {
		panic(err)
	}
	return template
}

// RenderedTemplate holds the output of [Template.Render]. Fields left
// unset in the source [TemplateConfig] are returned empty.
type RenderedTemplate struct {
	Subject string
	HTML    string
	Text    string
}

// Render executes the template against data, returning the rendered
// subject, HTML and text.
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
