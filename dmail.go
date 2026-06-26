package dmail

import "log/slog"

type Client struct {
	config    Config
	sender    string
	transport Transport
	logger    *slog.Logger
}

type Option func(*Client)

func WithLogger(logger *slog.Logger) Option {
	return func(c *Client) {
		if logger != nil {
			c.logger = logger
		}
	}
}

func WithTransport(transport Transport) Option {
	return func(c *Client) {
		if transport != nil {
			c.transport = transport
		}
	}
}

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

func NewFromEnv(options ...Option) (*Client, error) {
	return New(ConfigFromEnv(), options...)
}

func (c *Client) Send(email Email) error {
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
	if err := c.transport.Deliver(envelope); err != nil {
		log.Error("dmail: delivery failed", slog.String("error", err.Error()))
		return err
	}

	log.Info("dmail: email sent")
	return nil
}
