package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	count := 50
	if len(os.Args) > 1 {
		if parsed, err := strconv.Atoi(os.Args[1]); err == nil && parsed > 0 {
			count = parsed
		}
	}

	secret := []byte("pickup-secret-key")
	now := time.Now()
	tokens := make([]string, 0, count)

	for userID := 1; userID <= count; userID++ {
		claims := jwt.MapClaims{
			"user_id": userID,
			"role":    "passenger",
			"iss":     "pickup",
			"sub":     fmt.Sprintf("%d", userID),
			"iat":     now.Unix(),
			"exp":     now.Add(24 * time.Hour).Unix(),
			"nbf":     now.Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := token.SignedString(secret)
		if err != nil {
			panic(err)
		}
		tokens = append(tokens, signed)
	}

	payload, err := json.MarshalIndent(tokens, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(payload))
}
