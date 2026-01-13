package mailer

import "github.com/tkahng/playground/internal/tools/utils"

type LogMailer struct {
}

// Send implements Mailer.
func (l *LogMailer) Send(message *Message) error {
	utils.PrettyPrintJSON(message)
	return nil
}

var _ Mailer = (*LogMailer)(nil)
