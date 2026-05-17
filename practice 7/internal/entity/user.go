package entity

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID        uuid.UUID      `json:"ID" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CreatedAt time.Time      `json:"CreatedAt"`
	UpdatedAt time.Time      `json:"UpdatedAt"`
	DeletedAt gorm.DeletedAt `json:"DeletedAt" gorm:"index"`
	Username  string         `json:"Username" gorm:"unique"`
	Email     string         `json:"Email" gorm:"unique"`
	Password  string         `json:"Password"`
	Role      string         `json:"Role"`
	Verified  bool           `json:"Verified"`
}
