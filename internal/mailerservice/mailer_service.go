package mailservice

import "context"

type Mailerservice interface {
	SendVerificationEmail(ctx context.Context, email string) error
}
