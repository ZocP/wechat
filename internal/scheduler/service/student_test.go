package service

import (
	"testing"
	"time"

	"pickup/internal/scheduler/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm/clause"
)

func TestStudentService_CreateRequest_Success(t *testing.T) {
	db := newTestDB(t)
	svc := NewStudentService(db)

	res, err := svc.CreateRequest(1, CreateRequestInput{
		FlightNo:            "AA101",
		ArrivalDate:         "2026-03-01",
		Terminal:            "T5",
		CheckedBags:         2,
		CarryOnBags:         1,
		ExpectedArrivalTime: "2026-03-01 10:30:00",
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, models.RequestStatusPending, res.Status)
	assert.Equal(t, 90, res.PickupBuffer)
	require.NotNil(t, res.ArrivalTimeAPI)
	require.NotNil(t, res.CalcPickupTime)
	assert.Equal(t, res.ArrivalTimeAPI.Add(90*time.Minute), *res.CalcPickupTime)
}

func TestStudentService_CreateRequest_InvalidInput(t *testing.T) {
	db := newTestDB(t)
	svc := NewStudentService(db)

	_, err := svc.CreateRequest(1, CreateRequestInput{
		FlightNo:            "AA101",
		ArrivalDate:         "bad-date",
		Terminal:            "T1",
		ExpectedArrivalTime: "2026-03-01 10:30:00",
	})
	assert.Error(t, err)

	_, err = svc.CreateRequest(1, CreateRequestInput{
		FlightNo:            "AA101",
		ArrivalDate:         "2026-03-01",
		Terminal:            "T1",
		ExpectedArrivalTime: "bad-time",
	})
	assert.Error(t, err)
}

func TestStudentService_CreateRequest_OnlyOncePerUser(t *testing.T) {
	db := newTestDB(t)
	svc := NewStudentService(db)

	_, err := svc.CreateRequest(7, CreateRequestInput{
		FlightNo:            "AA101",
		ArrivalDate:         "2026-03-01",
		Terminal:            "T1",
		ExpectedArrivalTime: "2026-03-01 10:30:00",
	})
	require.NoError(t, err)

	_, err = svc.CreateRequest(7, CreateRequestInput{
		FlightNo:            "AA102",
		ArrivalDate:         "2026-03-02",
		Terminal:            "T5",
		ExpectedArrivalTime: "2026-03-02 11:30:00",
	})
	assert.ErrorContains(t, err, "user already has a request")
}

func TestStudentService_UpdatePendingRequest_EdgeCases(t *testing.T) {
	db := newTestDB(t)
	svc := NewStudentService(db)

	arrival := time.Date(2026, 3, 1, 10, 30, 0, 0, time.UTC)
	pickup := arrival.Add(45 * time.Minute)
	req := models.Request{
		UserID:         1,
		FlightNo:       "AA101",
		ArrivalDate:    time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		Terminal:       "T1",
		Status:         models.RequestStatusPending,
		ArrivalTimeAPI: &arrival,
		PickupBuffer:   45,
		CalcPickupTime: &pickup,
	}
	require.NoError(t, db.Omit(clause.Associations).Create(&req).Error)

	timeStr := "2026-03-01 11:00:00"
	terminal := "T5"
	updated, err := svc.UpdatePendingRequest(1, req.ID, UpdateRequestInput{
		Terminal:            &terminal,
		ExpectedArrivalTime: &timeStr,
	})
	require.NoError(t, err)
	assert.Equal(t, 90, updated.PickupBuffer)
	require.NotNil(t, updated.ArrivalTimeAPI)
	require.NotNil(t, updated.CalcPickupTime)
	assert.Equal(t, updated.ArrivalTimeAPI.Add(90*time.Minute), *updated.CalcPickupTime)

	req2 := models.Request{UserID: 1, FlightNo: "AA102", ArrivalDate: time.Now(), Terminal: "T1", Status: models.RequestStatusAssigned}
	require.NoError(t, db.Omit(clause.Associations).Create(&req2).Error)
	_, err = svc.UpdatePendingRequest(1, req2.ID, UpdateRequestInput{})
	assert.ErrorContains(t, err, "only pending request can be updated")
}

func TestStudentService_ListMyRequests_HideShiftForNonPublished(t *testing.T) {
	db := newTestDB(t)
	svc := NewStudentService(db)

	driver := models.Driver{Name: "d1", CarModel: "SUV", MaxSeats: 4, MaxChecked: 4, MaxCarryOn: 4}
	require.NoError(t, db.Create(&driver).Error)
	shift := models.Shift{DriverID: driver.ID, DepartureTime: time.Now(), Status: models.ShiftStatusDraft}
	require.NoError(t, db.Create(&shift).Error)

	r1 := models.Request{UserID: 99, FlightNo: "AA1", ArrivalDate: time.Now(), Terminal: "T1", Status: models.RequestStatusPending}
	r2 := models.Request{UserID: 99, FlightNo: "AA2", ArrivalDate: time.Now(), Terminal: "T1", Status: models.RequestStatusPublished}
	require.NoError(t, db.Omit(clause.Associations).Create(&r1).Error)
	require.NoError(t, db.Omit(clause.Associations).Create(&r2).Error)
	require.NoError(t, db.Table("shift_requests").Create(map[string]any{"shift_id": shift.ID, "request_id": r1.ID}).Error)
	require.NoError(t, db.Table("shift_requests").Create(map[string]any{"shift_id": shift.ID, "request_id": r2.ID}).Error)

	list, err := svc.ListMyRequests(99)
	require.NoError(t, err)
	require.Len(t, list, 2)

	for _, item := range list {
		if item.Status == models.RequestStatusPublished {
			require.NotNil(t, item.Shift)
			assert.Equal(t, driver.ID, item.Shift.Driver.ID)
		} else {
			assert.Nil(t, item.Shift)
		}
	}
}
