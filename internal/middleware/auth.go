package middleware

import (
	"net/http"
	"strings"

	"pickup/internal/model"
	"pickup/internal/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware JWT认证中间件
func AuthMiddleware(jwtUtil *utils.JWTUtil) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, model.MsgUnauthorized))
			c.Abort()
			return
		}

		// 检查Bearer前缀
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, model.MsgUnauthorized))
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := jwtUtil.ParseToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, model.MsgUnauthorized))
			c.Abort()
			return
		}

		// 将用户信息存储到上下文
		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)
		c.Next()
	}
}

// RequireRole 角色权限中间件
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, model.MsgUnauthorized))
			c.Abort()
			return
		}

		role, ok := userRole.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, model.MsgInternalError))
			c.Abort()
			return
		}

		// 检查角色是否在允许的角色列表中
		for _, allowedRole := range roles {
			if role == allowedRole {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, model.NewErrorResponse(model.CodeForbidden, model.MsgForbidden))
		c.Abort()
	}
}

// GetUserID 从上下文获取用户ID
func GetUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	id, ok := userID.(uint)
	return id, ok
}

// GetUserRole 从上下文获取用户角色
func GetUserRole(c *gin.Context) (string, bool) {
	userRole, exists := c.Get("user_role")
	if !exists {
		return "", false
	}
	role, ok := userRole.(string)
	return role, ok
}
