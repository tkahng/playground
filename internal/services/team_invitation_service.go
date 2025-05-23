package services

import (
	"context"
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mailer"
	"github.com/tkahng/authgo/internal/tools/security"
)

type TeamInvitationMailParams struct {
	Email           string
	InvitedByEmail  string
	TeamName        string
	TokenHash       string
	ConfirmationURL string
}

type TeamInvitationService interface {
	WorkerService
	CreateInvitation(
		ctx context.Context,
		teamId uuid.UUID,
		userId uuid.UUID,
		email string,
		role models.TeamMemberRole,
		resend bool,
	) error
	CheckValidInvitation(
		ctx context.Context,
		userId uuid.UUID,
		invitationToken string,
	) (bool, error)
	AcceptInvitation(
		ctx context.Context,
		userId uuid.UUID,
		invitationToken string,
	) error
	RejectInvitation(
		ctx context.Context,
		userId uuid.UUID,
		invitationToken string,
	) error

	CancelInvitation(
		ctx context.Context,
		teamId uuid.UUID,
		userId uuid.UUID,
		invitationId uuid.UUID,
	) error

	FindInvitations(
		ctx context.Context,
		teamId uuid.UUID,
	) ([]*models.TeamInvitation, error)

	SendInvitationEmail(
		ctx context.Context,
		params *TeamInvitationMailParams,
	) error
}

type TeamInvitationStore interface {
	FindTeamByID(ctx context.Context, teamId uuid.UUID) (*models.Team, error)
	CreateTeamMember(ctx context.Context, teamId, userId uuid.UUID, role models.TeamMemberRole, hasBillingAccess bool) (*models.TeamMember, error)
	DeleteTeamMember(ctx context.Context, teamId, userId uuid.UUID) error
	FindTeamMemberByTeamAndUserId(
		ctx context.Context,
		teamId uuid.UUID,
		userId uuid.UUID,
	) (*models.TeamMember, error)
	FindUserByID(ctx context.Context, userId uuid.UUID) (*models.User, error)
	CreateInvitation(
		ctx context.Context,
		invitation *models.TeamInvitation,
	) error
	UpdateInvitation(
		ctx context.Context,
		invitation *models.TeamInvitation,
	) error
	FindInvitationByToken(
		ctx context.Context,
		token string,
	) (*models.TeamInvitation, error)
	FindInvitationByID(
		ctx context.Context,
		invitationId uuid.UUID,
	) (*models.TeamInvitation, error)
	FindTeamInvitations(
		ctx context.Context,
		teamId uuid.UUID,
	) ([]*models.TeamInvitation, error)
	FindPendingInvitation(
		ctx context.Context,
		teamId uuid.UUID,
		email string,
	) (*models.TeamInvitation, error)
}

type InvitationService struct {
	WorkerService
	mailer   MailService
	store    TeamInvitationStore
	settings conf.AppOptions
}

func (i *InvitationService) CreateConfirmationUrl(tokenhash string) (string, error) {
	path, err := mailer.GetPathParams(
		"/team-invitation",
		tokenhash,
		string(shared.TokenTypesInviteToken),
		i.settings.Meta.AppUrl,
	)
	if err != nil {
		return "", err
	}
	appUrl, err := url.Parse(i.settings.Meta.AppUrl)
	if err != nil {
		return "", err
	}
	return appUrl.ResolveReference(path).String(), nil
}

// SendInvitationEmail implements TeamInvitationService.
func (i *InvitationService) SendInvitationEmail(ctx context.Context, params *TeamInvitationMailParams) error {
	if params == nil {
		return fmt.Errorf("params is nil")
	}
	if params.Email == "" {
		return fmt.Errorf("email is empty")
	}
	if params.TeamName == "" {
		return fmt.Errorf("team name is empty")
	}

	confUrl, err := i.CreateConfirmationUrl(params.TokenHash)
	if err != nil {
		return err
	}
	params.ConfirmationURL = confUrl
	body := mailer.GetTemplate("body", string(mailer.DefaultTeamInviteMail), params)
	param := &mailer.AllEmailParams{}
	param.CommonParams = &mailer.CommonParams{
		ConfirmationURL: params.ConfirmationURL,
		Email:           params.Email,
		SiteURL:         i.settings.Meta.AppUrl,
		Token:           params.TokenHash,
	}
	param.Message = &mailer.Message{
		To:      params.Email,
		Subject: fmt.Sprintf("Invitation to join %s", params.TeamName),
		Body:    body,
	}
	return i.mailer.SendMail(param)
}

var _ TeamInvitationService = (*InvitationService)(nil)

func NewInvitationService(
	store TeamInvitationStore,
	mailer MailService,
	settings conf.AppOptions,
	workerService WorkerService,
) TeamInvitationService {
	return &InvitationService{
		WorkerService: workerService,
		store:         store,
		mailer:        mailer,
		settings:      settings,
	}
}
func (i *InvitationService) CancelInvitation(
	ctx context.Context,
	teamId uuid.UUID,
	userId uuid.UUID,
	invitationId uuid.UUID,
) error {
	member, err := i.store.FindTeamMemberByTeamAndUserId(ctx, teamId, userId)
	if err != nil {
		return err
	}
	if member == nil {
		return fmt.Errorf("user is not a member of the team")
	}
	if member.Role != models.TeamMemberRoleOwner {
		return fmt.Errorf("user is not an owner of the team")
	}
	invitation, err := i.store.FindInvitationByID(ctx, invitationId)
	if err != nil {
		return err
	}
	if invitation == nil {
		return fmt.Errorf("invitation not found")
	}
	if invitation.TeamID != teamId {
		return fmt.Errorf("invitation does not match team")
	}
	invitation.Status = models.TeamInvitationStatusCanceled

	return i.store.UpdateInvitation(ctx, invitation)
}

// CheckValidInvitation implements TeamInvitationService.
func (i *InvitationService) CheckValidInvitation(
	ctx context.Context,
	userId uuid.UUID,
	invitationToken string,
) (bool, error) {
	invite, err := i.store.FindInvitationByToken(ctx, invitationToken)
	if err != nil {
		return false, err
	}
	if invite == nil {
		return false, fmt.Errorf("invitation not found")
	}
	user, err := i.store.FindUserByID(ctx, userId)
	if err != nil {
		return false, err
	}
	if user == nil {
		return false, fmt.Errorf("user not found")
	}
	if invite.Email != user.Email {
		return false, fmt.Errorf("user does not match invitation")
	}
	if invite.Status != models.TeamInvitationStatusPending {
		return false, fmt.Errorf("invitation is not pending")
	}
	return true, nil
}

// AcceptInvitation implements TeamInvitationService.
func (i *InvitationService) AcceptInvitation(
	ctx context.Context,
	userId uuid.UUID,
	invitationToken string,
) error {
	invite, err := i.store.FindInvitationByToken(ctx, invitationToken)
	if err != nil {
		return err
	}
	if invite == nil {
		return fmt.Errorf("invitation not found")
	}
	user, err := i.store.FindUserByID(ctx, userId)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}
	if invite.Email != user.Email {
		return fmt.Errorf("user does not match invitation")
	}
	if invite.Status != models.TeamInvitationStatusPending {
		return fmt.Errorf("invitation is not pending")
	}
	invite.Status = models.TeamInvitationStatusAccepted
	_, err = i.store.CreateTeamMember(ctx, invite.TeamID, user.ID, invite.Role, false)
	if err != nil {
		return err
	}
	err = i.store.UpdateInvitation(ctx, invite)
	if err != nil {
		return err
	}
	return nil
}

// CreateInvitation implements TeamInvitationService.
func (i *InvitationService) CreateInvitation(
	ctx context.Context,
	teamId uuid.UUID,
	invitingUserId uuid.UUID,
	inviteeEmail string,
	role models.TeamMemberRole,
	resend bool,
) error {

	member, err := i.store.FindTeamMemberByTeamAndUserId(ctx, teamId, invitingUserId)
	if err != nil {
		return err
	}
	if member == nil {
		return fmt.Errorf("user is not a member of the team")
	}
	if user, err := i.store.FindUserByID(ctx, invitingUserId); err != nil {
		return err
	} else if user == nil {
		return fmt.Errorf("user not found")
	} else {
		member.User = user
	}
	if team, err := i.store.FindTeamByID(ctx, teamId); err != nil {
		return err
	} else if team == nil {
		return fmt.Errorf("team not found")
	} else {
		member.Team = team
	}
	invitation := new(models.TeamInvitation)
	existingInvite, err := i.store.FindPendingInvitation(ctx, teamId, inviteeEmail)
	if err != nil {
		return err
	}
	if existingInvite == nil {
		token := security.GenerateTokenKey()
		invitation.Status = models.TeamInvitationStatusPending
		invitation.Token = token
		invitation.Email = inviteeEmail
		invitation.Role = role
		invitation.TeamID = teamId
		invitation.InviterMemberID = member.ID
		invitation.ExpiresAt = i.settings.Auth.InviteToken.Expires()
		err = i.store.CreateInvitation(ctx, invitation)
		if err != nil {
			return err
		}
	} else {
		if !resend {
			return fmt.Errorf("invitation already exists")
		}
		existingInvite.Status = models.TeamInvitationStatusPending
		existingInvite.Role = role
		existingInvite.ExpiresAt = i.settings.Auth.InviteToken.Expires()
		err = i.store.UpdateInvitation(ctx, existingInvite)
		if err != nil {
			return err
		}
		invitation = existingInvite
	}

	i.FireAndForget(
		func() {
			ctx := context.Background()

			err := i.SendInvitationEmail(ctx, &TeamInvitationMailParams{
				Email:          invitation.Email,
				InvitedByEmail: member.User.Email,
				TeamName:       member.Team.Name,
				TokenHash:      invitation.Token,
			})
			if err != nil {
				fmt.Printf("failed to send invitation email: %v", err)
			}
		},
	)
	return nil
}

// FindInvitations implements TeamInvitationService.
func (i *InvitationService) FindInvitations(ctx context.Context, teamId uuid.UUID) ([]*models.TeamInvitation, error) {
	invitations, err := i.store.FindTeamInvitations(ctx, teamId)
	if err != nil {
		return nil, err
	}
	return invitations, nil
}

// RejectInvitation implements TeamInvitationService.
func (i *InvitationService) RejectInvitation(ctx context.Context, userId uuid.UUID, invitationToken string) error {
	invite, err := i.store.FindInvitationByToken(ctx, invitationToken)
	if err != nil {
		return err
	}
	if invite == nil {
		return fmt.Errorf("invitation not found")
	}
	user, err := i.store.FindUserByID(ctx, userId)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}
	if invite.Email != user.Email {
		return fmt.Errorf("user does not match invitation")
	}
	invite.Status = models.TeamInvitationStatusDeclined
	return nil
}
