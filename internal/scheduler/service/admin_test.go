package service

import (
	"testing"
	"time"

	"pickup/internal/scheduler/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm/clause"
)

func TestAdminService_CoreFlows(t *testing.T) {
	db := newTestDB(t)
	assigner := NewShiftAssignmentService(db)
	svc := NewAdminService(db, assigner)

	driver, err := svc.CreateDriver(DriverDTO{Name: "d1", CarModel: "SUV", MaxSeats: 4, MaxChecked: 4, MaxCarryOn: 4})
	require.NoError(t, err)
	list, err := svc.ListDrivers()
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, driver.ID, list[0].ID)

	shift, err := svc.CreateShift(driver.ID, time.Now())
	require.NoError(t, err)

	req := models.Request{UserID: 1, FlightNo: "AA1", ArrivalDate: time.Now(), Terminal: "T1", Status: models.RequestStatusPending}
	require.NoError(t, db.Omit(clause.Associations).Create(&req).Error)
	require.NoError(t, db.Table("shift_requests").Create(map[string]any{"shift_id": shift.ID, "request_id": req.ID}).Error)

	require.NoError(t, svc.PublishShift(shift.ID))
	var updatedReq models.Request
	require.NoError(t, db.First(&updatedReq, req.ID).Error)
	assert.Equal(t, models.RequestStatusPublished, updatedReq.Status)

	require.NoError(t, svc.RemoveStudent(shift.ID, req.ID))
	require.NoError(t, db.First(&updatedReq, req.ID).Error)
	assert.Equal(t, models.RequestStatusPending, updatedReq.Status)

	pending, err := svc.PendingRequests()
	require.NoError(t, err)
	require.Len(t, pending, 1)

	dashboard, err := svc.DashboardShifts()
	require.NoError(t, err)
	require.Len(t, dashboard, 1)
	assert.Equal(t, driver.ID, dashboard[0].Driver.ID)
}

func TestAdminService_AssignStaff_EdgeCases(t *testing.T) {
	db := newTestDB(t)
	svc := NewAdminService(db, NewShiftAssignmentService(db))

	driver := models.Driver{Name: "d", CarModel: "SUV", MaxSeats: 4, MaxChecked: 4, MaxCarryOn: 4}
	require.NoError(t, db.Create(&driver).Error)
	shift := models.Shift{DriverID: driver.ID, DepartureTime: time.Now(), Status: models.ShiftStatusDraft}
	require.NoError(t, db.Create(&shift).Error)

	student := models.User{OpenID: "u-student", Name: "stu", Role: models.UserRoleStudent}
	require.NoError(t, db.Create(&student).Error)
	err := svc.AssignStaff(shift.ID, student.ID)
	assert.ErrorContains(t, err, "user is not staff")

	staff := models.User{OpenID: "u-staff", Name: "staff", Role: models.UserRoleStaff}
	require.NoError(t, db.Create(&staff).Error)
	require.NoError(t, svc.AssignStaff(shift.ID, staff.ID))
	require.NoError(t, svc.RemoveStaff(shift.ID, staff.ID))
}

func TestAdminService_UpdateDriverAndShift(t *testing.T) {
	db := newTestDB(t)
	svc := NewAdminService(db, NewShiftAssignmentService(db))

	driver, err := svc.CreateDriver(DriverDTO{Name: "d1", CarModel: "SUV", MaxSeats: 4, MaxChecked: 4, MaxCarryOn: 4})
	require.NoError(t, err)

	updatedDriver, err := svc.UpdateDriver(driver.ID, DriverDTO{Name: "d2", CarModel: "Van", MaxSeats: 6, MaxChecked: 6, MaxCarryOn: 6})
	require.NoError(t, err)
	assert.Equal(t, "d2", updatedDriver.Name)
	assert.Equal(t, "Van", updatedDriver.CarModel)

	shift, err := svc.CreateShift(driver.ID, time.Now())
	require.NoError(t, err)

	newDriver, err := svc.CreateDriver(DriverDTO{Name: "d3", CarModel: "Sedan", MaxSeats: 3, MaxChecked: 2, MaxCarryOn: 2})
	require.NoError(t, err)

	newTime := time.Now().Add(time.Hour)
	updatedShift, err := svc.UpdateShift(shift.ID, ShiftUpdateDTO{DriverID: &newDriver.ID, DepartureTime: &newTime})
	require.NoError(t, err)
	assert.Equal(t, newDriver.ID, updatedShift.DriverID)

	_, err = svc.UpdateShift(shift.ID, ShiftUpdateDTO{})
	assert.ErrorContains(t, err, "no fields to update")
}

func TestAdminService_UserRoleManagement(t *testing.T) {
	db := newTestDB(t)
	svc := NewAdminService(db, NewShiftAssignmentService(db))

	student := models.User{OpenID: "u-stu", Name: "stu", Role: models.UserRoleStudent}
	admin := models.User{OpenID: "u-admin", Name: "adm", Role: models.UserRoleAdmin}
	require.NoError(t, db.Create(&student).Error)
	require.NoError(t, db.Create(&admin).Error)

	users, err := svc.ListUsers()
	require.NoError(t, err)
	require.Len(t, users, 2)

	updated, err := svc.SetUserStaff(student.ID)
	require.NoError(t, err)
	assert.Equal(t, models.UserRoleStaff, updated.Role)

	updated, err = svc.UnsetUserStaff(student.ID)
	require.NoError(t, err)
	assert.Equal(t, models.UserRoleStudent, updated.Role)

	_, err = svc.SetUserStaff(admin.ID)
	assert.ErrorContains(t, err, "cannot change admin role")

	_, err = svc.UnsetUserStaff(admin.ID)
	assert.ErrorContains(t, err, "cannot change admin role")
}
