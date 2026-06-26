package dmail_test

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/danixts/dmail"
)

func selfSignedCert(t *testing.T) (tls.Certificate, *x509.CertPool) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
		DNSNames:     []string{"localhost"},
	}
	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create cert: %v", err)
	}
	parsed, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("parse cert: %v", err)
	}
	pool := x509.NewCertPool()
	pool.AddCert(parsed)
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key, Leaf: parsed}, pool
}

func startFakeSMTP(t *testing.T, cert tls.Certificate, authAdvert string, captured *string) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = listener.Close() })

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		serveFakeSMTP(conn, cert, authAdvert, captured)
	}()
	return listener.Addr().String()
}

func serveFakeSMTP(conn net.Conn, cert tls.Certificate, authAdvert string, captured *string) {
	defer func() { _ = conn.Close() }()
	reader := bufio.NewReader(conn)
	write := func(s string) { _, _ = conn.Write([]byte(s + "\r\n")) }
	readLine := func() (string, bool) {
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", false
		}
		return strings.TrimRight(line, "\r\n"), true
	}

	write("220 fake ESMTP ready")
	for {
		line, ok := readLine()
		if !ok {
			return
		}
		switch {
		case strings.HasPrefix(line, "EHLO"), strings.HasPrefix(line, "HELO"):
			write("250-fake greets you")
			write("250-STARTTLS")
			write("250 " + authAdvert)
		case line == "STARTTLS":
			write("220 ready to start TLS")
			tlsConn := tls.Server(conn, &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12})
			if err := tlsConn.Handshake(); err != nil {
				return
			}
			conn = tlsConn
			reader = bufio.NewReader(conn)
		case strings.HasPrefix(line, "AUTH PLAIN"):
			write("235 authenticated")
		case line == "AUTH LOGIN":
			write("334 VXNlcm5hbWU6")
			if _, ok := readLine(); !ok {
				return
			}
			write("334 UGFzc3dvcmQ6")
			if _, ok := readLine(); !ok {
				return
			}
			write("235 authenticated")
		case strings.HasPrefix(line, "MAIL FROM"):
			write("250 ok")
		case strings.HasPrefix(line, "RCPT TO"):
			write("250 ok")
		case line == "DATA":
			write("354 end data with <CR><LF>.<CR><LF>")
			var body strings.Builder
			for {
				dataLine, ok := readLine()
				if !ok {
					return
				}
				if dataLine == "." {
					break
				}
				body.WriteString(dataLine + "\n")
			}
			if captured != nil {
				*captured = body.String()
			}
			write("250 queued")
		case line == "QUIT":
			write("221 bye")
			return
		default:
			write("250 ok")
		}
	}
}

func clientForServer(t *testing.T, addr string, pool *x509.CertPool) *dmail.Client {
	t.Helper()
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("split host port: %v", err)
	}
	client, err := dmail.New(dmail.Config{
		Host:      host,
		Port:      port,
		Username:  "user",
		Password:  "pass",
		From:      "no-reply@example.com",
		TLSConfig: &tls.Config{ServerName: "localhost", RootCAs: pool, MinVersion: tls.VersionTLS12},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return client
}

func TestDeliverWithLoginAuth(t *testing.T) {
	cert, pool := selfSignedCert(t)
	var captured string
	addr := startFakeSMTP(t, cert, "AUTH LOGIN", &captured)

	client := clientForServer(t, addr, pool)
	if err := client.Send(dmail.Email{To: []string{"to@example.com"}, Subject: "Hi", Text: "Body"}); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if !strings.Contains(captured, "Subject: Hi") {
		t.Errorf("server did not receive message:\n%s", captured)
	}
}

func TestDeliverWithPlainAuth(t *testing.T) {
	cert, pool := selfSignedCert(t)
	var captured string
	addr := startFakeSMTP(t, cert, "AUTH LOGIN PLAIN", &captured)

	client := clientForServer(t, addr, pool)
	if err := client.Send(dmail.Email{To: []string{"to@example.com"}, Subject: "Hi", HTML: "<b>Hi</b>"}); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if !strings.Contains(captured, "text/html") {
		t.Errorf("expected html part:\n%s", captured)
	}
}

func TestDeliverDialError(t *testing.T) {
	client, err := dmail.New(dmail.Config{
		Host:     "127.0.0.1",
		Port:     "1",
		Username: "u",
		Password: "p",
		From:     "from@example.com",
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := client.Send(dmail.Email{To: []string{"a@example.com"}, Text: "x"}); err == nil {
		t.Error("expected dial error")
	}
}
