package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTUtil JWT工具
type JWTUtil struct {
	secret     []byte
	expireTime time.Duration
	issuer     string
}

// Claims JWT声明
type Claims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// NewJWTUtil 创建JWT工具
func NewJWTUtil(secret string, expireTime time.Duration, issuer string) *JWTUtil {
	return &JWTUtil{
		secret:     []byte(secret),
		expireTime: expireTime,
		issuer:     issuer,
	}
}

// GenerateToken 生成JWT令牌
func (j *JWTUtil) GenerateToken(userID uint, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   fmt.Sprintf("%d", userID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.expireTime)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

// ParseToken 解析JWT令牌
func (j *JWTUtil) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshToken 刷新令牌
func (j *JWTUtil) RefreshToken(tokenString string) (string, error) {
	claims, err := j.ParseToken(tokenString)
	if err != nil {
		return "", err
	}

	// 检查令牌是否即将过期（剩余时间少于一半）
	if time.Until(claims.ExpiresAt.Time) < j.expireTime/2 {
		return j.GenerateToken(claims.UserID, claims.Role)
	}

	return tokenString, nil
}
