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

// OrderHandler 订单处理器
type OrderHandler struct {
	orderService service.OrderService
	logger       *zap.Logger
}

// NewOrderHandler 创建订单处理器
func NewOrderHandler(orderService service.OrderService, logger *zap.Logger) *OrderHandler {
	return &OrderHandler{orderService: orderService, logger: logger}
}

// RegisterRoutes 注册订单相关路由
func (h *OrderHandler) RegisterRoutes(r *gin.RouterGroup) {
	orders := r.Group("/orders")
	{
		orders.POST("", h.CreateOrder)
		orders.GET("", h.GetMyOrders)
		orders.GET("/:id", h.GetOrder)
	}

	admin := r.Group("/admin")
	{
		admin.POST("/orders/:id/notify", h.NotifyOrder)
	}
}

// CreateOrder 创建订单
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "未授权"))
		return
	}

	var req service.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("invalid order request", zap.Error(err))
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeInvalidParams, "参数错误"))
		return
	}

	order, err := h.orderService.CreateOrder(userID, &req)
	if err != nil {
		h.logger.Error("create order failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "创建订单失败"))
		return
	}

	c.JSON(http.StatusCreated, model.NewSuccessResponse(order))
}

// GetOrder 获取订单详情
func (h *OrderHandler) GetOrder(c *gin.Context) {
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

	order, err := h.orderService.GetOrder(id, userID)
	if err != nil {
		h.logger.Error("get order failed", zap.Error(err))
		c.JSON(http.StatusNotFound, model.NewErrorResponse(model.CodeNotFound, "获取订单详情失败"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(order))
}

// GetMyOrders 获取我的订单列表
func (h *OrderHandler) GetMyOrders(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "未授权"))
		return
	}

	page, pageSize, offset := parsePagination(c)

	orders, err := h.orderService.GetUserOrders(userID)
	if err != nil {
		h.logger.Error("get user orders failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "获取订单列表失败"))
		return
	}

	total := len(orders)
	if offset >= total {
		c.Header("X-Page", strconv.Itoa(page))
		c.Header("X-Page-Size", strconv.Itoa(pageSize))
		c.Header("X-Total-Count", strconv.Itoa(total))
		c.JSON(http.StatusOK, model.NewSuccessResponse([]*model.PickupOrder{}))
		return
	}

	end := offset + pageSize
	if end > total {
		end = total
	}

	c.Header("X-Page", strconv.Itoa(page))
	c.Header("X-Page-Size", strconv.Itoa(pageSize))
	c.Header("X-Total-Count", strconv.Itoa(total))
	orders = orders[offset:end]

	c.JSON(http.StatusOK, model.NewSuccessResponse(orders))
}

// NotifyOrder 管理端通知订单
func (h *OrderHandler) NotifyOrder(c *gin.Context) {
	role, _ := middleware.GetUserRole(c)
	if role != string(model.RoleAdmin) && role != string(model.RoleDispatcher) {
		c.JSON(http.StatusForbidden, model.NewErrorResponse(model.CodeForbidden, "无权限执行该操作"))
		return
	}

	id, err := parseUintParam(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeInvalidParams, "无效的ID"))
		return
	}

	if err := h.orderService.UpdateOrderStatus(id, model.OrderStatusNotified); err != nil {
		h.logger.Error("notify order failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "通知订单失败"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(nil))
}
