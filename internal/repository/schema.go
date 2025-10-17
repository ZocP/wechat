package repository

import "gorm.io/gorm"

// ColumnInfo 描述数据库列信息
type ColumnInfo struct {
	TableName     string  `gorm:"column:table_name"`
	ColumnName    string  `gorm:"column:column_name"`
	DataType      string  `gorm:"column:data_type"`
	ColumnType    string  `gorm:"column:column_type"`
	ColumnKey     string  `gorm:"column:column_key"`
	IsNullable    string  `gorm:"column:is_nullable"`
	ColumnDefault *string `gorm:"column:column_default"`
	ColumnComment string  `gorm:"column:column_comment"`
}

// SchemaRepository 定义数据库结构仓储接口
type SchemaRepository interface {
	GetAllColumns() ([]ColumnInfo, error)
}

// schemaRepository 数据库结构仓储实现
type schemaRepository struct {
	db *gorm.DB
}

// NewSchemaRepository 创建数据库结构仓储
func NewSchemaRepository(db *gorm.DB) SchemaRepository {
	return &schemaRepository{db: db}
}

// GetAllColumns 查询当前数据库的全部字段
func (r *schemaRepository) GetAllColumns() ([]ColumnInfo, error) {
	var columns []ColumnInfo
	query := `SELECT TABLE_NAME, COLUMN_NAME, DATA_TYPE, COLUMN_TYPE, COLUMN_KEY, IS_NULLABLE, COLUMN_DEFAULT, COLUMN_COMMENT
FROM INFORMATION_SCHEMA.COLUMNS
WHERE TABLE_SCHEMA = DATABASE()
ORDER BY TABLE_NAME, ORDINAL_POSITION`
	if err := r.db.Raw(query).Scan(&columns).Error; err != nil {
		return nil, err
	}
	return columns, nil
}
