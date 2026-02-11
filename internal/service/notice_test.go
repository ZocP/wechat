package service

import (
	"errors"
	"testing"
	"time"

	"pickup/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// ===== Mock Repositories =====

type mockNoticeRepo struct {
	mock.Mock
}

func (m *mockNoticeRepo) Create(notice *model.Notice) error {
	args := m.Called(notice)
	return args.Error(0)
}

func (m *mockNoticeRepo) GetByID(id uint) (*model.Notice, error) {
	args := m.Called(id)
	if v := args.Get(0); v != nil {
		return v.(*model.Notice), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockNoticeRepo) GetVisibleNotices() ([]*model.Notice, error) {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*model.Notice), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockNoticeRepo) GetByFlightNo(flightNo string) ([]*model.Notice, error) {
	args := m.Called(flightNo)
	if v := args.Get(0); v != nil {
		return v.([]*model.Notice), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockNoticeRepo) Update(notice *model.Notice) error {
	args := m.Called(notice)
	return args.Error(0)
}

func (m *mockNoticeRepo) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

// ===== Notice Service Tests =====

func newTestNoticeService(repo *mockNoticeRepo) NoticeService {
	return NewNoticeService(repo, zap.NewNop())
}

func TestNoticeService_CreateNotice_Success(t *testing.T) {
	repo := new(mockNoticeRepo)
	svc := newTestNoticeService(repo)

	repo.On("Create", mock.AnythingOfType("*model.Notice")).Return(nil).Once()

	now := time.Now()
	req := &CreateNoticeRequest{
		FlightNo:       "CA1234",
		Terminal:       "T1",
		PickupBatch:    "B1",
		ArrivalAirport: "PEK",
		MeetingPoint:   "Gate A",
		GuideText:      "Follow signs",
		MapURL:         "https://example.com/map",
		ContactName:    "张三",
		ContactPhone:   "13800138000",
		VisibleFrom:    now,
		VisibleTo:      now.Add(2 * time.Hour),
	}

	notice, err := svc.CreateNotice(1, req)
	require.NoError(t, err)
	assert.NotNil(t, notice)
	assert.Equal(t, "CA1234", notice.FlightNo)
	assert.Equal(t, "T1", notice.Terminal)
	assert.Equal(t, uint(1), notice.CreatedBy)
	repo.AssertExpectations(t)
}

func TestNoticeService_CreateNotice_RepoError(t *testing.T) {
	repo := new(mockNoticeRepo)
	svc := newTestNoticeService(repo)

	repo.On("Create", mock.Anything).Return(errors.New("db error")).Once()

	req := &CreateNoticeRequest{
		FlightNo:       "CA1234",
		Terminal:       "T1",
		PickupBatch:    "B1",
		ArrivalAirport: "PEK",
		MeetingPoint:   "Gate A",
		VisibleFrom:    time.Now(),
		VisibleTo:      time.Now().Add(time.Hour),
	}

	notice, err := svc.CreateNotice(1, req)
	assert.Error(t, err)
	assert.Nil(t, notice)
	assert.Contains(t, err.Error(), "创建消息失败")
	repo.AssertExpectations(t)
}

func TestNoticeService_GetNotice_Success(t *testing.T) {
	repo := new(mockNoticeRepo)
	svc := newTestNoticeService(repo)

	expected := &model.Notice{ID: 1, FlightNo: "CA1234"}
	repo.On("GetByID", uint(1)).Return(expected, nil).Once()

	notice, err := svc.GetNotice(1)
	require.NoError(t, err)
	assert.Equal(t, expected, notice)
	repo.AssertExpectations(t)
}

func TestNoticeService_GetNotice_NotFound(t *testing.T) {
	repo := new(mockNoticeRepo)
	svc := newTestNoticeService(repo)

	repo.On("GetByID", uint(999)).Return(nil, errors.New("not found")).Once()

	notice, err := svc.GetNotice(999)
	assert.Error(t, err)
	assert.Nil(t, notice)
	repo.AssertExpectations(t)
}

func TestNoticeService_GetVisibleNotices_Success(t *testing.T) {
	repo := new(mockNoticeRepo)
	svc := newTestNoticeService(repo)

	expected := []*model.Notice{{ID: 1}, {ID: 2}}
	repo.On("GetVisibleNotices").Return(expected, nil).Once()

	notices, err := svc.GetVisibleNotices()
	require.NoError(t, err)
	assert.Len(t, notices, 2)
	repo.AssertExpectations(t)
}

func TestNoticeService_GetVisibleNotices_Error(t *testing.T) {
	repo := new(mockNoticeRepo)
	svc := newTestNoticeService(repo)

	repo.On("GetVisibleNotices").Return(nil, errors.New("db error")).Once()

	notices, err := svc.GetVisibleNotices()
	assert.Error(t, err)
	assert.Nil(t, notices)
	repo.AssertExpectations(t)
}

func TestNoticeService_GetNoticesByFlightNo_Success(t *testing.T) {
	repo := new(mockNoticeRepo)
	svc := newTestNoticeService(repo)

	expected := []*model.Notice{{ID: 1, FlightNo: "CA1234"}}
	repo.On("GetByFlightNo", "CA1234").Return(expected, nil).Once()

	notices, err := svc.GetNoticesByFlightNo("CA1234")
	require.NoError(t, err)
	assert.Len(t, notices, 1)
	repo.AssertExpectations(t)
}

func TestNoticeService_UpdateNotice_Success(t *testing.T) {
	repo := new(mockNoticeRepo)
	svc := newTestNoticeService(repo)

	existing := &model.Notice{ID: 1, FlightNo: "CA1234", Terminal: "T1", CreatedBy: 1}
	repo.On("GetByID", uint(1)).Return(existing, nil).Once()
	repo.On("Update", mock.AnythingOfType("*model.Notice")).Return(nil).Once()

	newTerminal := "T2"
	req := &UpdateNoticeRequest{Terminal: &newTerminal}

	notice, err := svc.UpdateNotice(1, 1, req)
	require.NoError(t, err)
	assert.Equal(t, "T2", notice.Terminal)
	assert.Equal(t, "CA1234", notice.FlightNo) // Unchanged
	repo.AssertExpectations(t)
}

func TestNoticeService_UpdateNotice_Forbidden(t *testing.T) {
	repo := new(mockNoticeRepo)
	svc := newTestNoticeService(repo)

	existing := &model.Notice{ID: 1, CreatedBy: 1}
	repo.On("GetByID", uint(1)).Return(existing, nil).Once()

	newTerminal := "T2"
	req := &UpdateNoticeRequest{Terminal: &newTerminal}

	notice, err := svc.UpdateNotice(1, 999, req) // Different user
	assert.Error(t, err)
	assert.Nil(t, notice)
	assert.Contains(t, err.Error(), "无权限")
	repo.AssertExpectations(t)
}

func TestNoticeService_UpdateNotice_NotFound(t *testing.T) {
	repo := new(mockNoticeRepo)
	svc := newTestNoticeService(repo)

	repo.On("GetByID", uint(999)).Return(nil, errors.New("not found")).Once()

	req := &UpdateNoticeRequest{}
	notice, err := svc.UpdateNotice(999, 1, req)
	assert.Error(t, err)
	assert.Nil(t, notice)
	repo.AssertExpectations(t)
}

func TestNoticeService_UpdateNotice_AllFields(t *testing.T) {
	repo := new(mockNoticeRepo)
	svc := newTestNoticeService(repo)

	existing := &model.Notice{ID: 1, CreatedBy: 1}
	repo.On("GetByID", uint(1)).Return(existing, nil).Once()
	repo.On("Update", mock.AnythingOfType("*model.Notice")).Return(nil).Once()

	flightNo := "MU5678"
	terminal := "T2"
	pickupBatch := "B2"
	arrivalAirport := "SHA"
	meetingPoint := "Gate B"
	guideText := "New guide"
	mapURL := "https://new.map"
	contactName := "李四"
	contactPhone := "13900139000"
	visibleFrom := time.Now()
	visibleTo := time.Now().Add(3 * time.Hour)

	req := &UpdateNoticeRequest{
		FlightNo:       &flightNo,
		Terminal:       &terminal,
		PickupBatch:    &pickupBatch,
		ArrivalAirport: &arrivalAirport,
		MeetingPoint:   &meetingPoint,
		GuideText:      &guideText,
		MapURL:         &mapURL,
		ContactName:    &contactName,
		ContactPhone:   &contactPhone,
		VisibleFrom:    &visibleFrom,
		VisibleTo:      &visibleTo,
	}

	notice, err := svc.UpdateNotice(1, 1, req)
	require.NoError(t, err)
	assert.Equal(t, "MU5678", notice.FlightNo)
	assert.Equal(t, "T2", notice.Terminal)
	assert.Equal(t, "B2", notice.PickupBatch)
	assert.Equal(t, "SHA", notice.ArrivalAirport)
	assert.Equal(t, "Gate B", notice.MeetingPoint)
	assert.Equal(t, "New guide", notice.GuideText)
	assert.Equal(t, "https://new.map", notice.MapURL)
	assert.Equal(t, "李四", notice.ContactName)
	assert.Equal(t, "13900139000", notice.ContactPhone)
	repo.AssertExpectations(t)
}

func TestNoticeService_DeleteNotice_Success(t *testing.T) {
	repo := new(mockNoticeRepo)
	svc := newTestNoticeService(repo)

	existing := &model.Notice{ID: 1, CreatedBy: 1}
	repo.On("GetByID", uint(1)).Return(existing, nil).Once()
	repo.On("Delete", uint(1)).Return(nil).Once()

	err := svc.DeleteNotice(1, 1)
	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestNoticeService_DeleteNotice_Forbidden(t *testing.T) {
	repo := new(mockNoticeRepo)
	svc := newTestNoticeService(repo)

	existing := &model.Notice{ID: 1, CreatedBy: 1}
	repo.On("GetByID", uint(1)).Return(existing, nil).Once()

	err := svc.DeleteNotice(1, 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "无权限")
	repo.AssertExpectations(t)
}

func TestNoticeService_DeleteNotice_NotFound(t *testing.T) {
	repo := new(mockNoticeRepo)
	svc := newTestNoticeService(repo)

	repo.On("GetByID", uint(999)).Return(nil, errors.New("not found")).Once()

	err := svc.DeleteNotice(999, 1)
	assert.Error(t, err)
	repo.AssertExpectations(t)
}

func TestNoticeService_DeleteNotice_RepoError(t *testing.T) {
	repo := new(mockNoticeRepo)
	svc := newTestNoticeService(repo)

	existing := &model.Notice{ID: 1, CreatedBy: 1}
	repo.On("GetByID", uint(1)).Return(existing, nil).Once()
	repo.On("Delete", uint(1)).Return(errors.New("db error")).Once()

	err := svc.DeleteNotice(1, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "删除消息失败")
	repo.AssertExpectations(t)
}

func TestNoticeService_CreateNotice_InvalidTimeWindow(t *testing.T) {
	repo := new(mockNoticeRepo)
	svc := newTestNoticeService(repo)

	now := time.Now()
	req := &CreateNoticeRequest{
		FlightNo:       "CA1234",
		Terminal:       "T1",
		PickupBatch:    "B1",
		ArrivalAirport: "PEK",
		MeetingPoint:   "Gate A",
		VisibleFrom:    now.Add(2 * time.Hour),
		VisibleTo:      now, // VisibleTo before VisibleFrom
	}

	notice, err := svc.CreateNotice(1, req)
	assert.Error(t, err)
	assert.Nil(t, notice)
	assert.Contains(t, err.Error(), "结束可见时间必须晚于开始可见时间")
}

func TestNoticeService_CreateNotice_EqualTimeWindow(t *testing.T) {
	repo := new(mockNoticeRepo)
	svc := newTestNoticeService(repo)

	now := time.Now()
	req := &CreateNoticeRequest{
		FlightNo:       "CA1234",
		Terminal:       "T1",
		PickupBatch:    "B1",
		ArrivalAirport: "PEK",
		MeetingPoint:   "Gate A",
		VisibleFrom:    now,
		VisibleTo:      now, // Equal times should also fail
	}

	notice, err := svc.CreateNotice(1, req)
	assert.Error(t, err)
	assert.Nil(t, notice)
	assert.Contains(t, err.Error(), "结束可见时间必须晚于开始可见时间")
}
