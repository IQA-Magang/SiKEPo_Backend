package models

import (
	"time"

	"gorm.io/gorm"
)

// Status peminjaman. Nama statusnya sama dengan yang dipakai frontend.
//
// Alur sampai disetujui:
//
//	MENUNGGU_MANAGER_PEMINJAM, MENUNGGU_PENGELOLA, MENUNGGU_MANAGER_LAB, DISETUJUI
//
// Tahap Manajer Peminjam dilewati bila orangnya sama dengan Manager Lab pemilik.
// Penolakan di tahap mana pun mengakhiri pengajuan dengan status DITOLAK.
const (
	StatusPeminjamanMenungguManajerPeminjam = "MENUNGGU_MANAGER_PEMINJAM"
	StatusPeminjamanMenungguPengelola       = "MENUNGGU_PENGELOLA"
	StatusPeminjamanMenungguManagerLab      = "MENUNGGU_MANAGER_LAB"
	StatusPeminjamanDisetujui               = "DISETUJUI"
	StatusPeminjamanDitolak                 = "DITOLAK"
)

// StatusPeminjamanAktif adalah status yang masih memegang jadwal alat.
// Pengajuan baru tidak boleh bertumpuk dengan peminjaman pada status ini.
var StatusPeminjamanAktif = []string{
	StatusPeminjamanMenungguManajerPeminjam,
	StatusPeminjamanMenungguPengelola,
	StatusPeminjamanMenungguManagerLab,
	StatusPeminjamanDisetujui,
}

// Status keputusan pada setiap tahap persetujuan.
const (
	TahapPending  = "pending"
	TahapApproved = "approved"
	TahapRejected = "rejected"
	TahapSkipped  = "skipped"
)

type Peminjaman struct {
	ID   uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Kode string `gorm:"column:kode;size:30;index" json:"kode"`

	PeralatanID uint64 `gorm:"column:peralatan_id;not null;index" json:"peralatan_id"`
	PeminjamID  uint64 `gorm:"column:peminjam_id;not null;index" json:"peminjam_id"`

	// Data pengajuan
	NamaOperator          string    `gorm:"column:nama_operator;size:100" json:"nama_operator"`
	TujuanPenggunaan      string    `gorm:"column:tujuan_penggunaan;type:text;not null" json:"tujuan_penggunaan"`
	NomorSPK              string    `gorm:"column:nomor_spk;size:100" json:"nomor_spk"`
	LokasiPenggunaan      string    `gorm:"column:lokasi_penggunaan;size:255;not null" json:"lokasi_penggunaan"`
	IsEksternal           bool      `gorm:"column:is_eksternal;not null;default:false" json:"is_eksternal"`
	RencanaTanggalKeluar  time.Time `gorm:"column:rencana_tanggal_keluar;type:date;not null;index" json:"rencana_tanggal_keluar"`
	RencanaTanggalKembali time.Time `gorm:"column:rencana_tanggal_kembali;type:date;not null;index" json:"rencana_tanggal_kembali"`
	KebutuhanKelengkapan  string    `gorm:"column:kebutuhan_kelengkapan;type:text" json:"kebutuhan_kelengkapan"`
	KebutuhanAksesori     string    `gorm:"column:kebutuhan_aksesori;type:text" json:"kebutuhan_aksesori"`
	Catatan               string    `gorm:"column:catatan;type:text" json:"catatan"`

	Status string `gorm:"column:status;size:30;not null;index" json:"status"`

	// Lab saat pengajuan dibuat. Disimpan agar riwayat tidak berubah bila user
	// atau alat berpindah lab.
	LabPeminjamID uint64 `gorm:"column:lab_peminjam_id;not null" json:"lab_peminjam_id"`
	LabPemilikID  uint64 `gorm:"column:lab_pemilik_id;not null;index" json:"lab_pemilik_id"`

	// Tahap 1: Manajer Peminjam, yaitu manager lab asal peminjam.
	ManajerPeminjamID      uint64     `gorm:"column:manajer_peminjam_id;not null;index" json:"manajer_peminjam_id"`
	ManajerPeminjamStatus  string     `gorm:"column:manajer_peminjam_status;size:10;not null;default:'pending'" json:"manajer_peminjam_status"`
	ManajerPeminjamCatatan string     `gorm:"column:manajer_peminjam_catatan;type:text" json:"manajer_peminjam_catatan"`
	ManajerPeminjamAt      *time.Time `gorm:"column:manajer_peminjam_at" json:"manajer_peminjam_at"`

	// Tahap 2: Pengelola Peralatan.
	PengelolaID      uint64     `gorm:"column:pengelola_id;not null;index" json:"pengelola_id"`
	PengelolaStatus  string     `gorm:"column:pengelola_status;size:10;not null;default:'pending'" json:"pengelola_status"`
	PengelolaCatatan string     `gorm:"column:pengelola_catatan;type:text" json:"pengelola_catatan"`
	PengelolaAt      *time.Time `gorm:"column:pengelola_at" json:"pengelola_at"`
	AlatAlternatifID *uint64    `gorm:"column:alat_alternatif_id" json:"alat_alternatif_id"`

	// Tahap 3: Manager Lab pemilik peralatan.
	ManagerLabID      uint64     `gorm:"column:manager_lab_id;not null;index" json:"manager_lab_id"`
	ManagerLabStatus  string     `gorm:"column:manager_lab_status;size:10;not null;default:'pending'" json:"manager_lab_status"`
	ManagerLabCatatan string     `gorm:"column:manager_lab_catatan;type:text" json:"manager_lab_catatan"`
	ManagerLabAt      *time.Time `gorm:"column:manager_lab_at" json:"manager_lab_at"`

	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`

	// Relasi
	Peralatan       *Peralatan `gorm:"foreignKey:PeralatanID;references:ID" json:"-"`
	Peminjam        *User      `gorm:"foreignKey:PeminjamID;references:UserID" json:"-"`
	LabPeminjam     *Labs      `gorm:"foreignKey:LabPeminjamID;references:ID" json:"-"`
	LabPemilik      *Labs      `gorm:"foreignKey:LabPemilikID;references:ID" json:"-"`
	ManajerPeminjam *User      `gorm:"foreignKey:ManajerPeminjamID;references:UserID" json:"-"`
	Pengelola       *User      `gorm:"foreignKey:PengelolaID;references:UserID" json:"-"`
	ManagerLab      *User      `gorm:"foreignKey:ManagerLabID;references:UserID" json:"-"`
	AlatAlternatif  *Peralatan `gorm:"foreignKey:AlatAlternatifID;references:ID" json:"-"`
}

func (Peminjaman) TableName() string {
	return "peminjaman"
}
