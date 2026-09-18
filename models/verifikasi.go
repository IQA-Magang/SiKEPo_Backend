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
	Status            string         `gorm:"column:status;type:enum('Draft','Diajukan','Disetujui','Ditolak');default:'Draft';index" json:"status"`
	Keputusan         *string        `gorm:"column:keputusan;type:enum('Layak','Tidak Layak')" json:"keputusan"`
	TindakLanjut      string         `gorm:"column:tindak_lanjut;type:text" json:"tindak_lanjut"`
	PICID             *uint64        `gorm:"column:pic_id;index" json:"pic_id"`
	PICSignature      string         `gorm:"column:pic_signature;type:text" json:"pic_signature"`
	PICSignedAt       *time.Time     `gorm:"column:pic_signed_at" json:"pic_signed_at"`
	VerifiedBy        *uint64        `gorm:"column:verified_by;index" json:"verified_by"`
	VerifiedAt        *time.Time     `gorm:"column:verified_at" json:"verified_at"`
	ManagerSignature  string         `gorm:"column:manager_signature;type:text" json:"manager_signature"`
	ManagerSignedAt   *time.Time     `gorm:"column:manager_signed_at" json:"manager_signed_at"`
	DokumenVerifikasi string         `gorm:"column:dokumen_verifikasi;type:varchar(255)" json:"dokumen_verifikasi"`
	Catatan           string         `gorm:"column:catatan;type:text" json:"catatan"`
	CreatedAt         time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`

	// =====================================================
	// RELATION
	// =====================================================

	Peralatan       *Peralatan        `gorm:"foreignKey:IDPeralatan;references:ID" json:"peralatan,omitempty"`
	VerifiedByUser  *User             `gorm:"foreignKey:VerifiedBy;references:UserID" json:"verified_by_user,omitempty"`
	PICUser         *User             `gorm:"foreignKey:PICID;references:UserID" json:"pic_user,omitempty"`
	HasilVerifikasi []HasilVerifikasi `gorm:"foreignKey:IDVerifikasi;references:IDVerifikasi" json:"hasil_verifikasi,omitempty"`
}

func (Verifikasi) TableName() string {
	return "verifikasi"
}
