package apis

import (
	"context"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/contextstore"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
)

type UserAccountOutput struct {
	ID                uuid.UUID            `db:"id,pk" json:"id"`
	UserID            uuid.UUID            `db:"user_id" json:"user_id"`
	Type              models.ProviderTypes `db:"type" json:"type" enum:"oauth,credentials"`
	Provider          models.Providers     `db:"provider" json:"provider" enum:"google,apple,facebook,github,credentials"`
	ProviderAccountID string               `db:"provider_account_id" json:"provider_account_id"`
	CreatedAt         time.Time            `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time            `db:"updated_at" json:"updated_at"`
}

func FromModelUserAccountOutput(u *models.UserAccount) *UserAccountOutput {
	if u == nil {
		return nil
	}
	return &UserAccountOutput{
		ID:                u.ID,
		UserID:            u.UserID,
		Type:              u.Type,
		Provider:          u.Provider,
		ProviderAccountID: u.ProviderAccountID,
		CreatedAt:         u.CreatedAt,
		UpdatedAt:         u.UpdatedAt,
	}
}

type UserWithAccounts struct {
	*ApiUser
	Accounts []*UserAccountOutput `json:"accounts"`
}
type MeOutput struct {
	Body *UserWithAccounts
}

func (api *Api) Me(ctx context.Context, input *struct{}) (*MeOutput, error) {
	claims := contextstore.GetContextUserInfo(ctx)
	if claims == nil {
		return nil, huma.Error404NotFound("User not found")
	}
	user, err := api.app.Adapter().User().FindUserByID(ctx, claims.User.ID)
	if err != nil {
		return nil, err
	}
	accounts, err := api.app.Adapter().UserAccount().GetUserAccounts(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	var acc []*models.UserAccount
	if len(accounts) > 0 {
		acc = accounts[0]
	}

	return &MeOutput{
		Body: &UserWithAccounts{
			ApiUser:  FromUserModel(user),
			Accounts: mapper.Map(acc, FromModelUserAccountOutput),
		},
	}, nil

}

func (api *Api) MeUpdate(ctx context.Context, input *struct {
	Body shared.UpdateMeInput
}) (*struct{}, error) {
	claims := contextstore.GetContextUserInfo(ctx)
	if claims == nil {
		return nil, huma.Error404NotFound("User not found")
	}
	adapter := api.app.Adapter()
	user, err := adapter.User().FindUserByID(ctx, claims.User.ID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, huma.Error404NotFound("User not found")
	}
	user.Name = input.Body.Name
	user.Image = input.Body.Image
	err = adapter.User().UpdateUser(ctx, user)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (api *Api) MeDelete(ctx context.Context, input *struct{}) (*struct{}, error) {
	claims := contextstore.GetContextUserInfo(ctx)
	if claims == nil {
		return nil, huma.Error404NotFound("User not found")
	}
	checker := api.app.Checker()
	ok, err := checker.CannotBeSuperUserID(ctx, claims.User.ID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, huma.Error403Forbidden("You cannot delete the super user")
	}
	// Check if the user has any active subscriptions
	ok, err = checker.CannotHaveValidUserSubscription(ctx, claims.User.ID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, huma.Error403Forbidden("You cannot delete a user with active subscriptions")
	}
	err = api.app.Adapter().User().DeleteUser(ctx, claims.User.ID)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
