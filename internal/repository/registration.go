package repository

import (
	"pickup/internal/model"

	"gorm.io/gorm"
)

// RegistrationRepository 报名仓储接口
type RegistrationRepository interface {
	Create(registration *model.Registration) error
	GetByID(id uint) (*model.Registration, error)
	GetByUserID(userID uint) ([]*model.Registration, error)
	Update(registration *model.Registration) error
	Delete(id uint) error
}

// registrationRepository 报名仓储实现
type registrationRepository struct {
	db *gorm.DB
}

// NewRegistrationRepository 创建报名仓储
func NewRegistrationRepository(db *gorm.DB) RegistrationRepository {
	return &registrationRepository{db: db}
}

// Create 创建报名
func (r *registrationRepository) Create(registration *model.Registration) error {
	return r.db.Create(registration).Error
}

// GetByID 根据ID获取报名
func (r *registrationRepository) GetByID(id uint) (*model.Registration, error) {
	var registration model.Registration
	err := r.db.Preload("User").First(&registration, id).Error
	if err != nil {
		return nil, err
	}
	return &registration, nil
}

// GetByUserID 根据用户ID获取报名列表
func (r *registrationRepository) GetByUserID(userID uint) ([]*model.Registration, error) {
	var registrations []*model.Registration
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&registrations).Error
	if err != nil {
		return nil, err
	}
	return registrations, nil
}

// Update 更新报名
func (r *registrationRepository) Update(registration *model.Registration) error {
	return r.db.Save(registration).Error
}

// Delete 删除报名
func (r *registrationRepository) Delete(id uint) error {
	return r.db.Delete(&model.Registration{}, id).Error
}
