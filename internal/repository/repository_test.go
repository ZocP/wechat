package repository

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"pickup/internal/model"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// setupTestDB creates a GORM DB backed by sqlmock
func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	require.NoError(t, err)

	t.Cleanup(func() { db.Close() })
	return gormDB, mock
}

// ==================== UserRepository Tests ====================

func TestNewUserRepository(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewUserRepository(db)
	assert.NotNil(t, repo)
}

func TestUserRepository_Create(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewUserRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `users`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	user := &model.User{Phone: "13800138000", OpenID: "openid1", Role: model.RolePassenger, Status: model.UserStatusActive}
	err := repo.Create(user)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Create_Error(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewUserRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `users`").WillReturnError(errors.New("duplicate key"))
	mock.ExpectRollback()

	user := &model.User{Phone: "13800138000", OpenID: "openid1"}
	err := repo.Create(user)
	assert.Error(t, err)
}

func TestUserRepository_GetByID(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewUserRepository(db)

	rows := sqlmock.NewRows([]string{"id", "openid", "phone", "nickname", "role", "status"}).
		AddRow(1, "openid1", "13800138000", "test", "passenger", "active")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE `users`.`id` = ? ORDER BY `users`.`id` LIMIT ?")).
		WithArgs(1, 1).
		WillReturnRows(rows)

	user, err := repo.GetByID(1)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "13800138000", user.Phone)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewUserRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE `users`.`id` = ? ORDER BY `users`.`id` LIMIT ?")).
		WithArgs(999, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	user, err := repo.GetByID(999)
	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestUserRepository_GetByOpenID(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewUserRepository(db)

	rows := sqlmock.NewRows([]string{"id", "openid", "phone"}).
		AddRow(1, "openid1", "13800138000")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE openid = ? ORDER BY `users`.`id` LIMIT ?")).
		WithArgs("openid1", 1).
		WillReturnRows(rows)

	user, err := repo.GetByOpenID("openid1")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByOpenID_NotFound(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewUserRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE openid = ? ORDER BY `users`.`id` LIMIT ?")).
		WithArgs("nonexistent", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	user, err := repo.GetByOpenID("nonexistent")
	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestUserRepository_GetByPhone(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewUserRepository(db)

	rows := sqlmock.NewRows([]string{"id", "openid", "phone"}).
		AddRow(1, "openid1", "13800138000")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE phone = ? ORDER BY `users`.`id` LIMIT ?")).
		WithArgs("13800138000", 1).
		WillReturnRows(rows)

	user, err := repo.GetByPhone("13800138000")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Update(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewUserRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `users`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	user := &model.User{ID: 1, Phone: "13800138000", OpenID: "openid1", Nickname: "updated"}
	err := repo.Update(user)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_UpdateLastLoginAt(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewUserRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `users` SET `last_login_at`=NOW()")).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.UpdateLastLoginAt(1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== OrderRepository Tests ====================

func TestNewOrderRepository(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewOrderRepository(db)
	assert.NotNil(t, repo)
}

func TestOrderRepository_Create(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewOrderRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `pickup_orders`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	order := &model.PickupOrder{PassengerID: 1, RegistrationID: 1, PriceTotal: 10000, Status: model.OrderStatusCreated}
	err := repo.Create(order)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_GetByID(t *testing.T) {
	db, mock := setupTestDB(t)
	mock.MatchExpectationsInOrder(false)
	repo := NewOrderRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "passenger_id", "registration_id", "status", "price_total", "created_at"}).
		AddRow(1, 1, 1, "created", 10000, now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `pickup_orders` WHERE `pickup_orders`.`id` = ? ORDER BY `pickup_orders`.`id` LIMIT ?")).
		WithArgs(1, 1).
		WillReturnRows(rows)
	// Preload queries - order may vary
	mock.ExpectQuery("SELECT \\* FROM `users`").WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery("SELECT \\* FROM `registrations`").WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery("SELECT \\* FROM `assignments`").WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery("SELECT \\* FROM `payment_orders`").WillReturnRows(sqlmock.NewRows([]string{"id"}))

	order, err := repo.GetByID(1)
	assert.NoError(t, err)
	assert.NotNil(t, order)
	assert.Equal(t, int64(10000), order.PriceTotal)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_GetByID_NotFound(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewOrderRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `pickup_orders` WHERE `pickup_orders`.`id` = ? ORDER BY `pickup_orders`.`id` LIMIT ?")).
		WithArgs(999, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	order, err := repo.GetByID(999)
	assert.Error(t, err)
	assert.Nil(t, order)
}

func TestOrderRepository_GetByUserID(t *testing.T) {
	db, mock := setupTestDB(t)
	mock.MatchExpectationsInOrder(false)
	repo := NewOrderRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "passenger_id", "registration_id", "status", "price_total", "created_at"}).
		AddRow(1, 1, 1, "created", 10000, now).
		AddRow(2, 1, 2, "paid", 20000, now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `pickup_orders` WHERE passenger_id = ?")).
		WithArgs(1).
		WillReturnRows(rows)
	mock.ExpectQuery("SELECT \\* FROM `registrations`").WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery("SELECT \\* FROM `assignments`").WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery("SELECT \\* FROM `payment_orders`").WillReturnRows(sqlmock.NewRows([]string{"id"}))

	orders, err := repo.GetByUserID(1)
	assert.NoError(t, err)
	assert.Len(t, orders, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_GetByRegistrationID(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewOrderRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "passenger_id", "registration_id", "status", "created_at"}).
		AddRow(1, 1, 5, "created", now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `pickup_orders` WHERE registration_id = ?")).
		WithArgs(uint(5), 1).
		WillReturnRows(rows)

	order, err := repo.GetByRegistrationID(5)
	assert.NoError(t, err)
	assert.NotNil(t, order)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_Update(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewOrderRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `pickup_orders`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	order := &model.PickupOrder{ID: 1, PassengerID: 1, RegistrationID: 1, Status: model.OrderStatusPaid}
	err := repo.Update(order)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_UpdateStatus(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewOrderRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `pickup_orders` SET `status`=?")).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.UpdateStatus(1, model.OrderStatusPaid)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOrderRepository_Delete(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewOrderRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM `pickup_orders`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.Delete(1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== NoticeRepository Tests ====================

func TestNewNoticeRepository(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewNoticeRepository(db)
	assert.NotNil(t, repo)
}

func TestNoticeRepository_Create(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewNoticeRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `notices`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	notice := &model.Notice{FlightNo: "CA123", Terminal: "T2", PickupBatch: "B1", ArrivalAirport: "PEK", MeetingPoint: "Gate 5"}
	err := repo.Create(notice)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNoticeRepository_GetByID(t *testing.T) {
	db, mock := setupTestDB(t)
	mock.MatchExpectationsInOrder(false)
	repo := NewNoticeRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "flight_no", "terminal", "created_by", "created_at"}).
		AddRow(1, "CA123", "T2", 1, now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `notices` WHERE `notices`.`id` = ? ORDER BY `notices`.`id` LIMIT ?")).
		WithArgs(1, 1).
		WillReturnRows(rows)
	mock.ExpectQuery("SELECT \\* FROM `users`").WillReturnRows(sqlmock.NewRows([]string{"id"}))

	notice, err := repo.GetByID(1)
	assert.NoError(t, err)
	assert.NotNil(t, notice)
	assert.Equal(t, "CA123", notice.FlightNo)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNoticeRepository_GetByID_NotFound(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewNoticeRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `notices` WHERE `notices`.`id` = ? ORDER BY `notices`.`id` LIMIT ?")).
		WithArgs(999, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	notice, err := repo.GetByID(999)
	assert.Error(t, err)
	assert.Nil(t, notice)
}

func TestNoticeRepository_GetVisibleNotices(t *testing.T) {
	db, mock := setupTestDB(t)
	mock.MatchExpectationsInOrder(false)
	repo := NewNoticeRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "flight_no", "terminal", "created_by", "created_at"}).
		AddRow(1, "CA123", "T1", 1, now).
		AddRow(2, "MU456", "T2", 1, now)

	mock.ExpectQuery("SELECT \\* FROM `notices` WHERE").
		WillReturnRows(rows)
	mock.ExpectQuery("SELECT \\* FROM `users`").WillReturnRows(sqlmock.NewRows([]string{"id"}))

	notices, err := repo.GetVisibleNotices()
	assert.NoError(t, err)
	assert.Len(t, notices, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNoticeRepository_GetByFlightNo(t *testing.T) {
	db, mock := setupTestDB(t)
	mock.MatchExpectationsInOrder(false)
	repo := NewNoticeRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "flight_no", "created_by", "created_at"}).
		AddRow(1, "CA123", 1, now)

	mock.ExpectQuery("SELECT \\* FROM `notices` WHERE").
		WillReturnRows(rows)
	mock.ExpectQuery("SELECT \\* FROM `users`").WillReturnRows(sqlmock.NewRows([]string{"id"}))

	notices, err := repo.GetByFlightNo("CA123")
	assert.NoError(t, err)
	assert.Len(t, notices, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNoticeRepository_Update(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewNoticeRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `notices`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	notice := &model.Notice{ID: 1, FlightNo: "CA123", Terminal: "T2", PickupBatch: "B1", ArrivalAirport: "PEK", MeetingPoint: "Gate 5"}
	err := repo.Update(notice)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNoticeRepository_Delete(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewNoticeRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM `notices`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.Delete(1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== PaymentRepository Tests ====================

func TestNewPaymentRepository(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewPaymentRepository(db)
	assert.NotNil(t, repo)
}

func TestPaymentRepository_Create(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPaymentRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `payment_orders`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	payment := &model.PaymentOrder{OrderID: 1, UserID: 1, Amount: 10000}
	err := repo.Create(payment)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPaymentRepository_GetByID(t *testing.T) {
	db, mock := setupTestDB(t)
	mock.MatchExpectationsInOrder(false)
	repo := NewPaymentRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "order_id", "user_id", "amount", "created_at"}).
		AddRow(1, 1, 1, 10000, now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `payment_orders` WHERE `payment_orders`.`id` = ? ORDER BY `payment_orders`.`id` LIMIT ?")).
		WithArgs(1, 1).
		WillReturnRows(rows)
	mock.ExpectQuery("SELECT \\* FROM `pickup_orders`").WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery("SELECT \\* FROM `users`").WillReturnRows(sqlmock.NewRows([]string{"id"}))

	payment, err := repo.GetByID(1)
	assert.NoError(t, err)
	assert.NotNil(t, payment)
	assert.Equal(t, int64(10000), payment.Amount)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPaymentRepository_GetByOrderID(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPaymentRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "order_id", "user_id", "amount", "created_at"}).
		AddRow(1, 5, 1, 10000, now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `payment_orders` WHERE order_id = ?")).
		WithArgs(uint(5), 1).
		WillReturnRows(rows)

	payment, err := repo.GetByOrderID(5)
	assert.NoError(t, err)
	assert.NotNil(t, payment)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPaymentRepository_GetByTransactionID(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPaymentRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "order_id", "wx_transaction_id", "created_at"}).
		AddRow(1, 1, "tx123", now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `payment_orders` WHERE wx_transaction_id = ?")).
		WithArgs("tx123", 1).
		WillReturnRows(rows)

	payment, err := repo.GetByTransactionID("tx123")
	assert.NoError(t, err)
	assert.NotNil(t, payment)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPaymentRepository_Update(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPaymentRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `payment_orders`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	payment := &model.PaymentOrder{ID: 1, OrderID: 1, Amount: 20000}
	err := repo.Update(payment)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPaymentRepository_UpdateState(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPaymentRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `payment_orders`").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.UpdateState(1, model.PaymentStatePaid)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== RegistrationRepository Tests ====================

func TestNewRegistrationRepository(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewRegistrationRepository(db)
	assert.NotNil(t, repo)
}

func TestRegistrationRepository_Create(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRegistrationRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `registrations`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	reg := &model.Registration{UserID: 1, FlightNo: "CA123"}
	err := repo.Create(reg)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRegistrationRepository_GetByID(t *testing.T) {
	db, mock := setupTestDB(t)
	mock.MatchExpectationsInOrder(false)
	repo := NewRegistrationRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "user_id", "flight_no", "created_at"}).
		AddRow(1, 1, "CA123", now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `registrations` WHERE `registrations`.`id` = ? ORDER BY `registrations`.`id` LIMIT ?")).
		WithArgs(1, 1).
		WillReturnRows(rows)
	mock.ExpectQuery("SELECT \\* FROM `users`").WillReturnRows(sqlmock.NewRows([]string{"id"}))

	reg, err := repo.GetByID(1)
	assert.NoError(t, err)
	assert.NotNil(t, reg)
	assert.Equal(t, "CA123", reg.FlightNo)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRegistrationRepository_GetByUserID(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRegistrationRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "user_id", "flight_no", "created_at"}).
		AddRow(1, 1, "CA123", now).
		AddRow(2, 1, "MU456", now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `registrations` WHERE user_id = ?")).
		WithArgs(1).
		WillReturnRows(rows)

	regs, err := repo.GetByUserID(1)
	assert.NoError(t, err)
	assert.Len(t, regs, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRegistrationRepository_Update(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRegistrationRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `registrations`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	reg := &model.Registration{ID: 1, UserID: 1, FlightNo: "CA123"}
	err := repo.Update(reg)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRegistrationRepository_Delete(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewRegistrationRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM `registrations`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.Delete(1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== AssignmentRepository Tests ====================

func TestNewAssignmentRepository(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewAssignmentRepository(db)
	assert.NotNil(t, repo)
}

func TestAssignmentRepository_Create(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewAssignmentRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `assignments`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	assignment := &model.Assignment{OrderID: 1, DriverID: 2}
	err := repo.Create(assignment)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAssignmentRepository_GetByID(t *testing.T) {
	db, mock := setupTestDB(t)
	mock.MatchExpectationsInOrder(false)
	repo := NewAssignmentRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "order_id", "driver_id", "created_at"}).
		AddRow(1, 1, 2, now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `assignments` WHERE `assignments`.`id` = ? ORDER BY `assignments`.`id` LIMIT ?")).
		WithArgs(1, 1).
		WillReturnRows(rows)
	mock.ExpectQuery("SELECT \\* FROM `pickup_orders`").WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery("SELECT \\* FROM `drivers`").WillReturnRows(sqlmock.NewRows([]string{"id"}))

	assignment, err := repo.GetByID(1)
	assert.NoError(t, err)
	assert.NotNil(t, assignment)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAssignmentRepository_GetByOrderID(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewAssignmentRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "order_id", "driver_id", "created_at"}).
		AddRow(1, 5, 2, now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `assignments` WHERE order_id = ?")).
		WithArgs(uint(5), 1).
		WillReturnRows(rows)

	assignment, err := repo.GetByOrderID(5)
	assert.NoError(t, err)
	assert.NotNil(t, assignment)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAssignmentRepository_GetByDriverID(t *testing.T) {
	db, mock := setupTestDB(t)
	mock.MatchExpectationsInOrder(false)
	repo := NewAssignmentRepository(db)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "order_id", "driver_id", "created_at"}).
		AddRow(1, 1, 3, now).
		AddRow(2, 2, 3, now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `assignments` WHERE driver_id = ?")).
		WithArgs(3).
		WillReturnRows(rows)
	mock.ExpectQuery("SELECT \\* FROM `pickup_orders`").WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery("SELECT \\* FROM `drivers`").WillReturnRows(sqlmock.NewRows([]string{"id"}))

	assignments, err := repo.GetByDriverID(3)
	assert.NoError(t, err)
	assert.Len(t, assignments, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAssignmentRepository_Update(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewAssignmentRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `assignments`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	assignment := &model.Assignment{ID: 1, OrderID: 1, DriverID: 2}
	err := repo.Update(assignment)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAssignmentRepository_UpdateStatus_Accepted(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewAssignmentRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `assignments`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.UpdateStatus(1, model.AssignmentStatusAccepted)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAssignmentRepository_UpdateStatus_Rejected(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewAssignmentRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `assignments`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.UpdateStatus(1, model.AssignmentStatusRejected)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== SchemaRepository Tests ====================

func TestNewSchemaRepository(t *testing.T) {
	db, _ := setupTestDB(t)
	repo := NewSchemaRepository(db)
	assert.NotNil(t, repo)
}

func TestSchemaRepository_GetAllColumns(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewSchemaRepository(db)

	rows := sqlmock.NewRows([]string{"table_name", "column_name", "data_type", "column_type", "column_key", "is_nullable", "column_default", "column_comment"}).
		AddRow("users", "id", "bigint", "bigint unsigned", "PRI", "NO", nil, "primary key").
		AddRow("users", "phone", "varchar", "varchar(20)", "", "NO", nil, "phone number")

	mock.ExpectQuery("SELECT TABLE_NAME, COLUMN_NAME").WillReturnRows(rows)

	columns, err := repo.GetAllColumns()
	assert.NoError(t, err)
	assert.Len(t, columns, 2)
	assert.Equal(t, "users", columns[0].TableName)
	assert.Equal(t, "id", columns[0].ColumnName)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSchemaRepository_GetAllColumns_Error(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewSchemaRepository(db)

	mock.ExpectQuery("SELECT TABLE_NAME, COLUMN_NAME").WillReturnError(errors.New("db error"))

	columns, err := repo.GetAllColumns()
	assert.Error(t, err)
	assert.Nil(t, columns)
}

// ==================== Provide Tests ====================

func TestProvide(t *testing.T) {
	opt := Provide()
	assert.NotNil(t, opt)
}

// ==================== Error Cases for Repositories ====================

func TestOrderRepository_Create_Error(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewOrderRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `pickup_orders`").WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	order := &model.PickupOrder{PassengerID: 1, RegistrationID: 1}
	err := repo.Create(order)
	assert.Error(t, err)
}

func TestNoticeRepository_GetVisibleNotices_Error(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewNoticeRepository(db)

	mock.ExpectQuery("SELECT \\* FROM `notices`").WillReturnError(errors.New("query failed"))

	notices, err := repo.GetVisibleNotices()
	assert.Error(t, err)
	assert.Nil(t, notices)
}

func TestPaymentRepository_GetByID_NotFound(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewPaymentRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `payment_orders`")).
		WillReturnError(gorm.ErrRecordNotFound)

	payment, err := repo.GetByID(999)
	assert.Error(t, err)
	assert.Nil(t, payment)
}

func TestUserRepository_GetByPhone_NotFound(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewUserRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE phone = ?")).
		WithArgs("nonexistent", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	user, err := repo.GetByPhone("nonexistent")
	assert.Error(t, err)
	assert.Nil(t, user)
}

// Helper to make sqlmock.NewRows with typed columns (avoids boilerplate)
func newNullableRow() *sql.Rows {
	return nil
}
