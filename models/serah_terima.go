package models

import "time"

// SerahTerima adalah lembar TLKM13/F/010 untuk serah terima keluar. Satu
// peminjaman punya paling banyak satu lembar.
//
// Pengelola mengisi dan menandatangani lebih dulu. Peminjam lalu
// mengonfirmasi terima dan menandatangani di lembar yang sama. Tanda tangan
// keduanya wajib berupa PNG.
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

	// Diisi Pengelola saat pengembalian (kolom Kembali pada F/010). Peminjam
	// yang menyerahkan dan Pengelola yang menerima sama-sama menandatangani.
	KembaliDiterimaOleh *uint64    `gorm:"column:kembali_diterima_oleh" json:"kembali_diterima_oleh"`
	KembaliAt           *time.Time `gorm:"column:kembali_at" json:"kembali_at"`
	KembaliPengelolaTTD string     `gorm:"column:kembali_pengelola_ttd;type:mediumtext" json:"-"`
	KembaliPeminjamTTD  string     `gorm:"column:kembali_peminjam_ttd;type:mediumtext" json:"-"`
	AdaKejadian         bool       `gorm:"column:ada_kejadian;not null;default:false" json:"ada_kejadian"`
	UraianKejadian      string     `gorm:"column:uraian_kejadian;type:text" json:"uraian_kejadian"`
	KembaliAdaTS        bool       `gorm:"column:kembali_ada_ts;not null;default:false" json:"kembali_ada_ts"`
	KembaliCatatan      string     `gorm:"column:kembali_catatan;type:text" json:"kembali_catatan"`
	TindakLanjut        string     `gorm:"column:tindak_lanjut;size:20" json:"tindak_lanjut"`
	AlasanTindakLanjut  string     `gorm:"column:alasan_tindak_lanjut;type:text" json:"alasan_tindak_lanjut"`

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
	Hasil         string `gorm:"column:hasil;size:3;not null" json:"hasil"`
	Keterangan    string `gorm:"column:keterangan;type:text" json:"keterangan"`

	// Hasil dan keterangan saat kembali. Kosong sampai alat dikembalikan.
	HasilKembali      string `gorm:"column:hasil_kembali;size:3" json:"hasil_kembali"`
	KeteranganKembali string `gorm:"column:keterangan_kembali;type:text" json:"keterangan_kembali"`
}

// Tindak lanjut setelah alat kembali. Nilainya mengikuti istilah flowchart.
const (
	// TindakLanjutDoNotUse: ada kejadian, butir TS, atau kembali setelah jatuh
	// tempo pemeriksaan. Alat dihentikan dan ditangani sesuai IK/012.
	TindakLanjutDoNotUse = "DO_NOT_USE"

	// TindakLanjutPerluVerifikasi: alat wajib diverifikasi ulang (IK/003)
	// sebelum dipakai lagi.
	TindakLanjutPerluVerifikasi = "PERLU_VERIFIKASI"
)

func (SerahTerimaItem) TableName() string {
	return "serah_terima_item"
}

// Hasil pemeriksaan: S sesuai, TS tidak sesuai, N/A tidak relevan.
const (
	HasilSesuai      = "S"
	HasilTidakSesuai = "TS"
	HasilNA          = "N/A"
)

// ButirLembarSerahTerima mendefinisikan satu butir pemeriksaan pada lembar
// TLKM13/F/010. BolehNA bernilai true untuk butir yang boleh dikosongkan,
// yaitu butir yang di form bertanda "bila perlu".
type ButirLembarSerahTerima struct {
	Nomor   int
	Butir   string
	BolehNA bool

	// KeteranganWajibJikaS bernilai true untuk butir fungsi khusus. Bila
	// dinyatakan sesuai, jenis pemeriksaan yang dilakukan wajib ditulis.
	KeteranganWajibJikaS bool
}

// DaftarButirSerahTerima adalah 7 butir bagian B TLKM13/F/010. Butir 4, 6, dan
// 7 memuat isian bebas (daftar kelengkapan dan jenis fungsi khusus) yang
// ditulis pada keterangan.
var DaftarButirSerahTerima = []ButirLembarSerahTerima{
	{1, "Identitas peralatan (nomor aset, nomor seri) sesuai dengan TLKM13/F/001", false, false},
	{2, "Label status terpasang, terbaca, dan masih berlaku", false, false},
	{3, "Kondisi fisik baik (casing, layar, tombol, konektor/port)", false, false},
	{4, "Kelengkapan sesuai daftar", false, false},
	{5, "Fungsi dasar normal (menyala, swauji/inisialisasi berhasil)", false, false},
	{6, "Pemeriksaan fungsi khusus (bila perlu)", true, true},
	{7, "Pemeriksaan fungsi khusus (bila perlu)", true, true},
}
