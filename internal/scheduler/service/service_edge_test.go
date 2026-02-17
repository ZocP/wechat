package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pickup/internal/config"
	"pickup/internal/scheduler/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm/clause"
)

func TestAdminService_ErrorBranchesAndAssignStudent(t *testing.T) {
	t.Run("assign student delegate", func(t *testing.T) {
		db := newTestDB(t)
		svc := NewAdminService(db, NewShiftAssignmentService(db))

		driver := models.Driver{Name: "d", CarModel: "SUV", MaxSeats: 5, MaxChecked: 5, MaxCarryOn: 5}
		require.NoError(t, db.Create(&driver).Error)
		shift := models.Shift{DriverID: driver.ID, DepartureTime: time.Now(), Status: models.ShiftStatusDraft}
		require.NoError(t, db.Create(&shift).Error)
		req := models.Request{UserID: 1, FlightNo: "AA1", ArrivalDate: time.Now(), Terminal: "T1", Status: models.RequestStatusPending}
		require.NoError(t, db.Omit(clause.Associations).Create(&req).Error)

		res, err := svc.AssignStudent(shift.ID, req.ID)
		require.NoError(t, err)
		assert.Empty(t, res.Warning)
	})

	t.Run("create driver error", func(t *testing.T) {
		db := newTestDB(t)
		svc := NewAdminService(db, NewShiftAssignmentService(db))
		require.NoError(t, db.Exec("DROP TABLE drivers").Error)
		_, err := svc.CreateDriver(DriverDTO{Name: "d", CarModel: "SUV", MaxSeats: 1, MaxChecked: 1, MaxCarryOn: 1})
		assert.Error(t, err)
	})

	t.Run("create shift error", func(t *testing.T) {
		db := newTestDB(t)
		svc := NewAdminService(db, NewShiftAssignmentService(db))
		require.NoError(t, db.Exec("DROP TABLE shifts").Error)
		_, err := svc.CreateShift(1, time.Now())
		assert.Error(t, err)
	})

	t.Run("remove student error", func(t *testing.T) {
		db := newTestDB(t)
		svc := NewAdminService(db, NewShiftAssignmentService(db))
		require.NoError(t, db.Exec("DROP TABLE shift_requests").Error)
		err := svc.RemoveStudent(1, 1)
		assert.Error(t, err)
	})

	t.Run("publish shift error", func(t *testing.T) {
		db := newTestDB(t)
		svc := NewAdminService(db, NewShiftAssignmentService(db))
		require.NoError(t, db.Exec("DROP TABLE shifts").Error)
		err := svc.PublishShift(1)
		assert.Error(t, err)
	})

	t.Run("assign staff not found", func(t *testing.T) {
		db := newTestDB(t)
		svc := NewAdminService(db, NewShiftAssignmentService(db))
		err := svc.AssignStaff(1, 999)
		assert.Error(t, err)
	})
}

func TestAuthService_ErrorBranches(t *testing.T) {
	t.Run("wechat login errcode", func(t *testing.T) {
		db := newTestDB(t)
		mux := http.NewServeMux()
		mux.HandleFunc("/sns/jscode2session", func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]any{"errcode": 40029, "errmsg": "invalid code"})
		})
		server := httptest.NewServer(mux)
		defer server.Close()

		svc := NewAuthService(db, &config.WechatConfig{AppID: "a", AppSecret: "b"}, &config.JWTConfig{Secret: "s", ExpireTime: time.Hour, Issuer: "i"}, zap.NewNop())
		svc.wechatClient.SetBaseURL(server.URL)

		_, err := svc.LoginWithWechatCode("bad")
		assert.Error(t, err)
	})

	t.Run("login db error", func(t *testing.T) {
		db := newTestDB(t)
		mux := http.NewServeMux()
		mux.HandleFunc("/sns/jscode2session", func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]any{"openid": "oid", "session_key": "sk", "errcode": 0})
		})
		server := httptest.NewServer(mux)
		defer server.Close()

		svc := NewAuthService(db, &config.WechatConfig{AppID: "a", AppSecret: "b"}, &config.JWTConfig{Secret: "s", ExpireTime: time.Hour, Issuer: "i"}, zap.NewNop())
		svc.wechatClient.SetBaseURL(server.URL)
		require.NoError(t, db.Exec("DROP TABLE users").Error)

		_, err := svc.LoginWithWechatCode("code")
		assert.Error(t, err)
	})

	t.Run("bind phone token errcode", func(t *testing.T) {
		db := newTestDB(t)
		require.NoError(t, db.Create(&models.User{OpenID: "oid-x", Name: "u", Role: models.UserRoleStudent}).Error)

		mux := http.NewServeMux()
		mux.HandleFunc("/cgi-bin/token", func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]any{"errcode": 40164, "errmsg": "ip not in whitelist"})
		})
		server := httptest.NewServer(mux)
		defer server.Close()

		svc := NewAuthService(db, &config.WechatConfig{AppID: "a", AppSecret: "b"}, &config.JWTConfig{Secret: "s", ExpireTime: time.Hour, Issuer: "i"}, zap.NewNop())
		svc.wechatClient.SetBaseURL(server.URL)

		var user models.User
		require.NoError(t, db.Where("open_id = ?", "oid-x").First(&user).Error)
		err := svc.BindPhone(user.ID, "code")
		assert.Error(t, err)
	})

	t.Run("bind phone api errcode", func(t *testing.T) {
		db := newTestDB(t)
		require.NoError(t, db.Create(&models.User{OpenID: "oid-y", Name: "u", Role: models.UserRoleStudent}).Error)

		mux := http.NewServeMux()
		mux.HandleFunc("/cgi-bin/token", func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "t", "expires_in": 7200})
		})
		mux.HandleFunc("/wxa/business/getuserphonenumber", func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]any{"errcode": 40029, "errmsg": "invalid code"})
		})
		server := httptest.NewServer(mux)
		defer server.Close()

		svc := NewAuthService(db, &config.WechatConfig{AppID: "a", AppSecret: "b"}, &config.JWTConfig{Secret: "s", ExpireTime: time.Hour, Issuer: "i"}, zap.NewNop())
		svc.wechatClient.SetBaseURL(server.URL)

		var user models.User
		require.NoError(t, db.Where("open_id = ?", "oid-y").First(&user).Error)
		err := svc.BindPhone(user.ID, "code")
		assert.Error(t, err)
	})
}

func TestShiftAssignmentService_ExtraBranches(t *testing.T) {
	t.Run("non-overload branch", func(t *testing.T) {
		db := newTestDB(t)
		svc := NewShiftAssignmentService(db)

		driver := models.Driver{Name: "d", CarModel: "SUV", MaxSeats: 5, MaxChecked: 5, MaxCarryOn: 5}
		require.NoError(t, db.Create(&driver).Error)
		shift := models.Shift{DriverID: driver.ID, DepartureTime: time.Now(), Status: models.ShiftStatusDraft}
		require.NoError(t, db.Create(&shift).Error)
		req := models.Request{UserID: 1, FlightNo: "A1", ArrivalDate: time.Now(), Terminal: "T1", Status: models.RequestStatusPending}
		require.NoError(t, db.Omit(clause.Associations).Create(&req).Error)

		res, err := svc.AssignStudentToShift(context.Background(), shift.ID, req.ID)
		require.NoError(t, err)
		assert.Empty(t, res.Warning)
	})

	t.Run("count query error branch", func(t *testing.T) {
		db := newTestDB(t)
		svc := NewShiftAssignmentService(db)

		driver := models.Driver{Name: "d", CarModel: "SUV", MaxSeats: 5, MaxChecked: 5, MaxCarryOn: 5}
		require.NoError(t, db.Create(&driver).Error)
		shift := models.Shift{DriverID: driver.ID, DepartureTime: time.Now(), Status: models.ShiftStatusDraft}
		require.NoError(t, db.Create(&shift).Error)
		req := models.Request{UserID: 1, FlightNo: "A1", ArrivalDate: time.Now(), Terminal: "T1", Status: models.RequestStatusPending}
		require.NoError(t, db.Omit(clause.Associations).Create(&req).Error)
		require.NoError(t, db.Exec("DROP TABLE shift_requests").Error)

		_, err := svc.AssignStudentToShift(context.Background(), shift.ID, req.ID)
		assert.Error(t, err)
	})
}

func TestStudentService_ExtraBranches(t *testing.T) {
	t.Run("list requests db error", func(t *testing.T) {
		db := newTestDB(t)
		svc := NewStudentService(db)
		require.NoError(t, db.Exec("DROP TABLE requests").Error)
		_, err := svc.ListMyRequests(1)
		assert.Error(t, err)
	})

	t.Run("update parse and not found errors", func(t *testing.T) {
		db := newTestDB(t)
		svc := NewStudentService(db)

		_, err := svc.UpdatePendingRequest(1, 999, UpdateRequestInput{})
		assert.Error(t, err)

		req := models.Request{UserID: 1, FlightNo: "AA1", ArrivalDate: time.Now(), Terminal: "T1", Status: models.RequestStatusPending}
		require.NoError(t, db.Omit(clause.Associations).Create(&req).Error)

		badDate := "bad-date"
		_, err = svc.UpdatePendingRequest(1, req.ID, UpdateRequestInput{ArrivalDate: &badDate})
		assert.Error(t, err)

		badTime := "bad-time"
		_, err = svc.UpdatePendingRequest(1, req.ID, UpdateRequestInput{ExpectedArrivalTime: &badTime})
		assert.Error(t, err)
	})
}
