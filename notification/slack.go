package notification

import "fmt"

type SlackNotifier struct {
	WebhookURL string
}

func (s *SlackNotifier) Notify() {
	fmt.Println("slack notify")
}
