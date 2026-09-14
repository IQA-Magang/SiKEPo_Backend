package repositories

import (
	"encoding/json"
	"errors"

	"backend/models"

	"gorm.io/gorm"
)

type PeralatanRepository interface {
	CreatePeralatan(req *models.CreatePeralatanRequest) error
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

func (r *peralatanRepository) CreatePeralatan(req *models.CreatePeralatanRequest) error {
	// Memulai Database Transaction
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Mapping dan Simpan ke Tabel Master (Peralatan)
		peralatan := models.Peralatan{
			NomorAset:           req.NomorAset,
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
}
