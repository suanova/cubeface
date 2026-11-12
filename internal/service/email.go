package service

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/suanova/medrouter/internal/config"
)

type EmailService struct {
	cfg *config.Config
}

func NewEmailService(cfg *config.Config) *EmailService {
	return &EmailService{cfg: cfg}
}

func (e *EmailService) SendVerificationCode(to, code, purpose string) error {
	subject := fmt.Sprintf("%s verification code", e.cfg.SystemName)
	body := fmt.Sprintf(
		"Hello,\n\nYour %s verification code is:\n\n    %s\n\nThis code expires in %d minutes. "+
			"If you did not request it, you can safely ignore this email.\n\n— %s\n",
		e.cfg.SystemName, code, int(e.cfg.VerificationCodeTTL.Minutes()), e.cfg.SystemName,
	)
	if e.cfg.EmailDebugMode {
		log.Printf("WARN [email-debug] verification email suppressed to=%s purpose=%s code=%s", to, purpose, code)
		return nil
	}
	return e.Send(to, subject, body)
}

// Send delivers a plain-text message over SMTP using one of three transport
// modes: implicit TLS (465), STARTTLS (587) or none (local relays only).
func (e *EmailService) Send(to, subject, body string) error {
	from, err := mail.ParseAddress(e.cfg.SMTPFrom)
	if err != nil {
		return fmt.Errorf("invalid SMTP_FROM: %w", err)
	}
	if e.cfg.SMTPHost == "" {
		return fmt.Errorf("SMTP_HOST is not configured")
	}

	hostPort := net.JoinHostPort(e.cfg.SMTPHost, strconv.Itoa(e.cfg.SMTPPort))
	dialer := &net.Dialer{Timeout: e.cfg.SMTPTimeout}
	tlsConfig := &tls.Config{
		ServerName: e.cfg.SMTPHost,
		MinVersion: tls.VersionTLS12,
	}

	var conn net.Conn
	var client *smtp.Client
	mode := strings.ToLower(e.cfg.SMTPTLSMode)

	switch mode {
	case "implicit", "tls", "ssl":
		conn, err = tls.DialWithDialer(dialer, "tcp", hostPort, tlsConfig)
		if err != nil {
			return fmt.Errorf("smtp implicit tls dial: %w", err)
		}
		client, err = smtp.NewClient(conn, e.cfg.SMTPHost)
		if err != nil {
			conn.Close()
			return fmt.Errorf("smtp client: %w", err)
		}
	case "starttls":
		conn, err = dialer.Dial("tcp", hostPort)
		if err != nil {
			return fmt.Errorf("smtp dial: %w", err)
		}
		client, err = smtp.NewClient(conn, e.cfg.SMTPHost)
		if err != nil {
			conn.Close()
			return fmt.Errorf("smtp client: %w", err)
		}
		if ok, _ := client.Extension("STARTTLS"); !ok {
			client.Close()
			return fmt.Errorf("smtp server does not offer STARTTLS")
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			client.Close()
			return fmt.Errorf("smtp starttls: %w", err)
		}
	case "none":
		conn, err = dialer.Dial("tcp", hostPort)
		if err != nil {
			return fmt.Errorf("smtp dial: %w", err)
		}
		client, err = smtp.NewClient(conn, e.cfg.SMTPHost)
		if err != nil {
			conn.Close()
			return fmt.Errorf("smtp client: %w", err)
		}
	default:
		return fmt.Errorf("unsupported SMTP_TLS_MODE %q (use implicit, starttls or none)", e.cfg.SMTPTLSMode)
	}
	defer client.Close()
	_ = conn.SetDeadline(time.Now().Add(e.cfg.SMTPTimeout))

	if e.cfg.SMTPUsername != "" {
		auth := smtp.PlainAuth("", e.cfg.SMTPUsername, e.cfg.SMTPPassword, e.cfg.SMTPHost)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}
	if err := client.Mail(from.Address); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt to: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := w.Write([]byte(e.buildMessage(from, to, subject, body))); err != nil {
		w.Close()
		return fmt.Errorf("smtp write body: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp finish body: %w", err)
	}
	if err := client.Quit(); err != nil {
		return fmt.Errorf("smtp quit: %w", err)
	}
	return nil
}

func (e *EmailService) buildMessage(from *mail.Address, to, subject, body string) string {
	headerFrom := from.Address
	if from.Name != "" {
		headerFrom = from.String()
	}
	body = strings.ReplaceAll(strings.ReplaceAll(body, "\r\n", "\n"), "\n", "\r\n")
	headers := []string{
		"From: " + headerFrom,
		"To: " + to,
		"Subject: " + sanitizeHeader(subject),
		"Date: " + time.Now().UTC().Format(time.RFC1123Z),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"Content-Transfer-Encoding: 8bit",
	}
	return strings.Join(headers, "\r\n") + "\r\n\r\n" + body
}

func sanitizeHeader(v string) string {
	return strings.NewReplacer("\r", " ", "\n", " ").Replace(v)
}
