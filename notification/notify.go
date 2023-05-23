package notification

import (
	"bytes"
	"fmt"
	"os"
)

const (
	StatusSuccess      = "success"
	StatusError        = "error"
	StatusDiffDetected = "diff_detected"
)

type Notifier interface {
	Notify(p NotifyParams)
}

type Status string

type NotifyParams struct {
	Status Status
	Buffer *bytes.Buffer
}

func CreateNotifier(option string) Notifier {
	switch option {
	case "slack":
		return &SlackNotifier{
			WebhookURL: os.Getenv("SLACK_WEBHOOK_URL"),
		}
	case "slack_bot":
		return &SlackBotNotifier{
			SlackBotToken: os.Getenv("SLACK_BOT_TOKEN"),
		}
	default:
		return &SlackBotNotifier{
			SlackBotToken: os.Getenv("SLACK_BOT_TOKEN"),
		}
	}
}

func NewStatus(s string) (Status, error) {
	switch s {
	case string(StatusSuccess), string(StatusError):
		return Status(s), nil
	default:
		return "", fmt.Errorf("invalid status: %s", s)
	}
}
