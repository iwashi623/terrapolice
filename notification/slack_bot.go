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

func (notifer *SlackBotNotifier) Notify(ctx context.Context, params *NotifyParams) error {
	c := slack.New(notifer.SlackBotToken)
	return notifer.notify(ctx, c, params)
}

func (notifer *SlackBotNotifier) notify(ctx context.Context, client *slack.Client, params *NotifyParams) error {
	c := slack.New(notifer.SlackBotToken)
	color, ok := StatusColor[params.Status]
	if !ok {
		return fmt.Errorf("invalid status: %s", params.Status)
	}

	message, ok := StatusMessage[params.Status]
	if !ok {
		return fmt.Errorf("invalid status: %s", params.Status)
	}

	_, _, err := c.PostMessageContext(ctx, notifer.SlackChannel, slack.MsgOptionBlocks(
		slack.NewSectionBlock(
			&slack.TextBlockObject{
				Type: "mrkdwn",
				Text: fmt.Sprintf(":large_%s_square: ", color) + "*terrapolice run result*" + "\n" +
					"*Directory*: " + params.Directory + "\n" +
					"*Run command*: terraform " + params.Command + "\n" +
					"*Result*: " + message + "\n",
			},
			nil,
			nil,
		),
	))
	if err != nil {
		return fmt.Errorf("error posting message: %w", err)
	}
	return nil
}
