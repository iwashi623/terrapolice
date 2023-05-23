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

func (s *SlackBotNotifier) Notify(ctx context.Context, params *NotifyParams) error {
	c := slack.New(s.SlackBotToken)
	switch params.Status {
	case StatusSuccess:
		return notifySuccessMessage(c, params)
	case StatusError:
		return notifyErrorMessage(c, params)
	case StatusDiffDetected:
		return notifyDiffDetectedMessage(c, params)
	default:
		return fmt.Errorf("invalid status: %s", params.Status)
	}
}

func notifySuccessMessage(c *slack.Client, params *NotifyParams) error {
	fmt.Println("slack bot notify success")
	fmt.Println(string(params.Buffer.String()))
	return nil
}

func notifyErrorMessage(c *slack.Client, params *NotifyParams) error {
	fmt.Println("slack bot notify error")
	fmt.Println(string(params.Buffer.String()))
	return nil
}

func notifyDiffDetectedMessage(c *slack.Client, params *NotifyParams) error {
	fmt.Println("slack bot notify diff detected")
	fmt.Println(string(params.Buffer.String()))
	return nil
}
