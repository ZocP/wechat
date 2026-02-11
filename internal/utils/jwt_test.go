package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestJWT() *JWTUtil {
	return NewJWTUtil("test-secret-key-for-unit-tests", 24*time.Hour, "test-issuer")
}

func TestNewJWTUtil(t *testing.T) {
	util := NewJWTUtil("secret", 1*time.Hour, "issuer")
	assert.NotNil(t, util)
	assert.Equal(t, []byte("secret"), util.secret)
	assert.Equal(t, 1*time.Hour, util.expireTime)
	assert.Equal(t, "issuer", util.issuer)
}

func TestGenerateToken_Success(t *testing.T) {
	util := newTestJWT()
	token, err := util.GenerateToken(1, "passenger")
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestGenerateToken_DifferentUsers(t *testing.T) {
	util := newTestJWT()
	token1, err := util.GenerateToken(1, "passenger")
	require.NoError(t, err)
	token2, err := util.GenerateToken(2, "admin")
	require.NoError(t, err)
	assert.NotEqual(t, token1, token2)
}

func TestParseToken_Success(t *testing.T) {
	util := newTestJWT()
	token, err := util.GenerateToken(42, "admin")
	require.NoError(t, err)

	claims, err := util.ParseToken(token)
	require.NoError(t, err)
	assert.Equal(t, uint(42), claims.UserID)
	assert.Equal(t, "admin", claims.Role)
	assert.Equal(t, "test-issuer", claims.Issuer)
	assert.Equal(t, "42", claims.Subject)
}

func TestParseToken_InvalidToken(t *testing.T) {
	util := newTestJWT()
	_, err := util.ParseToken("invalid.token.string")
	assert.Error(t, err)
}

func TestParseToken_EmptyToken(t *testing.T) {
	util := newTestJWT()
	_, err := util.ParseToken("")
	assert.Error(t, err)
}

func TestParseToken_WrongSecret(t *testing.T) {
	util1 := NewJWTUtil("secret-1", 24*time.Hour, "issuer")
	util2 := NewJWTUtil("secret-2", 24*time.Hour, "issuer")

	token, err := util1.GenerateToken(1, "passenger")
	require.NoError(t, err)

	_, err = util2.ParseToken(token)
	assert.Error(t, err)
}

func TestParseToken_ExpiredToken(t *testing.T) {
	// Create a JWT with very short expiry
	util := NewJWTUtil("secret", 1*time.Millisecond, "issuer")
	token, err := util.GenerateToken(1, "passenger")
	require.NoError(t, err)

	// Wait for expiry
	time.Sleep(10 * time.Millisecond)

	_, err = util.ParseToken(token)
	assert.Error(t, err)
}

func TestParseToken_ClaimsFields(t *testing.T) {
	util := newTestJWT()
	// JWT NumericDate has second precision, so truncate boundaries
	before := time.Now().Truncate(time.Second)
	token, err := util.GenerateToken(100, "driver")
	require.NoError(t, err)
	after := time.Now().Truncate(time.Second).Add(time.Second) // round up

	claims, err := util.ParseToken(token)
	require.NoError(t, err)

	assert.Equal(t, uint(100), claims.UserID)
	assert.Equal(t, "driver", claims.Role)
	assert.Equal(t, "test-issuer", claims.Issuer)
	assert.Equal(t, "100", claims.Subject)

	// Check time claims are within expected range (second precision)
	assert.False(t, claims.IssuedAt.Time.Before(before), "IssuedAt should be >= before")
	assert.False(t, claims.IssuedAt.Time.After(after), "IssuedAt should be <= after")
	assert.True(t, claims.ExpiresAt.Time.After(before.Add(24*time.Hour-time.Second)))
}

func TestRefreshToken_NotExpiringSoon(t *testing.T) {
	util := newTestJWT() // 24h expiry
	token, err := util.GenerateToken(1, "passenger")
	require.NoError(t, err)

	// Token is fresh, should not be refreshed
	refreshed, err := util.RefreshToken(token)
	require.NoError(t, err)
	assert.Equal(t, token, refreshed, "Token should not be refreshed when not near expiry")
}

func TestRefreshToken_ExpiringSoon(t *testing.T) {
	// Create a JWT with 4 second expiry so it stays valid while past the halfway mark
	util := NewJWTUtil("secret", 4*time.Second, "issuer")
	token, err := util.GenerateToken(1, "passenger")
	require.NoError(t, err)

	// Wait for more than half the expiry (>2 seconds) but leave enough room before full expiry
	time.Sleep(2100 * time.Millisecond)

	refreshed, err := util.RefreshToken(token)
	require.NoError(t, err)
	assert.NotEqual(t, token, refreshed, "Token should be refreshed when near expiry")

	// Verify the refreshed token is valid
	claims, err := util.ParseToken(refreshed)
	require.NoError(t, err)
	assert.Equal(t, uint(1), claims.UserID)
	assert.Equal(t, "passenger", claims.Role)
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	util := newTestJWT()
	_, err := util.RefreshToken("invalid-token")
	assert.Error(t, err)
}

func TestRefreshToken_ExpiredToken(t *testing.T) {
	util := NewJWTUtil("secret", 1*time.Millisecond, "issuer")
	token, err := util.GenerateToken(1, "passenger")
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	_, err = util.RefreshToken(token)
	assert.Error(t, err)
}
