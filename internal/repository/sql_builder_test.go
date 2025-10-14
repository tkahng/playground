package repository_test

import (
	"fmt"
	"testing"

	"github.com/tkahng/playground/internal/repository"
)

func TestNewSQLBuilder_Success(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		builder := repository.NewSQLBuilder[A]()
		if builder == nil {
			t.Errorf("expected builder to not be nil")
		}
	})

	testCases := []struct {
		desc   string
		fn     func()
		errMsg string
	}{
		{
			desc:   "fail - no info field",
			errMsg: "first field must be info field",
			fn: func() {
				type a struct {
					Name string
				}
				_ = repository.NewSQLBuilder[a]()
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("expected panic")
				} else {
					errMsg := fmt.Sprintf("%s", r)
					if errMsg != tC.errMsg {
						t.Errorf("expected error message to be '%s', got %s", tC.errMsg, errMsg)
					}
				}
			}()
			tC.fn()
		})
	}

}

func TestNewSQLBuilder_Fail(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		desc   string
		fn     func()
		errMsg string
	}{
		{
			desc:   "fail - no info field",
			errMsg: "first field must be info field",
			fn: func() {
				type a struct {
					Name string
				}
				_ = repository.NewSQLBuilder[a]()
			},
		},
		{
			desc:   "fail - no info tag",
			errMsg: "db info value not set",
			fn: func() {
				type a struct {
					_    struct{}
					Name string
				}
				_ = repository.NewSQLBuilder[a]()
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("expected panic")
				} else {
					errMsg := fmt.Sprintf("%s", r)
					if errMsg != tC.errMsg {
						t.Errorf("expected error message to be '%s', got %s", tC.errMsg, errMsg)
					}
				}
			}()
			tC.fn()
		})
	}
}
