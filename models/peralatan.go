package models

import "time"

// =========================================================
// MASTER PERALATAN
// =========================================================

type Peralatan struct {
	ID uint `gorm:"primaryKey" json:"id"`

	NomorAset string `gorm:"unique;not null;type:varchar(50)" json:"nomor_aset"`

	NamaPeralatan string `gorm:"not null;type:varchar(150)" json:"nama_peralatan"`

	KategoriID uint `gorm:"not null" json:"kategori_id"`

	KelompokAsetID uint `gorm:"not null" json:"kelompok_aset_id"`

	RuanganID uint `gorm:"not null" json:"ruangan_id"`

	PICID uint `gorm:"not null" json:"pic_id"`

	Merek string `gorm:"type:varchar(100)" json:"merek"`

	TipeModel string `gorm:"type:varchar(100)" json:"tipe_model"`

	NomorSeri string `gorm:"type:varchar(100)" json:"nomor_seri"`

	Foto string `gorm:"type:varchar(255)" json:"foto"`

	// Status fisik / penggunaan alat
	StatusAlat string `gorm:"type:enum('Karantina','Aktif','Dipinjam','Dalam Kalibrasi','Rusak','Dihapuskan');default:'Karantina'" json:"status_alat"`

	// Status proses verifikasi
	StatusVerifikasi string `gorm:"type:enum('Belum Diverifikasi','Draft','Diajukan','Disetujui','Ditolak');default:'Belum Diverifikasi'" json:"status_verifikasi"`

	Keterangan string `gorm:"type:text" json:"keterangan"`

	KategoriPeralatanID uint `gorm:"not null" json:"kategori_peralatan_id"`

	KategoriPeralatan *KategoriPeralatan `gorm:"foreignKey:KategoriPeralatanID;references:ID" json:"kategori_peralatan,omitempty"`

	CreatedAt time.Time `json:"created_at"`

	UpdatedAt time.Time `json:"updated_at"`

	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}

func (Peralatan) TableName() string {
	return "peralatan"
}

// =========================================================
// DETAIL ALAT UKUR
// =========================================================

type DetailAlatUkur struct {
	PeralatanID uint `gorm:"primaryKey" json:"peralatan_id"`

	PerantiLunakVersi string `gorm:"type:varchar(100)" json:"peranti_lunak_versi"`

	MetodeKelayakan string `gorm:"type:enum('kalibrasi internal','kalibrasi eksternal','verifikasi fungsi (metode tertentu)','verifikasi fungsi (uji banding)')" json:"metode_kelayakan"`

	NoSertifikat string `gorm:"type:varchar(100)" json:"no_sertifikat"`

	TglKalibrasi *time.Time `json:"tgl_kalibrasi"`

	TglJatuhTempo *time.Time `json:"tgl_jatuh_tempo"`

	IntervalBulan int `json:"interval_bulan"`

	FungsiSbgAlatStandar bool `gorm:"default:false" json:"fungsi_sbg_alat_standar"`

	JenisLabel string `gorm:"type:enum('calibration','limited calibration','do not use')" json:"jenis_label"`

	StatusKelayakan string `gorm:"type:enum('Layak','Terbatas','Tidak layak');default:'Layak'" json:"status_kelayakan"`
}

func (DetailAlatUkur) TableName() string {
	return "detail_alat_ukur"
}

// =========================================================
// DETAIL ALAT BANTU
// =========================================================

type DetailAlatBantu struct {
	PeralatanID uint `gorm:"primaryKey" json:"peralatan_id"`

	FungsiKegunaan string `gorm:"type:varchar(255)" json:"fungsi_kegunaan"`

	PerantiLunakVersi string `gorm:"type:varchar(100)" json:"peranti_lunak_versi"`

	JenisPemeriksaanBerkala string `gorm:"type:enum('kalibrasi','verifikasi fungsi','pemeriksaan lain')" json:"jenis_pemeriksaan_berkala"`

	KriteriaPemeriksaan string `gorm:"type:text" json:"kriteria_pemeriksaan"`

	TglPemeriksaanTerakhir *time.Time `json:"tgl_pemeriksaan_terakhir"`

	TglJatuhTempo *time.Time `json:"tgl_jatuh_tempo"`

	IntervalBulan int `json:"interval_bulan"`

	FungsiSbgAlatStandar bool `gorm:"default:false" json:"fungsi_sbg_alat_standar"`

	KarakteristikAcuan string `gorm:"type:varchar(255)" json:"karakteristik_acuan"`

	JadwalKarakterisasiUlang string `gorm:"type:varchar(100)" json:"jadwal_karakterisasi_ulang"`
}

func (DetailAlatBantu) TableName() string {
	return "detail_alat_bantu"
}

// =========================================================
// DETAIL ARTEFAK ACUAN
// =========================================================

type DetailArtefakAcuan struct {
	PeralatanID uint `gorm:"primaryKey" json:"peralatan_id"`

	JenisDeskripsi string `gorm:"type:varchar(255)" json:"jenis_deskripsi"`

	KarakteristikYangDiacu string `gorm:"type:enum('visual','dimension','kinerja funngsional','visual & dimension','dimension & kinerja')" json:"karakteristik_yang_diacu"`

	NilaiSpesifikasiKarakterisasi string `gorm:"type:varchar(255)" json:"nilai_spesifikasi_karakterisasi"`

	MetodeKarakterisasi string `gorm:"type:varchar(150)" json:"metode_karakterisasi"`

	NoLaporanKarakterisasi string `gorm:"type:varchar(100)" json:"no_laporan_karakterisasi"`

	TglKarakterisasiTerakhir *time.Time `json:"tgl_karakterisasi_terakhir"`

	TglKarakterisasiUlang *time.Time `json:"tgl_karakterisasi"`

	IntervalBulan int `json:"interval_bulan"`

	KondisiPenyimpanan string `gorm:"type:varchar(150)" json:"kondisi_penyimpanan"`

	Status string `gorm:"type:enum('aktif','karantina','dihapuskan')" json:"status"`
}

func (DetailArtefakAcuan) TableName() string {
	return "detail_artefak_acuan"
}

// =========================================================
// DETAIL KOMPONEN PENDUKUNG
// =========================================================

type DetailKomponenPendukung struct {
	PeralatanID uint `gorm:"primaryKey" json:"peralatan_id"`

	Kategori string `gorm:"type:enum('data acuan','pereaksi','bahan habis pakai')" json:"sub_kategori"`

	DeskripsiSpesifikasi string `gorm:"type:text" json:"deskripsi_spesifikasi"`

	SumberPemasok string `gorm:"type:varchar(150)" json:"sumber_pemasok"`

	NoLotBatchEdisi string `gorm:"type:varchar(100)" json:"no_lot_batch_edisi"`

	SatuanKemasan string `gorm:"type:varchar(100)" json:"satuan_kemasan"`

	TglTerimaTerbit *time.Time `json:"tgl_terima_terbit"`

	TglKedaluwarsa *time.Time `json:"tgl_kedaluwarsa"`

	KondisiPenyimpanan string `gorm:"type:varchar(150)" json:"kondisi_penyimpanan"`

	StatusKetersediaan string `gorm:"type:enum('Berlaku','Tersedia','Stok cukup','Stok menipis','Kedaluwarsa','Habis');default:'Tersedia'" json:"status_ketersediaan"`
}

func (DetailKomponenPendukung) TableName() string {
	return "detail_komponen_pendukung"
}

// =========================================================
// CREATE PERALATAN REQUEST
// =========================================================

type CreatePeralatanRequest struct {
	NamaPeralatan string `json:"nama_peralatan" validate:"required"`

	KategoriPeralatanID uint `json:"kategori_id" validate:"required"`

	KelompokAsetID uint `json:"kelompok_aset_id" validate:"required"`

	RuanganID uint `json:"ruangan_id" validate:"required"`

	PICID uint `json:"pic_id" validate:"required"`

	Merek string `json:"merek"`

	TipeModel string `json:"tipe_model"`

	NomorSeri string `json:"nomor_seri"`

	Foto string `json:"foto"`

	StatusAlat string `json:"status_alat"`

	Keterangan string `json:"keterangan"`

	Detail map[string]interface{} `json:"detail"`
}