package apis

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/tkahng/playground/internal/contextstore"
	"github.com/tkahng/playground/internal/tools/mailer"
	"github.com/tkahng/playground/internal/workers"
)

type EmailVerificationInput struct {
	Token string `json:"token" form:"token" query:"token" required:"true"`
}

type EmailVerificationRequestInput struct {
	Email string `json:"email" form:"email" required:"true"`
}

func (api *Api) RequestVerification(ctx context.Context, input *struct{}) (*struct{}, error) {
	jobService := api.App().JobService()
	claims := contextstore.GetContextUserInfo(ctx)
	if claims == nil {
		return nil, huma.Error404NotFound("User not found")
	}
	if claims.User.EmailVerifiedAt != nil {
		return nil, huma.Error404NotFound("Email already verified")
	}
	// user :=
	err := jobService.EnqueueOtpMailJob(ctx, &workers.OtpEmailJobArgs{
		UserID: claims.User.ID,
		Type:   mailer.EmailTypeVerify,
	})

	if err != nil {
		return nil, err
	}
	return nil, nil
}
