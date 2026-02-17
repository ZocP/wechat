package controllers

import (
	"net/http"
	"strconv"

	"pickup/internal/scheduler/middlewares"
	"pickup/internal/scheduler/service"

	"github.com/gin-gonic/gin"
)

type StudentController struct {
	svc *service.StudentService
}

func NewStudentController(svc *service.StudentService) *StudentController {
	return &StudentController{svc: svc}
}

func (ctl *StudentController) CreateRequest(c *gin.Context) {
	userID, ok := middlewares.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var input service.CreateRequestInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := ctl.svc.CreateRequest(userID, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, res)
}

func (ctl *StudentController) MyRequests(c *gin.Context) {
	userID, ok := middlewares.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	res, err := ctl.svc.ListMyRequests(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (ctl *StudentController) UpdateRequest(c *gin.Context) {
	userID, ok := middlewares.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input service.UpdateRequestInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := ctl.svc.UpdatePendingRequest(userID, uint(id64), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}
