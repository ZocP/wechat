package handler

import (
	"net/http"
	"strconv"

	"pickup/internal/middleware"
	"pickup/internal/model"
	"pickup/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// NoticeHandler 消息板处理器
type NoticeHandler struct {
	noticeService service.NoticeService
	logger        *zap.Logger
}

// NewNoticeHandler 创建消息板处理器
func NewNoticeHandler(noticeService service.NoticeService, logger *zap.Logger) *NoticeHandler {
	return &NoticeHandler{
		noticeService: noticeService,
		logger:        logger,
	}
}

// RegisterRoutes 注册路由
func (h *NoticeHandler) RegisterRoutes(r *gin.RouterGroup) {
	notices := r.Group("/notices")
	{
		notices.GET("", h.GetVisibleNotices)
		notices.GET("/flight/:flight_no", h.GetNoticesByFlightNo)
		notices.GET("/:id", h.GetNotice)
	}

	// 管理端路由
	admin := r.Group("/admin")
	// admin.Use(middleware.AuthMiddleware(nil), middleware.RequireRole("admin", "dispatcher")) // 暂时移除认证中间件
	{
		admin.POST("/notices", h.CreateNotice)
		admin.PUT("/notices/:id", h.UpdateNotice)
		admin.DELETE("/notices/:id", h.DeleteNotice)
	}
}

// CreateNotice 创建消息（管理端）
func (h *NoticeHandler) CreateNotice(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req service.CreateNoticeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("invalid request", zap.Error(err))
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeInvalidParams, "参数错误"))
		return
	}

	notice, err := h.noticeService.CreateNotice(userID, &req)
	if err != nil {
		h.logger.Error("create notice failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "创建消息失败"))
		return
	}

	c.JSON(http.StatusCreated, model.NewSuccessResponse(notice))
}

// GetNotice 获取消息详情
func (h *NoticeHandler) GetNotice(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeInvalidParams, "无效的ID"))
		return
	}

	notice, err := h.noticeService.GetNotice(uint(id))
	if err != nil {
		h.logger.Error("get notice failed", zap.Error(err))
		c.JSON(http.StatusNotFound, model.NewErrorResponse(model.CodeNotFound, "获取消息详情失败"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(notice))
}

// GetVisibleNotices 获取当前可见的消息
func (h *NoticeHandler) GetVisibleNotices(c *gin.Context) {
	notices, err := h.noticeService.GetVisibleNotices()
	if err != nil {
		h.logger.Error("get visible notices failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "获取消息列表失败"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(notices))
}

// GetNoticesByFlightNo 根据航班号获取消息
func (h *NoticeHandler) GetNoticesByFlightNo(c *gin.Context) {
	flightNo := c.Param("flight_no")
	if flightNo == "" {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeInvalidParams, "航班号不能为空"))
		return
	}

	notices, err := h.noticeService.GetNoticesByFlightNo(flightNo)
	if err != nil {
		h.logger.Error("get notices by flight failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "获取航班消息失败"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(notices))
}

// UpdateNotice 更新消息（管理端）
func (h *NoticeHandler) UpdateNotice(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeInvalidParams, "无效的ID"))
		return
	}

	var req service.UpdateNoticeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("invalid request", zap.Error(err))
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeInvalidParams, "参数错误"))
		return
	}

	notice, err := h.noticeService.UpdateNotice(uint(id), userID, &req)
	if err != nil {
		h.logger.Error("update notice failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "更新消息失败"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(notice))
}

// DeleteNotice 删除消息（管理端）
func (h *NoticeHandler) DeleteNotice(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeInvalidParams, "无效的ID"))
		return
	}

	err = h.noticeService.DeleteNotice(uint(id), userID)
	if err != nil {
		h.logger.Error("delete notice failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "删除消息失败"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(nil))
}
