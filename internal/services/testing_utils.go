package services

import (
	"testing"

	"github.com/tkahng/playground/internal/models"
)

func TestHelperCreateTeamCustomer(t *testing.T, srv PaymentService, team *models.Team, user *models.User) *models.StripeCustomer {
	cus, err := srv.CreateTeamCustomer(t.Context(), team, user)
	if err != nil {
		t.Fatal(err)
	}
	return cus
}
