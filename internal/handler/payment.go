package handler

import (
	"net/http"

	"pickup/internal/middleware"
	"pickup/internal/model"
	"pickup/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// PaymentHandler 支付处理器
type PaymentHandler struct {
	paymentService service.PaymentService
	logger         *zap.Logger
}

// NewPaymentHandler 创建支付处理器
func NewPaymentHandler(paymentService service.PaymentService, logger *zap.Logger) *PaymentHandler {
	return &PaymentHandler{paymentService: paymentService, logger: logger}
}

// RegisterRoutes 注册支付相关路由
func (h *PaymentHandler) RegisterRoutes(r *gin.RouterGroup) {
	pay := r.Group("/pay")
	{
		pay.POST("/prepare", h.PreparePayment)
		pay.POST("/notify", h.PaymentNotify)
	}
}

type preparePaymentRequest struct {
	OrderID uint `json:"order_id" binding:"required"`
}

// PreparePayment 准备支付
func (h *PaymentHandler) PreparePayment(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "未授权"))
		return
	}

	var req preparePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("invalid prepare payment request", zap.Error(err))
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeInvalidParams, "参数错误"))
		return
	}

	resp, err := h.paymentService.PreparePayment(userID, req.OrderID)
	if err != nil {
		h.logger.Error("prepare payment failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodePaymentFailed, "准备支付失败"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(resp))
}

// PaymentNotify 支付回调
func (h *PaymentHandler) PaymentNotify(c *gin.Context) {
	var req service.PaymentNotifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("invalid payment notify request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"code": "FAIL", "message": "invalid request"})
		return
	}

	if err := h.paymentService.HandlePaymentNotify(&req); err != nil {
		h.logger.Error("handle payment notify failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": "processing failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "OK"})
}
