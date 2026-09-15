package models

import (
	"time"

	"gorm.io/gorm"
)

type KategoriPeralatan struct {
	ID           uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	NamaKategori string         `gorm:"column:nama_kategori;size:100;not null;unique" json:"nama_kategori"`
	Description  string         `gorm:"column:description;size:255" json:"description"`
	CreatedAt    time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at" json:"-"`
}

func (KategoriPeralatan) TableName() string {
	return "kategori_peralatan"
}
