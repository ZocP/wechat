package models

// Driver 运力池/车辆表
type Driver struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	Name       string `gorm:"type:varchar(64);not null" json:"name"`
	CarModel   string `gorm:"column:car_model;type:varchar(64);not null" json:"car_model"`
	MaxSeats   int    `gorm:"column:max_seats;not null" json:"max_seats"`
	MaxChecked int    `gorm:"column:max_checked;not null" json:"max_checked"`
	MaxCarryOn int    `gorm:"column:max_carry_on;not null" json:"max_carry_on"`
}

func (Driver) TableName() string {
	return "drivers"
}
