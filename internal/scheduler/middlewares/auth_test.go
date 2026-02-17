package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pickup/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTAuthAndRequireRoles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtUtil := utils.NewJWTUtil("secret", time.Hour, "issuer")
	token, err := jwtUtil.GenerateToken(123, "admin")
	require.NoError(t, err)

	r := gin.New()
	r.GET("/p", JWTAuth(jwtUtil), RequireRoles("admin"), func(c *gin.Context) {
		uid, ok := UserID(c)
		if !ok || uid != 123 {
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/p", nil)
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusUnauthorized, w1.Code)

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/p", nil)
	req2.Header.Set("Authorization", "Bearer bad-token")
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusUnauthorized, w2.Code)

	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodGet, "/p", nil)
	req3.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusOK, w3.Code)
}

func TestRequireRolesForbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtUtil := utils.NewJWTUtil("secret", time.Hour, "issuer")
	token, err := jwtUtil.GenerateToken(123, "student")
	require.NoError(t, err)

	r := gin.New()
	r.GET("/p", JWTAuth(jwtUtil), RequireRoles("admin"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/p", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireRolesAdminSuperuser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtUtil := utils.NewJWTUtil("secret", time.Hour, "issuer")
	token, err := jwtUtil.GenerateToken(123, "admin")
	require.NoError(t, err)

	r := gin.New()
	r.GET("/p", JWTAuth(jwtUtil), RequireRoles("student"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/p", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireRoles_StaffAllowedWhenRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtUtil := utils.NewJWTUtil("secret", time.Hour, "issuer")
	token, err := jwtUtil.GenerateToken(123, "staff")
	require.NoError(t, err)

	r := gin.New()
	r.GET("/p", JWTAuth(jwtUtil), RequireRoles("staff"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/p", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAssignStaff_AdminOnlyChain(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtUtil := utils.NewJWTUtil("secret", time.Hour, "issuer")
	staffToken, err := jwtUtil.GenerateToken(1, "staff")
	require.NoError(t, err)
	adminToken, err := jwtUtil.GenerateToken(2, "admin")
	require.NoError(t, err)

	r := gin.New()
	r.GET("/assign-staff", JWTAuth(jwtUtil), RequireRoles("staff"), RequireRoles("admin"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/assign-staff", nil)
	req1.Header.Set("Authorization", "Bearer "+staffToken)
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusForbidden, w1.Code)

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/assign-staff", nil)
	req2.Header.Set("Authorization", "Bearer "+adminToken)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
}
