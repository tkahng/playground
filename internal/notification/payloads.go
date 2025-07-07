package notification

type NotificationContent struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type NotificationPayload[T NotificationDataKinder] struct {
	Notification NotificationContent `json:"notification" required:"true"`
	Data         T                   `json:"data" required:"true"`
}

type NotificationDataKinder interface {
	Kind() string
}

func NewNotificationPayload[T NotificationDataKinder](title, body string, data T) *NotificationPayload[T] {
	return &NotificationPayload[T]{
		Notification: NotificationContent{
			Title: title,
			Body:  body,
		},
		Data: data,
	}
}

type NotificationPayloadManager interface {
}

func RegisterPayloadKind(manager NotificationPayloadManager, kind string) {
}
