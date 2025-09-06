package mailer

import "sync"

type TestMailer struct {
	Mailer   Mailer
	Messages []*Message
	Wg       *sync.WaitGroup
}

// Send implements Mailer.
func (t *TestMailer) Send(message *Message) error {
	t.Messages = append(t.Messages, message)
	err := t.Mailer.Send(message)
	if t.Wg != nil {
		t.Wg.Done()
	}
	return err
}

var _ Mailer = (*TestMailer)(nil)
