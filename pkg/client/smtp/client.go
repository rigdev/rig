package smtp

import (
	"context"
	"errors"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/rigdev/rig-go-api/api/v1/project/settings"
	"github.com/rigdev/rig/pkg/gateway/email"
)

type smtpProvider struct {
	provider *settings.EmailProvider
	err      error
}

func New(provider *settings.EmailProvider) *smtpProvider {
	if provider.GetInstance().GetSmtp().GetHost() == "" {
		return &smtpProvider{err: errors.New("host is required")}
	}
	if provider.GetInstance().GetSmtp().GetPort() == 0 {
		return &smtpProvider{err: errors.New("port is required")}
	}
	return &smtpProvider{
		provider: provider,
	}
}

// Test sends an email to the testEmail account.
func (e *smtpProvider) Test(ctx context.Context, testEmail string) error {
	if e.err != nil {
		return e.err
	}

	m := &email.Message{
		From: &email.Recipient{
			Email: e.provider.GetFrom(),
			Name:  "Test",
		},
		To: []*email.Recipient{
			{
				Email: testEmail,
				Name:  "Test",
			},
		},
		Subject:  "Smtp provider configured",
		HTMLPart: "<p>Your Smtp provider has successfully been configured with Rig.</p>",
	}

	msg := buildMessage(m)

	addr := fmt.Sprintf("%s:%d", e.provider.GetInstance().GetSmtp().GetHost(), e.provider.GetInstance().GetSmtp().GetPort())

	auth := smtp.PlainAuth("", e.provider.GetCredentials().GetPublicKey(), e.provider.GetCredentials().GetPrivateKey(), e.provider.GetInstance().GetSmtp().GetHost())
	err := smtp.SendMail(addr, auth, e.provider.GetFrom(), []string{testEmail}, []byte(msg))

	if err != nil {
		return err
	}
	return nil
}

func (e *smtpProvider) Send(ctx context.Context, mail *email.Message) error {
	if e.err != nil {
		return e.err
	}

	toEmails := make([]string, len(mail.To))
	for i, r := range mail.To {
		toEmails[i] = r.Email
	}

	if mail.From == nil {
		mail.From = &email.Recipient{
			Email: e.provider.GetFrom(),
		}
	}

	msg := buildMessage(mail)

	addr := fmt.Sprintf("%s:%d", e.provider.GetInstance().GetSmtp().GetHost(), e.provider.GetInstance().GetSmtp().GetPort())

	auth := smtp.PlainAuth("", e.provider.GetCredentials().GetPublicKey(), e.provider.GetCredentials().GetPrivateKey(), e.provider.GetInstance().GetSmtp().GetHost())
	err := smtp.SendMail(addr, auth, e.provider.GetFrom(), toEmails, []byte(msg))
	if err != nil {
		return err
	}

	return nil
}

func buildMessage(m *email.Message) string {
	emails := make([]string, len(m.To))
	for i, r := range m.To {
		emails[i] = r.Email
	}

	msg := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\r\n"
	msg += fmt.Sprintf("From: %s\r\n", m.From.Email)
	msg += fmt.Sprintf("To: %s\r\n", strings.Join(emails, ";"))
	msg += fmt.Sprintf("Subject: %s\r\n", m.Subject)
	msg += fmt.Sprintf("\r\n%s\r\n", m.HTMLPart)

	return msg
}
