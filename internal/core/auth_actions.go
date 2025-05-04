package core

import (
	"context"

	"github.com/google/uuid"

	"github.com/tkahng/authgo/internal/shared"
)

type AuthActions interface {
	HandlePasswordResetRequest(ctx context.Context, email string) error
	HandleAccessToken(ctx context.Context, token string) (*shared.UserInfo, error)
	HandleRefreshToken(ctx context.Context, token string) (*shared.UserInfoTokens, error)
	HandleVerificationToken(ctx context.Context, token string) error
	HandlePasswordResetToken(ctx context.Context, token, password string) error
	CheckResetPasswordToken(ctx context.Context, token string) error
	VerifyStateToken(ctx context.Context, token string) (*ProviderStateClaims, error)
	CreateAndPersistStateToken(ctx context.Context, payload *ProviderStatePayload) (string, error)
	VerifyAndParseOtpToken(ctx context.Context, emailType EmailType, token string) (*OtpClaims, error)
	Authenticate(ctx context.Context, params *shared.AuthenticationInput) (*shared.User, error)
	// CreateAuthTokens(ctx context.Context, payload *shared.UserInfo) (*shared.UserInfoTokens, error)
	CreateAuthTokensFromEmail(ctx context.Context, email string) (*shared.UserInfoTokens, error)
	SendOtpEmail(emailType EmailType, ctx context.Context, user *shared.User) error
	Signout(ctx context.Context, token string) error
	ResetPassword(ctx context.Context, userId uuid.UUID, oldPassword, newPassword string) error

	// CreateAndSaveRefreshToken(ctx context.Context, user *RefreshTokenPayload) (string, error)
	// CreateAndSaveVerificationToken(ctx context.Context, user *OtpPayload) (string, error)
	// CreateAndSavePasswordResetToken(ctx context.Context, user *OtpPayload) (string, error)
	// CreateAndSaveStateToken(ctx context.Context, user *ProviderStatePayload) (string, error)
	// Signin(ctx context.Context, email string, password string) (*shared.AuthenticatedDTO, error)
	// Signup(ctx context.Context, email string, password string) (*shared.AuthenticatedDTO, error)
	// OAuth2Signin(ctx context.Context, code string, state string) (*shared.AuthenticatedDTO, error)
}
