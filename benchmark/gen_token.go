//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	secret := []byte("pickup-secret-key")
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id": 1,
		"role":    "passenger",
		"iss":     "pickup",
		"sub":     "1",
		"iat":     now.Unix(),
		"exp":     now.Add(24 * time.Hour).Unix(),
		"nbf":     now.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(secret)
	if err != nil {
		panic(err)
	}
	fmt.Print(signed)
}
