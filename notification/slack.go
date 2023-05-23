package notification

import (
	"context"
	"fmt"
)

type SlackNotifier struct {
	WebhookURL string
}

type SlackNotifyParams struct {
	Status string
}

func (s *SlackNotifier) Notify(ctx context.Context, params *NotifyParams) error {
	fmt.Println("slack notify")
	return nil
}
