package helpers

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

//GetHMACSecret returns HMAC secret for the JWT generation and validation
func GetHMACSecret() []byte {
	return []byte("blablabla")
}

//GenerateJWTToken for auth requests
func GenerateJWTToken() (string, error) {
	now := time.Now()
	expiresAt := now.AddDate(0, 1, 0)

	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		ExpiresAt: expiresAt.Unix(),
		Issuer:    "faktor zwei GmbH",
		IssuedAt:  now.Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	return token.SignedString(GetHMACSecret())
}
