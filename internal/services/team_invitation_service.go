package services

import (
	"context"
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/conf"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/stores"
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
	CreateInvitation(
		ctx context.Context,
		teamId uuid.UUID,
		invitingUserId uuid.UUID,
		inviteeEmail string,
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

	sendInvitationEmail(
		ctx context.Context,
		params *TeamInvitationMailParams,
	) error
}

var _ TeamInvitationService = (*InvitationService)(nil)

type InvitationService struct {
	routine RoutineService
	mailer  MailService
	// store    TeamInvitationStore
	adapter  stores.StorageAdapterInterface
	settings conf.AppOptions
}

func (i *InvitationService) CreateConfirmationUrl(tokenhash string) (string, error) {
	path, err := mailer.GetPathParams(
		"/team-invitation",
		tokenhash,
		string(models.TokenTypesInviteToken),
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

// sendInvitationEmail implements TeamInvitationService.
func (i *InvitationService) sendInvitationEmail(ctx context.Context, params *TeamInvitationMailParams) error {
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
	body := mailer.GenerateBody("body", string(mailer.DefaultTeamInviteMail), params)
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

func NewInvitationService(
	adapter stores.StorageAdapterInterface,
	mailer MailService,
	settings conf.AppOptions,
	workerService RoutineService,
) TeamInvitationService {
	return &InvitationService{
		routine:  workerService,
		mailer:   mailer,
		settings: settings,
		adapter:  adapter,
	}
}
func (i *InvitationService) CancelInvitation(
	ctx context.Context,
	teamId uuid.UUID,
	userId uuid.UUID,
	invitationId uuid.UUID,
) error {
	member, err := i.adapter.TeamMember().FindTeamMember(ctx, &stores.TeamMemberFilter{
		TeamIds: []uuid.UUID{teamId},
		UserIds: []uuid.UUID{userId},
	})
	if err != nil {
		return err
	}
	if member == nil {
		return fmt.Errorf("user is not a member of the team")
	}
	if member.Role != models.TeamMemberRoleOwner {
		return fmt.Errorf("user is not an owner of the team")
	}
	invitation, err := i.adapter.TeamInvitation().FindInvitationByID(ctx, invitationId)
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

	return i.adapter.TeamInvitation().UpdateInvitation(ctx, invitation)
}

// CheckValidInvitation implements TeamInvitationService.
func (i *InvitationService) CheckValidInvitation(
	ctx context.Context,
	userId uuid.UUID,
	invitationToken string,
) (bool, error) {
	invite, err := i.adapter.TeamInvitation().FindInvitationByToken(ctx, invitationToken)
	if err != nil {
		return false, err
	}
	if invite == nil {
		return false, fmt.Errorf("invitation not found")
	}
	user, err := i.adapter.User().FindUserByID(ctx, userId)
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
	invite, err := i.adapter.TeamInvitation().FindInvitationByToken(ctx, invitationToken)
	if err != nil {
		return err
	}
	if invite == nil {
		return fmt.Errorf("invitation not found")
	}
	user, err := i.adapter.User().FindUserByID(ctx, userId)
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
	_, err = i.adapter.TeamMember().CreateTeamMember(ctx, invite.TeamID, user.ID, invite.Role, false)
	if err != nil {
		return err
	}
	err = i.adapter.TeamInvitation().UpdateInvitation(ctx, invite)
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

	member, err := i.adapter.TeamMember().FindTeamMember(ctx, &stores.TeamMemberFilter{
		TeamIds: []uuid.UUID{teamId},
		UserIds: []uuid.UUID{invitingUserId},
	})
	if err != nil {
		return err
	}
	if member == nil {
		return fmt.Errorf("user is not a member of the team")
	}
	if user, err := i.adapter.User().FindUserByID(ctx, invitingUserId); err != nil {
		return err
	} else if user == nil {
		return fmt.Errorf("user not found")
	} else {
		member.User = user
	}
	if team, err := i.adapter.TeamGroup().FindTeamByID(ctx, teamId); err != nil {
		return err
	} else if team == nil {
		return fmt.Errorf("team not found")
	} else {
		member.Team = team
	}
	invitation := new(models.TeamInvitation)
	existingInvite, err := i.adapter.TeamInvitation().FindPendingInvitation(ctx, teamId, inviteeEmail)
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
		err = i.adapter.TeamInvitation().CreateInvitation(ctx, invitation)
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
		err = i.adapter.TeamInvitation().UpdateInvitation(ctx, existingInvite)
		if err != nil {
			return err
		}
		invitation = existingInvite
	}

	i.routine.FireAndForget(
		func() {
			ctx := context.Background()

			err := i.sendInvitationEmail(ctx, &TeamInvitationMailParams{
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
	invitations, err := i.adapter.TeamInvitation().FindTeamInvitations(ctx, &stores.TeamInvitationFilter{
		TeamIds: []uuid.UUID{teamId},
	})
	if err != nil {
		return nil, err
	}
	return invitations, nil
}

// RejectInvitation implements TeamInvitationService.
func (i *InvitationService) RejectInvitation(ctx context.Context, userId uuid.UUID, invitationToken string) error {
	invite, err := i.adapter.TeamInvitation().FindInvitationByToken(ctx, invitationToken)
	if err != nil {
		return err
	}
	if invite == nil {
		return fmt.Errorf("invitation not found")
	}
	user, err := i.adapter.User().FindUserByID(ctx, userId)
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
