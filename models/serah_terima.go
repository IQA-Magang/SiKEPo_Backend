package models

import "time"

// SerahTerima adalah lembar Lampiran A untuk serah terima keluar. Satu
// peminjaman punya paling banyak satu lembar.
//
// Pengelola mengisi dan menandatangani lebih dulu. Peminjam lalu
// mengonfirmasi terima dan menandatangani di lembar yang sama.
type SerahTerima struct {
	ID           uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	PeminjamanID uint64 `gorm:"column:peminjaman_id;not null;uniqueIndex" json:"peminjaman_id"`

	// Diisi Pengelola
	DiperiksaOleh uint64    `gorm:"column:diperiksa_oleh;not null" json:"diperiksa_oleh"`
	DiperiksaAt   time.Time `gorm:"column:diperiksa_at;not null" json:"diperiksa_at"`
	PengelolaTTD  string    `gorm:"column:pengelola_ttd;type:mediumtext" json:"-"`
	AdaTS         bool      `gorm:"column:ada_ts;not null;default:false" json:"ada_ts"`
	Catatan       string    `gorm:"column:catatan;type:text" json:"catatan"`

	// Diisi peminjam saat konfirmasi terima
	DiterimaOleh *uint64    `gorm:"column:diterima_oleh" json:"diterima_oleh"`
	DiterimaAt   *time.Time `gorm:"column:diterima_at" json:"diterima_at"`
	PeminjamTTD  string     `gorm:"column:peminjam_ttd;type:mediumtext" json:"-"`

	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`

	Items []SerahTerimaItem `gorm:"foreignKey:SerahTerimaID;references:ID" json:"-"`
}

func (SerahTerima) TableName() string {
	return "serah_terima"
}

// SerahTerimaItem adalah satu butir pemeriksaan pada lembar Lampiran A.
type SerahTerimaItem struct {
	ID            uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	SerahTerimaID uint64 `gorm:"column:serah_terima_id;not null;index" json:"serah_terima_id"`
	NomorButir    int    `gorm:"column:nomor_butir;not null" json:"nomor_butir"`
	Butir         string `gorm:"column:butir;size:255;not null" json:"butir"`
	Hasil         string `gorm:"column:hasil;size:2;not null" json:"hasil"`
	Keterangan    string `gorm:"column:keterangan;type:text" json:"keterangan"`
}

func (SerahTerimaItem) TableName() string {
	return "serah_terima_item"
}

// Hasil pemeriksaan: S sesuai, TS tidak sesuai, TB tidak berlaku.
const (
	HasilSesuai       = "S"
	HasilTidakSesuai  = "TS"
	HasilTidakBerlaku = "TB"
)

// ButirLampiranA mendefinisikan satu butir Lampiran A. BolehTB bernilai true
// untuk butir yang di IK bertanda "bila ada" atau "bila relevan".
type ButirLampiranA struct {
	Nomor   int
	Butir   string
	BolehTB bool
}

// DaftarButirLampiranA adalah 10 butir Lampiran A TLKM13/IK/005.
var DaftarButirLampiranA = []ButirLampiranA{
	{1, "Identitas peralatan sesuai TLKM13/F/001", false},
	{2, "Label status terpasang, terbaca, dan berlaku sampai rencana tanggal kembali", false},
	{3, "Segel atau penguncian pengaturan utuh (bila ada)", true},
	{4, "Kondisi fisik casing, layar, tombol, dan konektor/port baik", false},
	{5, "Kelengkapan aksesori, kabel, adaptor, catu daya/baterai sesuai daftar kelengkapan", false},
	{6, "Fungsi dasar: menyala normal, swauji/inisialisasi berhasil, tidak ada pesan galat", false},
	{7, "Versi peranti lunak/firmware sesuai yang tercatat (bila relevan)", true},
	{8, "Konektor/antarmuka optik bersih dan tertutup pelindung (bila relevan)", true},
	{9, "Wadah atau kemasan pelindung dalam kondisi baik", false},
	{10, "Dokumen pendukung: manual pengoperasian dan salinan sertifikat atau laporan verifikasi", false},
}
