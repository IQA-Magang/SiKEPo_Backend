package models

import "time"

type HasilVerifikasi struct {
	IDHasil      uint64      `gorm:"column:id_hasil;primaryKey;autoIncrement" json:"id_hasil"`
	IDVerifikasi uint64      `gorm:"column:id_verifikasi;not null;index" json:"id_verifikasi"`
	Identitas    string      `gorm:"column:identitas;type:varchar(10);not null" json:"identitas"`
	Kelengkapan  string      `gorm:"column:kelengkapan;type:varchar(10);not null" json:"kelengkapan"`
	Firmware     string      `gorm:"column:firmware;type:varchar(10);not null" json:"firmware"`
	KondisiFisik string      `gorm:"column:kondisi_fisik;type:varchar(10);not null" json:"kondisi_fisik"`
	Segel        string      `gorm:"column:segel;type:varchar(10);not null" json:"segel"`
	FungsiAwal   string      `gorm:"column:fungsi_awal;type:varchar(10);not null" json:"fungsi_awal"`
	Metrologi    string      `gorm:"column:metrologi;type:varchar(10);not null" json:"metrologi"`
	Sertifikat   string      `gorm:"column:sertifikat;type:varchar(10);not null" json:"sertifikat"`
	Catatan      string      `gorm:"column:catatan;type:text" json:"catatan"`
	CreatedAt    time.Time   `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time   `gorm:"column:updated_at" json:"updated_at"`
	Verifikasi   *Verifikasi `gorm:"foreignKey:IDVerifikasi;references:IDVerifikasi" json:"verifikasi,omitempty"`
}

func (HasilVerifikasi) TableName() string {
	return "hasil_verifikasi"
}
