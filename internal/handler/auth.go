package handler

import (
	"net/http"

	"pickup/internal/middleware"
	"pickup/internal/model"
	"pickup/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService service.AuthService
	logger      *zap.Logger
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authService service.AuthService, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

// WechatLoginRequest 微信登录请求
type WechatLoginRequest struct {
	Code      string `json:"code" binding:"required"`       // wx.login获取的code
	PhoneCode string `json:"phone_code" binding:"required"` // getPhoneNumber获取的code
}

// RegisterRoutes 注册路由
func (h *AuthHandler) RegisterRoutes(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	{
		auth.POST("/wechat/login", h.WechatLogin)
		auth.GET("/me", h.GetMe) // 暂时移除认证中间件，需要JWT工具实例
	}
}

// WechatLogin 微信登录
func (h *AuthHandler) WechatLogin(c *gin.Context) {
	var req WechatLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("invalid request", zap.Error(err))
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeInvalidParams, "参数错误"))
		return
	}

	response, err := h.authService.WechatLogin(req.Code, req.PhoneCode)
	if err != nil {
		h.logger.Error("wechat login failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeWechatAuthFailed, "微信登录失败"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(response))
}

// GetMe 获取当前用户信息
func (h *AuthHandler) GetMe(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "未授权"))
		return
	}

	user, err := h.authService.GetUserInfo(userID)
	if err != nil {
		h.logger.Error("get user info failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "获取用户信息失败"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(user))
}
