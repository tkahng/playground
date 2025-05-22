package services

import (
	"context"
	"fmt"
	"net/url"
	"time"

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
		invitationToken string,
		userId uuid.UUID,
	) (bool, error)
	AcceptInvitation(
		ctx context.Context,
		invitationToken string,
		userId uuid.UUID,
	) error
	RejectInvitation(
		ctx context.Context,
		invitationToken string,
		userId uuid.UUID,
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
	if params.ConfirmationURL == "" {
		return fmt.Errorf("confirmation URL is empty")
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

func NewInvitationService(store TeamInvitationStore, mailer MailService, settings conf.AppOptions) TeamInvitationService {
	return &InvitationService{
		store:    store,
		mailer:   mailer,
		settings: settings,
	}
}

// CheckValidInvitation implements TeamInvitationService.
func (i *InvitationService) CheckValidInvitation(ctx context.Context, invitationToken string, userId uuid.UUID) (bool, error) {
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
func (i *InvitationService) AcceptInvitation(ctx context.Context, invitationToken string, userId uuid.UUID) error {
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
	userId uuid.UUID,
	email string,
	role models.TeamMemberRole,
	resend bool,
) error {
	token := security.GenerateTokenKey()

	var invitation = &models.TeamInvitation{
		TeamID: teamId,
		Email:  email,
		Role:   role,
		Token:  token,
		Status: models.TeamInvitationStatusPending,
	}
	member, err := i.store.FindTeamMemberByTeamAndUserId(ctx, teamId, userId)
	if err != nil {
		return err
	}
	if member == nil {
		return fmt.Errorf("user is not a member of the team")
	}
	invitation.InviterMemberID = member.ID
	invitation.ExpiresAt = time.Now().Add(24 * time.Hour)
	invitation.CreatedAt = time.Now()
	invitation.UpdatedAt = time.Now()

	return i.store.CreateInvitation(ctx, invitation)
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
func (i *InvitationService) RejectInvitation(ctx context.Context, invitationToken string, userId uuid.UUID) error {
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
