package mailer

type EmailParams struct {
	Token      string
	Type       string
	RedirectTo string
}

func withDefault(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

// type
// func ProcessTemplateData()
