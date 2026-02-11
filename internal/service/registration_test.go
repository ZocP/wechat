package service

import (
	"errors"
	"testing"

	"pickup/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// ===== Registration Service Tests =====

func newTestRegistrationService(repo *mockRegistrationRepo) RegistrationService {
	return NewRegistrationService(repo, zap.NewNop())
}

func TestRegistrationService_CreateRegistration_Success(t *testing.T) {
	repo := new(mockRegistrationRepo)
	svc := newTestRegistrationService(repo)

	repo.On("Create", mock.AnythingOfType("*model.Registration")).Return(nil).Once()

	req := &CreateRegistrationRequest{
		Name:          "张三",
		Phone:         "13800138000",
		WechatID:      "wxid_test",
		FlightNo:      "CA1234",
		ArrivalDate:   "2024-06-15",
		ArrivalTime:   "14:30",
		DepartureCity: "北京",
		Companions:    2,
		LuggageCount:  3,
		PickupMethod:  "group",
		Notes:         "测试备注",
	}

	reg, err := svc.CreateRegistration(1, req)
	require.NoError(t, err)
	assert.NotNil(t, reg)
	assert.Equal(t, "张三", reg.Name)
	assert.Equal(t, "13800138000", reg.Phone)
	assert.Equal(t, "CA1234", reg.FlightNo)
	assert.Equal(t, "14:30:00", reg.ArrivalTime)
	assert.Equal(t, model.PickupMethodGroup, reg.PickupMethod)
	assert.Equal(t, model.RegistrationStatusDraft, reg.Status)
	assert.Equal(t, uint(1), reg.UserID)
	repo.AssertExpectations(t)
}

func TestRegistrationService_CreateRegistration_InvalidDate(t *testing.T) {
	repo := new(mockRegistrationRepo)
	svc := newTestRegistrationService(repo)

	req := &CreateRegistrationRequest{
		Name:          "张三",
		Phone:         "13800138000",
		FlightNo:      "CA1234",
		ArrivalDate:   "invalid-date",
		ArrivalTime:   "14:30",
		DepartureCity: "北京",
		PickupMethod:  "group",
	}

	reg, err := svc.CreateRegistration(1, req)
	assert.Error(t, err)
	assert.Nil(t, reg)
	assert.Contains(t, err.Error(), "日期格式")
}

func TestRegistrationService_CreateRegistration_InvalidTime(t *testing.T) {
	repo := new(mockRegistrationRepo)
	svc := newTestRegistrationService(repo)

	req := &CreateRegistrationRequest{
		Name:          "张三",
		Phone:         "13800138000",
		FlightNo:      "CA1234",
		ArrivalDate:   "2024-06-15",
		ArrivalTime:   "25:99", // Invalid time
		DepartureCity: "北京",
		PickupMethod:  "group",
	}

	reg, err := svc.CreateRegistration(1, req)
	assert.Error(t, err)
	assert.Nil(t, reg)
	assert.Contains(t, err.Error(), "时间格式")
}

func TestRegistrationService_CreateRegistration_RepoError(t *testing.T) {
	repo := new(mockRegistrationRepo)
	svc := newTestRegistrationService(repo)

	repo.On("Create", mock.Anything).Return(errors.New("db error")).Once()

	req := &CreateRegistrationRequest{
		Name:          "张三",
		Phone:         "13800138000",
		FlightNo:      "CA1234",
		ArrivalDate:   "2024-06-15",
		ArrivalTime:   "14:30",
		DepartureCity: "北京",
		PickupMethod:  "group",
	}

	reg, err := svc.CreateRegistration(1, req)
	assert.Error(t, err)
	assert.Nil(t, reg)
	assert.Contains(t, err.Error(), "创建报名失败")
}

func TestRegistrationService_GetRegistration_Success(t *testing.T) {
	repo := new(mockRegistrationRepo)
	svc := newTestRegistrationService(repo)

	expected := &model.Registration{ID: 1, UserID: 1, Name: "张三"}
	repo.On("GetByID", uint(1)).Return(expected, nil).Once()

	reg, err := svc.GetRegistration(1, 1)
	require.NoError(t, err)
	assert.Equal(t, expected, reg)
}

func TestRegistrationService_GetRegistration_WrongUser(t *testing.T) {
	repo := new(mockRegistrationRepo)
	svc := newTestRegistrationService(repo)

	existing := &model.Registration{ID: 1, UserID: 2}
	repo.On("GetByID", uint(1)).Return(existing, nil).Once()

	reg, err := svc.GetRegistration(1, 1) // User 1 tries to access user 2's reg
	assert.Error(t, err)
	assert.Nil(t, reg)
	assert.Contains(t, err.Error(), "无权限")
}

func TestRegistrationService_GetRegistration_NotFound(t *testing.T) {
	repo := new(mockRegistrationRepo)
	svc := newTestRegistrationService(repo)

	repo.On("GetByID", uint(999)).Return(nil, errors.New("not found")).Once()

	reg, err := svc.GetRegistration(999, 1)
	assert.Error(t, err)
	assert.Nil(t, reg)
}

func TestRegistrationService_UpdateRegistration_Success(t *testing.T) {
	repo := new(mockRegistrationRepo)
	svc := newTestRegistrationService(repo)

	existing := &model.Registration{
		ID:     1,
		UserID: 1,
		Name:   "张三",
		Status: model.RegistrationStatusDraft,
	}
	repo.On("GetByID", uint(1)).Return(existing, nil).Once()
	repo.On("Update", mock.AnythingOfType("*model.Registration")).Return(nil).Once()

	newName := "李四"
	req := &UpdateRegistrationRequest{Name: &newName}

	reg, err := svc.UpdateRegistration(1, 1, req)
	require.NoError(t, err)
	assert.Equal(t, "李四", reg.Name)
}

func TestRegistrationService_UpdateRegistration_WrongUser(t *testing.T) {
	repo := new(mockRegistrationRepo)
	svc := newTestRegistrationService(repo)

	existing := &model.Registration{ID: 1, UserID: 2, Status: model.RegistrationStatusDraft}
	repo.On("GetByID", uint(1)).Return(existing, nil).Once()

	newName := "李四"
	req := &UpdateRegistrationRequest{Name: &newName}

	reg, err := svc.UpdateRegistration(1, 1, req)
	assert.Error(t, err)
	assert.Nil(t, reg)
	assert.Contains(t, err.Error(), "无权限")
}

func TestRegistrationService_UpdateRegistration_NotDraft(t *testing.T) {
	repo := new(mockRegistrationRepo)
	svc := newTestRegistrationService(repo)

	existing := &model.Registration{
		ID:     1,
		UserID: 1,
		Status: model.RegistrationStatusSubmitted, // Not draft
	}
	repo.On("GetByID", uint(1)).Return(existing, nil).Once()

	newName := "李四"
	req := &UpdateRegistrationRequest{Name: &newName}

	reg, err := svc.UpdateRegistration(1, 1, req)
	assert.Error(t, err)
	assert.Nil(t, reg)
	assert.Contains(t, err.Error(), "草稿状态")
}

func TestRegistrationService_UpdateRegistration_AllFields(t *testing.T) {
	repo := new(mockRegistrationRepo)
	svc := newTestRegistrationService(repo)

	existing := &model.Registration{
		ID:     1,
		UserID: 1,
		Status: model.RegistrationStatusDraft,
	}
	repo.On("GetByID", uint(1)).Return(existing, nil).Once()
	repo.On("Update", mock.AnythingOfType("*model.Registration")).Return(nil).Once()

	name := "李四"
	phone := "13900139000"
	wechatID := "wxid_new"
	flightNo := "MU5678"
	arrivalDate := "2024-07-20"
	arrivalTime := "16:00"
	departureCity := "上海"
	companions := 3
	luggageCount := 5
	pickupMethod := "private"
	notes := "新备注"

	req := &UpdateRegistrationRequest{
		Name:          &name,
		Phone:         &phone,
		WechatID:      &wechatID,
		FlightNo:      &flightNo,
		ArrivalDate:   &arrivalDate,
		ArrivalTime:   &arrivalTime,
		DepartureCity: &departureCity,
		Companions:    &companions,
		LuggageCount:  &luggageCount,
		PickupMethod:  &pickupMethod,
		Notes:         &notes,
	}

	reg, err := svc.UpdateRegistration(1, 1, req)
	require.NoError(t, err)
	assert.Equal(t, "李四", reg.Name)
	assert.Equal(t, "13900139000", reg.Phone)
	assert.Equal(t, "MU5678", reg.FlightNo)
	assert.Equal(t, "16:00:00", reg.ArrivalTime)
	assert.Equal(t, model.PickupMethodPrivate, reg.PickupMethod)
	assert.Equal(t, 3, reg.Companions)
	assert.Equal(t, 5, reg.LuggageCount)
}

func TestRegistrationService_UpdateRegistration_InvalidDate(t *testing.T) {
	repo := new(mockRegistrationRepo)
	svc := newTestRegistrationService(repo)

	existing := &model.Registration{ID: 1, UserID: 1, Status: model.RegistrationStatusDraft}
	repo.On("GetByID", uint(1)).Return(existing, nil).Once()

	badDate := "not-a-date"
	req := &UpdateRegistrationRequest{ArrivalDate: &badDate}

	reg, err := svc.UpdateRegistration(1, 1, req)
	assert.Error(t, err)
	assert.Nil(t, reg)
	assert.Contains(t, err.Error(), "日期格式")
}

func TestRegistrationService_UpdateRegistration_InvalidTime(t *testing.T) {
	repo := new(mockRegistrationRepo)
	svc := newTestRegistrationService(repo)

	existing := &model.Registration{ID: 1, UserID: 1, Status: model.RegistrationStatusDraft}
	repo.On("GetByID", uint(1)).Return(existing, nil).Once()

	badTime := "99:99"
	req := &UpdateRegistrationRequest{ArrivalTime: &badTime}

	reg, err := svc.UpdateRegistration(1, 1, req)
	assert.Error(t, err)
	assert.Nil(t, reg)
	assert.Contains(t, err.Error(), "时间格式")
}

func TestRegistrationService_GetUserRegistrations_Success(t *testing.T) {
	repo := new(mockRegistrationRepo)
	svc := newTestRegistrationService(repo)

	expected := []*model.Registration{{ID: 1, UserID: 1}, {ID: 2, UserID: 1}}
	repo.On("GetByUserID", uint(1)).Return(expected, nil).Once()

	regs, err := svc.GetUserRegistrations(1)
	require.NoError(t, err)
	assert.Len(t, regs, 2)
}

func TestRegistrationService_GetUserRegistrations_Error(t *testing.T) {
	repo := new(mockRegistrationRepo)
	svc := newTestRegistrationService(repo)

	repo.On("GetByUserID", uint(1)).Return(nil, errors.New("db error")).Once()

	regs, err := svc.GetUserRegistrations(1)
	assert.Error(t, err)
	assert.Nil(t, regs)
}

func TestRegistrationService_DeleteRegistration_Success(t *testing.T) {
	repo := new(mockRegistrationRepo)
	svc := newTestRegistrationService(repo)

	existing := &model.Registration{ID: 1, UserID: 1, Status: model.RegistrationStatusDraft}
	repo.On("GetByID", uint(1)).Return(existing, nil).Once()
	repo.On("Delete", uint(1)).Return(nil).Once()

	err := svc.DeleteRegistration(1, 1)
	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestRegistrationService_DeleteRegistration_WrongUser(t *testing.T) {
	repo := new(mockRegistrationRepo)
	svc := newTestRegistrationService(repo)

	existing := &model.Registration{ID: 1, UserID: 2, Status: model.RegistrationStatusDraft}
	repo.On("GetByID", uint(1)).Return(existing, nil).Once()

	err := svc.DeleteRegistration(1, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "无权限")
}

func TestRegistrationService_DeleteRegistration_NotDraft(t *testing.T) {
	repo := new(mockRegistrationRepo)
	svc := newTestRegistrationService(repo)

	existing := &model.Registration{ID: 1, UserID: 1, Status: model.RegistrationStatusSubmitted}
	repo.On("GetByID", uint(1)).Return(existing, nil).Once()

	err := svc.DeleteRegistration(1, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "草稿状态")
}

func TestRegistrationService_DeleteRegistration_NotFound(t *testing.T) {
	repo := new(mockRegistrationRepo)
	svc := newTestRegistrationService(repo)

	repo.On("GetByID", uint(999)).Return(nil, errors.New("not found")).Once()

	err := svc.DeleteRegistration(999, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "获取报名信息失败")
}

func TestRegistrationService_DeleteRegistration_RepoError(t *testing.T) {
	repo := new(mockRegistrationRepo)
	svc := newTestRegistrationService(repo)

	existing := &model.Registration{ID: 1, UserID: 1, Status: model.RegistrationStatusDraft}
	repo.On("GetByID", uint(1)).Return(existing, nil).Once()
	repo.On("Delete", uint(1)).Return(errors.New("db error")).Once()

	err := svc.DeleteRegistration(1, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "删除报名失败")
}

// ===== Validation Tests =====

func TestRegistrationService_CreateRegistration_NegativeCompanions(t *testing.T) {
	repo := new(mockRegistrationRepo)
	svc := newTestRegistrationService(repo)

	req := &CreateRegistrationRequest{
		Name:          "张三",
		Phone:         "13800138000",
		FlightNo:      "CA1234",
		ArrivalDate:   "2024-06-15",
		ArrivalTime:   "14:30",
		DepartureCity: "北京",
		Companions:    -1,
		LuggageCount:  1,
		PickupMethod:  "group",
	}

	reg, err := svc.CreateRegistration(1, req)
	assert.Error(t, err)
	assert.Nil(t, reg)
	assert.Contains(t, err.Error(), "负数")
}

func TestRegistrationService_CreateRegistration_NegativeLuggage(t *testing.T) {
	repo := new(mockRegistrationRepo)
	svc := newTestRegistrationService(repo)

	req := &CreateRegistrationRequest{
		Name:          "张三",
		Phone:         "13800138000",
		FlightNo:      "CA1234",
		ArrivalDate:   "2024-06-15",
		ArrivalTime:   "14:30",
		DepartureCity: "北京",
		Companions:    0,
		LuggageCount:  -5,
		PickupMethod:  "group",
	}

	reg, err := svc.CreateRegistration(1, req)
	assert.Error(t, err)
	assert.Nil(t, reg)
	assert.Contains(t, err.Error(), "负数")
}

func TestRegistrationService_CreateRegistration_InvalidPickupMethod(t *testing.T) {
	repo := new(mockRegistrationRepo)
	svc := newTestRegistrationService(repo)

	req := &CreateRegistrationRequest{
		Name:          "张三",
		Phone:         "13800138000",
		FlightNo:      "CA1234",
		ArrivalDate:   "2024-06-15",
		ArrivalTime:   "14:30",
		DepartureCity: "北京",
		Companions:    0,
		LuggageCount:  1,
		PickupMethod:  "invalid_method",
	}

	reg, err := svc.CreateRegistration(1, req)
	assert.Error(t, err)
	assert.Nil(t, reg)
	assert.Contains(t, err.Error(), "无效的接送方式")
}

func TestRegistrationService_UpdateRegistration_NegativeCompanions(t *testing.T) {
	repo := new(mockRegistrationRepo)
	svc := newTestRegistrationService(repo)

	existing := &model.Registration{ID: 1, UserID: 1, Status: model.RegistrationStatusDraft}
	repo.On("GetByID", uint(1)).Return(existing, nil).Once()

	neg := -1
	req := &UpdateRegistrationRequest{Companions: &neg}
	reg, err := svc.UpdateRegistration(1, 1, req)
	assert.Error(t, err)
	assert.Nil(t, reg)
	assert.Contains(t, err.Error(), "负数")
}

func TestRegistrationService_UpdateRegistration_InvalidPickupMethod(t *testing.T) {
	repo := new(mockRegistrationRepo)
	svc := newTestRegistrationService(repo)

	existing := &model.Registration{ID: 1, UserID: 1, Status: model.RegistrationStatusDraft}
	repo.On("GetByID", uint(1)).Return(existing, nil).Once()

	bad := "xyz"
	req := &UpdateRegistrationRequest{PickupMethod: &bad}
	reg, err := svc.UpdateRegistration(1, 1, req)
	assert.Error(t, err)
	assert.Nil(t, reg)
	assert.Contains(t, err.Error(), "无效的接送方式")
}

func TestIsValidPickupMethod(t *testing.T) {
	assert.True(t, isValidPickupMethod("group"))
	assert.True(t, isValidPickupMethod("private"))
	assert.True(t, isValidPickupMethod("shuttle"))
	assert.False(t, isValidPickupMethod(""))
	assert.False(t, isValidPickupMethod("invalid"))
	assert.False(t, isValidPickupMethod("GROUP")) // case-sensitive
}
