package models

import "time"

// Status pengajuan perpanjangan.
const (
	// PerpanjanganMenunggu: menunggu keputusan Manager Lab.
	PerpanjanganMenunggu = "menunggu"

	PerpanjanganDisetujui = "disetujui"
	PerpanjanganDitolak   = "ditolak"

	// PerpanjanganDibatalkan: alat sudah dikembalikan sebelum Manager Lab memutuskan.
	PerpanjanganDibatalkan = "dibatalkan"
)

// Perpanjangan adalah satu pengajuan perpanjangan tanggal kembali pada
// peminjaman yang sedang berjalan. Satu peminjaman boleh punya beberapa
// pengajuan, tetapi hanya satu yang menunggu pada satu waktu.
//
// Hanya Manager Lab pemilik alat yang memutuskan. Pengelola cukup diberi tahu.
type Perpanjangan struct {
	ID           uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	PeminjamanID uint64 `gorm:"column:peminjaman_id;not null;index" json:"peminjaman_id"`

	DiajukanOleh uint64 `gorm:"column:diajukan_oleh;not null" json:"diajukan_oleh"`

	TglKembaliLama time.Time `gorm:"column:tgl_kembali_lama;type:date;not null" json:"tgl_kembali_lama"`
	TglKembaliBaru time.Time `gorm:"column:tgl_kembali_baru;type:date;not null" json:"tgl_kembali_baru"`
	Alasan         string    `gorm:"column:alasan;type:text;not null" json:"alasan"`

	Status string `gorm:"column:status;size:12;not null;index" json:"status"`

	DiputuskanOleh   *uint64    `gorm:"column:diputuskan_oleh" json:"diputuskan_oleh"`
	DiputuskanAt     *time.Time `gorm:"column:diputuskan_at" json:"diputuskan_at"`
	CatatanKeputusan string     `gorm:"column:catatan_keputusan;type:text" json:"catatan_keputusan"`
	TTD              string     `gorm:"column:ttd;type:mediumtext" json:"-"`

	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`

	// Relasi
	Pengaju *User `gorm:"foreignKey:DiajukanOleh;references:UserID" json:"-"`
	Pemutus *User `gorm:"foreignKey:DiputuskanOleh;references:UserID" json:"-"`
}

func (Perpanjangan) TableName() string {
	return "perpanjangan_peminjaman"
}
