package mailer

import (
	"sync"
)

type TestMailer struct {
	Mailer   Mailer
	Messages []*Message
	Wg       *sync.WaitGroup
}

func (t *TestMailer) GetMessages() []*Message {
	var m []*Message
	m = append(m, t.Messages...)
	t.Messages = nil
	return m
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
func NewTestMailer() *TestMailer {
	return &TestMailer{
		Mailer: &LogMailer{},
		Wg:     nil,
	}
}

var _ Mailer = (*TestMailer)(nil)
