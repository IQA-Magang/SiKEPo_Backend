package models

import (
	"time"

	"gorm.io/gorm"
)

type HasilVerifikasi struct {
	IDHasil      uint64         `gorm:"column:id_hasil;primaryKey;autoIncrement" json:"id_hasil"`
	IDVerifikasi uint64         `gorm:"column:id_verifikasi;not null;index" json:"id_verifikasi"`
	Identitas    string         `gorm:"column:identitas;type:enum('S','TS','TB');not null" json:"identitas"`
	Kelengkapan  string         `gorm:"column:kelengkapan;type:enum('S','TS','TB');not null" json:"kelengkapan"`
	Firmware     string         `gorm:"column:firmware;type:enum('S','TS','TB');not null" json:"firmware"`
	KondisiFisik string         `gorm:"column:kondisi_fisik;type:enum('S','TS','TB');not null" json:"kondisi_fisik"`
	Segel        string         `gorm:"column:segel;type:enum('S','TS','TB');not null" json:"segel"`
	FungsiAwal   string         `gorm:"column:fungsi_awal;type:enum('S','TS','TB');not null" json:"fungsi_awal"`
	Metrologi    string         `gorm:"column:metrologi;type:enum('S','TS','TB');not null" json:"metrologi"`
	Sertifikat   string         `gorm:"column:sertifikat;type:enum('S','TS','TB');not null" json:"sertifikat"`
	Catatan      string         `gorm:"column:catatan;type:text" json:"catatan"`
	CreatedAt    time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at" json:"-"`
	Verifikasi   *Verifikasi    `gorm:"foreignKey:IDVerifikasi;references:IDVerifikasi" json:"verifikasi,omitempty"`
}

func (HasilVerifikasi) TableName() string {
	return "hasil_verifikasi"
}
