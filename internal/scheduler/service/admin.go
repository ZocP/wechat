package service

import (
	"context"
	"errors"
	"time"

	"pickup/internal/scheduler/models"

	"gorm.io/gorm"
)

type AdminService struct {
	db       *gorm.DB
	assigner *ShiftAssignmentService
}

type DriverDTO struct {
	Name       string
	CarModel   string
	MaxSeats   int
	MaxChecked int
	MaxCarryOn int
}

type ShiftUpdateDTO struct {
	DriverID      *uint
	DepartureTime *time.Time
}

func NewAdminService(db *gorm.DB, assigner *ShiftAssignmentService) *AdminService {
	return &AdminService{db: db, assigner: assigner}
}

func (s *AdminService) ListDrivers() ([]models.Driver, error) {
	var drivers []models.Driver
	err := s.db.Find(&drivers).Error
	return drivers, err
}

func (s *AdminService) ListUsers() ([]models.User, error) {
	var users []models.User
	err := s.db.Order("id ASC").Find(&users).Error
	return users, err
}

func (s *AdminService) CreateDriver(input DriverDTO) (*models.Driver, error) {
	driver := models.Driver{
		Name:       input.Name,
		CarModel:   input.CarModel,
		MaxSeats:   input.MaxSeats,
		MaxChecked: input.MaxChecked,
		MaxCarryOn: input.MaxCarryOn,
	}
	if err := s.db.Create(&driver).Error; err != nil {
		return nil, err
	}
	return &driver, nil
}

func (s *AdminService) UpdateDriver(driverID uint, input DriverDTO) (*models.Driver, error) {
	updates := map[string]any{
		"name":         input.Name,
		"car_model":    input.CarModel,
		"max_seats":    input.MaxSeats,
		"max_checked":  input.MaxChecked,
		"max_carry_on": input.MaxCarryOn,
	}
	if err := s.db.Model(&models.Driver{}).Where("id = ?", driverID).Updates(updates).Error; err != nil {
		return nil, err
	}

	var driver models.Driver
	if err := s.db.First(&driver, driverID).Error; err != nil {
		return nil, err
	}
	return &driver, nil
}

func (s *AdminService) DashboardShifts() ([]models.Shift, error) {
	var shifts []models.Shift
	err := s.db.
		Where("status IN ?", []models.ShiftStatus{models.ShiftStatusDraft, models.ShiftStatusPublished}).
		Preload("Driver").
		Preload("Requests").
		Preload("Staffs").
		Find(&shifts).Error
	return shifts, err
}

func (s *AdminService) PendingRequests() ([]models.Request, error) {
	var reqs []models.Request
	err := s.db.Where("status = ?", models.RequestStatusPending).Find(&reqs).Error
	return reqs, err
}

func (s *AdminService) CreateShift(driverID uint, departureTime time.Time) (*models.Shift, error) {
	shift := models.Shift{DriverID: driverID, DepartureTime: departureTime, Status: models.ShiftStatusDraft}
	if err := s.db.Create(&shift).Error; err != nil {
		return nil, err
	}
	return &shift, nil
}

func (s *AdminService) UpdateShift(shiftID uint, input ShiftUpdateDTO) (*models.Shift, error) {
	updates := map[string]any{}
	if input.DriverID != nil {
		updates["driver_id"] = *input.DriverID
	}
	if input.DepartureTime != nil {
		updates["departure_time"] = *input.DepartureTime
	}
	if len(updates) == 0 {
		return nil, errors.New("no fields to update")
	}
	if err := s.db.Model(&models.Shift{}).Where("id = ?", shiftID).Updates(updates).Error; err != nil {
		return nil, err
	}

	var shift models.Shift
	if err := s.db.Preload("Driver").First(&shift, shiftID).Error; err != nil {
		return nil, err
	}
	return &shift, nil
}

func (s *AdminService) AssignStudent(shiftID, requestID uint) (AssignStudentResult, error) {
	return s.assigner.AssignStudentToShift(context.Background(), shiftID, requestID)
}

func (s *AdminService) RemoveStudent(shiftID, requestID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("shift_id = ? AND request_id = ?", shiftID, requestID).Delete(&models.ShiftRequest{}).Error; err != nil {
			return err
		}
		return tx.Table("requests").Where("id = ?", requestID).Update("status", models.RequestStatusPending).Error
	})
}

func (s *AdminService) AssignStaff(shiftID, staffID uint) error {
	var user models.User
	if err := s.db.First(&user, staffID).Error; err != nil {
		return err
	}
	if user.Role != models.UserRoleStaff && user.Role != models.UserRoleAdmin {
		return errors.New("user is not staff")
	}
	return s.db.Table("shift_staffs").Create(map[string]any{"shift_id": shiftID, "staff_id": staffID}).Error
}

func (s *AdminService) RemoveStaff(shiftID, staffID uint) error {
	return s.db.Where("shift_id = ? AND staff_id = ?", shiftID, staffID).Delete(&models.ShiftStaff{}).Error
}

func (s *AdminService) SetUserStaff(userID uint) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	if user.Role == models.UserRoleAdmin {
		return nil, errors.New("cannot change admin role")
	}
	user.Role = models.UserRoleStaff
	if err := s.db.Model(&models.User{}).Where("id = ?", userID).Update("role", user.Role).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *AdminService) UnsetUserStaff(userID uint) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	if user.Role == models.UserRoleAdmin {
		return nil, errors.New("cannot change admin role")
	}
	user.Role = models.UserRoleStudent
	if err := s.db.Model(&models.User{}).Where("id = ?", userID).Update("role", user.Role).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *AdminService) PublishShift(shiftID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Shift{}).Where("id = ?", shiftID).Update("status", models.ShiftStatusPublished).Error; err != nil {
			return err
		}
		return tx.Table("requests").
			Where("id IN (SELECT request_id FROM shift_requests WHERE shift_id = ?)", shiftID).
			Update("status", models.RequestStatusPublished).Error
	})
}
