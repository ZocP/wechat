package controllers

import (
	"net/http"
	"strconv"
	"time"

	"pickup/internal/scheduler/service"

	"github.com/gin-gonic/gin"
)

type AdminController struct {
	svc *service.AdminService
}

func NewAdminController(svc *service.AdminService) *AdminController {
	return &AdminController{svc: svc}
}

type createShiftRequest struct {
	DriverID      uint   `json:"driver_id" binding:"required"`
	DepartureTime string `json:"departure_time" binding:"required"`
}

type createDriverRequest struct {
	Name       string `json:"name" binding:"required"`
	CarModel   string `json:"car_model" binding:"required"`
	MaxSeats   int    `json:"max_seats" binding:"required"`
	MaxChecked int    `json:"max_checked" binding:"required"`
	MaxCarryOn int    `json:"max_carry_on" binding:"required"`
}

type assignStudentRequest struct {
	RequestID uint `json:"request_id" binding:"required"`
}

type assignStaffRequest struct {
	StaffID uint `json:"staff_id" binding:"required"`
}

type updateDriverRequest struct {
	Name       string `json:"name" binding:"required"`
	CarModel   string `json:"car_model" binding:"required"`
	MaxSeats   int    `json:"max_seats" binding:"required"`
	MaxChecked int    `json:"max_checked" binding:"required"`
	MaxCarryOn int    `json:"max_carry_on" binding:"required"`
}

type updateShiftRequest struct {
	DriverID      *uint   `json:"driver_id"`
	DepartureTime *string `json:"departure_time"`
}

func (ctl *AdminController) ListDrivers(c *gin.Context) {
	res, err := ctl.svc.ListDrivers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (ctl *AdminController) CreateDriver(c *gin.Context) {
	var input createDriverRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	driver, err := ctl.svc.CreateDriver(service.DriverDTO{
		Name:       input.Name,
		CarModel:   input.CarModel,
		MaxSeats:   input.MaxSeats,
		MaxChecked: input.MaxChecked,
		MaxCarryOn: input.MaxCarryOn,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, driver)
}

func (ctl *AdminController) UpdateDriver(c *gin.Context) {
	driverID, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid driver id"})
		return
	}

	var input updateDriverRequest
	if err = c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	driver, err := ctl.svc.UpdateDriver(driverID, service.DriverDTO{
		Name:       input.Name,
		CarModel:   input.CarModel,
		MaxSeats:   input.MaxSeats,
		MaxChecked: input.MaxChecked,
		MaxCarryOn: input.MaxCarryOn,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, driver)
}

func (ctl *AdminController) Dashboard(c *gin.Context) {
	res, err := ctl.svc.DashboardShifts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (ctl *AdminController) PendingRequests(c *gin.Context) {
	res, err := ctl.svc.PendingRequests()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (ctl *AdminController) ListUsers(c *gin.Context) {
	res, err := ctl.svc.ListUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (ctl *AdminController) SetStaff(c *gin.Context) {
	userID, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	user, err := ctl.svc.SetUserStaff(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (ctl *AdminController) UnsetStaff(c *gin.Context) {
	userID, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	user, err := ctl.svc.UnsetUserStaff(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (ctl *AdminController) CreateShift(c *gin.Context) {
	var req createShiftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	t, err := time.Parse("2006-01-02 15:04:05", req.DepartureTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid departure_time"})
		return
	}
	shift, err := ctl.svc.CreateShift(req.DriverID, t)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, shift)
}

func (ctl *AdminController) UpdateShift(c *gin.Context) {
	shiftID, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid shift id"})
		return
	}

	var req updateShiftRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var departureTime *time.Time
	if req.DepartureTime != nil {
		parsed, parseErr := time.Parse("2006-01-02 15:04:05", *req.DepartureTime)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid departure_time"})
			return
		}
		departureTime = &parsed
	}

	shift, err := ctl.svc.UpdateShift(shiftID, service.ShiftUpdateDTO{
		DriverID:      req.DriverID,
		DepartureTime: departureTime,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, shift)
}

func (ctl *AdminController) AssignStudent(c *gin.Context) {
	shiftID, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid shift id"})
		return
	}
	var req assignStudentRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := ctl.svc.AssignStudent(shiftID, req.RequestID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (ctl *AdminController) RemoveStudent(c *gin.Context) {
	shiftID, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid shift id"})
		return
	}
	var req assignStudentRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err = ctl.svc.RemoveStudent(shiftID, req.RequestID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (ctl *AdminController) AssignStaff(c *gin.Context) {
	shiftID, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid shift id"})
		return
	}
	var req assignStaffRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err = ctl.svc.AssignStaff(shiftID, req.StaffID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (ctl *AdminController) RemoveStaff(c *gin.Context) {
	shiftID, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid shift id"})
		return
	}
	var req assignStaffRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err = ctl.svc.RemoveStaff(shiftID, req.StaffID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (ctl *AdminController) PublishShift(c *gin.Context) {
	shiftID, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid shift id"})
		return
	}
	if err = ctl.svc.PublishShift(shiftID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func parseID(raw string) (uint, error) {
	id64, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id64), nil
}
