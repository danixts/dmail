# dmail

A simple, dependency-free Go package to send email over SMTP through
**Azure Communication Services (ACS)**.

- No external dependencies (standard library only)
- HTML + plain text (multipart/alternative)
- CC, BCC, Reply-To
- Attachments of any type (MIME auto-detected)
- Handles the `AUTH LOGIN` mechanism ACS requires (net/smtp only ships PlainAuth)
- Structured logging via `log/slog`
- Pluggable transport (easy to test)
- Go 1.26.4

## Install

```bash
go get github.com/danixts/dmail
```

## Quick start

Set the environment variables:

```
ACS_SMTP_HOST=smtp.azurecomm.net
ACS_SMTP_PORT=587
ACS_SMTP_USER=mail@yourdomain.com
ACS_SMTP_PASS=<client-secret>
ACS_SENDER=no-reply@yourdomain.com
ACS_SENDER_NAME=Your Brand
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

## Explicit config

```go
client, err := dmail.New(dmail.Config{
	User:       "mail@yourdomain.com",
	Pass:       "<client-secret>",
	SenderAddr: "no-reply@yourdomain.com",
	SenderName: "Your Brand",
})
```

`Host` and `Port` default to ACS values when omitted.

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

The SMTP transport is behind the `Transport` interface, so you can inject a fake
in tests:

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
| `NewFromEnv(...Option) (*Client, error)` | Create a client from `ACS_*` env vars |
| `WithLogger(*slog.Logger) Option` | Inject a custom logger |
| `WithTransport(Transport) Option` | Inject a custom transport |
| `(*Client).Send(Email) error` | Send an email |
| `AttachFile(path) (Attachment, error)` | Read a file and detect its MIME type |
| `Config` | Host, Port, User, Pass, SenderAddr, SenderName |
| `Email` | To, Cc, Bcc, ReplyTo, Subject, Text, HTML, Attachments |
| `Attachment` | Filename, Content (`[]byte`), ContentType (optional) |
| `Transport` | `Deliver(Envelope) error` — the delivery port |

## Architecture

Files are split by responsibility (ports & adapters):

| File | Responsibility |
|------|----------------|
| `dmail.go` | `Client` facade that wires everything together |
| `config.go` | Config: data, defaults, validation, env loading |
| `message.go` | Domain types: `Email`, `Attachment` |
| `mime.go` | MIME message rendering |
| `auth.go` | SMTP `AUTH LOGIN` mechanism |
| `transport.go` | `Transport` port + SMTP adapter |

## Requirements

`SenderAddr` must be a verified MailFrom on your ACS domain, and the domain must
have SPF/DKIM verified in Azure Communication Services.

## License

MIT — see [LICENSE](LICENSE).
