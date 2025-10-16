package test

import (
	"testing"

	"github.com/tkahng/playground/internal/models"
)

func GetFirstItem[T any](t *testing.T, items []T) T {
	if len(items) <= 0 {
		t.Errorf("items lengths is zero.")
	}
	return items[0]
}

func CheckSliceLength[M any](t *testing.T, got []*M, expected int) {
	if len(got) != expected {
		t.Errorf("%s: check slice length got = %d, want %d", t.Name(), len(got), expected)
	}
}

func CheckUserOrderByName(t *testing.T, got []*models.User) {
	for i := 1; i < len(got)-1; i++ {
		firstName, secondName := *got[i].Name, *got[i+1].Name
		if firstName > secondName {
			t.Errorf("users are not in order. first name %s > second name %s", firstName, secondName)
		}
	}
}
func CheckUserAccountOrderByName(t *testing.T, got []*models.UserAccount) {
	for i := 1; i < len(got)-1; i++ {
		firstName, secondName := got[i].Provider.String(), got[i+1].Provider.String()
		if firstName > secondName {
			t.Errorf("users are not in order. first name %s > second name %s", firstName, secondName)
		}
	}
}
