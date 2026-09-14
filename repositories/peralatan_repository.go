package repositories

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"backend/models"

	"gorm.io/gorm"
)

type PeralatanRepository interface {
	CreatePeralatan(req *models.CreatePeralatanRequest) (string, error)
	FindByID(id uint) (*models.Peralatan, error)
}

type peralatanRepository struct {
	db *gorm.DB
}

func NewPeralatanRepository(db *gorm.DB) PeralatanRepository {
	return &peralatanRepository{db}
}

func (r *peralatanRepository) FindByID(id uint) (*models.Peralatan, error) {
	var peralatan models.Peralatan
	if err := r.db.First(&peralatan, id).Error; err != nil {
		return nil, err
	}

	return &peralatan, nil
}

func (r *peralatanRepository) CreatePeralatan(req *models.CreatePeralatanRequest) (string, error) {
	// Memulai Database Transaction
	var nomorAset string
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var kelompokAsset models.KelompokAsset
		if err := tx.First(&kelompokAsset, req.KelompokAsetID).Error; err != nil {
			return err
		}

		prefix := prefixNomorAset(kelompokAsset.Kode, kelompokAsset.Nama)
		if prefix == "" {
			return errors.New("kode atau nama kelompok asset harus memiliki minimal 3 huruf")
		}

		var nomorUrut int
		if err := tx.Model(&models.Peralatan{}).
			Where("nomor_aset LIKE ?", prefix+"-%").
			Select("COALESCE(MAX(CAST(SUBSTRING_INDEX(nomor_aset, '-', -1) AS UNSIGNED)), 0)").
			Scan(&nomorUrut).Error; err != nil {
			return err
		}

		nomorUrut++
		nomorAset = fmt.Sprintf("%s-%03d", prefix, nomorUrut)

		// 1. Mapping dan Simpan ke Tabel Master (Peralatan)
		peralatan := models.Peralatan{
			NomorAset:           nomorAset,
			NamaPeralatan:       req.NamaPeralatan,
			KategoriPeralatanID: req.KategoriPeralatanID,
			KelompokAsetID:      req.KelompokAsetID,
			RuanganID:           req.RuanganID,
			PICID:               req.PICID,
			Merek:               req.Merek,
			TipeModel:           req.TipeModel,
			NomorSeri:           req.NomorSeri,
			Foto:                req.Foto,
			StatusAlat:          req.StatusAlat, // Bisa dikirim dari frontend, atau hardcode "Karantina"
			Keterangan:          req.Keterangan,
		}

		if peralatan.StatusAlat == "" {
			peralatan.StatusAlat = "Aktif" // Default value jika kosong
		}

		// Insert ke tabel peralatan
		if err := tx.Create(&peralatan).Error; err != nil {
			return err
		}

		// Convert map[string]interface{} ke JSON bytes agar bisa di-unmarshal ke Struct spesifik
		detailBytes, err := json.Marshal(req.Detail)
		if err != nil {
			return errors.New("gagal memproses data detail peralatan")
		}

		// 2. Routing Simpan ke Tabel Detail berdasarkan KategoriPeralatanID
		switch req.KategoriPeralatanID {
		case 1: // ALAT UKUR (Sheet 1)
			var detail models.DetailAlatUkur
			if err := json.Unmarshal(detailBytes, &detail); err != nil {
				return err
			}
			detail.PeralatanID = peralatan.ID

			// Hitung Tgl Jatuh Tempo jika Tgl Kalibrasi dan Interval diisi
			if detail.TglKalibrasi != nil && detail.IntervalBulan > 0 {
				jatuhTempo := detail.TglKalibrasi.AddDate(0, detail.IntervalBulan, 0)
				detail.TglJatuhTempo = &jatuhTempo
			}

			if err := tx.Create(&detail).Error; err != nil {
				return err
			}

		case 2: // ALAT BANTU (Sheet 2)
			var detail models.DetailAlatBantu
			if err := json.Unmarshal(detailBytes, &detail); err != nil {
				return err
			}
			detail.PeralatanID = peralatan.ID

			// Hitung Tgl Jatuh Tempo Pemeriksaan
			if detail.TglPemeriksaanTerakhir != nil && detail.IntervalBulan > 0 {
				jatuhTempo := detail.TglPemeriksaanTerakhir.AddDate(0, detail.IntervalBulan, 0)
				detail.TglJatuhTempo = &jatuhTempo
			}

			if err := tx.Create(&detail).Error; err != nil {
				return err
			}

		case 3: // ARTEFAK ACUAN (Sheet 3)
			var detail models.DetailArtefakAcuan
			if err := json.Unmarshal(detailBytes, &detail); err != nil {
				return err
			}
			detail.PeralatanID = peralatan.ID

			// Hitung Jadwal Ulang Karakterisasi
			if detail.TglKarakterisasiTerakhir != nil && detail.IntervalBulan > 0 {
				jatuhTempo := detail.TglKarakterisasiTerakhir.AddDate(0, detail.IntervalBulan, 0)
				detail.TglJatuhTempo = &jatuhTempo
			}

			if err := tx.Create(&detail).Error; err != nil {
				return err
			}

		case 4: // KOMPONEN PENDUKUNG (Sheet 4)
			var detail models.DetailKomponenPendukung
			if err := json.Unmarshal(detailBytes, &detail); err != nil {
				return err
			}
			detail.PeralatanID = peralatan.ID

			// Komponen pendukung menggunakan input Tgl Kedaluwarsa langsung dari user/pabrik,
			// tidak perlu dihitung otomatis dengan interval.
			if err := tx.Create(&detail).Error; err != nil {
				return err
			}

		default:
			return errors.New("kategori_id tidak valid atau tidak didukung")
		}

		// Jika semua berhasil, return nil untuk Commit transaksi
		return nil
	})
	return nomorAset, err
}

func prefixNomorAset(kode, nama string) string {
	prefix := kode
	if strings.TrimSpace(prefix) == "" {
		prefix = nama
	}

	prefix = strings.ToUpper(prefix)
	prefix = regexp.MustCompile(`[^A-Z0-9]`).ReplaceAllString(prefix, "")
	if len(prefix) < 3 {
		return ""
	}

	return prefix[:3]
}
