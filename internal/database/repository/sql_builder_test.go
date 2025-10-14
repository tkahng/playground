package repository_test

import (
	"fmt"
	"testing"

	"github.com/tkahng/playground/internal/database/repository"
)

func TestNewSQLBuilder_Success(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		builder := repository.NewSQLBuilder[A]()
		if builder == nil {
			t.Errorf("expected builder to not be nil")
		}
	})

}

func TestNewSQLBuilder_Fail(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		desc   string
		fn     func()
		errMsg string
	}{
		{
			desc:   "no info field",
			errMsg: "first field must be info field",
			fn: func() {
				type a struct {
					Name string
				}
				_ = repository.NewSQLBuilder[a]()
			},
		},
		{
			desc:   "no info tag",
			errMsg: "db info value not set",
			fn: func() {
				type a struct {
					_    struct{}
					Name string
				}
				_ = repository.NewSQLBuilder[a]()
			},
		},
		{
			desc:   "src not set at struct field",
			errMsg: "src not set at struct field name name of a",
			fn: func() {
				type a struct {
					_    struct{} `db:"a"`
					Name string   `db:"name" table:"b"`
				}
				_ = repository.NewSQLBuilder[a]()
			},
		},
		{
			desc:   "dest not set at struct field",
			errMsg: "dest not set at struct field name name of a",
			fn: func() {
				type a struct {
					_    struct{} `db:"a"`
					Name string   `db:"name" table:"b" src:"id"`
				}
				_ = repository.NewSQLBuilder[a]()
			},
		},
		{
			desc:   "through_dest not set at struct field",
			errMsg: "through_dest not set at struct field name name of a",
			fn: func() {
				type a struct {
					_    struct{} `db:"a"`
					Name string   `db:"name" table:"b" src:"id" dest:"id" through:"ab"`
				}
				_ = repository.NewSQLBuilder[a]()
			},
		},
		{
			desc:   "through_src not set at struct field",
			errMsg: "through_src not set at struct field name name of a",
			fn: func() {
				type a struct {
					_    struct{} `db:"a"`
					Name string   `db:"name" table:"b" src:"id" dest:"id" through:"ab" through_dest:"id"`
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
