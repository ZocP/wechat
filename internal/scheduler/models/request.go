package models

import "time"

// Request 学生接机需求表
type Request struct {
	ID             uint          `gorm:"primaryKey" json:"id"`
	UserID         uint          `gorm:"column:user_id;not null;index:idx_requests_user_id" json:"user_id"`
	FlightNo       string        `gorm:"column:flight_no;type:varchar(20);not null;index:idx_requests_flight_no" json:"flight_no"`
	ArrivalDate    time.Time     `gorm:"column:arrival_date;type:date;not null;index:idx_requests_arrival_date" json:"arrival_date"`
	Terminal       string        `gorm:"type:varchar(10);not null" json:"terminal"`
	CheckedBags    int           `gorm:"column:checked_bags;not null;default:0" json:"checked_bags"`
	CarryOnBags    int           `gorm:"column:carry_on_bags;not null;default:0" json:"carry_on_bags"`
	Status         RequestStatus `gorm:"type:enum('pending','assigned','published');not null;default:'pending';index:idx_requests_status" json:"status"`
	ArrivalTimeAPI *time.Time    `gorm:"column:arrival_time_api;type:datetime" json:"arrival_time_api,omitempty"`
	PickupBuffer   int           `gorm:"column:pickup_buffer;not null;default:45" json:"pickup_buffer"`
	CalcPickupTime *time.Time    `gorm:"column:calc_pickup_time;type:datetime" json:"calc_pickup_time,omitempty"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`

	User   *User   `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"user,omitempty"`
	Shifts []Shift `gorm:"many2many:shift_requests;joinForeignKey:RequestID;joinReferences:ShiftID" json:"-"`
	Shift  *Shift  `gorm:"-" json:"shift,omitempty"`
}

func (Request) TableName() string {
	return "requests"
}
