package service

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"strings"
)

type EmailMessage struct {
	To      string
	Subject string
	Body    string
}

type EmailSender interface {
	Send(ctx context.Context, message EmailMessage) error
}

type logEmailSender struct {
	from string
}

type smtpEmailSender struct {
	address  string
	host     string
	username string
	password string
	from     string
}

func NewLogEmailSender(from string) EmailSender {
	return &logEmailSender{from: from}
}

func NewSMTPEmailSender(host, port, username, password, from string) EmailSender {
	return &smtpEmailSender{
		address:  net.JoinHostPort(host, port),
		host:     host,
		username: username,
		password: password,
		from:     from,
	}
}

func (s *logEmailSender) Send(ctx context.Context, message EmailMessage) error {
	_ = ctx

	log.Printf(
		"notification email [from=%s to=%s subject=%q body=%q]",
		s.from,
		message.To,
		message.Subject,
		message.Body,
	)

	return nil
}

func (s *smtpEmailSender) Send(ctx context.Context, message EmailMessage) error {
	_ = ctx

	if strings.TrimSpace(message.To) == "" {
		return fmt.Errorf("email recipient is empty")
	}
	if strings.TrimSpace(s.host) == "" || strings.TrimSpace(s.from) == "" {
		return fmt.Errorf("smtp sender is not configured")
	}

	var auth smtp.Auth
	if s.username != "" || s.password != "" {
		auth = smtp.PlainAuth("", s.username, s.password, s.host)
	}

	return smtp.SendMail(
		s.address,
		auth,
		s.from,
		[]string{message.To},
		[]byte(buildEmailPayload(s.from, message)),
	)
}

func buildEmailPayload(from string, message EmailMessage) string {
	headers := []string{
		fmt.Sprintf("From: %s", from),
		fmt.Sprintf("To: %s", message.To),
		fmt.Sprintf("Subject: %s", message.Subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
	}

	return strings.Join(headers, "\r\n") + "\r\n\r\n" + message.Body
}
