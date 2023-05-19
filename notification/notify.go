package notification

import "os"

type Notifier interface {
	Notify()
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
