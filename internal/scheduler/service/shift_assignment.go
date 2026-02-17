package service

import (
	"context"
	"errors"

	"pickup/internal/scheduler/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrRequestNotPending = errors.New("request status is not pending")
	ErrShiftNotFound     = errors.New("shift not found")
	ErrRequestNotFound   = errors.New("request not found")
)

type AssignStudentResult struct {
	Warning string `json:"warning,omitempty"`
}

type ShiftAssignmentService struct {
	db *gorm.DB
}

func NewShiftAssignmentService(db *gorm.DB) *ShiftAssignmentService {
	return &ShiftAssignmentService{db: db}
}

// AssignStudentToShift 将学生需求加入班次。
// 核心保障：
// 1. 锁定 Shift + Request 行（FOR UPDATE）
// 2. 校验 Request 为 pending
// 3. 原子写入 shift_requests + Request.status=assigned
// 4. 若软超载返回 warning=capacity_overload，但仍提交事务
func (s *ShiftAssignmentService) AssignStudentToShift(ctx context.Context, shiftID, requestID uint) (AssignStudentResult, error) {
	result := AssignStudentResult{}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var shift models.Shift
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("Driver").
			First(&shift, shiftID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrShiftNotFound
			}
			return err
		}

		var req models.Request
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&req, requestID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrRequestNotFound
			}
			return err
		}

		if req.Status != models.RequestStatusPending {
			return ErrRequestNotPending
		}

		var boundRequestCount int64
		if err := tx.Model(&models.ShiftRequest{}).
			Where("shift_id = ?", shiftID).
			Count(&boundRequestCount).Error; err != nil {
			return err
		}

		var staffCount int64
		if err := tx.Model(&models.ShiftStaff{}).
			Where("shift_id = ?", shiftID).
			Count(&staffCount).Error; err != nil {
			return err
		}

		type baggageAggregate struct {
			Checked int64
			CarryOn int64
		}
		var aggregate baggageAggregate
		if err := tx.Table("shift_requests sr").
			Select("COALESCE(SUM(r.checked_bags), 0) AS checked, COALESCE(SUM(r.carry_on_bags), 0) AS carry_on").
			Joins("JOIN requests r ON r.id = sr.request_id").
			Where("sr.shift_id = ?", shiftID).
			Scan(&aggregate).Error; err != nil {
			return err
		}

		totalSeats := int(boundRequestCount+1) + int(staffCount)
		totalChecked := int(aggregate.Checked) + req.CheckedBags
		totalCarryOn := int(aggregate.CarryOn) + req.CarryOnBags

		if totalSeats > shift.Driver.MaxSeats || totalChecked > shift.Driver.MaxChecked || totalCarryOn > shift.Driver.MaxCarryOn {
			result.Warning = "capacity_overload"
		}

		if err := tx.Table("shift_requests").Create(map[string]any{
			"shift_id":   shiftID,
			"request_id": requestID,
		}).Error; err != nil {
			return err
		}

		if err := tx.Table("requests").
			Where("id = ?", requestID).
			Update("status", models.RequestStatusAssigned).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return AssignStudentResult{}, err
	}
	return result, nil
}
