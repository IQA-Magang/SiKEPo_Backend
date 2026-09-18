package models

import (
	"time"

	"gorm.io/gorm"
)

type KelompokAsset struct {
	ID        uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	LabID     uint64         `gorm:"column:lab_id;not null;index" json:"lab_id"`
	PICID     uint64         `gorm:"column:pic_id;not null;index" json:"pic_id"`
	Kode      string         `gorm:"column:kode;size:50;not null" json:"kode"`
	Nama      string         `gorm:"column:nama;size:100;not null" json:"nama"`
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`

	// =========================
	// RELATION
	// =========================

	Lab *Labs `gorm:"foreignKey:LabID;references:ID" json:"lab,omitempty"`
	PIC *User `gorm:"foreignKey:PICID;references:UserID" json:"pic,omitempty"`
}

func (KelompokAsset) TableName() string {
	return "kelompok_asset"
}
