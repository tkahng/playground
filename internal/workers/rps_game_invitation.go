package workers

import (
	"context"

	"github.com/tkahng/playground/internal/jobs"
)

type RpsGameInvitationJobArgs struct {
	Email           string
	InvitedByEmail  string
	TokenHash       string
	ConfirmationURL string
}

func (j RpsGameInvitationJobArgs) Kind() string {
	return "rps_game_invitation_mail"
}

type RpsGameInvitationJobWorker jobs.Worker[RpsGameInvitationJobArgs]

type RpsGameInvitationWorker struct {
	mail OtpMailServiceInterface
}

// Work implements jobs.Worker.
func (t *RpsGameInvitationWorker) Work(ctx context.Context, job *jobs.Job[RpsGameInvitationJobArgs]) error {
	return t.mail.SendRpsGameInvitationEmail(ctx, &job.Args)
}

func NewRpsGameInvitationWorker(otpMailService OtpMailServiceInterface) jobs.Worker[RpsGameInvitationJobArgs] {
	return &RpsGameInvitationWorker{
		mail: otpMailService,
	}
}
