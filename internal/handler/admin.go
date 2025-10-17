package handler

import (
	"net/http"

	"pickup/internal/model"
	"pickup/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AdminHandler 管理端辅助接口
type AdminHandler struct {
	schemaService service.SchemaService
	logger        *zap.Logger
}

// NewAdminHandler 创建管理端处理器
func NewAdminHandler(schemaService service.SchemaService, logger *zap.Logger) *AdminHandler {
	return &AdminHandler{
		schemaService: schemaService,
		logger:        logger,
	}
}

// RegisterRoutes 注册管理端路由
func (h *AdminHandler) RegisterRoutes(r *gin.RouterGroup) {
	admin := r.Group("/admin")
	{
		admin.GET("/exports/database-fields", h.ExportDatabaseFields)
	}
}

// ExportDatabaseFields 导出数据库字段信息
func (h *AdminHandler) ExportDatabaseFields(c *gin.Context) {
	data, err := h.schemaService.ExportSchemaJSON()
	if err != nil {
		h.logger.Error("export database fields failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "导出数据库字段失败"))
		return
	}

	c.Header("Content-Disposition", "attachment; filename=database_fields.json")
	c.Data(http.StatusOK, "application/json", data)
}
