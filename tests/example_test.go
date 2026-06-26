package dmail_test

import (
	"log/slog"
	"os"

	"github.com/danixts/dmail"
)

func ExampleClient_Send() {
	client, err := dmail.NewFromEnv()
	if err != nil {
		slog.Error("config", "error", err)
		return
	}

	_ = client.Send(dmail.Email{
		To:      []string{"someone@example.com"},
		Subject: "Hello",
		HTML:    "<h1>Hi</h1>",
		Text:    "Hi",
	})
}

func ExampleClient_Send_template() {
	client, err := dmail.NewFromEnv()
	if err != nil {
		slog.Error("config", "error", err)
		return
	}

	invoice := dmail.MustTemplate(dmail.TemplateConfig{
		Subject: "Invoice {{.Number}}",
		HTML:    "<h2>Hi {{.Name}}</h2><p>Total: {{.Total}}</p>",
	})

	type Invoice struct {
		Name   string
		Number string
		Total  string
	}

	_ = client.Send(dmail.Email{
		To:       []string{"client@example.com"},
		Template: invoice,
		Data:     Invoice{Name: "Ariel", Number: "F-001", Total: "100 USD"},
	})
}

func ExampleAttachFile() {
	client, err := dmail.NewFromEnv(dmail.WithLogger(slog.New(slog.NewJSONHandler(os.Stdout, nil))))
	if err != nil {
		return
	}

	file, err := dmail.AttachFile("invoice.pdf")
	if err != nil {
		return
	}

	_ = client.Send(dmail.Email{
		To:          []string{"client@example.com"},
		Subject:     "Your invoice",
		HTML:        "<p>Attached.</p>",
		Attachments: []dmail.Attachment{file},
	})
}
