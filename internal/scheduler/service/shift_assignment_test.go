package service

import (
	"context"
	"testing"
	"time"

	"pickup/internal/scheduler/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm/clause"
)

func TestShiftAssignmentService_AssignStudentToShift_Errors(t *testing.T) {
	db := newTestDB(t)
	svc := NewShiftAssignmentService(db)

	_, err := svc.AssignStudentToShift(context.Background(), 999, 1)
	assert.ErrorIs(t, err, ErrShiftNotFound)

	driver := models.Driver{Name: "d", CarModel: "SUV", MaxSeats: 4, MaxChecked: 4, MaxCarryOn: 4}
	require.NoError(t, db.Create(&driver).Error)
	shift := models.Shift{DriverID: driver.ID, DepartureTime: time.Now(), Status: models.ShiftStatusDraft}
	require.NoError(t, db.Create(&shift).Error)

	_, err = svc.AssignStudentToShift(context.Background(), shift.ID, 999)
	assert.ErrorIs(t, err, ErrRequestNotFound)

	req := models.Request{UserID: 1, FlightNo: "AA1", ArrivalDate: time.Now(), Terminal: "T1", Status: models.RequestStatusAssigned}
	require.NoError(t, db.Omit(clause.Associations).Create(&req).Error)
	_, err = svc.AssignStudentToShift(context.Background(), shift.ID, req.ID)
	assert.ErrorIs(t, err, ErrRequestNotPending)
}

func TestShiftAssignmentService_AssignStudentToShift_SuccessAndOverload(t *testing.T) {
	db := newTestDB(t)
	svc := NewShiftAssignmentService(db)

	driver := models.Driver{Name: "d", CarModel: "SUV", MaxSeats: 1, MaxChecked: 1, MaxCarryOn: 1}
	require.NoError(t, db.Create(&driver).Error)
	shift := models.Shift{DriverID: driver.ID, DepartureTime: time.Now(), Status: models.ShiftStatusDraft}
	require.NoError(t, db.Create(&shift).Error)

	staff := models.User{OpenID: "staff-openid", Name: "staff", Role: models.UserRoleStaff}
	require.NoError(t, db.Create(&staff).Error)
	require.NoError(t, db.Table("shift_staffs").Create(map[string]any{"shift_id": shift.ID, "staff_id": staff.ID}).Error)

	req := models.Request{UserID: 1, FlightNo: "AA1", ArrivalDate: time.Now(), Terminal: "T1", Status: models.RequestStatusPending, CheckedBags: 1, CarryOnBags: 1}
	require.NoError(t, db.Omit(clause.Associations).Create(&req).Error)

	res, err := svc.AssignStudentToShift(context.Background(), shift.ID, req.ID)
	require.NoError(t, err)
	assert.Equal(t, "capacity_overload", res.Warning)

	var saved models.ShiftRequest
	require.NoError(t, db.Where("shift_id = ? AND request_id = ?", shift.ID, req.ID).First(&saved).Error)
	var updated models.Request
	require.NoError(t, db.First(&updated, req.ID).Error)
	assert.Equal(t, models.RequestStatusAssigned, updated.Status)
}
