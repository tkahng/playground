package auth

import (
	"encoding/json"
	"time"
)

// wrapFactory is a helper that wraps a Provider specific factory
// function and returns its result as Provider interface.
func wrapFactory[T ProviderConfig](factory func() T) ProviderFactoryFunc {
	return func() ProviderConfig {
		return factory()
	}
}

// ProviderFactoryFunc defines a function for initializing a new OAuth2 provider.
type ProviderFactoryFunc func() ProviderConfig

// Providers defines a map with all of the available OAuth2 providers.
//
// To register a new provider append a new entry in the map.
var Providers = map[string]ProviderFactoryFunc{}

// NewProviderByName returns a new preconfigured provider instance by its name identifier.
func NewProviderByName(name string) ProviderConfig {
	factory, ok := Providers[name]
	if !ok {
		return nil
	}

	return factory()
}

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
