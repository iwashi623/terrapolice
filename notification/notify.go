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

	SlackBotUserName = "terrapolice"
	SlackBotIconURL  = "https://raw.githubusercontent.com/iwashi623/terrapolice/main/images/terrapolice.png"
)

var StatusColor = map[Status]string{
	StatusSuccess:      "green",
	StatusError:        "red",
	StatusDiffDetected: "yellow",
}

var StatusMessage = map[Status]string{
	StatusSuccess:      "差分なしです:saluting_face:",
	StatusError:        "実行時にエラーが発生しました:alert:",
	StatusDiffDetected: "差分を検知したよ:eyes:",
}

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

func NewNotifier(option string) (Notifier, error) {
	switch option {
	case "slack":
		return &SlackNotifier{
			WebhookURL: os.Getenv("SLACK_WEBHOOK_URL"),
		}, nil
	case "slack_bot":
		return &SlackBotNotifier{
			SlackBotToken: os.Getenv("SLACK_BOT_TOKEN"),
			SlackChannel:  os.Getenv("SLACK_CHANNEL"),
		}, nil
	default:
		return nil, fmt.Errorf("invalid Notification option: %s", option)
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
