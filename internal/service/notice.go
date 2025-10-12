package service

import (
	"fmt"
	"time"

	"pickup/internal/model"
	"pickup/internal/repository"

	"go.uber.org/zap"
)

// NoticeService 消息板服务接口
type NoticeService interface {
	CreateNotice(userID uint, req *CreateNoticeRequest) (*model.Notice, error)
	GetNotice(id uint) (*model.Notice, error)
	GetVisibleNotices() ([]*model.Notice, error)
	GetNoticesByFlightNo(flightNo string) ([]*model.Notice, error)
	UpdateNotice(id uint, userID uint, req *UpdateNoticeRequest) (*model.Notice, error)
	DeleteNotice(id uint, userID uint) error
}

// CreateNoticeRequest 创建消息请求
type CreateNoticeRequest struct {
	FlightNo       string    `json:"flight_no" binding:"required"`
	Terminal       string    `json:"terminal" binding:"required"`
	PickupBatch    string    `json:"pickup_batch" binding:"required"`
	ArrivalAirport string    `json:"arrival_airport" binding:"required"`
	MeetingPoint   string    `json:"meeting_point" binding:"required"`
	GuideText      string    `json:"guide_text"`
	MapURL         string    `json:"map_url"`
	ContactName    string    `json:"contact_name"`
	ContactPhone   string    `json:"contact_phone"`
	VisibleFrom    time.Time `json:"visible_from" binding:"required"`
	VisibleTo      time.Time `json:"visible_to" binding:"required"`
}

// UpdateNoticeRequest 更新消息请求
type UpdateNoticeRequest struct {
	FlightNo       *string    `json:"flight_no"`
	Terminal       *string    `json:"terminal"`
	PickupBatch    *string    `json:"pickup_batch"`
	ArrivalAirport *string    `json:"arrival_airport"`
	MeetingPoint   *string    `json:"meeting_point"`
	GuideText      *string    `json:"guide_text"`
	MapURL         *string    `json:"map_url"`
	ContactName    *string    `json:"contact_name"`
	ContactPhone   *string    `json:"contact_phone"`
	VisibleFrom    *time.Time `json:"visible_from"`
	VisibleTo      *time.Time `json:"visible_to"`
}

// noticeService 消息板服务实现
type noticeService struct {
	noticeRepo repository.NoticeRepository
	logger     *zap.Logger
}

// NewNoticeService 创建消息板服务
func NewNoticeService(
	noticeRepo repository.NoticeRepository,
	logger *zap.Logger,
) NoticeService {
	return &noticeService{
		noticeRepo: noticeRepo,
		logger:     logger,
	}
}

// CreateNotice 创建消息
func (s *noticeService) CreateNotice(userID uint, req *CreateNoticeRequest) (*model.Notice, error) {
	notice := &model.Notice{
		FlightNo:       req.FlightNo,
		Terminal:       req.Terminal,
		PickupBatch:    req.PickupBatch,
		ArrivalAirport: req.ArrivalAirport,
		MeetingPoint:   req.MeetingPoint,
		GuideText:      req.GuideText,
		MapURL:         req.MapURL,
		ContactName:    req.ContactName,
		ContactPhone:   req.ContactPhone,
		VisibleFrom:    req.VisibleFrom,
		VisibleTo:      req.VisibleTo,
		CreatedBy:      userID,
	}

	if err := s.noticeRepo.Create(notice); err != nil {
		s.logger.Error("failed to create notice", zap.Error(err))
		return nil, fmt.Errorf("创建消息失败: %w", err)
	}

	s.logger.Info("notice created", zap.Uint("notice_id", notice.ID))
	return notice, nil
}

// GetNotice 获取消息详情
func (s *noticeService) GetNotice(id uint) (*model.Notice, error) {
	notice, err := s.noticeRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("获取消息详情失败: %w", err)
	}
	return notice, nil
}

// GetVisibleNotices 获取当前可见的消息
func (s *noticeService) GetVisibleNotices() ([]*model.Notice, error) {
	notices, err := s.noticeRepo.GetVisibleNotices()
	if err != nil {
		return nil, fmt.Errorf("获取可见消息失败: %w", err)
	}
	return notices, nil
}

// GetNoticesByFlightNo 根据航班号获取消息
func (s *noticeService) GetNoticesByFlightNo(flightNo string) ([]*model.Notice, error) {
	notices, err := s.noticeRepo.GetByFlightNo(flightNo)
	if err != nil {
		return nil, fmt.Errorf("获取航班消息失败: %w", err)
	}
	return notices, nil
}

// UpdateNotice 更新消息
func (s *noticeService) UpdateNotice(id uint, userID uint, req *UpdateNoticeRequest) (*model.Notice, error) {
	notice, err := s.noticeRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("获取消息详情失败: %w", err)
	}

	// 检查权限：只有创建者可以修改
	if notice.CreatedBy != userID {
		return nil, fmt.Errorf("无权限修改该消息")
	}

	// 更新字段
	if req.FlightNo != nil {
		notice.FlightNo = *req.FlightNo
	}
	if req.Terminal != nil {
		notice.Terminal = *req.Terminal
	}
	if req.PickupBatch != nil {
		notice.PickupBatch = *req.PickupBatch
	}
	if req.ArrivalAirport != nil {
		notice.ArrivalAirport = *req.ArrivalAirport
	}
	if req.MeetingPoint != nil {
		notice.MeetingPoint = *req.MeetingPoint
	}
	if req.GuideText != nil {
		notice.GuideText = *req.GuideText
	}
	if req.MapURL != nil {
		notice.MapURL = *req.MapURL
	}
	if req.ContactName != nil {
		notice.ContactName = *req.ContactName
	}
	if req.ContactPhone != nil {
		notice.ContactPhone = *req.ContactPhone
	}
	if req.VisibleFrom != nil {
		notice.VisibleFrom = *req.VisibleFrom
	}
	if req.VisibleTo != nil {
		notice.VisibleTo = *req.VisibleTo
	}

	if err := s.noticeRepo.Update(notice); err != nil {
		s.logger.Error("failed to update notice", zap.Error(err))
		return nil, fmt.Errorf("更新消息失败: %w", err)
	}

	s.logger.Info("notice updated", zap.Uint("notice_id", notice.ID))
	return notice, nil
}

// DeleteNotice 删除消息
func (s *noticeService) DeleteNotice(id uint, userID uint) error {
	notice, err := s.noticeRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("获取消息详情失败: %w", err)
	}

	// 检查权限：只有创建者可以删除
	if notice.CreatedBy != userID {
		return fmt.Errorf("无权限删除该消息")
	}

	if err := s.noticeRepo.Delete(id); err != nil {
		s.logger.Error("failed to delete notice", zap.Error(err))
		return fmt.Errorf("删除消息失败: %w", err)
	}

	s.logger.Info("notice deleted", zap.Uint("notice_id", id))
	return nil
}
