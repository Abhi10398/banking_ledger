package db_model

import (
	"github.com/google/uuid"
)

type Account struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name     string    `gorm:"column:name;not null"                           json:"name"`
	Balance  int64     `gorm:"column:balance;not null;default:0"              json:"balance"`
	Currency string    `gorm:"column:currency;not null;default:'INR'"         json:"currency"`
	BaseModel
}

func (Account) TableName() string {
	return "accounts"
}
