package notification

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"
)

type SlackNotifier struct {
	WebhookURL string
}

type SlackNotifyParams struct {
	Status string
}

func (notifer *SlackNotifier) Notify(ctx context.Context, params *NotifyParams) error {
	color, ok := StatusColor[params.Status]
	if !ok {
		return fmt.Errorf("invalid status: %s", params.Status)
	}

	message, ok := StatusMessage[params.Status]
	if !ok {
		return fmt.Errorf("invalid status: %s", params.Status)
	}

	msg := slack.WebhookMessage{
		Username: SlackBotUserName,
		IconURL:  SlackBotIconURL,
		Text: fmt.Sprintf(":large_%s_square: ", color) + "*terrapolice run result*" + "\n" +
			"*Directory*: " + params.Directory + "\n" +
			"*Run command*: terraform " + params.Command + "\n" +
			"*Result*: " + message + "\n",
	}

	err := slack.PostWebhook(notifer.WebhookURL, &msg)
	if err != nil {
		return fmt.Errorf("error posting message: %w", err)
	}
	return nil
}
