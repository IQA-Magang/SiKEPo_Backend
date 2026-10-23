
package models

import "time"

type JadwalPemeriksaanBerkala struct {
	ID                  uint       `gorm:"column:id;primaryKey" json:"id"`
	PeralatanID         uint       `gorm:"column:peralatan_id;not null" json:"peralatan_id"`
	JenisJadwal         string     `gorm:"column:jenis_jadwal;not null" json:"jenis_jadwal"`
	PeranJadwal         *string    `gorm:"column:peran_jadwal" json:"peran_jadwal"`

	MetodeKalibrasi     *string    `gorm:"column:metode_kalibrasi" json:"metode_kalibrasi"`
	NomorSertifikat     *string    `gorm:"column:nomor_sertifikat" json:"nomor_sertifikat"`
	AlatStandarBendaUji *string    `gorm:"column:alat_standar_benda_uji" json:"alat_standar_benda_uji"`
	FrekuensiPenggunaan *string    `gorm:"column:frekuensi_penggunaan" json:"frekuensi_penggunaan"`
	JenisPemeliharaan   *string    `gorm:"column:jenis_pemeliharaan" json:"jenis_pemeliharaan"`
	MetodeVerifikasi    *string    `gorm:"column:metode_verifikasi" json:"metode_verifikasi"`
	NomorLaporan        *string    `gorm:"column:nomor_laporan" json:"nomor_laporan"`

	IntervalBulan       int        `gorm:"column:interval_bulan;not null" json:"interval_bulan"`
	TanggalTerakhir     *time.Time `gorm:"column:tanggal_terakhir;type:date" json:"tanggal_terakhir"`
	TanggalJatuhTempo   *time.Time `gorm:"column:tanggal_jatuh_tempo;type:date" json:"tanggal_jatuh_tempo"`
	TanggalRealisasi    *time.Time `gorm:"column:tanggal_realisasi;type:date" json:"tanggal_realisasi"`
	PIC                 *string    `gorm:"column:pic" json:"pic"`
	Keterangan          *string    `gorm:"column:keterangan;type:text" json:"keterangan"`

	CreatedAt           time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt           time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

func (JadwalPemeriksaanBerkala) TableName() string {
	return "jadwal_pemeriksaan_berkala"
}
