package models

import (
	"time"

	"gorm.io/gorm"
)

type Labs struct {
	ID        uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	NamaLabs  string         `gorm:"column:nama_labs;size:150;not null;unique" json:"nama_labs"`
	KodeLabs  string         `gorm:"column:kode_labs;size:50;not null;unique" json:"kode_labs"`
	ManagerID *uint64        `gorm:"column:manager_id" json:"manager_id"`
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at" json:"-"`

	// Relasi
	Manager *User `gorm:"foreignKey:ManagerID;references:UserID" json:"manager,omitempty"`
}

func (Labs) TableName() string {
	return "labs"
}
