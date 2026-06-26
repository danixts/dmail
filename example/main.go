package main

import (
	"bufio"
	"log/slog"
	"os"
	"strings"

	"github.com/danixts/dmail"
)

func loadEnv(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
			continue
		}
		pair := strings.SplitN(line, "=", 2)
		key := strings.TrimSpace(pair[0])
		if _, ok := os.LookupEnv(key); !ok {
			os.Setenv(key, strings.TrimSpace(pair[1]))
		}
	}
}

func main() {
	loadEnv("../.env")

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	client, err := dmail.NewFromEnv(dmail.WithLogger(logger))
	if err != nil {
		logger.Error("cannot create client", slog.String("error", err.Error()))
		os.Exit(1)
	}

	attachment, err := dmail.AttachFile("sample.txt")
	if err != nil {
		logger.Error("cannot read attachment", slog.String("error", err.Error()))
		os.Exit(1)
	}

	err = client.Send(dmail.Email{
		To:      []string{"codedany789@gmail.com"},
		Cc:      []string{"arielcristhian32@gmail.com"},
		Subject: "dmail package test",
		Text:    "If you read this, the dmail package works.",
		HTML: `<div style="font-family:Arial,sans-serif;color:#333">
  <h2 style="color:#6b21a8">dmail works! 🎉</h2>
  <p>Sent with <b>github.com/danixts/dmail</b> in a few lines, with CC and attachment.</p>
</div>`,
		Attachments: []dmail.Attachment{attachment},
	})
	if err != nil {
		os.Exit(1)
	}
}
