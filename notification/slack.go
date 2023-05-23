package notification

import "fmt"

type SlackNotifier struct {
	WebhookURL string
}

type SlackNotifyParams struct {
	Status string
}

func (s *SlackNotifier) Notify(p NotifyParams) {
	fmt.Println("slack notify")
}
