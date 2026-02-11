package model

import (
	"database/sql/driver"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ===== Table Name Tests =====

func TestTableNames(t *testing.T) {
	tests := []struct {
		model    interface{ TableName() string }
		expected string
	}{
		{&User{}, "users"},
		{&Registration{}, "registrations"},
		{&PickupOrder{}, "pickup_orders"},
		{&PaymentOrder{}, "payment_orders"},
		{&Assignment{}, "assignments"},
		{&Driver{}, "drivers"},
		{&Vehicle{}, "vehicles"},
		{&Notice{}, "notices"},
		{&ConsentLog{}, "consent_logs"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.model.TableName())
		})
	}
}

// ===== Response Tests =====

func TestNewSuccessResponse(t *testing.T) {
	data := map[string]string{"key": "value"}
	before := time.Now()
	resp := NewSuccessResponse(data)
	after := time.Now()

	assert.Equal(t, CodeSuccess, resp.Code)
	assert.Equal(t, "success", resp.Message)
	assert.NotNil(t, resp.Data)
	assert.True(t, resp.Time.After(before) || resp.Time.Equal(before))
	assert.True(t, resp.Time.Before(after) || resp.Time.Equal(after))
}

func TestNewSuccessResponse_NilData(t *testing.T) {
	resp := NewSuccessResponse(nil)
	assert.Equal(t, CodeSuccess, resp.Code)
	assert.Nil(t, resp.Data)
}

func TestNewErrorResponse(t *testing.T) {
	before := time.Now()
	resp := NewErrorResponse(CodeInvalidParams, "参数错误")
	after := time.Now()

	assert.Equal(t, CodeInvalidParams, resp.Code)
	assert.Equal(t, "参数错误", resp.Message)
	assert.Nil(t, resp.Data)
	assert.True(t, resp.Time.After(before) || resp.Time.Equal(before))
	assert.True(t, resp.Time.Before(after) || resp.Time.Equal(after))
}

func TestNewErrorResponse_AllCodes(t *testing.T) {
	codes := []struct {
		code    int
		message string
	}{
		{CodeSuccess, MsgSuccess},
		{CodeInvalidParams, MsgInvalidParams},
		{CodeUnauthorized, MsgUnauthorized},
		{CodeForbidden, MsgForbidden},
		{CodeNotFound, MsgNotFound},
		{CodeConflict, MsgConflict},
		{CodeInternalError, MsgInternalError},
		{CodeWechatAuthFailed, MsgWechatAuthFailed},
		{CodePaymentFailed, MsgPaymentFailed},
		{CodeOrderStatusError, MsgOrderStatusError},
		{CodeDriverNotAvailable, MsgDriverNotAvailable},
	}

	for _, tc := range codes {
		resp := NewErrorResponse(tc.code, tc.message)
		assert.Equal(t, tc.code, resp.Code)
		assert.Equal(t, tc.message, resp.Message)
	}
}

func TestAPIResponse_JSONSerialization(t *testing.T) {
	resp := NewSuccessResponse(map[string]int{"count": 42})
	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var parsed APIResponse
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)
	assert.Equal(t, 0, parsed.Code)
	assert.Equal(t, "success", parsed.Message)
}

// ===== JSONB Tests =====

func TestJSONB_Scan_Nil(t *testing.T) {
	var j JSONB
	err := j.Scan(nil)
	require.NoError(t, err)
	assert.NotNil(t, j)
	assert.Empty(t, j)
}

func TestJSONB_Scan_ValidJSON(t *testing.T) {
	var j JSONB
	err := j.Scan([]byte(`{"key":"value","num":42}`))
	require.NoError(t, err)
	assert.Equal(t, "value", j["key"])
	assert.Equal(t, float64(42), j["num"])
}

func TestJSONB_Scan_InvalidJSON(t *testing.T) {
	var j JSONB
	err := j.Scan([]byte(`not json`))
	assert.Error(t, err)
}

func TestJSONB_Scan_InvalidType(t *testing.T) {
	var j JSONB
	err := j.Scan("string type")
	assert.Error(t, err)
}

func TestJSONB_Value_Nil(t *testing.T) {
	var j JSONB
	val, err := j.Value()
	require.NoError(t, err)
	assert.Nil(t, val)
}

func TestJSONB_Value_ValidData(t *testing.T) {
	j := JSONB{"key": "value"}
	val, err := j.Value()
	require.NoError(t, err)

	driverVal, ok := val.([]byte)
	require.True(t, ok)

	var parsed map[string]interface{}
	err = json.Unmarshal(driverVal, &parsed)
	require.NoError(t, err)
	assert.Equal(t, "value", parsed["key"])
}

func TestJSONB_ImplementsInterfaces(t *testing.T) {
	j := JSONB{"test": "data"}
	var _ driver.Valuer = j
}

// ===== Enum Constants Tests =====

func TestUserRoleConstants(t *testing.T) {
	assert.Equal(t, UserRole("passenger"), RolePassenger)
	assert.Equal(t, UserRole("driver"), RoleDriver)
	assert.Equal(t, UserRole("dispatcher"), RoleDispatcher)
	assert.Equal(t, UserRole("admin"), RoleAdmin)
}

func TestUserStatusConstants(t *testing.T) {
	assert.Equal(t, UserStatus("active"), UserStatusActive)
	assert.Equal(t, UserStatus("inactive"), UserStatusInactive)
	assert.Equal(t, UserStatus("blocked"), UserStatusBlocked)
}

func TestOrderStatusConstants(t *testing.T) {
	assert.Equal(t, OrderStatus("created"), OrderStatusCreated)
	assert.Equal(t, OrderStatus("paid"), OrderStatusPaid)
	assert.Equal(t, OrderStatus("assigned"), OrderStatusAssigned)
	assert.Equal(t, OrderStatus("notified"), OrderStatusNotified)
	assert.Equal(t, OrderStatus("completed"), OrderStatusCompleted)
	assert.Equal(t, OrderStatus("canceled"), OrderStatusCanceled)
}

func TestPaymentStateConstants(t *testing.T) {
	assert.Equal(t, PaymentState("pending"), PaymentStatePending)
	assert.Equal(t, PaymentState("paid"), PaymentStatePaid)
	assert.Equal(t, PaymentState("refunded"), PaymentStateRefunded)
	assert.Equal(t, PaymentState("closed"), PaymentStateClosed)
}

func TestAssignmentStatusConstants(t *testing.T) {
	assert.Equal(t, AssignmentStatus("assigned"), AssignmentStatusAssigned)
	assert.Equal(t, AssignmentStatus("accepted"), AssignmentStatusAccepted)
	assert.Equal(t, AssignmentStatus("rejected"), AssignmentStatusRejected)
}

func TestDriverStatusConstants(t *testing.T) {
	assert.Equal(t, DriverStatus("available"), DriverStatusAvailable)
	assert.Equal(t, DriverStatus("busy"), DriverStatusBusy)
	assert.Equal(t, DriverStatus("offline"), DriverStatusOffline)
}

func TestPickupMethodConstants(t *testing.T) {
	assert.Equal(t, PickupMethod("group"), PickupMethodGroup)
	assert.Equal(t, PickupMethod("private"), PickupMethodPrivate)
	assert.Equal(t, PickupMethod("shuttle"), PickupMethodShuttle)
}

func TestRegistrationStatusConstants(t *testing.T) {
	assert.Equal(t, RegistrationStatus("draft"), RegistrationStatusDraft)
	assert.Equal(t, RegistrationStatus("submitted"), RegistrationStatusSubmitted)
	assert.Equal(t, RegistrationStatus("confirmed"), RegistrationStatusConfirmed)
}

func TestConsentScopeConstants(t *testing.T) {
	assert.Equal(t, ConsentScope("phone"), ConsentScopePhone)
	assert.Equal(t, ConsentScope("wechat"), ConsentScopeWechat)
	assert.Equal(t, ConsentScope("location"), ConsentScopeLocation)
}

// ===== BeforeCreate Hook Tests =====

func TestUser_BeforeCreate_Defaults(t *testing.T) {
	user := &User{}
	err := user.BeforeCreate(nil)
	require.NoError(t, err)
	assert.Equal(t, RolePassenger, user.Role)
	assert.Equal(t, UserStatusActive, user.Status)
}

func TestUser_BeforeCreate_PreserveExisting(t *testing.T) {
	user := &User{
		Role:   RoleAdmin,
		Status: UserStatusBlocked,
	}
	err := user.BeforeCreate(nil)
	require.NoError(t, err)
	assert.Equal(t, RoleAdmin, user.Role)
	assert.Equal(t, UserStatusBlocked, user.Status)
}

func TestPickupOrder_BeforeCreate_Defaults(t *testing.T) {
	order := &PickupOrder{}
	err := order.BeforeCreate(nil)
	require.NoError(t, err)
	assert.Equal(t, OrderStatusCreated, order.Status)
	assert.Equal(t, "CNY", order.Currency)
}

func TestPickupOrder_BeforeCreate_PreserveExisting(t *testing.T) {
	order := &PickupOrder{
		Status:   OrderStatusPaid,
		Currency: "USD",
	}
	err := order.BeforeCreate(nil)
	require.NoError(t, err)
	assert.Equal(t, OrderStatusPaid, order.Status)
	assert.Equal(t, "USD", order.Currency)
}

func TestPaymentOrder_BeforeCreate_Defaults(t *testing.T) {
	payment := &PaymentOrder{}
	err := payment.BeforeCreate(nil)
	require.NoError(t, err)
	assert.Equal(t, PaymentStatePending, payment.State)
	assert.Equal(t, "CNY", payment.Currency)
}

func TestPaymentOrder_BeforeCreate_PreserveExisting(t *testing.T) {
	payment := &PaymentOrder{
		State:    PaymentStatePaid,
		Currency: "USD",
	}
	err := payment.BeforeCreate(nil)
	require.NoError(t, err)
	assert.Equal(t, PaymentStatePaid, payment.State)
	assert.Equal(t, "USD", payment.Currency)
}

func TestAssignment_BeforeCreate_Defaults(t *testing.T) {
	assignment := &Assignment{}
	before := time.Now()
	err := assignment.BeforeCreate(nil)
	require.NoError(t, err)
	assert.Equal(t, AssignmentStatusAssigned, assignment.Status)
	assert.True(t, assignment.AssignedAt.After(before) || assignment.AssignedAt.Equal(before))
}

func TestAssignment_BeforeCreate_PreserveExisting(t *testing.T) {
	fixedTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	assignment := &Assignment{
		Status:     AssignmentStatusAccepted,
		AssignedAt: fixedTime,
	}
	err := assignment.BeforeCreate(nil)
	require.NoError(t, err)
	assert.Equal(t, AssignmentStatusAccepted, assignment.Status)
	assert.Equal(t, fixedTime, assignment.AssignedAt)
}

func TestDriver_BeforeCreate_Defaults(t *testing.T) {
	d := &Driver{}
	err := d.BeforeCreate(nil)
	require.NoError(t, err)
	assert.Equal(t, DriverStatusOffline, d.Status)
}

func TestDriver_BeforeCreate_PreserveExisting(t *testing.T) {
	d := &Driver{Status: DriverStatusAvailable}
	err := d.BeforeCreate(nil)
	require.NoError(t, err)
	assert.Equal(t, DriverStatusAvailable, d.Status)
}
