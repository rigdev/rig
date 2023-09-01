package twilio

import (
	"context"
	"errors"

	"github.com/rigdev/rig-go-api/api/v1/project/settings"
	"github.com/twilio/twilio-go"

	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

type twilioTextSender struct {
	client   *twilio.RestClient
	provider *settings.TextProvider
	err      error
}

// New implements text.Provider interface using the Twilio client.
func New(provider *settings.TextProvider) *twilioTextSender {
	if provider.GetCredentials().GetPublicKey() == "" {
		return &twilioTextSender{err: errors.New("missing required Twilio public key")}
	} else if provider.GetCredentials().GetPrivateKey() == "" {
		return &twilioTextSender{err: errors.New("missing required Twilio private key")}
	}
	return &twilioTextSender{
		client: twilio.NewRestClientWithParams(twilio.ClientParams{
			Username: provider.GetCredentials().GetPublicKey(),
			Password: provider.GetCredentials().GetPrivateKey(),
		}),
		provider: provider,
	}
}

// Test sends a text message to the phoneNuber.
func (tw *twilioTextSender) Test(ctx context.Context, phoneNumber string) error {
	params := &openapi.CreateMessageParams{}
	params.SetTo(phoneNumber)
	params.SetFrom(tw.provider.From)
	params.SetBody("Twilio provider successfully configured to run with Rig")
	if _, err := tw.client.Api.CreateMessage(params); err != nil {
		return err
	}
	return nil
}

// SendText sends a text message using the Twilio client.
func (tw *twilioTextSender) SendText(ctx context.Context, to, body string) error {
	params := &openapi.CreateMessageParams{}
	params.SetTo(to)
	params.SetFrom(tw.provider.From)
	params.SetBody(body)
	if _, err := tw.client.Api.CreateMessage(params); err != nil {
		return err
	}
	return nil
}
