package auth

import (
	"encoding/json"
	"time"
)

// AuthUser defines a standardized OAuth2 user data structure.
type AuthUser struct {
	Expiry       time.Time      `json:"expiry"`
	RawUser      map[string]any `json:"rawUser"`
	Id           string         `json:"id"`
	Name         string         `json:"name"`
	Username     string         `json:"username"`
	Email        string         `json:"email"`
	AvatarURL    string         `json:"avatarURL"`
	AccessToken  string         `json:"accessToken"`
	RefreshToken string         `json:"refreshToken"`

	// @todo
	// deprecated: use AvatarURL instead
	// AvatarUrl will be removed after dropping v0.22 support
	AvatarUrl string `json:"avatarUrl"`
}

// MarshalJSON implements the [json.Marshaler] interface.
//
// @todo remove after dropping v0.22 support
func (au AuthUser) MarshalJSON() ([]byte, error) {
	type alias AuthUser // prevent recursion

	au2 := alias(au)
	au2.AvatarUrl = au.AvatarURL // ensure that the legacy field is populated

	return json.Marshal(au2)
}
