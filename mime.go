package dmail

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"mime"
	"mime/multipart"
	"net/textproto"
	"strings"
	"time"
)

type mimeEntity struct {
	header textproto.MIMEHeader
	body   []byte
}

func renderMessage(sender string, email Email) []byte {
	content := buildContent(email)

	var message bytes.Buffer
	for _, header := range messageHeaders(sender, email) {
		message.WriteString(header + "\r\n")
	}
	for _, name := range []string{"Content-Type", "Content-Transfer-Encoding"} {
		if value := content.header.Get(name); value != "" {
			message.WriteString(name + ": " + value + "\r\n")
		}
	}
	message.WriteString("\r\n")
	message.Write(content.body)
	return message.Bytes()
}

func messageHeaders(sender string, email Email) []string {
	headers := []string{
		"From: " + sanitizeHeader(sender),
		"To: " + joinAddresses(email.To),
	}
	if len(email.Cc) > 0 {
		headers = append(headers, "Cc: "+joinAddresses(email.Cc))
	}
	if email.ReplyTo != "" {
		headers = append(headers, "Reply-To: "+sanitizeHeader(email.ReplyTo))
	}
	return append(headers,
		"Subject: "+encodeHeader(sanitizeHeader(email.Subject)),
		"Date: "+time.Now().Format(time.RFC1123Z),
		"MIME-Version: 1.0",
	)
}

func joinAddresses(addresses []string) string {
	cleaned := make([]string, len(addresses))
	for i, address := range addresses {
		cleaned[i] = sanitizeHeader(address)
	}
	return strings.Join(cleaned, ", ")
}

func sanitizeHeader(value string) string {
	return strings.NewReplacer("\r", "", "\n", "").Replace(value)
}

func buildContent(email Email) mimeEntity {
	content := buildBody(email)
	if len(email.Attachments) == 0 {
		return content
	}
	parts := []mimeEntity{content}
	for _, attachment := range email.Attachments {
		parts = append(parts, buildAttachment(attachment))
	}
	return combineParts("mixed", parts)
}

func buildBody(email Email) mimeEntity {
	var parts []mimeEntity
	if strings.TrimSpace(email.Text) != "" {
		parts = append(parts, buildText("text/plain", email.Text))
	}
	if strings.TrimSpace(email.HTML) != "" {
		parts = append(parts, buildText("text/html", email.HTML))
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return combineParts("alternative", parts)
}

func buildText(mediaType, body string) mimeEntity {
	header := textproto.MIMEHeader{}
	header.Set("Content-Type", mediaType+`; charset="UTF-8"`)
	header.Set("Content-Transfer-Encoding", "8bit")
	return mimeEntity{header: header, body: []byte(body)}
}

func buildAttachment(attachment Attachment) mimeEntity {
	header := textproto.MIMEHeader{}
	header.Set("Content-Type", fmt.Sprintf("%s; name=%q", attachment.resolvedContentType(), attachment.Filename))
	header.Set("Content-Transfer-Encoding", "base64")
	header.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", attachment.Filename))
	return mimeEntity{header: header, body: encodeBase64(attachment.Content)}
}

func combineParts(subtype string, parts []mimeEntity) mimeEntity {
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)
	for _, part := range parts {
		partWriter, err := writer.CreatePart(part.header)
		if err != nil {
			continue
		}
		_, _ = partWriter.Write(part.body)
	}
	_ = writer.Close()

	header := textproto.MIMEHeader{}
	header.Set("Content-Type", fmt.Sprintf("multipart/%s; boundary=%q", subtype, writer.Boundary()))
	return mimeEntity{header: header, body: buffer.Bytes()}
}

func formatSender(name, address string) string {
	if name == "" {
		return address
	}
	return encodeHeader(name) + " <" + address + ">"
}

func encodeHeader(value string) string {
	return mime.QEncoding.Encode("UTF-8", value)
}

func encodeBase64(data []byte) []byte {
	encoded := base64.StdEncoding.EncodeToString(data)
	var wrapped strings.Builder
	for len(encoded) > 76 {
		wrapped.WriteString(encoded[:76] + "\r\n")
		encoded = encoded[76:]
	}
	wrapped.WriteString(encoded)
	return []byte(wrapped.String())
}
