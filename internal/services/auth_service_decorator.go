package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/auth/oauth"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
)

type AuthServiceDecorator struct {
	delegate                       BaseAuthService
	MailFunc                       func() MailService
	PasswordFunc                   func() PasswordService
	StoreFunc                      func() AuthStore
	TokenFunc                      func() JwtService
	CreateOAuthUrlFunc             func(ctx context.Context, provider shared.Providers, redirectUrl string) (string, error)
	AuthenticateFunc               func(ctx context.Context, params *shared.AuthenticationInput) (*models.User, error)
	CheckResetPasswordTokenFunc    func(ctx context.Context, token string) error
	HandleAccessTokenFunc          func(ctx context.Context, token string) (*shared.UserInfo, error)
	HandleRefreshTokenFunc         func(ctx context.Context, token string) (*shared.UserInfoTokens, error)
	HandlePasswordResetRequestFunc func(ctx context.Context, email string) error
	HandlePasswordResetTokenFunc   func(ctx context.Context, token string, password string) error
	HandleVerificationTokenFunc    func(ctx context.Context, token string) error
	SignoutFunc                    func(ctx context.Context, token string) error
	ResetPasswordFunc              func(ctx context.Context, userId uuid.UUID, oldPassword string, newPassword string) error
	SendOtpEmailFunc               func(emailType EmailType, ctx context.Context, user *models.User) error
	VerifyAndParseOtpTokenFunc     func(ctx context.Context, emailType EmailType, token string) (*shared.OtpClaims, error)
	VerifyStateTokenFunc           func(ctx context.Context, token string) (*shared.ProviderStateClaims, error)
	CreateAndPersistStateTokenFunc func(ctx context.Context, payload *shared.ProviderStatePayload) (string, error)
	CreateAuthTokensFromEmailFunc  func(ctx context.Context, email string) (*shared.UserInfoTokens, error)
	FetchAuthUserFunc              func(ctx context.Context, code string, parsedState *shared.ProviderStateClaims) (*oauth.AuthUser, error)
	FireAndForgetFunc              func(f func())
}

// Mail implements AuthService.
func (a *AuthServiceDecorator) Mail() MailService {
	if a.MailFunc != nil {
		return a.MailFunc()
	}
	return a.delegate.Mail()
}

// Password implements AuthService.
func (a *AuthServiceDecorator) Password() PasswordService {
	if a.PasswordFunc != nil {
		return a.PasswordFunc()
	}
	return a.delegate.Password()
}

// Store implements AuthService.
func (a *AuthServiceDecorator) Store() AuthStore {
	if a.StoreFunc != nil {
		return a.StoreFunc()
	}
	return a.delegate.Store()
}

// Token implements AuthService.
func (a *AuthServiceDecorator) Token() JwtService {
	if a.TokenFunc != nil {
		return a.TokenFunc()
	}
	return a.delegate.Token()
}

// CreateOAuthUrl implements AuthService.
func (a *AuthServiceDecorator) CreateOAuthUrl(ctx context.Context, provider shared.Providers, redirectUrl string) (string, error) {
	if a.CreateOAuthUrlFunc != nil {
		return a.CreateOAuthUrlFunc(ctx, provider, redirectUrl)
	}
	return a.delegate.CreateOAuthUrl(ctx, provider, redirectUrl)
}

// Authenticate implements AuthService.
func (a *AuthServiceDecorator) Authenticate(ctx context.Context, params *shared.AuthenticationInput) (*models.User, error) {
	if a.AuthenticateFunc != nil {
		return a.AuthenticateFunc(ctx, params)
	}
	return a.delegate.Authenticate(ctx, params)
}

// CheckResetPasswordToken implements AuthService.
func (a *AuthServiceDecorator) CheckResetPasswordToken(ctx context.Context, token string) error {
	if a.CheckResetPasswordTokenFunc != nil {
		return a.CheckResetPasswordTokenFunc(ctx, token)
	}
	return a.delegate.CheckResetPasswordToken(ctx, token)
}

// CreateAndPersistStateToken implements AuthService.
func (a *AuthServiceDecorator) CreateAndPersistStateToken(ctx context.Context, payload *shared.ProviderStatePayload) (string, error) {
	if a.CreateAndPersistStateTokenFunc != nil {
		return a.CreateAndPersistStateTokenFunc(ctx, payload)
	}
	return a.delegate.CreateAndPersistStateToken(ctx, payload)
}

// CreateAuthTokensFromEmail implements AuthService.
func (a *AuthServiceDecorator) CreateAuthTokensFromEmail(ctx context.Context, email string) (*shared.UserInfoTokens, error) {
	if a.CreateAuthTokensFromEmailFunc != nil {
		return a.CreateAuthTokensFromEmailFunc(ctx, email)
	}
	return a.delegate.CreateAuthTokensFromEmail(ctx, email)
}

// FetchAuthUser implements AuthService.
func (a *AuthServiceDecorator) FetchAuthUser(ctx context.Context, code string, parsedState *shared.ProviderStateClaims) (*oauth.AuthUser, error) {
	if a.FetchAuthUserFunc != nil {
		return a.FetchAuthUserFunc(ctx, code, parsedState)
	}
	return a.delegate.FetchAuthUser(ctx, code, parsedState)
}

// FireAndForget implements AuthService.
func (a *AuthServiceDecorator) FireAndForget(f func()) {
	if a.FireAndForgetFunc != nil {
		a.FireAndForgetFunc(f)
		return
	}
	a.delegate.routine.FireAndForget(f)
}

// HandleAccessToken implements AuthService.
func (a *AuthServiceDecorator) HandleAccessToken(ctx context.Context, token string) (*shared.UserInfo, error) {
	if a.HandleAccessTokenFunc != nil {
		return a.HandleAccessTokenFunc(ctx, token)
	}
	return a.delegate.HandleAccessToken(ctx, token)
}

// HandlePasswordResetRequest implements AuthService.
func (a *AuthServiceDecorator) HandlePasswordResetRequest(ctx context.Context, email string) error {
	if a.HandlePasswordResetRequestFunc != nil {
		return a.HandlePasswordResetRequestFunc(ctx, email)
	}
	return a.delegate.HandlePasswordResetRequest(ctx, email)
}

// HandlePasswordResetToken implements AuthService.
func (a *AuthServiceDecorator) HandlePasswordResetToken(ctx context.Context, token string, password string) error {
	if a.HandlePasswordResetTokenFunc != nil {
		return a.HandlePasswordResetTokenFunc(ctx, token, password)
	}
	return a.delegate.HandlePasswordResetToken(ctx, token, password)
}

// HandleRefreshToken implements AuthService.
func (a *AuthServiceDecorator) HandleRefreshToken(ctx context.Context, token string) (*shared.UserInfoTokens, error) {
	if a.HandleRefreshTokenFunc != nil {
		return a.HandleRefreshTokenFunc(ctx, token)
	}
	return a.delegate.HandleRefreshToken(ctx, token)
}

// HandleVerificationToken implements AuthService.
func (a *AuthServiceDecorator) HandleVerificationToken(ctx context.Context, token string) error {
	if a.HandleVerificationTokenFunc != nil {
		return a.HandleVerificationTokenFunc(ctx, token)
	}
	return a.delegate.HandleVerificationToken(ctx, token)
}

// ResetPassword implements AuthService.
func (a *AuthServiceDecorator) ResetPassword(ctx context.Context, userId uuid.UUID, oldPassword string, newPassword string) error {
	if a.ResetPasswordFunc != nil {
		return a.ResetPasswordFunc(ctx, userId, oldPassword, newPassword)
	}
	return a.delegate.ResetPassword(ctx, userId, oldPassword, newPassword)
}

// SendOtpEmail implements AuthService.
func (a *AuthServiceDecorator) SendOtpEmail(emailType EmailType, ctx context.Context, user *models.User) error {
	if a.SendOtpEmailFunc != nil {
		return a.SendOtpEmailFunc(emailType, ctx, user)
	}
	return a.delegate.SendOtpEmail(emailType, ctx, user)
}

// Signout implements AuthService.
func (a *AuthServiceDecorator) Signout(ctx context.Context, token string) error {
	if a.SignoutFunc != nil {
		return a.SignoutFunc(ctx, token)
	}
	return a.delegate.Signout(ctx, token)
}

// VerifyAndParseOtpToken implements AuthService.
func (a *AuthServiceDecorator) VerifyAndParseOtpToken(ctx context.Context, emailType EmailType, token string) (*shared.OtpClaims, error) {
	if a.VerifyAndParseOtpTokenFunc != nil {
		return a.VerifyAndParseOtpTokenFunc(ctx, emailType, token)
	}
	return a.delegate.VerifyAndParseOtpToken(ctx, emailType, token)
}

// VerifyStateToken implements AuthService.
func (a *AuthServiceDecorator) VerifyStateToken(ctx context.Context, token string) (*shared.ProviderStateClaims, error) {
	if a.VerifyStateTokenFunc != nil {
		return a.VerifyStateTokenFunc(ctx, token)
	}
	return a.delegate.VerifyStateToken(ctx, token)
}

var _ AuthService = (*AuthServiceDecorator)(nil)

type AuthStoreDecorator struct {
	delegate                               AuthStore
	AssignUserRolesFunc                    func(ctx context.Context, userId uuid.UUID, roleNames ...string) error
	CreateUserFunc                         func(ctx context.Context, user *models.User) (*models.User, error)
	DeleteTokenFunc                        func(ctx context.Context, token string) error
	DeleteUserFunc                         func(ctx context.Context, id uuid.UUID) error
	FindUserAccountByUserIdAndProviderFunc func(ctx context.Context, userId uuid.UUID, provider models.Providers) (*models.UserAccount, error)
	FindUserByEmailFunc                    func(ctx context.Context, email string) (*models.User, error)
	GetTokenFunc                           func(ctx context.Context, token string) (*models.Token, error)
	GetUserInfoFunc                        func(ctx context.Context, email string) (*shared.UserInfo, error)
	LinkAccountFunc                        func(ctx context.Context, account *models.UserAccount) error
	SaveTokenFunc                          func(ctx context.Context, token *shared.CreateTokenDTO) error
	UnlinkAccountFunc                      func(ctx context.Context, userId uuid.UUID, provider models.Providers) error
	UpdateUserFunc                         func(ctx context.Context, user *models.User) error
	UpdateUserAccountFunc                  func(ctx context.Context, account *models.UserAccount) error
}

// AssignUserRoles implements AuthStore.
func (a *AuthStoreDecorator) AssignUserRoles(ctx context.Context, userId uuid.UUID, roleNames ...string) error {
	if a.AssignUserRolesFunc != nil {
		return a.AssignUserRolesFunc(ctx, userId, roleNames...)
	}
	return a.delegate.AssignUserRoles(ctx, userId, roleNames...)
}

// CreateUser implements AuthStore.
func (a *AuthStoreDecorator) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	if a.CreateUserFunc != nil {
		return a.CreateUserFunc(ctx, user)
	}
	return a.delegate.CreateUser(ctx, user)
}

// DeleteToken implements AuthStore.
func (a *AuthStoreDecorator) DeleteToken(ctx context.Context, token string) error {
	if a.DeleteTokenFunc != nil {
		return a.DeleteTokenFunc(ctx, token)
	}
	return a.delegate.DeleteToken(ctx, token)
}

// DeleteUser implements AuthStore.
func (a *AuthStoreDecorator) DeleteUser(ctx context.Context, id uuid.UUID) error {
	if a.DeleteUserFunc != nil {
		return a.DeleteUserFunc(ctx, id)
	}
	return a.delegate.DeleteUser(ctx, id)
}

// FindUserAccountByUserIdAndProvider implements AuthStore.
func (a *AuthStoreDecorator) FindUserAccountByUserIdAndProvider(ctx context.Context, userId uuid.UUID, provider models.Providers) (*models.UserAccount, error) {
	if a.FindUserAccountByUserIdAndProviderFunc != nil {
		return a.FindUserAccountByUserIdAndProviderFunc(ctx, userId, provider)
	}
	return a.delegate.FindUserAccountByUserIdAndProvider(ctx, userId, provider)
}

// FindUserByEmail implements AuthStore.
func (a *AuthStoreDecorator) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	panic("unimplemented")
}

// GetToken implements AuthStore.
func (a *AuthStoreDecorator) GetToken(ctx context.Context, token string) (*models.Token, error) {
	panic("unimplemented")
}

// GetUserInfo implements AuthStore.
func (a *AuthStoreDecorator) GetUserInfo(ctx context.Context, email string) (*shared.UserInfo, error) {
	panic("unimplemented")
}

// LinkAccount implements AuthStore.
func (a *AuthStoreDecorator) LinkAccount(ctx context.Context, account *models.UserAccount) error {
	panic("unimplemented")
}

// SaveToken implements AuthStore.
func (a *AuthStoreDecorator) SaveToken(ctx context.Context, token *shared.CreateTokenDTO) error {
	panic("unimplemented")
}

// UnlinkAccount implements AuthStore.
func (a *AuthStoreDecorator) UnlinkAccount(ctx context.Context, userId uuid.UUID, provider models.Providers) error {
	panic("unimplemented")
}

// UpdateUser implements AuthStore.
func (a *AuthStoreDecorator) UpdateUser(ctx context.Context, user *models.User) error {
	panic("unimplemented")
}

// UpdateUserAccount implements AuthStore.
func (a *AuthStoreDecorator) UpdateUserAccount(ctx context.Context, account *models.UserAccount) error {
	panic("unimplemented")
}

var _ AuthStore = (*AuthStoreDecorator)(nil)
