package models

import (
	"time"

	"gorm.io/gorm"
)

type Ruangan struct {
	ID          uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	NamaRuangan string         `gorm:"column:nama_ruangan;size:150;not null;unique" json:"nama_ruangan"`
	PICUserID   *uint64        `gorm:"column:pic_user_id" json:"pic_user_id"`
	CreatedAt   time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at" json:"-"`

	// Relasi
	PICUser *User `gorm:"foreignKey:PICUserID;references:UserID" json:"pic_user,omitempty"`
}

func (Ruangan) TableName() string {
	return "ruangan"
}
