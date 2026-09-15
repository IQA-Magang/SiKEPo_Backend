package models

import (
	"time"

	"gorm.io/gorm"
)

type Notification struct {
	ID        uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID    uint64         `gorm:"column:user_id;not null;index" json:"user_id"`
	Type      string         `gorm:"column:type;size:50;not null;default:'equipment_added'" json:"type"`
	Title     string         `gorm:"column:title;size:150;not null" json:"title"`
	Message   string         `gorm:"column:message;type:text;not null" json:"message"`
	IsRead    bool           `gorm:"column:is_read;default:false" json:"is_read"`
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at" json:"-"`
}

func (Notification) TableName() string {
	return "notifications"
}
