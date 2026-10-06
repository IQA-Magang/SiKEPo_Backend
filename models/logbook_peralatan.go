package models

import (
	"strings"
	"time"
)

// LogbookPeralatan adalah padanan digital TLKM13/F/010 Logbook Peralatan.
//
// Tabel ini bersifat append-only: baris hanya ditambah oleh modul lain di dalam
// transaksinya sendiri, dan tidak pernah diubah atau dihapus lewat API.
type LogbookPeralatan struct {
	ID          uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	PeralatanID uint64    `gorm:"column:peralatan_id;not null;index" json:"peralatan_id"`
	Jenis       string    `gorm:"column:jenis;size:30;not null;index" json:"jenis"`
	ReferensiID *uint64   `gorm:"column:referensi_id;index" json:"referensi_id"`
	Aksi        string    `gorm:"column:aksi;size:50;not null" json:"aksi"`
	Status      string    `gorm:"column:status;size:30" json:"status"`
	UserID      *uint64   `gorm:"column:user_id;index" json:"user_id"`
	Keterangan  string    `gorm:"column:keterangan;type:text" json:"keterangan"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`

	// Relasi
	User *User `gorm:"foreignKey:UserID;references:UserID" json:"-"`
}

func (LogbookPeralatan) TableName() string {
	return "logbook_peralatan"
}

// Jenis log, sama dengan pilihan pada dropdown Log Aktivitas Peralatan di frontend.
const (
	LogbookJenisVerifikasi  = "verifikasi"
	LogbookJenisPeminjaman  = "peminjaman"
	LogbookJenisPerpindahan = "perpindahan"
	LogbookJenisPenggunaan  = "penggunaan"
	LogbookJenisPemeriksaan = "pemeriksaan"
)

// IsValidLogbookJenis memeriksa apakah jenis log dikenal.
func IsValidLogbookJenis(jenis string) bool {
	switch jenis {
	case LogbookJenisVerifikasi,
		LogbookJenisPeminjaman,
		LogbookJenisPerpindahan,
		LogbookJenisPenggunaan,
		LogbookJenisPemeriksaan:
		return true
	}
	return false
}

// Kode aksi untuk jenis peminjaman.
const (
	LogbookAksiPengajuan               = "pengajuan"
	LogbookAksiDivalidasiManajerPinjam = "divalidasi_manajer_peminjam"
	LogbookAksiManajerPinjamDilewati   = "manajer_peminjam_dilewati"
	LogbookAksiDitolakManajerPinjam    = "ditolak_manajer_peminjam"
	LogbookAksiDiteruskanPengelola     = "diteruskan_pengelola"
	LogbookAksiDitolakPengelola        = "ditolak_pengelola"
	LogbookAksiDisetujuiManagerLab     = "disetujui_manager_lab"
	LogbookAksiDitolakManagerLab       = "ditolak_manager_lab"
	LogbookAksiChecklistKeluar         = "checklist_keluar"
	LogbookAksiDibatalkanChecklist     = "dibatalkan_checklist"
	LogbookAksiSerahTerimaKeluar       = "serah_terima_keluar"
)

var logbookJudul = map[string]string{
	LogbookAksiPengajuan:               "Pengajuan peminjaman",
	LogbookAksiDivalidasiManajerPinjam: "Divalidasi Manajer Peminjam",
	LogbookAksiManajerPinjamDilewati:   "Tahap Manajer Peminjam dilewati",
	LogbookAksiDitolakManajerPinjam:    "Ditolak Manajer Peminjam",
	LogbookAksiDiteruskanPengelola:     "Diteruskan Pengelola Peralatan",
	LogbookAksiDitolakPengelola:        "Ditolak Pengelola Peralatan",
	LogbookAksiDisetujuiManagerLab:     "Disetujui Manager Lab",
	LogbookAksiDitolakManagerLab:       "Ditolak Manager Lab",
	LogbookAksiChecklistKeluar:         "Pemeriksaan serah terima keluar",
	LogbookAksiDibatalkanChecklist:     "Peminjaman dibatalkan",
	LogbookAksiSerahTerimaKeluar:       "Serah terima keluar",
}

// LogbookJudul mengubah kode aksi menjadi judul yang ditampilkan. Judul tidak
// disimpan di database, sehingga redaksinya bisa diubah tanpa menyentuh data lama.
// Kode yang belum terdaftar tetap tampil, dengan garis bawah diganti spasi.
func LogbookJudul(aksi string) string {
	if judul, ok := logbookJudul[aksi]; ok {
		return judul
	}

	text := strings.ReplaceAll(aksi, "_", " ")
	if text == "" {
		return text
	}

	return strings.ToUpper(text[:1]) + text[1:]
}
