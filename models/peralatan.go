package models

import (
	"time"

	"gorm.io/gorm"
)

type Peralatan struct {
	ID                uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	RuanganID         uint64         `gorm:"column:ruangan_id;not null;index" json:"ruangan_id"`
	PICID             *uint64        `gorm:"column:pic_id;index" json:"pic_id"`
	NomorAset         string         `gorm:"column:nomor_aset;size:100;not null;index" json:"nomor_aset"`
	NamaPeralatan     string         `gorm:"column:nama_peralatan;size:150;not null" json:"nama_peralatan"`
	Merk              string         `gorm:"column:merk;size:100" json:"merk"`
	Model             string         `gorm:"column:model;size:100" json:"model"`
	NomorSeri         string         `gorm:"column:nomor_seri;size:100" json:"nomor_seri"`
	Jumlah            uint           `gorm:"column:jumlah;not null;default:1" json:"jumlah"`
	KategoriPeralatan string         `gorm:"column:kategori_peralatan;not null" json:"kategori_peralatan"`
	Kondisi           string         `gorm:"column:kondisi;not null;default:'sesuai'" json:"kondisi"`
	StatusKelayakan   string         `gorm:"column:status_kelayakan;not null;default:'pending'" json:"status_kelayakan"`
	Metode            string         `gorm:"column:metode" json:"metode"`
	JenisPakai        string         `gorm:"column:jenis_pakai" json:"jenis_pakai"`
	VerifiedAt        *time.Time     `gorm:"column:verified_at" json:"verified_at"`
	VerificationNote  string         `gorm:"column:verification_note;type:text" json:"verification_note"`
	InputBy           uint64         `gorm:"column:input_by;not null" json:"input_by"`
	VerifiedBy        *uint64        `gorm:"column:verified_by" json:"verified_by"`
	CreatedAt         time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"column:deleted_at" json:"-"`

	Ruangan        *Ruangan `gorm:"foreignKey:RuanganID;references:ID" json:"ruangan,omitempty"`
	PIC            *User    `gorm:"foreignKey:PICID;references:UserID" json:"pic,omitempty"`
	InputByUser    *User    `gorm:"foreignKey:InputBy;references:UserID" json:"input_by_user,omitempty"`
	VerifiedByUser *User    `gorm:"foreignKey:VerifiedBy;references:UserID" json:"verified_by_user,omitempty"`
}

func (Peralatan) TableName() string {
	return "peralatan"
}
