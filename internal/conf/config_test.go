//go:build !integration

package conf_test

import (
	"os"
	"testing"

	"github.com/tkahng/playground/internal/conf"
)

func TestDBConfig_GetUrl(t *testing.T) {
	tests := []struct {
		name   string // description of this test case
		setEnv func()
		want   string
	}{
		{
			name: "get default url succeeded",
			want: "postgres://postgres:postgres@localhost:5432/playground?sslmode=disable",
			setEnv: func() {

			},
		},
		{
			name: "get url from env succeeded",
			want: "postgres://user:password@somehost:5432/somedatabase?sslmode=disable",
			setEnv: func() {
				t.Setenv("DATABASE_USER", "user")
				t.Setenv("DATABASE_PASSWORD", "password")
				t.Setenv("DATABASE_HOST", "somehost")
				t.Setenv("DATABASE_DB", "somedatabase")
			},
		},
		{
			name: "get test db url from default",
			want: "postgres://postgres:postgres@localhost:5432/playground_test?sslmode=disable",
			setEnv: func() {
				t.Setenv("DATABASE_DB", "playground_test")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				os.Clearenv()
			}()
			tt.setEnv()
			c := conf.GetConfig[conf.DBConfig]()
			got := c.GetDatabaseURL()
			if got != tt.want {
				t.Errorf("GetUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDBConfig_DatabaseURL(t *testing.T) {
	tests := []struct {
		name   string // description of this test case
		setEnv func()
		want   string
	}{
		{
			name: "get default url succeeded",
			want: "postgres://postgres:postgres@localhost:5432/playground?sslmode=disable",
			setEnv: func() {

			},
		},
		{
			name: "get url from env succeeded",
			want: "postgres://user:password@somehost:5432/somedatabase?sslmode=disable",
			setEnv: func() {
				t.Setenv("DATABASE_USER", "user")
				t.Setenv("DATABASE_PASSWORD", "password")
				t.Setenv("DATABASE_HOST", "somehost")
				t.Setenv("DATABASE_DB", "somedatabase")
			},
		},
		{
			name: "get test db url from default",
			want: "postgres://postgres:postgres@localhost:5432/playground_test?sslmode=disable",
			setEnv: func() {
				t.Setenv("DATABASE_DB", "playground_test")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				os.Clearenv()
			}()
			tt.setEnv()
			c := conf.GetConfig[conf.DBConfig]()
			got := c.GetDatabaseURL()
			if got != tt.want {
				t.Errorf("GetUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}
