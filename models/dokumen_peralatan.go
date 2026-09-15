package models

import (
	"time"

	"gorm.io/gorm"
)

type DokumenPeralatan struct {
	ID          uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	NamaDokumen string         `gorm:"column:nama_dokumen;size:150;not null;unique" json:"nama_dokumen"`
	PathDokumen string         `gorm:"column:path_dokumen;size:255;not null" json:"path_dokumen"`
	PeralatanID uint64         `gorm:"column:peralatan_id;not null" json:"peralatan_id"`
	CreatedAt   time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at" json:"-"`

	// Relasi
	Peralatan *Peralatan `gorm:"foreignKey:PeralatanID;references:ID" json:"peralatan,omitempty"`
}

func (DokumenPeralatan) TableName() string {
	return "dokumen_peralatan"
}
