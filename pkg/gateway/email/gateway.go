package email

import (
	"context"
)

type Recipient struct {
	Name  string
	Email string
}

type Message struct {
	From     *Recipient
	To       []*Recipient
	Subject  string
	TextPart string
	HTMLPart string
}

type Gateway interface {
	Send(ctx context.Context, msg *Message) error
	Test(ctx context.Context, testEmail string) error
}
