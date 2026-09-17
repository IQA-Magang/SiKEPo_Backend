package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	UserID    uint64         `gorm:"column:user_id;primaryKey;autoIncrement" json:"user_id"`
	NIP       string         `gorm:"column:nip;size:30;not null;unique" json:"nip"`
	Name      string         `gorm:"column:name;size:100;not null" json:"name"`
	Email     string         `gorm:"column:email;size:150;not null;unique" json:"email"`
	Password  string         `gorm:"column:password;size:255;not null" json:"-"`
	Role      string         `gorm:"column:role;not null" json:"role"`
	Position  string         `gorm:"column:position;size:100;not null" json:"position"`
	PIC       bool           `gorm:"column:pic;not null;default:false" json:"pic"`
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at" json:"-"`
}

func (User) TableName() string {
	return "users"
}