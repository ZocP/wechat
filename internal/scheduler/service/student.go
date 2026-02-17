package service

import (
	"errors"
	"time"

	"pickup/internal/scheduler/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StudentService struct {
	db *gorm.DB
}

func NewStudentService(db *gorm.DB) *StudentService {
	return &StudentService{db: db}
}

type CreateRequestInput struct {
	FlightNo            string `json:"flight_no" binding:"required"`
	ArrivalDate         string `json:"arrival_date" binding:"required"`
	Terminal            string `json:"terminal" binding:"required"`
	CheckedBags         int    `json:"checked_bags"`
	CarryOnBags         int    `json:"carry_on_bags"`
	ExpectedArrivalTime string `json:"expected_arrival_time" binding:"required"`
}

type UpdateRequestInput struct {
	FlightNo            *string `json:"flight_no"`
	ArrivalDate         *string `json:"arrival_date"`
	Terminal            *string `json:"terminal"`
	CheckedBags         *int    `json:"checked_bags"`
	CarryOnBags         *int    `json:"carry_on_bags"`
	ExpectedArrivalTime *string `json:"expected_arrival_time"`
}

func pickupBufferByTerminal(terminal string) int {
	if terminal == "T5" {
		return 90
	}
	return 45
}

func (s *StudentService) CreateRequest(userID uint, input CreateRequestInput) (*models.Request, error) {
	var existingCount int64
	if err := s.db.Model(&models.Request{}).Where("user_id = ?", userID).Count(&existingCount).Error; err != nil {
		return nil, err
	}
	if existingCount > 0 {
		return nil, errors.New("user already has a request")
	}

	arrivalDate, err := time.Parse("2006-01-02", input.ArrivalDate)
	if err != nil {
		return nil, err
	}
	expectedArrivalTime, err := time.Parse("2006-01-02 15:04:05", input.ExpectedArrivalTime)
	if err != nil {
		return nil, err
	}
	buffer := pickupBufferByTerminal(input.Terminal)
	calcPickupTime := expectedArrivalTime.Add(time.Duration(buffer) * time.Minute)

	req := models.Request{
		UserID:         userID,
		FlightNo:       input.FlightNo,
		ArrivalDate:    arrivalDate,
		Terminal:       input.Terminal,
		CheckedBags:    input.CheckedBags,
		CarryOnBags:    input.CarryOnBags,
		Status:         models.RequestStatusPending,
		ArrivalTimeAPI: &expectedArrivalTime,
		PickupBuffer:   buffer,
		CalcPickupTime: &calcPickupTime,
	}
	if err := s.db.Omit(clause.Associations).Create(&req).Error; err != nil {
		return nil, err
	}
	return &req, nil
}

func (s *StudentService) ListMyRequests(userID uint) ([]models.Request, error) {
	var reqs []models.Request
	if err := s.db.Where("user_id = ?", userID).
		Preload("Shifts.Driver").
		Find(&reqs).Error; err != nil {
		return nil, err
	}
	for i := range reqs {
		if reqs[i].Status != models.RequestStatusPublished || len(reqs[i].Shifts) == 0 {
			reqs[i].Shift = nil
			continue
		}
		reqs[i].Shift = &reqs[i].Shifts[0]
	}
	return reqs, nil
}

func (s *StudentService) UpdatePendingRequest(userID, requestID uint, input UpdateRequestInput) (*models.Request, error) {
	var req models.Request
	if err := s.db.Where("id = ? AND user_id = ?", requestID, userID).First(&req).Error; err != nil {
		return nil, err
	}
	if req.Status != models.RequestStatusPending {
		return nil, errors.New("only pending request can be updated")
	}
	if input.FlightNo != nil {
		req.FlightNo = *input.FlightNo
	}
	if input.ArrivalDate != nil {
		arrivalDate, err := time.Parse("2006-01-02", *input.ArrivalDate)
		if err != nil {
			return nil, err
		}
		req.ArrivalDate = arrivalDate
	}
	if input.Terminal != nil {
		req.Terminal = *input.Terminal
		req.PickupBuffer = pickupBufferByTerminal(*input.Terminal)
		if req.ArrivalTimeAPI != nil {
			pickup := req.ArrivalTimeAPI.Add(time.Duration(req.PickupBuffer) * time.Minute)
			req.CalcPickupTime = &pickup
		}
	}
	if input.CheckedBags != nil {
		req.CheckedBags = *input.CheckedBags
	}
	if input.CarryOnBags != nil {
		req.CarryOnBags = *input.CarryOnBags
	}
	if input.ExpectedArrivalTime != nil {
		expectedArrivalTime, err := time.Parse("2006-01-02 15:04:05", *input.ExpectedArrivalTime)
		if err != nil {
			return nil, err
		}
		req.ArrivalTimeAPI = &expectedArrivalTime
		pickup := expectedArrivalTime.Add(time.Duration(req.PickupBuffer) * time.Minute)
		req.CalcPickupTime = &pickup
	}
	if err := s.db.Omit(clause.Associations).Save(&req).Error; err != nil {
		return nil, err
	}
	return &req, nil
}
