package notification

import (
	"fmt"
)

type SlackBotNotifier struct {
	SlackBotToken string
}

func (s *SlackBotNotifier) Notify(p NotifyParams) {
	fmt.Println(p.Buffer.String())
}
