package db_model

import "time"

// BaseModel provides the common created_at timestamp for all db models.
type BaseModel struct {
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}
