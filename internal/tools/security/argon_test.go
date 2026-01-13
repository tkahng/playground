package security

import (
	"testing"

	"github.com/alexedwards/argon2id"
)

func TestCreateHash(t *testing.T) {
	type args struct {
		password string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "1",
			args: args{
				password: "password",
			},
			wantErr: false,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CreateHash(tt.args.password, argon2id.DefaultParams)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestComparePasswordAndHash(t *testing.T) {
	type args struct {
		password string
		hash     string
	}
	tests := []struct {
		name      string
		args      args
		wantMatch bool
		wantErr   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMatch, err := ComparePasswordAndHash(tt.args.password, tt.args.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("ComparePasswordAndHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotMatch != tt.wantMatch {
				t.Errorf("ComparePasswordAndHash() = %v, want %v", gotMatch, tt.wantMatch)
			}
		})
	}
}
