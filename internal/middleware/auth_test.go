package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pickup/internal/model"
	"pickup/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newTestJWTUtil() *utils.JWTUtil {
	return utils.NewJWTUtil("test-secret-key", 24*time.Hour, "test-issuer")
}

func TestAuthMiddleware_Success(t *testing.T) {
	jwtUtil := newTestJWTUtil()
	token, err := jwtUtil.GenerateToken(42, "admin")
	require.NoError(t, err)

	router := gin.New()
	router.Use(AuthMiddleware(jwtUtil))
	router.GET("/test", func(c *gin.Context) {
		userID, exists := GetUserID(c)
		assert.True(t, exists)
		assert.Equal(t, uint(42), userID)

		role, exists := GetUserRole(c)
		assert.True(t, exists)
		assert.Equal(t, "admin", role)

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	jwtUtil := newTestJWTUtil()

	router := gin.New()
	router.Use(AuthMiddleware(jwtUtil))
	router.GET("/test", func(c *gin.Context) {
		t.Fatal("handler should not be called")
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_InvalidFormat(t *testing.T) {
	jwtUtil := newTestJWTUtil()

	testCases := []struct {
		name   string
		header string
	}{
		{"no bearer prefix", "token_value"},
		{"basic auth", "Basic dXNlcjpwYXNz"},
		{"empty bearer", "Bearer "},
		{"triple parts", "Bearer token extra"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()
			router.Use(AuthMiddleware(jwtUtil))
			router.GET("/test", func(c *gin.Context) {
				t.Fatal("handler should not be called")
			})

			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", tc.header)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	jwtUtil := newTestJWTUtil()

	router := gin.New()
	router.Use(AuthMiddleware(jwtUtil))
	router.GET("/test", func(c *gin.Context) {
		t.Fatal("handler should not be called")
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	jwtUtil := utils.NewJWTUtil("secret", 1*time.Millisecond, "issuer")
	token, err := jwtUtil.GenerateToken(1, "passenger")
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	router := gin.New()
	router.Use(AuthMiddleware(jwtUtil))
	router.GET("/test", func(c *gin.Context) {
		t.Fatal("handler should not be called")
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireRole_Allowed(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_role", "admin")
		c.Next()
	})
	router.Use(RequireRole("admin", "dispatcher"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireRole_Forbidden(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_role", "passenger")
		c.Next()
	})
	router.Use(RequireRole("admin", "dispatcher"))
	router.GET("/test", func(c *gin.Context) {
		t.Fatal("handler should not be called")
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireRole_NoRole(t *testing.T) {
	router := gin.New()
	router.Use(RequireRole("admin"))
	router.GET("/test", func(c *gin.Context) {
		t.Fatal("handler should not be called")
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireRole_InvalidRoleType(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_role", 12345) // Not a string
		c.Next()
	})
	router.Use(RequireRole("admin"))
	router.GET("/test", func(c *gin.Context) {
		t.Fatal("handler should not be called")
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetUserID_Exists(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", uint(42))

	id, ok := GetUserID(c)
	assert.True(t, ok)
	assert.Equal(t, uint(42), id)
}

func TestGetUserID_NotExists(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	id, ok := GetUserID(c)
	assert.False(t, ok)
	assert.Equal(t, uint(0), id)
}

func TestGetUserID_WrongType(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", "not_uint")

	id, ok := GetUserID(c)
	assert.False(t, ok)
	assert.Equal(t, uint(0), id)
}

func TestGetUserRole_Exists(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_role", "admin")

	role, ok := GetUserRole(c)
	assert.True(t, ok)
	assert.Equal(t, "admin", role)
}

func TestGetUserRole_NotExists(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	role, ok := GetUserRole(c)
	assert.False(t, ok)
	assert.Equal(t, "", role)
}

func TestGetUserRole_WrongType(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_role", 123)

	role, ok := GetUserRole(c)
	assert.False(t, ok)
	assert.Equal(t, "", role)
}

// TestAuthMiddleware_ResponseFormat verifies the error response format
func TestAuthMiddleware_ResponseFormat(t *testing.T) {
	jwtUtil := newTestJWTUtil()

	router := gin.New()
	router.Use(AuthMiddleware(jwtUtil))
	router.GET("/test", func(c *gin.Context) {})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp model.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, model.CodeUnauthorized, resp.Code)
	assert.Equal(t, model.MsgUnauthorized, resp.Message)
}
