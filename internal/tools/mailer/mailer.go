package mailer

type Message struct {
	From    string
	To      string
	Subject string
	Body    string
}

type Mailer interface {
	Send(message *Message) error
}
