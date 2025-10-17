package service

import (
	"encoding/json"
	"fmt"
	"sort"

	"pickup/internal/repository"

	"go.uber.org/zap"
)

// SchemaService 定义导出数据库字段的服务接口
type SchemaService interface {
	GetDatabaseSchema() ([]TableSchema, error)
	ExportSchemaJSON() ([]byte, error)
}

// TableSchema 表结构信息
type TableSchema struct {
	TableName string         `json:"table_name"`
	Columns   []ColumnSchema `json:"columns"`
}

// ColumnSchema 列结构信息
type ColumnSchema struct {
	ColumnName    string  `json:"column_name"`
	DataType      string  `json:"data_type"`
	ColumnType    string  `json:"column_type"`
	ColumnKey     string  `json:"column_key"`
	IsNullable    string  `json:"is_nullable"`
	ColumnDefault *string `json:"column_default"`
	ColumnComment string  `json:"column_comment"`
}

type schemaService struct {
	repo   repository.SchemaRepository
	logger *zap.Logger
}

// NewSchemaService 创建数据库结构服务
func NewSchemaService(repo repository.SchemaRepository, logger *zap.Logger) SchemaService {
	return &schemaService{repo: repo, logger: logger}
}

// GetDatabaseSchema 获取数据库结构信息
func (s *schemaService) GetDatabaseSchema() ([]TableSchema, error) {
	columns, err := s.repo.GetAllColumns()
	if err != nil {
		s.logger.Error("failed to load database columns", zap.Error(err))
		return nil, fmt.Errorf("获取数据库字段失败: %w", err)
	}

	tables := make(map[string][]ColumnSchema)
	for _, col := range columns {
		tables[col.TableName] = append(tables[col.TableName], ColumnSchema{
			ColumnName:    col.ColumnName,
			DataType:      col.DataType,
			ColumnType:    col.ColumnType,
			ColumnKey:     col.ColumnKey,
			IsNullable:    col.IsNullable,
			ColumnDefault: col.ColumnDefault,
			ColumnComment: col.ColumnComment,
		})
	}

	tableNames := make([]string, 0, len(tables))
	for tableName := range tables {
		tableNames = append(tableNames, tableName)
	}
	sort.Strings(tableNames)

	result := make([]TableSchema, 0, len(tableNames))
	for _, tableName := range tableNames {
		columnList := tables[tableName]
		result = append(result, TableSchema{
			TableName: tableName,
			Columns:   columnList,
		})
	}
	return result, nil
}

// ExportSchemaJSON 以 JSON 格式导出数据库字段
func (s *schemaService) ExportSchemaJSON() ([]byte, error) {
	schema, err := s.GetDatabaseSchema()
	if err != nil {
		return nil, err
	}

	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		s.logger.Error("failed to marshal schema", zap.Error(err))
		return nil, fmt.Errorf("生成 JSON 数据失败: %w", err)
	}
	return data, nil
}
