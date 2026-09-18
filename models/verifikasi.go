package models

import (
	"time"

	"gorm.io/gorm"
)

type Verifikasi struct {
	IDVerifikasi      uint64         `gorm:"column:id_verifikasi;primaryKey;autoIncrement" json:"id_verifikasi"`
	IDPeralatan       uint64         `gorm:"column:id_peralatan;not null;index" json:"id_peralatan"`
	TanggalVerifikasi time.Time      `gorm:"column:tanggal_verifikasi;not null" json:"tanggal_verifikasi"`
	KodeAktivitas     string         `gorm:"column:kode_aktivitas;size:100;not null" json:"kode_aktivitas"`
	IDKriteria        *uint64        `gorm:"column:id_kriteria;index" json:"id_kriteria"`
	Keputusan         *string        `gorm:"column:keputusan;type:enum('Layak','Tidak Layak')" json:"keputusan"`
	TindakLanjut      string         `gorm:"column:tindak_lanjut;type:text" json:"tindak_lanjut"`
	VerifiedBy        *uint64        `gorm:"column:verified_by;index" json:"verified_by"`
	VerifiedAt        *time.Time     `gorm:"column:verified_at" json:"verified_at"`
	Catatan           string         `gorm:"column:catatan;type:text" json:"catatan"`
	CreatedAt         time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"column:deleted_at" json:"-"`

	// Relasi ke peralatan
	Peralatan *Peralatan `gorm:"foreignKey:IDPeralatan;references:ID" json:"peralatan,omitempty"`
	// Relasi ke user sebagai Manager/pengesah
	VerifiedByUser *User `gorm:"foreignKey:VerifiedBy;references:UserID" json:"verified_by_user,omitempty"`
	// Relasi ke hasil verifikasi
	HasilVerifikasi []HasilVerifikasi `gorm:"foreignKey:IDVerifikasi;references:IDVerifikasi" json:"hasil_verifikasi,omitempty"`
}

func (Verifikasi) TableName() string {
	return "verifikasi"
}
