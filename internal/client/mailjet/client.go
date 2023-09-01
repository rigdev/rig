package mailjet

import (
	"context"
	"errors"

	"github.com/mailjet/mailjet-apiv3-go"
	"github.com/rigdev/rig-go-api/api/v1/project/settings"
	"github.com/rigdev/rig/internal/gateway/email"
)

type mailjetProvider struct {
	client   *mailjet.Client
	provider *settings.EmailProvider
	err      error
}

// New implements text.Provider interface using the Mailjet client.
func New(provider *settings.EmailProvider) *mailjetProvider {
	if provider.GetCredentials().GetPublicKey() == "" {
		return &mailjetProvider{err: errors.New("missing required Mailjet public key")}
	} else if provider.GetCredentials().GetPrivateKey() == "" {
		return &mailjetProvider{err: errors.New("missing required Mailjet private key")}
	}
	return &mailjetProvider{
		client:   mailjet.NewMailjetClient(provider.GetCredentials().GetPublicKey(), provider.GetCredentials().GetPrivateKey()),
		provider: provider,
	}
}

// Test sends an email to the testEmail account.
func (e *mailjetProvider) Test(ctx context.Context, testEmail string) error {
	if e.err != nil {
		return e.err
	}
	messagesInfo := []mailjet.InfoMessagesV31{
		{
			From: &mailjet.RecipientV31{
				Email: e.provider.From,
				Name:  "Mailjet provider configured",
			},
			To: &mailjet.RecipientsV31{
				mailjet.RecipientV31{
					Email: testEmail,
				},
			},
			Subject:  "Mailjet provider configured",
			TextPart: "Your Mailjet provider has successfully been configured with Rig.",
		},
	}
	messages := mailjet.MessagesV31{Info: messagesInfo}
	if _, err := e.client.SendMailV31(&messages); err != nil {
		return err
	}
	return nil
}

// SendEmail sends an email using the Mailjet client. Note that the client supports either text input or html.
func (e *mailjetProvider) Send(ctx context.Context, msg *email.Message) error {
	if e.err != nil {
		return e.err
	}

	im := mailjet.InfoMessagesV31{
		To:       &mailjet.RecipientsV31{},
		Subject:  msg.Subject,
		HTMLPart: msg.HTMLPart,
		TextPart: msg.TextPart,
	}
	for _, r := range msg.To {
		*im.To = append(*im.To, mailjet.RecipientV31{
			Email: r.Email,
			Name:  r.Name,
		})
	}

	if msg.From != nil {
		im.From = &mailjet.RecipientV31{
			Email: msg.From.Email,
			Name:  msg.From.Name,
		}
	} else {
		im.From = &mailjet.RecipientV31{
			Email: e.provider.From,
		}
	}

	messages := mailjet.MessagesV31{Info: []mailjet.InfoMessagesV31{im}}
	if _, err := e.client.SendMailV31(&messages); err != nil {
		return err
	}
	return nil
}
