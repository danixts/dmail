package dmail

import (
	"context"
	"log/slog"
)

// Client sends [Email] messages over SMTP.
type Client struct {
	config    Config
	sender    string
	transport Transport
	logger    *slog.Logger
}

// Option configures a [Client] created by [New] or [NewFromEnv].
type Option func(*Client)

// WithLogger sets the logger used by the [Client]. A nil logger is ignored.
func WithLogger(logger *slog.Logger) Option {
	return func(c *Client) {
		if logger != nil {
			c.logger = logger
		}
	}
}

// WithTransport overrides the [Transport] used to deliver messages, mainly
// for testing. A nil transport is ignored.
func WithTransport(transport Transport) Option {
	return func(c *Client) {
		if transport != nil {
			c.transport = transport
		}
	}
}

// New creates a [Client] from config, applying any options. It returns an
// error if config fails validation.
func New(config Config, options ...Option) (*Client, error) {
	config = config.withDefaults()
	if err := config.validate(); err != nil {
		return nil, err
	}

	client := &Client{
		config: config,
		sender: formatSender(config.FromName, config.From),
		logger: slog.Default(),
	}
	for _, option := range options {
		option(client)
	}
	if client.transport == nil {
		client.transport = newSMTPTransport(config)
	}
	return client, nil
}

// NewFromEnv creates a [Client] using [ConfigFromEnv].
func NewFromEnv(options ...Option) (*Client, error) {
	return New(ConfigFromEnv(), options...)
}

// Send renders email (applying its [Template] if set) and delivers it
// through the client's [Transport]. It stops as soon as ctx is done.
func (c *Client) Send(ctx context.Context, email Email) error {
	if err := email.applyTemplate(); err != nil {
		c.logger.Error("dmail: template render failed", slog.String("error", err.Error()))
		return err
	}

	log := c.logger.With(
		slog.Any("to", email.To),
		slog.Any("cc", email.Cc),
		slog.String("subject", email.Subject),
	)

	if err := email.validate(); err != nil {
		log.Error("dmail: email validation failed", slog.String("error", err.Error()))
		return err
	}

	envelope := Envelope{
		From:       c.config.From,
		Recipients: email.recipients(),
		Message:    renderMessage(c.sender, email),
	}

	log.Info("dmail: sending email", slog.Int("recipients", len(envelope.Recipients)))
	if err := c.transport.Deliver(ctx, envelope); err != nil {
		log.Error("dmail: delivery failed", slog.String("error", err.Error()))
		return err
	}

	log.Info("dmail: email sent")
	return nil
}
