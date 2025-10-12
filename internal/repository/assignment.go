package repository

import (
	"pickup/internal/model"

	"gorm.io/gorm"
)

// AssignmentRepository 分配仓储接口
type AssignmentRepository interface {
	Create(assignment *model.Assignment) error
	GetByID(id uint) (*model.Assignment, error)
	GetByOrderID(orderID uint) (*model.Assignment, error)
	GetByDriverID(driverID uint) ([]*model.Assignment, error)
	Update(assignment *model.Assignment) error
	UpdateStatus(assignmentID uint, status model.AssignmentStatus) error
}

// assignmentRepository 分配仓储实现
type assignmentRepository struct {
	db *gorm.DB
}

// NewAssignmentRepository 创建分配仓储
func NewAssignmentRepository(db *gorm.DB) AssignmentRepository {
	return &assignmentRepository{db: db}
}

// Create 创建分配
func (r *assignmentRepository) Create(assignment *model.Assignment) error {
	return r.db.Create(assignment).Error
}

// GetByID 根据ID获取分配
func (r *assignmentRepository) GetByID(id uint) (*model.Assignment, error) {
	var assignment model.Assignment
	err := r.db.Preload("Order").Preload("Driver").First(&assignment, id).Error
	if err != nil {
		return nil, err
	}
	return &assignment, nil
}

// GetByOrderID 根据订单ID获取分配
func (r *assignmentRepository) GetByOrderID(orderID uint) (*model.Assignment, error) {
	var assignment model.Assignment
	err := r.db.Where("order_id = ?", orderID).First(&assignment).Error
	if err != nil {
		return nil, err
	}
	return &assignment, nil
}

// GetByDriverID 根据司机ID获取分配列表
func (r *assignmentRepository) GetByDriverID(driverID uint) ([]*model.Assignment, error) {
	var assignments []*model.Assignment
	err := r.db.Where("driver_id = ?", driverID).
		Preload("Order").
		Preload("Driver").
		Order("created_at DESC").
		Find(&assignments).Error
	if err != nil {
		return nil, err
	}
	return assignments, nil
}

// Update 更新分配
func (r *assignmentRepository) Update(assignment *model.Assignment) error {
	return r.db.Save(assignment).Error
}

// UpdateStatus 更新分配状态
func (r *assignmentRepository) UpdateStatus(assignmentID uint, status model.AssignmentStatus) error {
	updates := map[string]interface{}{"status": status}
	now := gorm.Expr("NOW()")

	switch status {
	case model.AssignmentStatusAccepted:
		updates["accepted_at"] = now
	case model.AssignmentStatusRejected:
		updates["rejected_at"] = now
	}

	return r.db.Model(&model.Assignment{}).Where("id = ?", assignmentID).Updates(updates).Error
}
