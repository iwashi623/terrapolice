package notification

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"
)

var StatusColor = map[Status]string{
	StatusSuccess:      "green",
	StatusError:        "red",
	StatusDiffDetected: "yellow",
}

type SlackBotNotifier struct {
	SlackBotToken string
	SlackChannel  string
}

func (s *SlackBotNotifier) Notify(ctx context.Context, params *NotifyParams) error {
	c := slack.New(s.SlackBotToken)
	return s.notify(ctx, c, params)
}

func (s *SlackBotNotifier) notify(ctx context.Context, client *slack.Client, params *NotifyParams) error {
	c := slack.New(s.SlackBotToken)
	color, ok := StatusColor[params.Status]
	if !ok {
		return fmt.Errorf("invalid status: %s", params.Status)
	}

	_, _, err := c.PostMessageContext(ctx, s.SlackChannel, slack.MsgOptionBlocks(
		slack.NewSectionBlock(
			&slack.TextBlockObject{
				Type: "mrkdwn",
				Text: fmt.Sprintf(":large_%s_square: ", color) + "*terrapolice run result*" + "\n" +
					"result: " + string(params.Status) + "\n" +
					"run command: terraform " + params.Command + "\n" +
					"directory: " + params.Directory + "\n",
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
