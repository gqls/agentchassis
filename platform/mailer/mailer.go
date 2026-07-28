// Package mailer is the platform's ONE way to send email.
//
// WHY THIS EXISTS. Until 2026-07-28 there was no SMTP anywhere in the built
// code — `grep -rn "net/smtp" --include=*.go platform/ internal/ cmd/` returned
// nothing. The only working mailer in the estate lived in idea.uk's VM app under
// docs/agent_docs/docs024_key_docs_latest/idea.uk/golang_files/service.go, i.e.
// outside `go build`, untested by CI and undeployable by the image pipeline.
// `send_notification` (basic_actions.go) is NOT email — it produces a Kafka
// message, and the name has misled at least one reader.
//
// Every "we'll email you a link" journey needs this: idea.uk's paid report today,
// the gripper dossier next, contact forms after that. The gripper design's plan of
// record was to LIFT idea.uk's file, which would have made mailer #2. This package
// is that lift done once, in the build, so the third caller reuses rather than
// copies (features_open/024 A2).
//
// WHAT WAS PORTED, AND WHY EACH DETAIL IS LOAD-BEARING — every one of these is a
// scar from the idea.uk implementation, not a preference:
//
//   - Bodies are base64-encoded and wrapped at 76 columns. SMTP's hard line limit
//     is 998 octets and a long unbroken HTML line silently exceeded it; the
//     original comment records the resulting fold corrupting the message.
//   - Port 465 gets implicit TLS via tls.Dial; anything else takes the STARTTLS
//     path through smtp.SendMail. Getting this the wrong way round hangs.
//   - The dial and the whole conversation are bounded, so a network fault fails in
//     seconds rather than sitting on the OS TCP timeout.
//   - Headers go through mime.QEncoding so a non-ASCII subject or display name
//     survives.
//
// WHAT THIS ADDS over the ported original: address and header validation. A bare
// CR or LF in a recipient, subject or display name is refused rather than written
// into the message, because a shared package will eventually be handed a value
// that came from a form. The original interpolated `to` directly; that was safe
// there only because the value came from its own store.
//
// NOT IN SCOPE, deliberately: retries, queueing, bounce handling, templating. A
// caller that needs delivery guarantees should persist its own intent and retry —
// the platform already has work items for exactly that.
package mailer

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"os"
	"strings"
	"time"
)

const (
	// b64Wrap keeps every encoded body line far under SMTP's 998-octet limit.
	b64Wrap = 76
	// dialTimeout and convTimeout bound the connect and the whole SMTP
	// conversation respectively.
	dialTimeout = 10 * time.Second
	convTimeout = 30 * time.Second
)

// Config is the transport and identity. Host/Port/From are required; User/Pass
// are optional so an unauthenticated relay works.
type Config struct {
	Host     string
	Port     string
	User     string
	Pass     string
	From     string // envelope sender and From: address
	FromName string // optional display name
	ReplyTo  string // optional
}

// Message is one email. Text is required; HTML is optional and, when present,
// makes the message multipart/alternative.
type Message struct {
	To      []string
	Subject string
	Text    string
	HTML    string
}

// Sender is the seam callers depend on, so tests substitute a recorder rather
// than reaching for a network.
type Sender interface {
	Send(ctx context.Context, m Message) error
}

// SMTP sends over SMTP. Construct with New.
type SMTP struct{ cfg Config }

// New validates the config and returns a Sender.
func New(cfg Config) (*SMTP, error) {
	if strings.TrimSpace(cfg.Host) == "" {
		return nil, fmt.Errorf("mailer: Host is required")
	}
	if strings.TrimSpace(cfg.Port) == "" {
		return nil, fmt.Errorf("mailer: Port is required")
	}
	if err := validateAddr("From", cfg.From); err != nil {
		return nil, err
	}
	if cfg.ReplyTo != "" {
		if err := validateAddr("ReplyTo", cfg.ReplyTo); err != nil {
			return nil, err
		}
	}
	if err := validateHeaderValue("FromName", cfg.FromName); err != nil {
		return nil, err
	}
	return &SMTP{cfg: cfg}, nil
}

// FromEnv reads a config from <PREFIX>_HOST/_PORT/_USER/_PASS/_FROM/_FROM_NAME/
// _REPLY_TO. Mirrors the platform convention that a secret is named by env var
// and never carried in config (aiservice does the same with api_key_env_var).
func FromEnv(prefix string) (Config, error) {
	get := func(suffix string) string { return strings.TrimSpace(os.Getenv(prefix + suffix)) }
	cfg := Config{
		Host:     get("_HOST"),
		Port:     get("_PORT"),
		User:     get("_USER"),
		Pass:     get("_PASS"),
		From:     get("_FROM"),
		FromName: get("_FROM_NAME"),
		ReplyTo:  get("_REPLY_TO"),
	}
	if cfg.Port == "" {
		cfg.Port = "587"
	}
	if cfg.Host == "" || cfg.From == "" {
		return cfg, fmt.Errorf("mailer: %s_HOST and %s_FROM must be set", prefix, prefix)
	}
	return cfg, nil
}

// Send builds and delivers the message. It is synchronous: a caller that wants
// fire-and-forget owns that decision, because swallowing a delivery error is a
// choice worth making explicitly.
func (s *SMTP) Send(ctx context.Context, m Message) error {
	raw, err := Build(s.cfg, m)
	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return s.deliver(ctx, m.To, raw)
}

// Build renders the RFC 5322 message. Exported and pure so the interesting half
// — encoding, folding, header safety — is testable without a network.
func Build(cfg Config, m Message) ([]byte, error) {
	if len(m.To) == 0 {
		return nil, fmt.Errorf("mailer: no recipients")
	}
	for _, to := range m.To {
		if err := validateAddr("To", to); err != nil {
			return nil, err
		}
	}
	if err := validateHeaderValue("Subject", m.Subject); err != nil {
		return nil, err
	}
	if strings.TrimSpace(m.Text) == "" {
		return nil, fmt.Errorf("mailer: Text body is required (it is the fallback part)")
	}

	from := cfg.From
	if cfg.FromName != "" {
		from = fmt.Sprintf("%s <%s>", mime.QEncoding.Encode("utf-8", cfg.FromName), cfg.From)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "From: %s\r\n", from)
	fmt.Fprintf(&b, "To: %s\r\n", strings.Join(m.To, ", "))
	if cfg.ReplyTo != "" {
		fmt.Fprintf(&b, "Reply-To: %s\r\n", cfg.ReplyTo)
	}
	fmt.Fprintf(&b, "Subject: %s\r\n", mime.QEncoding.Encode("utf-8", m.Subject))
	fmt.Fprintf(&b, "Date: %s\r\n", time.Now().Format(time.RFC1123Z))
	mid, err := messageID(cfg.From)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(&b, "Message-ID: %s\r\n", mid)
	b.WriteString("MIME-Version: 1.0\r\n")

	if m.HTML == "" {
		b.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
		b.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")
		b.WriteString(b64Body(m.Text))
		return []byte(b.String()), nil
	}

	boundary, err := randomToken()
	if err != nil {
		return nil, err
	}
	// multipart/alternative: the client picks the richest part it can render, and
	// the plain part must come FIRST so a text-only client shows the fallback.
	fmt.Fprintf(&b, "Content-Type: multipart/alternative; boundary=%q\r\n\r\n", boundary)
	fmt.Fprintf(&b, "--%s\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n"+
		"Content-Transfer-Encoding: base64\r\n\r\n%s\r\n", boundary, b64Body(m.Text))
	fmt.Fprintf(&b, "--%s\r\nContent-Type: text/html; charset=\"UTF-8\"\r\n"+
		"Content-Transfer-Encoding: base64\r\n\r\n%s\r\n", boundary, b64Body(m.HTML))
	fmt.Fprintf(&b, "--%s--\r\n", boundary)
	return []byte(b.String()), nil
}

// UsesImplicitTLS reports whether this port takes the tls.Dial path rather than
// STARTTLS. Exported so the branch is testable without opening a socket — the
// original's equivalent decision was inline and therefore untested.
func UsesImplicitTLS(port string) bool { return port == "465" }

func (s *SMTP) deliver(ctx context.Context, to []string, msg []byte) error {
	addr := net.JoinHostPort(s.cfg.Host, s.cfg.Port)
	var auth smtp.Auth
	if s.cfg.User != "" {
		auth = smtp.PlainAuth("", s.cfg.User, s.cfg.Pass, s.cfg.Host)
	}

	if !UsesImplicitTLS(s.cfg.Port) {
		return smtp.SendMail(addr, auth, s.cfg.From, to, msg)
	}

	d := &net.Dialer{Timeout: dialTimeout}
	conn, err := tls.DialWithDialer(d, "tcp", addr, &tls.Config{ServerName: s.cfg.Host})
	if err != nil {
		return fmt.Errorf("mailer: dial %s: %w", addr, err)
	}
	// Bound the whole conversation, not just the connect.
	deadline := time.Now().Add(convTimeout)
	if dl, ok := ctx.Deadline(); ok && dl.Before(deadline) {
		deadline = dl
	}
	_ = conn.SetDeadline(deadline)

	c, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("mailer: smtp client: %w", err)
	}
	defer func() { _ = c.Close() }()

	if auth != nil {
		if err := c.Auth(auth); err != nil {
			return fmt.Errorf("mailer: auth: %w", err)
		}
	}
	if err := c.Mail(s.cfg.From); err != nil {
		return fmt.Errorf("mailer: MAIL FROM: %w", err)
	}
	for _, rcpt := range to {
		if err := c.Rcpt(rcpt); err != nil {
			return fmt.Errorf("mailer: RCPT TO %s: %w", rcpt, err)
		}
	}
	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("mailer: DATA: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("mailer: write body: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("mailer: close body: %w", err)
	}
	return c.Quit()
}

// b64Body base64-encodes a part and wraps it so no line approaches SMTP's
// 998-octet limit. Long unbroken HTML lines were corrupted by server-side
// folding before the original added this.
func b64Body(s string) string {
	enc := base64.StdEncoding.EncodeToString([]byte(s))
	var b strings.Builder
	for i := 0; i < len(enc); i += b64Wrap {
		end := i + b64Wrap
		if end > len(enc) {
			end = len(enc)
		}
		b.WriteString(enc[i:end])
		b.WriteString("\r\n")
	}
	return b.String()
}

// validateAddr refuses anything that cannot safely occupy an address header.
// Header injection is the reason: a bare CR or LF would end the header and let a
// caller append their own.
func validateAddr(field, addr string) error {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return fmt.Errorf("mailer: %s address is empty", field)
	}
	if err := validateHeaderValue(field, addr); err != nil {
		return err
	}
	at := strings.IndexByte(addr, '@')
	if at <= 0 || at == len(addr)-1 || strings.ContainsAny(addr, " <>,") {
		return fmt.Errorf("mailer: %s address %q is not a bare addr-spec", field, addr)
	}
	return nil
}

func validateHeaderValue(field, v string) error {
	if strings.ContainsAny(v, "\r\n") {
		return fmt.Errorf("mailer: %s contains a line break (header injection refused)", field)
	}
	return nil
}

func randomToken() (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("mailer: entropy: %w", err)
	}
	return hex.EncodeToString(raw[:]), nil
}

func messageID(from string) (string, error) {
	tok, err := randomToken()
	if err != nil {
		return "", err
	}
	domain := "localhost"
	if at := strings.LastIndexByte(from, '@'); at >= 0 && at < len(from)-1 {
		domain = from[at+1:]
	}
	return fmt.Sprintf("<%s@%s>", tok, domain), nil
}
