package notification

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"
)

type SlackBotNotifier struct {
	SlackBotToken string
	SlackChannel  string
}

func (s *SlackBotNotifier) Notify(ctx context.Context, params NotifyParams) {
	c := slack.New(s.SlackBotToken)
	if params.Status == StatusSuccess {
		notifySuccess(c, params)
	} else if params.Status == StatusDiffDetected {
		notifyDiffDetected(c)
	}
}

func notifySuccess(c *slack.Client, params NotifyParams) {
	fmt.Println(string(params.Buffer.String()))
}

func notifyError(c *slack.Client) {
	fmt.Println("slack bot notify error")
}

func notifyDiffDetected(c *slack.Client) {
	fmt.Println("slack bot notify diff detected")
}
