package controllers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"pickup/internal/config"
	"pickup/internal/scheduler/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func newControllerTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	ddls := []string{
		`CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, open_id TEXT NOT NULL UNIQUE, name TEXT NOT NULL, phone TEXT, role TEXT NOT NULL DEFAULT 'student', created_at DATETIME, updated_at DATETIME);`,
		`CREATE TABLE drivers (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, car_model TEXT NOT NULL, max_seats INTEGER NOT NULL, max_checked INTEGER NOT NULL, max_carry_on INTEGER NOT NULL);`,
		`CREATE TABLE requests (id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER NOT NULL, flight_no TEXT NOT NULL, arrival_date DATETIME NOT NULL, terminal TEXT NOT NULL, checked_bags INTEGER NOT NULL DEFAULT 0, carry_on_bags INTEGER NOT NULL DEFAULT 0, status TEXT NOT NULL DEFAULT 'pending', arrival_time_api DATETIME, pickup_buffer INTEGER NOT NULL DEFAULT 45, calc_pickup_time DATETIME, created_at DATETIME, updated_at DATETIME);`,
		`CREATE TABLE shifts (id INTEGER PRIMARY KEY AUTOINCREMENT, driver_id INTEGER NOT NULL, departure_time DATETIME NOT NULL, status TEXT NOT NULL DEFAULT 'draft', created_at DATETIME);`,
		`CREATE TABLE shift_requests (shift_id INTEGER NOT NULL, request_id INTEGER NOT NULL UNIQUE, PRIMARY KEY (shift_id, request_id));`,
		`CREATE TABLE shift_staffs (shift_id INTEGER NOT NULL, staff_id INTEGER NOT NULL, PRIMARY KEY (shift_id, staff_id));`,
	}
	for _, ddl := range ddls {
		require.NoError(t, db.Exec(ddl).Error)
	}
	return db
}

func TestAuthController_BadRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctl := NewAuthController(nil)
	r := gin.New()
	r.POST("/login", ctl.Login)
	r.POST("/bind", ctl.BindPhone)

	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("{}"))
	req1.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusBadRequest, w1.Code)

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/bind", strings.NewReader(`{"phone_code":"x"}`))
	req2.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusUnauthorized, w2.Code)

	r2 := gin.New()
	r2.POST("/bind", func(c *gin.Context) { c.Set("user_id", "bad"); ctl.BindPhone(c) })
	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodPost, "/bind", strings.NewReader(`{"phone_code":"x"}`))
	req3.Header.Set("Content-Type", "application/json")
	r2.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusUnauthorized, w3.Code)

	r3 := gin.New()
	r3.POST("/bind", func(c *gin.Context) { c.Set("user_id", uint(1)); ctl.BindPhone(c) })
	w4 := httptest.NewRecorder()
	req4 := httptest.NewRequest(http.MethodPost, "/bind", strings.NewReader(`{`))
	req4.Header.Set("Content-Type", "application/json")
	r3.ServeHTTP(w4, req4)
	assert.Equal(t, http.StatusBadRequest, w4.Code)
}

func TestStudentController_Flows(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newControllerTestDB(t)
	svc := service.NewStudentService(db)
	ctl := NewStudentController(svc)

	r := gin.New()
	r.POST("/requests", func(c *gin.Context) { c.Set("user_id", uint(1)); ctl.CreateRequest(c) })
	r.GET("/my", func(c *gin.Context) { c.Set("user_id", uint(1)); ctl.MyRequests(c) })
	r.PUT("/requests/:id", func(c *gin.Context) { c.Set("user_id", uint(1)); ctl.UpdateRequest(c) })

	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodPost, "/requests", strings.NewReader(`{"flight_no":"AA1","arrival_date":"2026-03-01","terminal":"T1","expected_arrival_time":"2026-03-01 10:00:00"}`))
	req1.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusCreated, w1.Code)

	w1b := httptest.NewRecorder()
	req1b := httptest.NewRequest(http.MethodPost, "/requests", strings.NewReader(`{"flight_no":"AA2","arrival_date":"2026-03-02","terminal":"T5","expected_arrival_time":"2026-03-02 11:00:00"}`))
	req1b.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w1b, req1b)
	assert.Equal(t, http.StatusBadRequest, w1b.Code)

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/my", nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodPut, "/requests/1", strings.NewReader(`{"terminal":"T5"}`))
	req3.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusOK, w3.Code)

	w4 := httptest.NewRecorder()
	req4 := httptest.NewRequest(http.MethodPut, "/requests/bad", strings.NewReader(`{}`))
	req4.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w4, req4)
	assert.Equal(t, http.StatusBadRequest, w4.Code)

	w5 := httptest.NewRecorder()
	req5 := httptest.NewRequest(http.MethodPost, "/requests", strings.NewReader(`{`))
	req5.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w5, req5)
	assert.Equal(t, http.StatusBadRequest, w5.Code)

	r2 := gin.New()
	r2.POST("/requests", ctl.CreateRequest)
	w6 := httptest.NewRecorder()
	req6 := httptest.NewRequest(http.MethodPost, "/requests", strings.NewReader(`{"flight_no":"AA1","arrival_date":"2026-03-01","terminal":"T1","expected_arrival_time":"2026-03-01 10:00:00"}`))
	req6.Header.Set("Content-Type", "application/json")
	r2.ServeHTTP(w6, req6)
	assert.Equal(t, http.StatusUnauthorized, w6.Code)

	r3 := gin.New()
	r3.GET("/my", ctl.MyRequests)
	w7 := httptest.NewRecorder()
	req7 := httptest.NewRequest(http.MethodGet, "/my", nil)
	r3.ServeHTTP(w7, req7)
	assert.Equal(t, http.StatusUnauthorized, w7.Code)

	require.NoError(t, db.Exec(`DROP TABLE requests`).Error)
	w8 := httptest.NewRecorder()
	req8 := httptest.NewRequest(http.MethodGet, "/my", nil)
	r.ServeHTTP(w8, req8)
	assert.Equal(t, http.StatusInternalServerError, w8.Code)
}

func TestAdminController_FlowsAndErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newControllerTestDB(t)
	assigner := service.NewShiftAssignmentService(db)
	svc := service.NewAdminService(db, assigner)
	ctl := NewAdminController(svc)

	r := gin.New()
	r.GET("/drivers", ctl.ListDrivers)
	r.POST("/drivers", ctl.CreateDriver)
	r.POST("/shifts", ctl.CreateShift)
	r.PUT("/drivers/:id", ctl.UpdateDriver)
	r.PUT("/shifts/:id", ctl.UpdateShift)
	r.POST("/shifts/:id/assign-student", ctl.AssignStudent)
	r.POST("/shifts/:id/remove-student", ctl.RemoveStudent)
	r.POST("/shifts/:id/assign-staff", ctl.AssignStaff)
	r.POST("/shifts/:id/remove-staff", ctl.RemoveStaff)
	r.POST("/shifts/:id/publish", ctl.PublishShift)
	r.GET("/dashboard", ctl.Dashboard)
	r.GET("/pending", ctl.PendingRequests)
	r.GET("/users", ctl.ListUsers)
	r.POST("/users/:id/set-staff", ctl.SetStaff)
	r.POST("/users/:id/unset-staff", ctl.UnsetStaff)

	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodPost, "/drivers", strings.NewReader(`{"name":"d1","car_model":"SUV","max_seats":4,"max_checked":4,"max_carry_on":4}`))
	req1.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusCreated, w1.Code)

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/shifts", strings.NewReader(`{"driver_id":1,"departure_time":"2026-03-01 12:00:00"}`))
	req2.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusCreated, w2.Code)

	require.NoError(t, db.Exec(`INSERT INTO requests(user_id,flight_no,arrival_date,terminal,status,checked_bags,carry_on_bags,pickup_buffer) VALUES (1,'AA1','2026-03-01','T1','pending',0,0,45)`).Error)

	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodPost, "/shifts/1/assign-student", strings.NewReader(`{"request_id":1}`))
	req3.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusOK, w3.Code)

	require.NoError(t, db.Exec(`INSERT INTO users(open_id,name,role) VALUES ('staff-1','staff','staff')`).Error)

	w4 := httptest.NewRecorder()
	req4 := httptest.NewRequest(http.MethodPost, "/shifts/1/assign-staff", strings.NewReader(`{"staff_id":1}`))
	req4.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w4, req4)
	assert.Equal(t, http.StatusOK, w4.Code)

	w5 := httptest.NewRecorder()
	req5 := httptest.NewRequest(http.MethodPost, "/shifts/1/publish", nil)
	r.ServeHTTP(w5, req5)
	assert.Equal(t, http.StatusOK, w5.Code)

	w5a := httptest.NewRecorder()
	req5a := httptest.NewRequest(http.MethodPost, "/shifts/1/remove-staff", strings.NewReader(`{"staff_id":1}`))
	req5a.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w5a, req5a)
	assert.Equal(t, http.StatusOK, w5a.Code)

	w6 := httptest.NewRecorder()
	req6 := httptest.NewRequest(http.MethodPost, "/shifts/1/remove-student", strings.NewReader(`{"request_id":1}`))
	req6.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w6, req6)
	assert.Equal(t, http.StatusOK, w6.Code)

	w7 := httptest.NewRecorder()
	req7 := httptest.NewRequest(http.MethodGet, "/drivers", nil)
	r.ServeHTTP(w7, req7)
	assert.Equal(t, http.StatusOK, w7.Code)

	w8 := httptest.NewRecorder()
	req8 := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	r.ServeHTTP(w8, req8)
	assert.Equal(t, http.StatusOK, w8.Code)

	w9 := httptest.NewRecorder()
	req9 := httptest.NewRequest(http.MethodGet, "/pending", nil)
	r.ServeHTTP(w9, req9)
	assert.Equal(t, http.StatusOK, w9.Code)

	w10 := httptest.NewRecorder()
	req10 := httptest.NewRequest(http.MethodPost, "/shifts/bad/assign-student", strings.NewReader(`{"request_id":1}`))
	req10.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w10, req10)
	assert.Equal(t, http.StatusBadRequest, w10.Code)

	w11 := httptest.NewRecorder()
	req11 := httptest.NewRequest(http.MethodPost, "/drivers", strings.NewReader(`{`))
	req11.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w11, req11)
	assert.Equal(t, http.StatusBadRequest, w11.Code)

	w12 := httptest.NewRecorder()
	req12 := httptest.NewRequest(http.MethodPost, "/shifts", strings.NewReader(`{`))
	req12.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w12, req12)
	assert.Equal(t, http.StatusBadRequest, w12.Code)

	w12a := httptest.NewRecorder()
	req12a := httptest.NewRequest(http.MethodPut, "/drivers/1", strings.NewReader(`{"name":"d2","car_model":"SUV","max_seats":5,"max_checked":5,"max_carry_on":5}`))
	req12a.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w12a, req12a)
	assert.Equal(t, http.StatusOK, w12a.Code)

	w12b := httptest.NewRecorder()
	req12b := httptest.NewRequest(http.MethodPut, "/shifts/1", strings.NewReader(`{"departure_time":"2026-03-01 13:00:00"}`))
	req12b.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w12b, req12b)
	assert.Equal(t, http.StatusOK, w12b.Code)

	w13 := httptest.NewRecorder()
	req13 := httptest.NewRequest(http.MethodPost, "/shifts/1/assign-student", strings.NewReader(`{`))
	req13.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w13, req13)
	assert.Equal(t, http.StatusBadRequest, w13.Code)

	w13a := httptest.NewRecorder()
	req13a := httptest.NewRequest(http.MethodPut, "/shifts/bad", strings.NewReader(`{}`))
	req13a.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w13a, req13a)
	assert.Equal(t, http.StatusBadRequest, w13a.Code)

	require.NoError(t, db.Exec(`DROP TABLE drivers`).Error)
	w14 := httptest.NewRecorder()
	req14 := httptest.NewRequest(http.MethodGet, "/drivers", nil)
	r.ServeHTTP(w14, req14)
	assert.Equal(t, http.StatusInternalServerError, w14.Code)

	w15 := httptest.NewRecorder()
	req15 := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	r.ServeHTTP(w15, req15)
	assert.Contains(t, []int{http.StatusOK, http.StatusInternalServerError}, w15.Code)

	w16 := httptest.NewRecorder()
	req16 := httptest.NewRequest(http.MethodGet, "/pending", nil)
	r.ServeHTTP(w16, req16)
	assert.Contains(t, []int{http.StatusOK, http.StatusInternalServerError}, w16.Code)

	require.NoError(t, db.Exec(`CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY AUTOINCREMENT, open_id TEXT NOT NULL UNIQUE, name TEXT NOT NULL, phone TEXT, role TEXT NOT NULL DEFAULT 'student', created_at DATETIME, updated_at DATETIME)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO users(open_id,name,role) VALUES ('u2','u2','student')`).Error)

	w17 := httptest.NewRecorder()
	req17 := httptest.NewRequest(http.MethodGet, "/users", nil)
	r.ServeHTTP(w17, req17)
	assert.Equal(t, http.StatusOK, w17.Code)

	w18 := httptest.NewRecorder()
	req18 := httptest.NewRequest(http.MethodPost, "/users/1/set-staff", nil)
	r.ServeHTTP(w18, req18)
	assert.Equal(t, http.StatusOK, w18.Code)

	w19 := httptest.NewRecorder()
	req19 := httptest.NewRequest(http.MethodPost, "/users/1/unset-staff", nil)
	r.ServeHTTP(w19, req19)
	assert.Equal(t, http.StatusOK, w19.Code)
}

func TestAdminController_ParseID(t *testing.T) {
	_, err := parseID("x")
	assert.Error(t, err)
	id, err := parseID("12")
	require.NoError(t, err)
	assert.Equal(t, uint(12), id)
}

func TestAuthController_LoginUnauthorizedPathWithService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newControllerTestDB(t)
	authSvc := service.NewAuthService(
		db,
		&config.WechatConfig{AppID: "x", AppSecret: "y"},
		&config.JWTConfig{Secret: "s", ExpireTime: time.Hour, Issuer: "i"},
		zap.NewNop(),
	)
	ctl := NewAuthController(authSvc)
	r := gin.New()
	r.POST("/login", ctl.Login)
	r.GET("/me", ctl.Me)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"code":"bad"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	r2 := gin.New()
	r2.POST("/bind", func(c *gin.Context) { c.Set("user_id", uint(1)); ctl.BindPhone(c) })
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/bind", strings.NewReader(`{"phone_code":"bad"}`))
	req2.Header.Set("Content-Type", "application/json")
	r2.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusBadRequest, w2.Code)

	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodGet, "/me", nil)
	r.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusUnauthorized, w3.Code)

	r3 := gin.New()
	r3.GET("/me", func(c *gin.Context) { c.Set("user_id", uint(1)); ctl.Me(c) })
	w4 := httptest.NewRecorder()
	req4 := httptest.NewRequest(http.MethodGet, "/me", nil)
	r3.ServeHTTP(w4, req4)
	assert.Equal(t, http.StatusUnauthorized, w4.Code)

	require.NoError(t, db.Exec(`INSERT INTO users(id,open_id,name,role) VALUES (1,'u1','user-1','student')`).Error)
	w5 := httptest.NewRecorder()
	req5 := httptest.NewRequest(http.MethodGet, "/me", nil)
	r3.ServeHTTP(w5, req5)
	assert.Equal(t, http.StatusOK, w5.Code)
}

func TestAdminController_ErrorBranchesDeep(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newControllerTestDB(t)
	assigner := service.NewShiftAssignmentService(db)
	svc := service.NewAdminService(db, assigner)
	ctl := NewAdminController(svc)

	require.NoError(t, db.Exec(`INSERT INTO drivers(name,car_model,max_seats,max_checked,max_carry_on) VALUES ('d1','SUV',4,4,4)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO shifts(driver_id,departure_time,status) VALUES (1,'2026-03-01 12:00:00','draft')`).Error)
	require.NoError(t, db.Exec(`INSERT INTO requests(user_id,flight_no,arrival_date,terminal,status,checked_bags,carry_on_bags,pickup_buffer) VALUES (1,'AA1','2026-03-01','T1','pending',0,0,45)`).Error)

	r := gin.New()
	r.POST("/shifts/:id/remove-student", ctl.RemoveStudent)
	r.POST("/shifts/:id/assign-staff", ctl.AssignStaff)
	r.POST("/shifts/:id/publish", ctl.PublishShift)

	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodPost, "/shifts/1/remove-student", strings.NewReader(`{`))
	req1.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusBadRequest, w1.Code)

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/shifts/1/assign-staff", strings.NewReader(`{`))
	req2.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusBadRequest, w2.Code)

	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodPost, "/shifts/bad/publish", nil)
	r.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusBadRequest, w3.Code)

	require.NoError(t, db.Exec(`INSERT INTO users(open_id,name,role) VALUES ('u1','student-user','student')`).Error)
	w4 := httptest.NewRecorder()
	req4 := httptest.NewRequest(http.MethodPost, "/shifts/1/assign-staff", strings.NewReader(`{"staff_id":1}`))
	req4.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w4, req4)
	assert.Equal(t, http.StatusBadRequest, w4.Code)

	require.NoError(t, db.Exec(`DROP TABLE shift_requests`).Error)
	w5 := httptest.NewRecorder()
	req5 := httptest.NewRequest(http.MethodPost, "/shifts/1/remove-student", strings.NewReader(`{"request_id":1}`))
	req5.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w5, req5)
	assert.Equal(t, http.StatusBadRequest, w5.Code)
}

func TestStudentController_UpdateErrorBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newControllerTestDB(t)
	svc := service.NewStudentService(db)
	ctl := NewStudentController(svc)

	require.NoError(t, db.Exec(`INSERT INTO requests(user_id,flight_no,arrival_date,terminal,status,checked_bags,carry_on_bags,pickup_buffer) VALUES (1,'AA1','2026-03-01','T1','pending',0,0,45)`).Error)

	r := gin.New()
	r.PUT("/requests/:id", ctl.UpdateRequest)

	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodPut, "/requests/1", strings.NewReader(`{"terminal":"T2"}`))
	req1.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusUnauthorized, w1.Code)

	r2 := gin.New()
	r2.PUT("/requests/:id", func(c *gin.Context) { c.Set("user_id", "bad"); ctl.UpdateRequest(c) })
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPut, "/requests/1", strings.NewReader(`{"terminal":"T2"}`))
	req2.Header.Set("Content-Type", "application/json")
	r2.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusUnauthorized, w2.Code)

	require.NoError(t, db.Exec(`DROP TABLE requests`).Error)
	r3 := gin.New()
	r3.PUT("/requests/:id", func(c *gin.Context) { c.Set("user_id", uint(1)); ctl.UpdateRequest(c) })
	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodPut, "/requests/1", strings.NewReader(`{"terminal":"T2"}`))
	req3.Header.Set("Content-Type", "application/json")
	r3.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusBadRequest, w3.Code)
}

func TestAdminController_PendingAndCreateShiftErrorBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newControllerTestDB(t)
	assigner := service.NewShiftAssignmentService(db)
	svc := service.NewAdminService(db, assigner)
	ctl := NewAdminController(svc)

	r := gin.New()
	r.GET("/pending", ctl.PendingRequests)
	r.POST("/shifts", ctl.CreateShift)

	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodPost, "/shifts", strings.NewReader(`{"driver_id":1,"departure_time":"bad-time"}`))
	req1.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusBadRequest, w1.Code)

	require.NoError(t, db.Exec(`DROP TABLE requests`).Error)
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/pending", nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusInternalServerError, w2.Code)
}
