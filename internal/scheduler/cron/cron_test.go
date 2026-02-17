package cron

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type fakeLifecycle struct {
	hooks []fx.Hook
}

func (f *fakeLifecycle) Append(h fx.Hook) {
	f.hooks = append(f.hooks, h)
}

func newCronDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`CREATE TABLE requests (id INTEGER PRIMARY KEY AUTOINCREMENT, flight_no TEXT, arrival_date DATETIME, status TEXT, terminal TEXT, arrival_time_api DATETIME, pickup_buffer INTEGER, calc_pickup_time DATETIME)`).Error)
	return db
}

func TestSyncFlightService_BasicBranches(t *testing.T) {
	db := newCronDB(t)
	svc := NewSyncFlightService(db, zap.NewNop())
	svc.baseURL = ""

	err := svc.SyncFlightData(context.Background())
	require.NoError(t, err)

	_, err = svc.fetchFlight(context.Background(), "AA1")
	assert.Error(t, err)

	err = svc.batchUpdateByFlightNo(context.Background(), nil)
	require.NoError(t, err)
}

func TestRegisterCron_HookLifecycle(t *testing.T) {
	lc := &fakeLifecycle{}
	svc := &SyncFlightService{baseURL: "", logger: zap.NewNop()}
	RegisterCron(lc, svc, zap.NewNop())
	require.Len(t, lc.hooks, 1)
	require.NoError(t, lc.hooks[0].OnStart(context.Background()))
	time.Sleep(10 * time.Millisecond)
	require.NoError(t, lc.hooks[0].OnStop(context.Background()))
}

func TestSyncFlightService_WithConfiguredBaseURLAndBatchUpdate(t *testing.T) {
	db := newCronDB(t)
	today := time.Now().Format("2006-01-02")
	require.NoError(t, db.Exec(`INSERT INTO requests(flight_no,arrival_date,status,terminal,pickup_buffer) VALUES ('AA100', ?, 'pending', 'T1', 45)`, today).Error)

	svc := NewSyncFlightService(db, zap.NewNop())
	svc.baseURL = "http://example.test"

	err := svc.SyncFlightData(context.Background())
	require.NoError(t, err)

	err = svc.batchUpdateByFlightNo(context.Background(), []FlightResult{{
		FlightNo:    "AA100",
		Terminal:    "T5",
		ArrivalTime: time.Now(),
	}})
	require.NoError(t, err)
}
