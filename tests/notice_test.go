package tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pickup/internal/handler"
	"pickup/internal/model"
	"pickup/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockNoticeService 模拟消息服务
type MockNoticeService struct {
	mock.Mock
}

func (m *MockNoticeService) CreateNotice(userID uint, req *service.CreateNoticeRequest) (*model.Notice, error) {
	args := m.Called(userID, req)
	var notice *model.Notice
	if v := args.Get(0); v != nil {
		notice = v.(*model.Notice)
	}
	return notice, args.Error(1)
}

func (m *MockNoticeService) GetNotice(id uint) (*model.Notice, error) {
	args := m.Called(id)
	var notice *model.Notice
	if v := args.Get(0); v != nil {
		notice = v.(*model.Notice)
	}
	return notice, args.Error(1)
}

func (m *MockNoticeService) GetVisibleNotices() ([]*model.Notice, error) {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*model.Notice), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockNoticeService) GetNoticesByFlightNo(flightNo string) ([]*model.Notice, error) {
	args := m.Called(flightNo)
	if v := args.Get(0); v != nil {
		return v.([]*model.Notice), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockNoticeService) UpdateNotice(id uint, userID uint, req *service.UpdateNoticeRequest) (*model.Notice, error) {
	args := m.Called(id, userID, req)
	var notice *model.Notice
	if v := args.Get(0); v != nil {
		notice = v.(*model.Notice)
	}
	return notice, args.Error(1)
}

func (m *MockNoticeService) DeleteNotice(id uint, userID uint) error {
	args := m.Called(id, userID)
	return args.Error(0)
}

func setupNoticeRouter(service service.NoticeService, withUser bool) *gin.Engine {
	router := gin.New()
	api := router.Group("/api/v1")
	if withUser {
		api.Use(func(c *gin.Context) {
			c.Set("user_id", uint(1))
		})
	}
	handler.NewNoticeHandler(service, zap.NewNop()).RegisterRoutes(api)
	return router
}

func sampleCreateNoticeBody() map[string]interface{} {
	now := time.Now()
	return map[string]interface{}{
		"flight_no":       "CA1234",
		"terminal":        "T1",
		"pickup_batch":    "B1",
		"arrival_airport": "PEK",
		"meeting_point":   "Gate A",
		"visible_from":    now.Format(time.RFC3339),
		"visible_to":      now.Add(2 * time.Hour).Format(time.RFC3339),
	}
}

func TestCreateNotice_Success(t *testing.T) {
	mockService := new(MockNoticeService)
	mockService.On("CreateNotice", uint(1), mock.AnythingOfType("*service.CreateNoticeRequest")).Return(&model.Notice{ID: 1}, nil).Once()
	router := setupNoticeRouter(mockService, true)

	data, _ := json.Marshal(sampleCreateNoticeBody())
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/notices", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)
}

func TestCreateNotice_Unauthorized(t *testing.T) {
	mockService := new(MockNoticeService)
	router := setupNoticeRouter(mockService, false)

	data, _ := json.Marshal(sampleCreateNoticeBody())
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/notices", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockService.AssertExpectations(t)
}

func TestCreateNotice_InvalidBody(t *testing.T) {
	mockService := new(MockNoticeService)
	router := setupNoticeRouter(mockService, true)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/notices", bytes.NewReader([]byte(`{"flight_no":1}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestCreateNotice_ServiceError(t *testing.T) {
	mockService := new(MockNoticeService)
	mockService.On("CreateNotice", uint(1), mock.AnythingOfType("*service.CreateNoticeRequest")).Return((*model.Notice)(nil), errors.New("failed")).Once()
	router := setupNoticeRouter(mockService, true)

	data, _ := json.Marshal(sampleCreateNoticeBody())
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/notices", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetNotice_Success(t *testing.T) {
	mockService := new(MockNoticeService)
	mockService.On("GetNotice", uint(1)).Return(&model.Notice{ID: 1}, nil).Once()
	router := setupNoticeRouter(mockService, false)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/notices/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetNotice_InvalidID(t *testing.T) {
	mockService := new(MockNoticeService)
	router := setupNoticeRouter(mockService, false)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/notices/abc", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetNotice_ServiceError(t *testing.T) {
	mockService := new(MockNoticeService)
	mockService.On("GetNotice", uint(1)).Return((*model.Notice)(nil), errors.New("not found")).Once()
	router := setupNoticeRouter(mockService, false)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/notices/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetVisibleNotices_Success(t *testing.T) {
	mockService := new(MockNoticeService)
	mockService.On("GetVisibleNotices").Return([]*model.Notice{{ID: 1}}, nil).Once()
	router := setupNoticeRouter(mockService, false)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/notices", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetVisibleNotices_ServiceError(t *testing.T) {
	mockService := new(MockNoticeService)
	mockService.On("GetVisibleNotices").Return(([]*model.Notice)(nil), errors.New("failed")).Once()
	router := setupNoticeRouter(mockService, false)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/notices", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetNoticesByFlight_Success(t *testing.T) {
	mockService := new(MockNoticeService)
	mockService.On("GetNoticesByFlightNo", "CA1234").Return([]*model.Notice{{ID: 1}}, nil).Once()
	router := setupNoticeRouter(mockService, false)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/notices/flight/CA1234", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetNoticesByFlight_MissingFlight(t *testing.T) {
	mockService := new(MockNoticeService)
	noticeHandler := handler.NewNoticeHandler(mockService, zap.NewNop())
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "flight_no", Value: ""}}
	c.Request, _ = http.NewRequest(http.MethodGet, "/api/v1/notices/flight/", nil)

	noticeHandler.GetNoticesByFlightNo(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetNoticesByFlight_ServiceError(t *testing.T) {
	mockService := new(MockNoticeService)
	mockService.On("GetNoticesByFlightNo", "CA1234").Return(([]*model.Notice)(nil), errors.New("failed")).Once()
	router := setupNoticeRouter(mockService, false)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/notices/flight/CA1234", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestUpdateNotice_Success(t *testing.T) {
	mockService := new(MockNoticeService)
	mockService.On("UpdateNotice", uint(1), uint(1), mock.AnythingOfType("*service.UpdateNoticeRequest")).Return(&model.Notice{ID: 1}, nil).Once()
	router := setupNoticeRouter(mockService, true)

	body := map[string]string{"terminal": "T2"}
	data, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/admin/notices/1", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestUpdateNotice_InvalidID(t *testing.T) {
	mockService := new(MockNoticeService)
	router := setupNoticeRouter(mockService, true)

	req, _ := http.NewRequest(http.MethodPut, "/api/v1/admin/notices/abc", bytes.NewReader([]byte(`{"terminal":"T2"}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestUpdateNotice_InvalidBody(t *testing.T) {
	mockService := new(MockNoticeService)
	router := setupNoticeRouter(mockService, true)

	req, _ := http.NewRequest(http.MethodPut, "/api/v1/admin/notices/1", bytes.NewReader([]byte(`{"visible_from":"bad"}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestUpdateNotice_ServiceError(t *testing.T) {
	mockService := new(MockNoticeService)
	mockService.On("UpdateNotice", uint(1), uint(1), mock.AnythingOfType("*service.UpdateNoticeRequest")).Return((*model.Notice)(nil), errors.New("failed")).Once()
	router := setupNoticeRouter(mockService, true)

	body := map[string]string{"terminal": "T2"}
	data, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/admin/notices/1", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestUpdateNotice_Unauthorized(t *testing.T) {
	mockService := new(MockNoticeService)
	router := setupNoticeRouter(mockService, false)

	body := map[string]string{"terminal": "T2"}
	data, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/admin/notices/1", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockService.AssertExpectations(t)
}

func TestDeleteNotice_Success(t *testing.T) {
	mockService := new(MockNoticeService)
	mockService.On("DeleteNotice", uint(1), uint(1)).Return(nil).Once()
	router := setupNoticeRouter(mockService, true)

	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/admin/notices/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestDeleteNotice_InvalidID(t *testing.T) {
	mockService := new(MockNoticeService)
	router := setupNoticeRouter(mockService, true)

	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/admin/notices/abc", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestDeleteNotice_ServiceError(t *testing.T) {
	mockService := new(MockNoticeService)
	mockService.On("DeleteNotice", uint(1), uint(1)).Return(errors.New("failed")).Once()
	router := setupNoticeRouter(mockService, true)

	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/admin/notices/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestDeleteNotice_Unauthorized(t *testing.T) {
	mockService := new(MockNoticeService)
	router := setupNoticeRouter(mockService, false)

	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/admin/notices/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockService.AssertExpectations(t)
}
