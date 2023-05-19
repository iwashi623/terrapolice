package notification

import "fmt"

type SlackBotNotifier struct {
	SlackBotToken string
}

func (s *SlackBotNotifier) Notify() {
	fmt.Println("slack notify")
}
