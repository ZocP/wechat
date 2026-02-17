package models

import "time"

type UserRole string

const (
	UserRoleStudent UserRole = "student"
	UserRoleStaff   UserRole = "staff"
	UserRoleAdmin   UserRole = "admin"
)

type RequestStatus string

const (
	RequestStatusPending   RequestStatus = "pending"
	RequestStatusAssigned  RequestStatus = "assigned"
	RequestStatusPublished RequestStatus = "published"
)

type ShiftStatus string

const (
	ShiftStatusDraft     ShiftStatus = "draft"
	ShiftStatusPublished ShiftStatus = "published"
)

// DateOnly 用于 GORM date 字段。
type DateOnly = time.Time
