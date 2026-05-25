package security

import (
	"testing"

	"github.com/alexedwards/argon2id"
)

func TestCreateHash(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{name: "valid password", password: "password", wantErr: false},
		{name: "empty password", password: "", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CreateHash(tt.password, argon2id.DefaultParams)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateHash() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestComparePasswordAndHash(t *testing.T) {
	hash, err := CreateHash("password", argon2id.DefaultParams)
	if err != nil {
		t.Fatalf("setup: CreateHash() error = %v", err)
	}

	tests := []struct {
		name      string
		password  string
		hash      string
		wantMatch bool
		wantErr   bool
	}{
		{name: "correct password", password: "password", hash: hash, wantMatch: true, wantErr: false},
		{name: "wrong password", password: "wrong", hash: hash, wantMatch: false, wantErr: false},
		{name: "invalid hash", password: "password", hash: "not-a-hash", wantMatch: false, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMatch, err := ComparePasswordAndHash(tt.password, tt.hash)
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
