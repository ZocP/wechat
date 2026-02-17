package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pickup/internal/scheduler/controllers"
	"pickup/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRegisterRoutes_HealthAndProtected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	jwtUtil := utils.NewJWTUtil("secret", time.Hour, "issuer")
	RegisterRoutes(r, &controllers.AuthController{}, &controllers.StudentController{}, &controllers.AdminController{}, jwtUtil)

	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/admin/drivers", nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusUnauthorized, w2.Code)
}
