package services

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/tkahng/playground/internal/conf"
)

func TestJwtService_CreateJwtToken(t *testing.T) {
	t.Parallel()
	service := NewJwtService()
	claims := jwt.MapClaims{"type": "access_token", "foo": "bar"}
	token, err := service.CreateJwtToken(claims, "secret")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestJwtService_ParseToken_InvalidToken(t *testing.T) {
	t.Parallel()
	service := NewJwtService()
	var data map[string]interface{}
	err := service.ParseToken("invalid.token", conf.TokenOption{Secret: "secret", Type: "access_token"}, &data)
	assert.Error(t, err)
}

func TestJwtService_ParseToken_ValidToken(t *testing.T) {
	t.Parallel()
	service := NewJwtService()
	claims := jwt.MapClaims{"type": "access_token", "foo": "bar"}
	token, err := service.CreateJwtToken(claims, "secret")
	assert.NoError(t, err)

	var data map[string]interface{}
	err = service.ParseToken(token, conf.TokenOption{Secret: "secret", Type: "access_token"}, &data)
	assert.NoError(t, err)
	assert.Equal(t, "bar", data["foo"])
}
