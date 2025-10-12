package repository

import (
	"time"

	"pickup/internal/model"

	"gorm.io/gorm"
)

// NoticeRepository 消息板仓储接口
type NoticeRepository interface {
	Create(notice *model.Notice) error
	GetByID(id uint) (*model.Notice, error)
	GetVisibleNotices() ([]*model.Notice, error)
	GetByFlightNo(flightNo string) ([]*model.Notice, error)
	Update(notice *model.Notice) error
	Delete(id uint) error
}

// noticeRepository 消息板仓储实现
type noticeRepository struct {
	db *gorm.DB
}

// NewNoticeRepository 创建消息板仓储
func NewNoticeRepository(db *gorm.DB) NoticeRepository {
	return &noticeRepository{db: db}
}

// Create 创建消息
func (r *noticeRepository) Create(notice *model.Notice) error {
	return r.db.Create(notice).Error
}

// GetByID 根据ID获取消息
func (r *noticeRepository) GetByID(id uint) (*model.Notice, error) {
	var notice model.Notice
	err := r.db.Preload("Creator").First(&notice, id).Error
	if err != nil {
		return nil, err
	}
	return &notice, nil
}

// GetVisibleNotices 获取当前可见的消息
func (r *noticeRepository) GetVisibleNotices() ([]*model.Notice, error) {
	var notices []*model.Notice
	now := time.Now()
	err := r.db.Where("visible_from <= ? AND visible_to >= ?", now, now).
		Preload("Creator").
		Order("created_at DESC").
		Find(&notices).Error
	if err != nil {
		return nil, err
	}
	return notices, nil
}

// GetByFlightNo 根据航班号获取消息
func (r *noticeRepository) GetByFlightNo(flightNo string) ([]*model.Notice, error) {
	var notices []*model.Notice
	now := time.Now()
	err := r.db.Where("flight_no = ? AND visible_from <= ? AND visible_to >= ?", flightNo, now, now).
		Preload("Creator").
		Order("created_at DESC").
		Find(&notices).Error
	if err != nil {
		return nil, err
	}
	return notices, nil
}

// Update 更新消息
func (r *noticeRepository) Update(notice *model.Notice) error {
	return r.db.Save(notice).Error
}

// Delete 删除消息
func (r *noticeRepository) Delete(id uint) error {
	return r.db.Delete(&model.Notice{}, id).Error
}
