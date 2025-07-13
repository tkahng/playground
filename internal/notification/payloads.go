package notification

type NotificationContent struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type NotificationPayload[T NotificationData] struct {
	Notification NotificationContent `json:"notification" required:"true"`
	Data         T                   `json:"data" required:"true"`
}

type NotificationData interface {
	Kind() string
}

func NewNotificationPayload[T NotificationData](title, body string, data T) *NotificationPayload[T] {
	return &NotificationPayload[T]{
		Notification: NotificationContent{
			Title: title,
			Body:  body,
		},
		Data: data,
	}
}
