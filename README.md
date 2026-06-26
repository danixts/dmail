# dmail

A small, dependency-free Go package to send email over any SMTP provider.

dmail speaks plain SMTP, so it works with Amazon SES, SendGrid, Mailgun,
Postmark, Brevo, Gmail / Google Workspace, Microsoft 365, Azure Communication
Services, or your own mail server — you only change the host, port and
credentials.

## Features

- No external dependencies — standard library only
- Works with any SMTP provider
- STARTTLS (port 587/25) and implicit TLS (port 465)
- Auto-negotiates `AUTH PLAIN` / `AUTH LOGIN` based on what the server offers
- HTML + plain text (multipart/alternative)
- CC, BCC, Reply-To
- Attachments of any type (MIME type auto-detected)
- Structured logging via `log/slog`
- Pluggable transport for easy testing
- Go 1.26+

## Install

```bash
go get github.com/danixts/dmail
```

## Quick start

Set the environment variables:

```
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=your-username
SMTP_PASSWORD=your-password
SMTP_FROM=no-reply@yourdomain.com
SMTP_FROM_NAME=Your Brand
```

```go
package main

import "github.com/danixts/dmail"

func main() {
	client, err := dmail.NewFromEnv()
	if err != nil {
		panic(err)
	}

	err = client.Send(dmail.Email{
		To:      []string{"someone@example.com"},
		Subject: "Hello",
		HTML:    "<h1>Hello world</h1>",
		Text:    "Hello world",
	})
	if err != nil {
		panic(err)
	}
}
```

Or pass the configuration explicitly:

```go
client, err := dmail.New(dmail.Config{
	Host:     "smtp.example.com",
	Port:     "587",
	Username: "your-username",
	Password: "your-password",
	From:     "no-reply@yourdomain.com",
	FromName: "Your Brand",
})
```

`Port` defaults to `587` when omitted.

## Provider settings

Any SMTP provider works. Set `Host`, `Port`, `Username` and `Password`
accordingly:

| Provider | Host | Port | Username | Password |
|----------|------|------|----------|----------|
| Amazon SES | `email-smtp.<region>.amazonaws.com` | 587 | SES SMTP username | SES SMTP password |
| SendGrid | `smtp.sendgrid.net` | 587 | `apikey` | your API key |
| Mailgun | `smtp.mailgun.org` | 587 | `postmaster@<domain>` | SMTP password |
| Postmark | `smtp.postmarkapp.com` | 587 | server token | server token |
| Brevo | `smtp-relay.brevo.com` | 587 | login | SMTP key |
| Gmail / Workspace | `smtp.gmail.com` | 587 | your address | app password |
| Microsoft 365 | `smtp.office365.com` | 587 | your address | password / app password |
| Azure Communication Services | `smtp.azurecomm.net` | 587 | SMTP username | Entra client secret |

> **Amazon SES note:** the SMTP username and password are the *SMTP credentials*
> generated in the SES console — not your AWS access key/secret. Use the host for
> your region, e.g. `email-smtp.us-east-1.amazonaws.com`.

### TLS

- Port **587** or **25** → dmail connects in plaintext and upgrades with
  **STARTTLS**.
- Port **465** → dmail connects with **implicit TLS** from the start.

The mode is chosen automatically from the port.

### Authentication

dmail inspects the mechanisms the server advertises and uses `AUTH PLAIN` when
available, falling back to `AUTH LOGIN` otherwise. This covers providers that
only support one or the other (for example, some only offer `LOGIN`).

## CC, BCC, Reply-To and attachments

`AttachFile` reads a file and auto-detects its MIME type from the extension:

```go
invoice, err := dmail.AttachFile("invoice.pdf")
if err != nil {
	// handle read error
}

client.Send(dmail.Email{
	To:          []string{"client@example.com"},
	Cc:          []string{"billing@yourdomain.com"},
	Bcc:         []string{"audit@yourdomain.com"},
	ReplyTo:     "support@yourdomain.com",
	Subject:     "Your invoice",
	HTML:        "<p>Your invoice is attached.</p>",
	Attachments: []dmail.Attachment{invoice},
})
```

Build attachments from in-memory bytes (ContentType is optional; detected from
the Filename extension when omitted):

```go
Attachments: []dmail.Attachment{
	{Filename: "report.csv", Content: csvBytes},
	{Filename: "logo.png", Content: pngBytes, ContentType: "image/png"},
}
```

## Logging

The client logs send attempts, successes and failures with `log/slog`. By
default it uses `slog.Default()`. Inject your own logger:

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
client, _ := dmail.NewFromEnv(dmail.WithLogger(logger))
```

Example output:

```
level=INFO msg="dmail: sending email" to=[client@example.com] cc=[] subject="Your invoice" recipients=1
level=INFO msg="dmail: email sent" to=[client@example.com] cc=[] subject="Your invoice"
```

`Send` also returns the error, so you can both log and handle it.

## Custom transport (testing)

The SMTP transport sits behind the `Transport` interface, so you can inject a
fake in tests without touching the network:

```go
type fakeTransport struct{ delivered dmail.Envelope }

func (f *fakeTransport) Deliver(e dmail.Envelope) error {
	f.delivered = e
	return nil
}

fake := &fakeTransport{}
client, _ := dmail.New(cfg, dmail.WithTransport(fake))
```

## API

| Function / type | Description |
|-----------------|-------------|
| `New(Config, ...Option) (*Client, error)` | Create a client with explicit config |
| `NewFromEnv(...Option) (*Client, error)` | Create a client from `SMTP_*` env vars |
| `WithLogger(*slog.Logger) Option` | Inject a custom logger |
| `WithTransport(Transport) Option` | Inject a custom transport |
| `(*Client).Send(Email) error` | Send an email |
| `AttachFile(path) (Attachment, error)` | Read a file and detect its MIME type |
| `Config` | Host, Port, Username, Password, From, FromName, TLSConfig (optional) |
| `Email` | To, Cc, Bcc, ReplyTo, Subject, Text, HTML, Attachments |
| `Attachment` | Filename, Content (`[]byte`), ContentType (optional) |
| `Transport` | `Deliver(Envelope) error` — the delivery port |

### Environment variables

| Variable | Maps to |
|----------|---------|
| `SMTP_HOST` | Config.Host |
| `SMTP_PORT` | Config.Port |
| `SMTP_USERNAME` | Config.Username |
| `SMTP_PASSWORD` | Config.Password |
| `SMTP_FROM` | Config.From |
| `SMTP_FROM_NAME` | Config.FromName |

## Architecture

Files are split by responsibility (ports & adapters):

| File | Responsibility |
|------|----------------|
| `dmail.go` | `Client` facade that wires everything together |
| `config.go` | Config: data, defaults, validation, env loading |
| `message.go` | Domain types: `Email`, `Attachment` |
| `mime.go` | MIME message rendering |
| `auth.go` | `AUTH LOGIN` mechanism |
| `transport.go` | `Transport` port + SMTP adapter (STARTTLS / implicit TLS, auth negotiation) |

## Development

Tests live in the `tests/` folder as black-box tests (`package dmail_test`),
including an in-memory fake SMTP server. Common tasks via the Makefile:

```bash
make fmt        # format
make vet        # go vet
make lint       # golangci-lint
make cover      # tests + coverage gate (fails under 80%)
make check      # everything
```

CI (GitHub Actions) runs formatting, vet, lint, build and tests with a coverage
gate on every push and pull request.

## License

MIT — see [LICENSE](LICENSE).
