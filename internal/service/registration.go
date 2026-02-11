package service

import (
	"fmt"
	"time"

	"pickup/internal/model"
	"pickup/internal/repository"

	"go.uber.org/zap"
)

// RegistrationService 报名服务接口
type RegistrationService interface {
	CreateRegistration(userID uint, req *CreateRegistrationRequest) (*model.Registration, error)
	GetRegistration(id uint, userID uint) (*model.Registration, error)
	UpdateRegistration(id uint, userID uint, req *UpdateRegistrationRequest) (*model.Registration, error)
	GetUserRegistrations(userID uint) ([]*model.Registration, error)
	DeleteRegistration(id uint, userID uint) error
}

// CreateRegistrationRequest 创建报名请求
type CreateRegistrationRequest struct {
	Name          string `json:"name" binding:"required"`
	Phone         string `json:"phone" binding:"required"`
	WechatID      string `json:"wechat_id"`
	FlightNo      string `json:"flight_no" binding:"required"`
	ArrivalDate   string `json:"arrival_date" binding:"required"`
	ArrivalTime   string `json:"arrival_time" binding:"required"`
	DepartureCity string `json:"departure_city" binding:"required"`
	Companions    int    `json:"companions"`
	LuggageCount  int    `json:"luggage_count"`
	PickupMethod  string `json:"pickup_method" binding:"required"`
	Notes         string `json:"notes"`
}

// UpdateRegistrationRequest 更新报名请求
type UpdateRegistrationRequest struct {
	Name          *string `json:"name"`
	Phone         *string `json:"phone"`
	WechatID      *string `json:"wechat_id"`
	FlightNo      *string `json:"flight_no"`
	ArrivalDate   *string `json:"arrival_date"`
	ArrivalTime   *string `json:"arrival_time"`
	DepartureCity *string `json:"departure_city"`
	Companions    *int    `json:"companions"`
	LuggageCount  *int    `json:"luggage_count"`
	PickupMethod  *string `json:"pickup_method"`
	Notes         *string `json:"notes"`
}

// registrationService 报名服务实现
type registrationService struct {
	registrationRepo repository.RegistrationRepository
	logger           *zap.Logger
}

// NewRegistrationService 创建报名服务
func NewRegistrationService(
	registrationRepo repository.RegistrationRepository,
	logger *zap.Logger,
) RegistrationService {
	return &registrationService{
		registrationRepo: registrationRepo,
		logger:           logger,
	}
}

// CreateRegistration 创建报名
func (s *registrationService) CreateRegistration(userID uint, req *CreateRegistrationRequest) (*model.Registration, error) {
	// 验证非负数字段
	if req.Companions < 0 {
		return nil, fmt.Errorf("同行人数不能为负数")
	}
	if req.LuggageCount < 0 {
		return nil, fmt.Errorf("行李数量不能为负数")
	}

	// 验证接送方式
	if !isValidPickupMethod(req.PickupMethod) {
		return nil, fmt.Errorf("无效的接送方式: %s", req.PickupMethod)
	}

	// 解析日期和时间
	arrivalDate, err := time.Parse("2006-01-02", req.ArrivalDate)
	if err != nil {
		return nil, fmt.Errorf("无效的到达日期格式")
	}

	arrivalTime, err := time.Parse("15:04", req.ArrivalTime)
	if err != nil {
		return nil, fmt.Errorf("无效的到达时间格式")
	}

	registration := &model.Registration{
		UserID:        userID,
		Name:          req.Name,
		Phone:         req.Phone,
		WechatID:      req.WechatID,
		FlightNo:      req.FlightNo,
		ArrivalDate:   arrivalDate,
		ArrivalTime:   arrivalTime.Format("15:04:05"),
		DepartureCity: req.DepartureCity,
		Companions:    req.Companions,
		LuggageCount:  req.LuggageCount,
		PickupMethod:  model.PickupMethod(req.PickupMethod),
		Notes:         req.Notes,
		Status:        model.RegistrationStatusDraft,
	}

	if err := s.registrationRepo.Create(registration); err != nil {
		s.logger.Error("failed to create registration", zap.Error(err))
		return nil, fmt.Errorf("创建报名失败: %w", err)
	}

	s.logger.Info("registration created", zap.Uint("registration_id", registration.ID))
	return registration, nil
}

// GetRegistration 获取报名信息
func (s *registrationService) GetRegistration(id uint, userID uint) (*model.Registration, error) {
	registration, err := s.registrationRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("获取报名信息失败: %w", err)
	}

	// 检查权限：只能查看自己的报名
	if registration.UserID != userID {
		return nil, fmt.Errorf("无权限访问该报名")
	}

	return registration, nil
}

// UpdateRegistration 更新报名
func (s *registrationService) UpdateRegistration(id uint, userID uint, req *UpdateRegistrationRequest) (*model.Registration, error) {
	registration, err := s.registrationRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("获取报名信息失败: %w", err)
	}

	// 检查权限：只能修改自己的报名
	if registration.UserID != userID {
		return nil, fmt.Errorf("无权限修改该报名")
	}

	// 检查状态：只有草稿状态可以修改
	if registration.Status != model.RegistrationStatusDraft {
		return nil, fmt.Errorf("只有草稿状态的报名可以修改")
	}

	// 更新字段
	if req.Name != nil {
		registration.Name = *req.Name
	}
	if req.Phone != nil {
		registration.Phone = *req.Phone
	}
	if req.WechatID != nil {
		registration.WechatID = *req.WechatID
	}
	if req.FlightNo != nil {
		registration.FlightNo = *req.FlightNo
	}
	if req.ArrivalDate != nil {
		arrivalDate, err := time.Parse("2006-01-02", *req.ArrivalDate)
		if err != nil {
			return nil, fmt.Errorf("无效的到达日期格式")
		}
		registration.ArrivalDate = arrivalDate
	}
	if req.ArrivalTime != nil {
		arrivalTime, err := time.Parse("15:04", *req.ArrivalTime)
		if err != nil {
			return nil, fmt.Errorf("无效的到达时间格式")
		}
		registration.ArrivalTime = arrivalTime.Format("15:04:05")
	}
	if req.DepartureCity != nil {
		registration.DepartureCity = *req.DepartureCity
	}
	if req.Companions != nil {
		if *req.Companions < 0 {
			return nil, fmt.Errorf("同行人数不能为负数")
		}
		registration.Companions = *req.Companions
	}
	if req.LuggageCount != nil {
		if *req.LuggageCount < 0 {
			return nil, fmt.Errorf("行李数量不能为负数")
		}
		registration.LuggageCount = *req.LuggageCount
	}
	if req.PickupMethod != nil {
		if !isValidPickupMethod(*req.PickupMethod) {
			return nil, fmt.Errorf("无效的接送方式: %s", *req.PickupMethod)
		}
		registration.PickupMethod = model.PickupMethod(*req.PickupMethod)
	}
	if req.Notes != nil {
		registration.Notes = *req.Notes
	}

	if err := s.registrationRepo.Update(registration); err != nil {
		s.logger.Error("failed to update registration", zap.Error(err))
		return nil, fmt.Errorf("更新报名失败: %w", err)
	}

	s.logger.Info("registration updated", zap.Uint("registration_id", registration.ID))
	return registration, nil
}

// GetUserRegistrations 获取用户报名列表
func (s *registrationService) GetUserRegistrations(userID uint) ([]*model.Registration, error) {
	registrations, err := s.registrationRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("获取报名列表失败: %w", err)
	}
	return registrations, nil
}

// DeleteRegistration 删除报名
func (s *registrationService) DeleteRegistration(id uint, userID uint) error {
	registration, err := s.registrationRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("获取报名信息失败: %w", err)
	}

	// 检查权限：只能删除自己的报名
	if registration.UserID != userID {
		return fmt.Errorf("无权限删除该报名")
	}

	// 检查状态：只有草稿状态可以删除
	if registration.Status != model.RegistrationStatusDraft {
		return fmt.Errorf("只有草稿状态的报名可以删除")
	}

	if err := s.registrationRepo.Delete(id); err != nil {
		s.logger.Error("failed to delete registration", zap.Error(err))
		return fmt.Errorf("删除报名失败: %w", err)
	}

	s.logger.Info("registration deleted", zap.Uint("registration_id", id))
	return nil
}

// isValidPickupMethod 验证接送方式是否有效
func isValidPickupMethod(method string) bool {
	switch model.PickupMethod(method) {
	case model.PickupMethodGroup, model.PickupMethodPrivate, model.PickupMethodShuttle:
		return true
	default:
		return false
	}
}
