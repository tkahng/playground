package workers

import (
	"context"

	"github.com/tkahng/authgo/internal/jobs"
)

type TeamInvitationJobArgs struct {
	Email           string
	InvitedByEmail  string
	TeamName        string
	TokenHash       string
	ConfirmationURL string
}

func (j TeamInvitationJobArgs) Kind() string {
	return "team_invitation_mail"
}

type TeamInvitationJob struct {
	Args TeamInvitationJobArgs
}

type TeamInvitationWorker struct {
	mail OtpMailServiceInterface
}

// Work implements jobs.Worker.
func (t *TeamInvitationWorker) Work(ctx context.Context, job *jobs.Job[TeamInvitationJobArgs]) error {
	return t.mail.SendTeamInvitationEmail(ctx, &job.Args)
}

func NewTeamInvitationWorker(otpMailService OtpMailServiceInterface) jobs.Worker[TeamInvitationJobArgs] {
	return &TeamInvitationWorker{
		mail: otpMailService,
	}
}

// type TeamInvitationWorkerInterface interface {
// 	SendInvitationEmail(ctx context.Context, params *TeamInvitationJobArgs) error
// }
