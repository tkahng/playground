package security

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tkahng/authgo/internal/tools/utils"
)

// ParseUnverifiedJWT parses JWT and returns its claims
// but DOES NOT verify the signature.
//
// It verifies only the exp, iat and nbf claims.
func ParseUnverifiedJWT(token string) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}

	parser := &jwt.Parser{}
	_, _, err := parser.ParseUnverified(token, claims)

	if err == nil {
		err = jwt.NewValidator(jwt.WithIssuedAt()).Validate(claims)
	}

	return claims, err
}

// ParseJWTMapClaims verifies and parses JWT and returns its claims.
func ParseJWTMapClaims(token string, verificationKey string) (jwt.MapClaims, error) {
	parser := jwt.NewParser(jwt.WithValidMethods([]string{"HS256"}))

	parsedToken, err := parser.Parse(token, func(t *jwt.Token) (any, error) {
		return []byte(verificationKey), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
		return claims, nil
	}

	return nil, errors.New("unable to parse token")
}

func MarshalToken[T any](token jwt.MapClaims, obj T) (T, error) {
	jsonData := utils.MarshalJSONByte(token)
	obja := obj
	if err := json.Unmarshal(jsonData, obja); err != nil {
		return obja, fmt.Errorf("error at error: %w", err)
	}
	return obja, nil
}

// func ParseJWTWithClaims[T any](token string, claims T, verificationKey string) (*TokenClaims, error) {
// 	parser := jwt.NewParser(jwt.WithValidMethods([]string{"HS256"}))

// 	parsedToken, err := parser.ParseWithClaims(token, &TokenClaims{}, func(t *jwt.Token) (any, error) {
// 		return []byte(verificationKey), nil
// 	})
// 	if err != nil {
// 		return nil, fmt.Errorf("error at error: %w", err)
// 	}

// 	if claims, ok := parsedToken.Claims.(*TokenClaims); ok && parsedToken.Valid {
// 		return claims, nil
// 	}

// 	return nil, errors.New("unable to parse token")
// }

// NewJWT generates and returns new HS256 signed JWT.
func NewJWT(payload jwt.MapClaims, signingKey string, duration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"exp": time.Now().Add(duration).Unix(),
	}

	for k, v := range payload {
		claims[k] = v
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(signingKey))
}

func NewJWTWithClaims(payload jwt.Claims, signingKey string) (string, error) {

	return jwt.NewWithClaims(jwt.SigningMethodHS256, payload).SignedString([]byte(signingKey))
}
