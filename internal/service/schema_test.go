package service

import (
	"errors"
	"testing"

	"pickup/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// ===== Mock Schema Repository =====

type mockSchemaRepo struct {
	mock.Mock
}

func (m *mockSchemaRepo) GetAllColumns() ([]repository.ColumnInfo, error) {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]repository.ColumnInfo), args.Error(1)
	}
	return nil, args.Error(1)
}

// ===== Schema Service Tests =====

func TestSchemaService_GetDatabaseSchema_Success(t *testing.T) {
	repo := new(mockSchemaRepo)
	svc := NewSchemaService(repo, zap.NewNop())

	columns := []repository.ColumnInfo{
		{TableName: "users", ColumnName: "id", DataType: "int", ColumnType: "int unsigned", ColumnKey: "PRI", IsNullable: "NO"},
		{TableName: "users", ColumnName: "name", DataType: "varchar", ColumnType: "varchar(64)", ColumnKey: "", IsNullable: "NO"},
		{TableName: "orders", ColumnName: "id", DataType: "int", ColumnType: "int unsigned", ColumnKey: "PRI", IsNullable: "NO"},
	}
	repo.On("GetAllColumns").Return(columns, nil).Once()

	schema, err := svc.GetDatabaseSchema()
	require.NoError(t, err)
	assert.Len(t, schema, 2) // 2 tables

	// Should be sorted alphabetically
	assert.Equal(t, "orders", schema[0].TableName)
	assert.Len(t, schema[0].Columns, 1)
	assert.Equal(t, "users", schema[1].TableName)
	assert.Len(t, schema[1].Columns, 2)
	repo.AssertExpectations(t)
}

func TestSchemaService_GetDatabaseSchema_Empty(t *testing.T) {
	repo := new(mockSchemaRepo)
	svc := NewSchemaService(repo, zap.NewNop())

	repo.On("GetAllColumns").Return([]repository.ColumnInfo{}, nil).Once()

	schema, err := svc.GetDatabaseSchema()
	require.NoError(t, err)
	assert.Empty(t, schema)
}

func TestSchemaService_GetDatabaseSchema_Error(t *testing.T) {
	repo := new(mockSchemaRepo)
	svc := NewSchemaService(repo, zap.NewNop())

	repo.On("GetAllColumns").Return(nil, errors.New("db error")).Once()

	schema, err := svc.GetDatabaseSchema()
	assert.Error(t, err)
	assert.Nil(t, schema)
	assert.Contains(t, err.Error(), "获取数据库字段失败")
}

func TestSchemaService_ExportSchemaJSON_Success(t *testing.T) {
	repo := new(mockSchemaRepo)
	svc := NewSchemaService(repo, zap.NewNop())

	columns := []repository.ColumnInfo{
		{TableName: "users", ColumnName: "id", DataType: "int", ColumnType: "int unsigned", ColumnKey: "PRI", IsNullable: "NO"},
	}
	repo.On("GetAllColumns").Return(columns, nil).Once()

	data, err := svc.ExportSchemaJSON()
	require.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Contains(t, string(data), "users")
	assert.Contains(t, string(data), "id")
}

func TestSchemaService_ExportSchemaJSON_Error(t *testing.T) {
	repo := new(mockSchemaRepo)
	svc := NewSchemaService(repo, zap.NewNop())

	repo.On("GetAllColumns").Return(nil, errors.New("db error")).Once()

	data, err := svc.ExportSchemaJSON()
	assert.Error(t, err)
	assert.Nil(t, data)
}

func TestSchemaService_ColumnDefaultPointer(t *testing.T) {
	repo := new(mockSchemaRepo)
	svc := NewSchemaService(repo, zap.NewNop())

	defaultVal := "0"
	columns := []repository.ColumnInfo{
		{TableName: "test", ColumnName: "col1", DataType: "int", ColumnDefault: &defaultVal},
		{TableName: "test", ColumnName: "col2", DataType: "int", ColumnDefault: nil},
	}
	repo.On("GetAllColumns").Return(columns, nil).Once()

	schema, err := svc.GetDatabaseSchema()
	require.NoError(t, err)
	assert.Len(t, schema, 1)
	assert.Len(t, schema[0].Columns, 2)
	assert.NotNil(t, schema[0].Columns[0].ColumnDefault)
	assert.Equal(t, "0", *schema[0].Columns[0].ColumnDefault)
	assert.Nil(t, schema[0].Columns[1].ColumnDefault)
}
