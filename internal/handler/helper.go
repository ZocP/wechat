package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func parseUintParam(value string) (uint, error) {
	id, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

func parsePagination(c *gin.Context) (page int, pageSize int, offset int) {
	const (
		defaultPage     = 1
		defaultPageSize = 20
		maxPageSize     = 100
	)

	page = defaultPage
	pageSize = defaultPageSize

	if value := c.Query("page"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if value := c.Query("page_size"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
			if parsed > maxPageSize {
				pageSize = maxPageSize
			} else {
				pageSize = parsed
			}
		}
	}

	offset = (page - 1) * pageSize
	return page, pageSize, offset
}
