# dmail

Send email over **any SMTP provider** from Go. Zero dependencies.

Works with Amazon SES, SendGrid, Mailgun, Gmail, Microsoft 365, Azure — just
change the host and credentials.

```bash
go get github.com/danixts/dmail
```

## Quick start

```go
client, _ := dmail.NewFromEnv()

client.Send(dmail.Email{
	To:      []string{"someone@example.com"},
	Subject: "Hello",
	HTML:    "<h1>Hi 👋</h1>",
	Text:    "Hi",
})
```

`NewFromEnv` reads:

```
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=your-username
SMTP_PASSWORD=your-password
SMTP_FROM=no-reply@yourdomain.com
SMTP_FROM_NAME=Your Brand
```

Or configure it directly:

```go
client, _ := dmail.New(dmail.Config{
	Host:     "smtp.example.com",
	Username: "user",
	Password: "pass",
	From:     "no-reply@yourdomain.com",
	FromName: "Your Brand",
})
```

## ✨ HTML templates with variables

Compile once, reuse for every email. Values are auto-escaped (no injection).

```go
invoice := dmail.MustTemplate(dmail.TemplateConfig{
	Subject: "Invoice {{.Number}}",
	HTML:    "<h2>Hi {{.Name}}</h2><p>Total: {{.Total}}</p>",
})

type Invoice struct {
	Name   string
	Number string
	Total  string
}

client.Send(dmail.Email{
	To:       []string{"client@example.com"},
	Template: invoice,
	Data:     Invoice{Name: "Ariel", Number: "F-001", Total: "100 USD"},
})
```

`Data` is any value — a struct (type-safe, recommended) or a `map[string]any`.
Need a preview? `invoice.Render(data)`.

## Attachments, CC, BCC, Reply-To

```go
file, _ := dmail.AttachFile("invoice.pdf") // MIME type auto-detected

client.Send(dmail.Email{
	To:          []string{"client@example.com"},
	Cc:          []string{"billing@yourdomain.com"},
	Bcc:         []string{"audit@yourdomain.com"},
	ReplyTo:     "support@yourdomain.com",
	Subject:     "Your invoice",
	HTML:        "<p>Attached.</p>",
	Attachments: []dmail.Attachment{file},
})
```

## Logging

Structured logs via `log/slog` (errors are also returned):

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
client, _ := dmail.NewFromEnv(dmail.WithLogger(logger))
```

## Providers

| Provider | Host | Port |
|----------|------|------|
| Amazon SES | `email-smtp.<region>.amazonaws.com` | 587 |
| SendGrid | `smtp.sendgrid.net` | 587 |
| Mailgun | `smtp.mailgun.org` | 587 |
| Postmark | `smtp.postmarkapp.com` | 587 |
| Gmail / Workspace | `smtp.gmail.com` | 587 |
| Microsoft 365 | `smtp.office365.com` | 587 |
| Azure Communication Services | `smtp.azurecomm.net` | 587 |

Port `587`/`25` uses STARTTLS, `465` uses implicit TLS — chosen automatically.
Auth negotiates `PLAIN` / `LOGIN` based on what the server offers.

> **Amazon SES:** use the *SMTP credentials* from the SES console (not your AWS
> access key).

## API

| | |
|--|--|
| `New(Config, ...Option)` / `NewFromEnv(...Option)` | create a client |
| `client.Send(Email)` | send an email |
| `WithLogger(*slog.Logger)` / `WithTransport(Transport)` | options |
| `AttachFile(path)` | read a file as an attachment |
| `NewTemplate` / `MustTemplate` / `template.Render` | HTML templates |

## Development

```bash
make check   # fmt + vet + lint + build + tests (80% coverage gate)
```

Tests live in `tests/` (black-box, with an in-memory fake SMTP server).

## License

MIT
