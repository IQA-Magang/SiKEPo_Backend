package models

import (
	"time"

	"gorm.io/gorm"
)

type DetailPeminjaman struct {
	ID uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`

	PeminjamanID uint64 `gorm:"column:peminjaman_id;not null;index" json:"peminjaman_id"`

	PeralatanID uint64 `gorm:"column:peralatan_id;not null;index" json:"peralatan_id"`

	Jumlah uint `gorm:"column:jumlah;not null;default:1" json:"jumlah"`

	Status string `gorm:"column:status;not null;default:pending" json:"status"`

	VerifiedBy *uint64 `gorm:"column:verified_by;index" json:"verified_by"`

	VerifiedAt *time.Time `gorm:"column:verified_at" json:"verified_at"`

	VerificationNote string `gorm:"column:verification_note;type:text" json:"verification_note"`

	KondisiSaatPinjam *string `gorm:"column:kondisi_saat_pinjam" json:"kondisi_saat_pinjam"`

	KondisiSaatKembali *string `gorm:"column:kondisi_saat_kembali" json:"kondisi_saat_kembali"`

	Catatan string `gorm:"column:catatan;type:text" json:"catatan"`

	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`

	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`

	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at" json:"-"`

	// Relation
	Peralatan *Peralatan `gorm:"foreignKey:PeralatanID;references:ID" json:"peralatan,omitempty"`

	VerifiedByUser *User `gorm:"foreignKey:VerifiedBy;references:UserID" json:"verified_by_user,omitempty"`
}

func (DetailPeminjaman) TableName() string {
	return "detail_peminjaman"
}