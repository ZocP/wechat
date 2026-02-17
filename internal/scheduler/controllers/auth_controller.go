package controllers

import (
	"net/http"

	"pickup/internal/scheduler/middlewares"
	"pickup/internal/scheduler/service"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authSvc *service.AuthService
}

func NewAuthController(authSvc *service.AuthService) *AuthController {
	return &AuthController{authSvc: authSvc}
}

type loginRequest struct {
	Code string `json:"code" binding:"required"`
}

type bindPhoneRequest struct {
	PhoneCode string `json:"phone_code" binding:"required"`
}

func (ctl *AuthController) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := ctl.authSvc.LoginWithWechatCode(req.Code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (ctl *AuthController) BindPhone(c *gin.Context) {
	userID, ok := middlewares.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req bindPhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := ctl.authSvc.BindPhone(userID, req.PhoneCode); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (ctl *AuthController) Me(c *gin.Context) {
	userID, ok := middlewares.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, err := ctl.authSvc.GetMe(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}
