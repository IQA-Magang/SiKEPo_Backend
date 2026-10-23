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
	FindAll() ([]models.Peralatan, error)

	CreatePeralatan(
		req *models.CreatePeralatanRequest,
	) (string, error)

	FindByID(
		id uint,
	) (*models.Peralatan, error)

	FindByNomorAset(
		nomorAset string,
	) (*models.Peralatan, error)

	UpdateFoto(
		id uint,
		foto string,
	) error
}

type peralatanRepository struct {
	db *gorm.DB
}

func NewPeralatanRepository(
	db *gorm.DB,
) PeralatanRepository {
	return &peralatanRepository{
		db: db,
	}
}

// =========================================================
// FIND ALL
// =========================================================

func (r *peralatanRepository) FindAll() (
	[]models.Peralatan,
	error,
) {
	var data []models.Peralatan

	err := r.db.
		Preload("KategoriPeralatan").
		Where("deleted_at IS NULL").
		Order("id DESC").
		Find(&data).
		Error

	return data, err
}

// =========================================================
// FIND BY ID
// =========================================================

func (r *peralatanRepository) FindByID(
	id uint,
) (*models.Peralatan, error) {

	var data models.Peralatan

	err := r.db.
		Preload("KategoriPeralatan").
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		First(&data).
		Error

	if err != nil {
		return nil, err
	}

	return &data, nil
}

// =========================================================
// FIND BY NOMOR ASET
// =========================================================

func (r *peralatanRepository) FindByNomorAset(
	nomorAset string,
) (*models.Peralatan, error) {

	var data models.Peralatan

	err := r.db.
		Preload("KategoriPeralatan").
		Where(
			"nomor_aset = ?",
			nomorAset,
		).
		First(&data).
		Error

	if err != nil {
		return nil, err
	}

	return &data, nil
}

// =========================================================
// UPDATE FOTO
// =========================================================

func (r *peralatanRepository) UpdateFoto(
	id uint,
	foto string,
) error {

	return r.db.
		Model(&models.Peralatan{}).
		Where("id = ?", id).
		Update("foto", foto).
		Error
}

// =========================================================
// CREATE PERALATAN
// =========================================================

func (r *peralatanRepository) CreatePeralatan(
	req *models.CreatePeralatanRequest,
) (string, error) {

	var nomorAset string

	err := r.db.Transaction(
		func(tx *gorm.DB) error {

			// =================================================
			// VALIDASI KELOMPOK ASSET
			// =================================================

			var kelompokAsset models.KelompokAsset

			if err := tx.
				Where(
					"id = ?",
					req.KelompokAsetID,
				).
				First(&kelompokAsset).
				Error; err != nil {

				return fmt.Errorf(
					"kelompok asset tidak ditemukan: %w",
					err,
				)
			}

			// =================================================
			// AMBIL RUANGAN
			// =================================================

			var ruangan models.Ruangan

			if err := tx.
				Preload("Labs").
				Where(
					"id = ?",
					req.RuanganID,
				).
				First(&ruangan).
				Error; err != nil {

				return fmt.Errorf(
					"ruangan tidak ditemukan: %w",
					err,
				)
			}

			// =================================================
			// VALIDASI LAB
			// =================================================

			if ruangan.Labs == nil {
				return errors.New(
					"ruangan belum terhubung dengan laboratorium",
				)
			}

			// =================================================
			// VALIDASI MANAGER LAB
			// =================================================

			if ruangan.Labs.ManagerID == nil {
				return errors.New(
					"laboratorium belum memiliki Manager Lab",
				)
			}

			// =================================================
			// PIC OTOMATIS = MANAGER LAB
			// =================================================
			//
			// Manager Lab dari lab yang menaungi ruangan
			// otomatis menjadi PIC peralatan.
			//
			// ManagerID bertipe uint64.
			// PICID pada Peralatan bertipe uint.
			// Karena itu dilakukan konversi uint().
			// =================================================

			picID := *ruangan.Labs.ManagerID

			// =================================================
			// PREFIX NOMOR ASET
			// =================================================

			prefix := prefixNomorAset(
				kelompokAsset.Kode,
				kelompokAsset.Nama,
			)

			if prefix == "" {
				return errors.New(
					"kode atau nama kelompok asset minimal 3 karakter",
				)
			}

			// =================================================
			// NOMOR URUT
			// =================================================

			var nomorUrut int

			err := tx.
				Model(&models.Peralatan{}).
				Where(
					"nomor_aset LIKE ?",
					prefix+"-%",
				).
				Select(
					"COALESCE(MAX(CAST(SUBSTRING_INDEX(nomor_aset, '-', -1) AS UNSIGNED)), 0)",
				).
				Scan(&nomorUrut).
				Error

			if err != nil {
				return err
			}

			nomorUrut++

			nomorAset = fmt.Sprintf(
				"%s-%03d",
				prefix,
				nomorUrut,
			)

			// =================================================
			// STATUS AWAL
			// =================================================

			statusAlat := strings.TrimSpace(
				req.StatusAlat,
			)

			if statusAlat == "" {
				statusAlat = "Karantina"
			}

			// =================================================
			// MASTER PERALATAN
			// =================================================

			peralatan := models.Peralatan{
				NomorAset: nomorAset,

				NamaPeralatan: req.NamaPeralatan,

				KategoriID: req.KategoriPeralatanID,

				KategoriPeralatanID: req.KategoriPeralatanID,

				KelompokAsetID: req.KelompokAsetID,

				RuanganID: req.RuanganID,

				// =================================================
				// PIC OTOMATIS = MANAGER LAB
				// =================================================
				PICID: uint(picID),

				Merek: req.Merek,

				TipeModel: req.TipeModel,

				NomorSeri: req.NomorSeri,

				Foto: req.Foto,

				StatusAlat: statusAlat,

				StatusVerifikasi: "Belum Diverifikasi",

				Keterangan: req.Keterangan,
			}

			// =================================================
			// SIMPAN MASTER PERALATAN
			// =================================================

			if err := tx.
				Create(&peralatan).
				Error; err != nil {

				return err
			}

			// =================================================
			// PROSES DETAIL PERALATAN
			// =================================================

			detailBytes, err :=
				json.Marshal(req.Detail)

			if err != nil {
				return errors.New(
					"gagal memproses detail peralatan",
				)
			}

			// =================================================
			// KATEGORI 1 - ALAT UKUR
			// =================================================

			switch req.KategoriPeralatanID {

			case 1:

				var detail models.DetailAlatUkur

				if err := json.Unmarshal(
					detailBytes,
					&detail,
				); err != nil {
					return err
				}

				// Default jenis label
				if detail.JenisLabel == "" {
					detail.JenisLabel = "calibration"
				}

				// Validasi jenis label
				if detail.JenisLabel != "calibration" &&
					detail.JenisLabel != "limited calibration" &&
					detail.JenisLabel != "do not use" {

					return fmt.Errorf(
						"jenis_label tidak valid: %q",
						detail.JenisLabel,
					)
				}

				detail.PeralatanID =
					peralatan.ID

				// Hitung tanggal jatuh tempo otomatis
				if detail.TglKalibrasi != nil &&
					detail.IntervalBulan > 0 {

					jatuhTempo :=
						detail.TglKalibrasi.AddDate(
							0,
							detail.IntervalBulan,
							0,
						)

					detail.TglJatuhTempo =
						&jatuhTempo
				}

				if err := tx.
					Create(&detail).
					Error; err != nil {

					return err
				}

			// =================================================
			// KATEGORI 2 - ALAT BANTU
			// =================================================

			case 2:

				var detail models.DetailAlatBantu

				if err := json.Unmarshal(
					detailBytes,
					&detail,
				); err != nil {
					return err
				}

				detail.PeralatanID =
					peralatan.ID

				// Hitung tanggal jatuh tempo otomatis
				if detail.TglPemeriksaanTerakhir != nil &&
					detail.IntervalBulan > 0 {

					jatuhTempo :=
						detail.TglPemeriksaanTerakhir.AddDate(
							0,
							detail.IntervalBulan,
							0,
						)

					detail.TglJatuhTempo =
						&jatuhTempo
				}

				if err := tx.
					Create(&detail).
					Error; err != nil {

					return err
				}

			// =================================================
			// KATEGORI 3 - ARTEFAK ACUAN
			// =================================================

			case 3:

				var detail models.DetailArtefakAcuan

				if err := json.Unmarshal(
					detailBytes,
					&detail,
				); err != nil {
					return err
				}

				detail.PeralatanID =
					peralatan.ID

				if err := tx.
					Create(&detail).
					Error; err != nil {

					return err
				}

			// =================================================
			// KATEGORI 4 - KOMPONEN PENDUKUNG
			// =================================================

			case 4:

				var detail models.DetailKomponenPendukung

				if err := json.Unmarshal(
					detailBytes,
					&detail,
				); err != nil {
					return err
				}

				detail.PeralatanID =
					peralatan.ID

				if err := tx.
					Create(&detail).
					Error; err != nil {

					return err
				}

			// =================================================
			// KATEGORI TIDAK VALID
			// =================================================

			default:

				return errors.New(
					"kategori_peralatan_id tidak valid",
				)
			}

			return nil
		},
	)

	// =========================================================
	// TRANSACTION ERROR
	// =========================================================

	if err != nil {
		return "", err
	}

	// =========================================================
	// RETURN NOMOR ASET
	// =========================================================

	return nomorAset, nil
}

// =========================================================
// PREFIX NOMOR ASET
// =========================================================

func prefixNomorAset(
	kode string,
	nama string,
) string {

	prefix := strings.TrimSpace(
		kode,
	)

	if prefix == "" {
		prefix = strings.TrimSpace(
			nama,
		)
	}

	prefix = strings.ToUpper(
		prefix,
	)

	prefix = regexp.
		MustCompile(
			`[^A-Z0-9]`,
		).
		ReplaceAllString(
			prefix,
			"",
		)

	if len(prefix) < 3 {
		return ""
	}

	return prefix[:3]
}
