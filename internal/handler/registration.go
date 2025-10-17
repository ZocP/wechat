package handler

import (
	"net/http"

	"pickup/internal/middleware"
	"pickup/internal/model"
	"pickup/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RegistrationHandler 报名处理器
type RegistrationHandler struct {
	registrationService service.RegistrationService
	logger              *zap.Logger
}

// NewRegistrationHandler 创建报名处理器
func NewRegistrationHandler(registrationService service.RegistrationService, logger *zap.Logger) *RegistrationHandler {
	return &RegistrationHandler{
		registrationService: registrationService,
		logger:              logger,
	}
}

// RegisterRoutes 注册报名相关路由
func (h *RegistrationHandler) RegisterRoutes(r *gin.RouterGroup) {
	registrations := r.Group("/registrations")
	{
		registrations.POST("", h.CreateRegistration)
		registrations.GET("", h.GetMyRegistrations)
		registrations.GET("/:id", h.GetRegistration)
		registrations.PUT("/:id", h.UpdateRegistration)
		registrations.DELETE("/:id", h.DeleteRegistration)
	}

	// 兼容 /registrations/my 路由
	registrations.GET("/my", h.GetMyRegistrations)
}

// CreateRegistration 创建报名
func (h *RegistrationHandler) CreateRegistration(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "未授权"))
		return
	}

	var req service.CreateRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("invalid registration request", zap.Error(err))
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeInvalidParams, "参数错误"))
		return
	}

	registration, err := h.registrationService.CreateRegistration(userID, &req)
	if err != nil {
		h.logger.Error("create registration failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "创建报名失败"))
		return
	}

	c.JSON(http.StatusCreated, model.NewSuccessResponse(registration))
}

// GetRegistration 获取报名详情
func (h *RegistrationHandler) GetRegistration(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "未授权"))
		return
	}

	id, err := parseUintParam(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeInvalidParams, "无效的ID"))
		return
	}

	registration, err := h.registrationService.GetRegistration(id, userID)
	if err != nil {
		h.logger.Error("get registration failed", zap.Error(err))
		c.JSON(http.StatusNotFound, model.NewErrorResponse(model.CodeNotFound, "获取报名详情失败"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(registration))
}

// GetMyRegistrations 获取我的报名列表
func (h *RegistrationHandler) GetMyRegistrations(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "未授权"))
		return
	}

	registrations, err := h.registrationService.GetUserRegistrations(userID)
	if err != nil {
		h.logger.Error("get user registrations failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "获取报名列表失败"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(registrations))
}

// UpdateRegistration 更新报名
func (h *RegistrationHandler) UpdateRegistration(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "未授权"))
		return
	}

	id, err := parseUintParam(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeInvalidParams, "无效的ID"))
		return
	}

	var req service.UpdateRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("invalid update registration request", zap.Error(err))
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeInvalidParams, "参数错误"))
		return
	}

	registration, err := h.registrationService.UpdateRegistration(id, userID, &req)
	if err != nil {
		h.logger.Error("update registration failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "更新报名失败"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(registration))
}

// DeleteRegistration 删除报名
func (h *RegistrationHandler) DeleteRegistration(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "未授权"))
		return
	}

	id, err := parseUintParam(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeInvalidParams, "无效的ID"))
		return
	}

	if err := h.registrationService.DeleteRegistration(id, userID); err != nil {
		h.logger.Error("delete registration failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "删除报名失败"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(nil))
}
