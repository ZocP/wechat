package cron

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"pickup/internal/scheduler/models"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type FlightResult struct {
	FlightNo    string
	Terminal    string
	ArrivalTime time.Time
}

type FlightAPIResponse struct {
	FlightNo    string `json:"flight_no"`
	Terminal    string `json:"terminal"`
	ArrivalTime string `json:"arrival_time"`
}

type SyncFlightService struct {
	db      *gorm.DB
	logger  *zap.Logger
	baseURL string
	client  *http.Client
}

func NewSyncFlightService(db *gorm.DB, logger *zap.Logger) *SyncFlightService {
	return &SyncFlightService{
		db:      db,
		logger:  logger,
		baseURL: strings.TrimSpace(os.Getenv("FLIGHT_API_URL")),
		client:  &http.Client{Timeout: 8 * time.Second},
	}
}

func (s *SyncFlightService) SyncFlightData(ctx context.Context) error {
	if strings.TrimSpace(s.baseURL) == "" {
		s.logger.Info("flight sync skipped: FLIGHT_API_URL not configured")
		return nil
	}

	var flightNos []string
	today := time.Now().Format("2006-01-02")
	if err := s.db.Model(&models.Request{}).
		Where("arrival_date = ? AND status IN ?", today, []models.RequestStatus{models.RequestStatusPending, models.RequestStatusAssigned}).
		Distinct().
		Pluck("flight_no", &flightNos).Error; err != nil {
		return err
	}
	if len(flightNos) == 0 {
		return nil
	}

	s.logger.Info("flight sync placeholder: API integration reserved for future", zap.Int("flight_count", len(flightNos)))
	return nil
}

func (s *SyncFlightService) fetchFlight(ctx context.Context, flightNo string) (*FlightResult, error) {
	return nil, fmt.Errorf("flight api integration placeholder")
}

func (s *SyncFlightService) batchUpdateByFlightNo(ctx context.Context, updates []FlightResult) error {
	if len(updates) == 0 {
		return nil
	}

	terminalCase := "CASE flight_no"
	arrivalCase := "CASE flight_no"
	bufferCase := "CASE flight_no"
	pickupCase := "CASE flight_no"
	args := make([]any, 0, len(updates)*7)
	inArgs := make([]string, 0, len(updates))

	for _, u := range updates {
		buffer := 45
		if u.Terminal == "T5" {
			buffer = 90
		}
		pickup := u.ArrivalTime.Add(time.Duration(buffer) * time.Minute)

		terminalCase += " WHEN ? THEN ?"
		args = append(args, u.FlightNo, u.Terminal)
		arrivalCase += " WHEN ? THEN ?"
		args = append(args, u.FlightNo, u.ArrivalTime)
		bufferCase += " WHEN ? THEN ?"
		args = append(args, u.FlightNo, buffer)
		pickupCase += " WHEN ? THEN ?"
		args = append(args, u.FlightNo, pickup)

		inArgs = append(inArgs, "?")
	}
	terminalCase += " END"
	arrivalCase += " END"
	bufferCase += " END"
	pickupCase += " END"

	for _, u := range updates {
		args = append(args, u.FlightNo)
	}

	query := fmt.Sprintf(
		"UPDATE requests SET terminal = %s, arrival_time_api = %s, pickup_buffer = %s, calc_pickup_time = %s WHERE flight_no IN (%s) AND status IN ('pending','assigned')",
		terminalCase,
		arrivalCase,
		bufferCase,
		pickupCase,
		strings.Join(inArgs, ","),
	)

	return s.db.WithContext(ctx).Exec(query, args...).Error
}
