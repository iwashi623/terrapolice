package notification

import (
	"bytes"
	"context"
	"fmt"
	"os"
)

const (
	StatusSuccess      = "success"
	StatusError        = "error"
	StatusDiffDetected = "diff_detected"
)

type Notifier interface {
	Notify(ctx context.Context, params *NotifyParams) error
}

type Status string

type NotifyParams struct {
	Status    Status
	Buffer    *bytes.Buffer
	Command   string
	Directory string
}

func NewNotifier(option string) Notifier {
	switch option {
	case "slack":
		return &SlackNotifier{
			WebhookURL: os.Getenv("SLACK_WEBHOOK_URL"),
		}
	case "slack_bot":
		return &SlackBotNotifier{
			SlackBotToken: os.Getenv("SLACK_BOT_TOKEN"),
			SlackChannel:  os.Getenv("SLACK_CHANNEL"),
		}
	default:
		return &SlackBotNotifier{
			SlackBotToken: os.Getenv("SLACK_BOT_TOKEN"),
			SlackChannel:  os.Getenv("SLACK_CHANNEL"),
		}
	}
}

func NewStatus(s string) (Status, error) {
	switch s {
	case StatusSuccess, StatusError, StatusDiffDetected:
		return Status(s), nil
	default:
		return "", fmt.Errorf("invalid status: %s", s)
	}
}
