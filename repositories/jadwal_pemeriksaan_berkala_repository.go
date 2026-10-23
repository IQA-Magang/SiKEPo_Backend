
package repositories

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"backend/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type JadwalPemeriksaanBerkalaRepository interface {
	CreatePair(
		peralatanID uint,
		jenisUtama string,
		frekuensi string,
		intervalUtama int,
		intervalTambahan int,
		tanggalTerakhir time.Time,
		metode string,
		nomorDokumen string,
		alatStandar string,
		jenisPemeliharaan string,
		pic string,
		keterangan string,
	) error

	FindAll() ([]models.JadwalPemeriksaanBerkala, error)
	FindByID(id uint) (*models.JadwalPemeriksaanBerkala, error)
	UpdateRealisasi(id uint, tanggal time.Time) error
}

type jadwalPemeriksaanBerkalaRepository struct {
	db *gorm.DB
}

func NewJadwalPemeriksaanBerkalaRepository(
	db *gorm.DB,
) JadwalPemeriksaanBerkalaRepository {
	return &jadwalPemeriksaanBerkalaRepository{db: db}
}

func (r *jadwalPemeriksaanBerkalaRepository) CreatePair(
	peralatanID uint,
	jenisUtama string,
	frekuensi string,
	intervalUtama int,
	intervalTambahan int,
	tanggalTerakhir time.Time,
	metode string,
	nomorDokumen string,
	alatStandar string,
	jenisPemeliharaan string,
	pic string,
	keterangan string,
) error {
	jenisUtama = strings.ToLower(strings.TrimSpace(jenisUtama))
	frekuensi = strings.ToLower(strings.TrimSpace(frekuensi))

	if jenisUtama != "kalibrasi" &&
		jenisUtama != "verifikasi_fungsi" {
		return errors.New(
			"pemeriksaan utama harus kalibrasi atau verifikasi_fungsi",
		)
	}

	if frekuensi != "tinggi" &&
		frekuensi != "sedang" &&
		frekuensi != "rendah" {
		return errors.New(
			"frekuensi harus tinggi, sedang, atau rendah",
		)
	}

	if intervalUtama <= 0 || intervalTambahan <= 0 {
		return errors.New(
			"interval pemeriksaan harus lebih dari 0 bulan",
		)
	}

	jenisTambahan := "pemeliharaan"
	if frekuensi == "tinggi" {
		jenisTambahan = "pengecekan_antara"
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		// Kunci baris peralatan selama transaksi untuk mengurangi
		// risiko dua request membuat pasangan jadwal bersamaan.
		var peralatan models.Peralatan

		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", peralatanID).
			First(&peralatan).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("peralatan tidak ditemukan")
			}
			return err
		}

		var jumlah int64
		if err := tx.Model(&models.JadwalPemeriksaanBerkala{}).
			Where("peralatan_id = ?", peralatanID).
			Count(&jumlah).Error; err != nil {
			return err
		}

		if jumlah > 0 {
			return errors.New(
				"peralatan sudah memiliki jadwal; gunakan fitur ubah jadwal",
			)
		}

		// Tabel jadwal tidak menyimpan nomor aset dan nama peralatan.
		// Identitas peralatan direferensikan melalui peralatan_id.

		freq := frekuensi
		tanggalUtama := tanggalTerakhir
		jatuhTempoUtama := tanggalTerakhir.AddDate(
			0, intervalUtama, 0,
		)

		tanggalTambahan := tanggalTerakhir
		jatuhTempoTambahan := tanggalTerakhir.AddDate(
			0, intervalTambahan, 0,
		)

		jadwalUtama := models.JadwalPemeriksaanBerkala{
			PeralatanID:       peralatanID,
			JenisJadwal:       jenisUtama,
			PeranJadwal:       stringPointer("utama"),
			IntervalBulan:     intervalUtama,
			TanggalTerakhir:   &tanggalUtama,
			TanggalJatuhTempo: &jatuhTempoUtama,
			FrekuensiPenggunaan: stringPointer(freq),
			PIC:               stringPointer(pic),
			Keterangan:        stringPointer(keterangan),
		}

		// Petakan data dokumen sesuai jenis pemeriksaan utama.
		if jenisUtama == "kalibrasi" {
			jadwalUtama.MetodeKalibrasi = stringPointer(metode)
			jadwalUtama.NomorSertifikat = stringPointer(nomorDokumen)
			jadwalUtama.AlatStandarBendaUji = stringPointer(alatStandar)
		} else {
			jadwalUtama.MetodeVerifikasi = stringPointer(metode)
			jadwalUtama.NomorLaporan = stringPointer(nomorDokumen)
		}

		jadwalTambahan := models.JadwalPemeriksaanBerkala{
			PeralatanID:         peralatanID,
			JenisJadwal:         jenisTambahan,
			PeranJadwal:         stringPointer("tambahan"),
			FrekuensiPenggunaan: stringPointer(freq),
			JenisPemeliharaan:   stringPointer(jenisPemeliharaan),
			IntervalBulan:       intervalTambahan,
			TanggalTerakhir:     &tanggalTambahan,
			TanggalJatuhTempo:   &jatuhTempoTambahan,
			PIC:                 stringPointer(pic),
			Keterangan:          stringPointer(keterangan),
		}

		if err := tx.Create(&jadwalUtama).Error; err != nil {
			return fmt.Errorf("gagal membuat jadwal utama: %w", err)
		}

		if err := tx.Create(&jadwalTambahan).Error; err != nil {
			return fmt.Errorf("gagal membuat jadwal tambahan: %w", err)
		}

		return nil
	})
}

func stringPointer(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}

func (r *jadwalPemeriksaanBerkalaRepository) FindAll() (
	[]models.JadwalPemeriksaanBerkala, error,
) {
	var data []models.JadwalPemeriksaanBerkala

	err := r.db.
		Order("tanggal_jatuh_tempo ASC").
		Order("id ASC").
		Find(&data).Error

	return data, err
}

func (r *jadwalPemeriksaanBerkalaRepository) FindByID(
	id uint,
) (*models.JadwalPemeriksaanBerkala, error) {
	var data models.JadwalPemeriksaanBerkala

	if err := r.db.First(&data, id).Error; err != nil {
		return nil, err
	}

	return &data, nil
}

func (r *jadwalPemeriksaanBerkalaRepository) UpdateRealisasi(
	id uint,
	tanggal time.Time,
) error {
	var jadwal models.JadwalPemeriksaanBerkala

	if err := r.db.First(&jadwal, id).Error; err != nil {
		return err
	}

	jatuhTempo := tanggal.AddDate(0, jadwal.IntervalBulan, 0)

	return r.db.Model(&jadwal).Updates(map[string]interface{}{
		"tanggal_realisasi":   tanggal,
		"tanggal_terakhir":    tanggal,
		"tanggal_jatuh_tempo": jatuhTempo,
	}).Error
}
