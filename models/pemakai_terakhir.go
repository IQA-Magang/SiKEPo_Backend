package models

import "time"

// PemakaiTerakhir adalah orang yang terakhir benar-benar menerima peralatan,
// yaitu peminjam dari serah terima keluar terbaru. Pengajuan yang ditolak,
// dibatalkan, atau belum diserahkan tidak dihitung.
//
// Bila alat sedang dipinjam, pemakai terakhir adalah orang yang sedang
// menggunakannya, dan SedangDigunakan bernilai true.
type PemakaiTerakhir struct {
	UserID          uint64    `json:"user_id"`
	Nama            string    `json:"nama"`
	TanggalKeluar   time.Time `json:"tanggal_keluar"`
	PeminjamanID    uint64    `json:"peminjaman_id"`
	Kode            string    `json:"kode"`
	Status          string    `json:"status"`
	SedangDigunakan bool      `json:"sedang_digunakan"`
}
