package repositories

import (
	"backend/models"

	"gorm.io/gorm"
)

type VerifikasiRepository struct {
	DB *gorm.DB
}

func NewVerifikasiRepository(db *gorm.DB) *VerifikasiRepository {
	return &VerifikasiRepository{
		DB: db,
	}
}

// =====================================================
// GET ALL
// =====================================================

func (r *VerifikasiRepository) GetAllVerifikasi() ([]models.Verifikasi, error) {

	var data []models.Verifikasi

	err := r.DB.
		Preload("Peralatan").
		Preload("PICUser").
		Preload("VerifiedByUser").
		Preload("HasilVerifikasi").
		Where("verifikasi.deleted_at IS NULL").
		Order("id_verifikasi DESC").
		Find(&data).
		Error

	return data, err
}

// =====================================================
// GET BY ID
// =====================================================

func (r *VerifikasiRepository) GetVerifikasiByID(
	id uint64,
) (*models.Verifikasi, error) {

	var data models.Verifikasi

	err := r.DB.
		Preload("Peralatan").
		Preload("PICUser").
		Preload("VerifiedByUser").
		Preload("HasilVerifikasi").
		Where("verifikasi.id_verifikasi = ?", id).
		First(&data).
		Error

	if err != nil {
		return nil, err
	}

	return &data, nil
}

// =====================================================
// GET BY PERALATAN
// =====================================================

func (r *VerifikasiRepository) GetByPeralatanID(
	peralatanID uint64,
) ([]models.Verifikasi, error) {

	var data []models.Verifikasi

	err := r.DB.
		Preload("Peralatan").
		Preload("PICUser").
		Preload("VerifiedByUser").
		Preload("HasilVerifikasi").
		Where("id_peralatan = ?", peralatanID).
		Where("deleted_at IS NULL").
		Order("tanggal_verifikasi DESC").
		Find(&data).
		Error

	return data, err
}

// =====================================================
// GET PENGAJUAN
// =====================================================

func (r *VerifikasiRepository) GetPengajuan() ([]models.Verifikasi, error) {

	var data []models.Verifikasi

	err := r.DB.
		Preload("Peralatan").
		Preload("PICUser").
		Preload("HasilVerifikasi").
		Where("verifikasi.status = ?", "Diajukan").
		Where("verifikasi.deleted_at IS NULL").
		Order("verifikasi.pic_signed_at DESC").
		Find(&data).
		Error

	return data, err
}

// =====================================================
// CREATE
// =====================================================

func (r *VerifikasiRepository) CreateVerifikasi(
	tx *gorm.DB,
	data *models.Verifikasi,
) error {

	return tx.Create(data).Error
}

// =====================================================
// CREATE HASIL
// =====================================================

func (r *VerifikasiRepository) CreateHasilVerifikasi(
	tx *gorm.DB,
	data *models.HasilVerifikasi,
) error {

	return tx.Create(data).Error
}

// =====================================================
// SIGN PIC
// =====================================================

func (r *VerifikasiRepository) SignPIC(
	tx *gorm.DB,
	id uint64,
	picID uint64,
	signature string,
) error {

	return tx.
		Model(&models.Verifikasi{}).
		Where("id_verifikasi = ?", id).
		Updates(map[string]interface{}{
			"status":        "Diajukan",
			"pic_id":        picID,
			"pic_signature": signature,
			"pic_signed_at": gorm.Expr("NOW()"),
		}).
		Error
}

// =====================================================
// APPROVE
// =====================================================

func (r *VerifikasiRepository) ApproveVerifikasi(
	tx *gorm.DB,
	id uint64,
	managerID uint64,
	signature string,
) error {

	return tx.
		Model(&models.Verifikasi{}).
		Where("id_verifikasi = ?", id).
		Updates(map[string]interface{}{
			"status":            "Disetujui",
			"keputusan":        "Layak",
			"verified_by":      managerID,
			"verified_at":      gorm.Expr("NOW()"),
			"manager_signature": signature,
			"manager_signed_at": gorm.Expr("NOW()"),
		}).
		Error
}

// =====================================================
// REJECT
// =====================================================

func (r *VerifikasiRepository) RejectVerifikasi(
	tx *gorm.DB,
	id uint64,
	managerID uint64,
	catatan string,
) error {

	return tx.
		Model(&models.Verifikasi{}).
		Where("id_verifikasi = ?", id).
		Updates(map[string]interface{}{
			"status":       "Ditolak",
			"keputusan":    "Tidak Layak",
			"verified_by":  managerID,
			"verified_at":  gorm.Expr("NOW()"),
			"tindak_lanjut": catatan,
		}).
		Error
}

// =====================================================
// DELETE
// =====================================================

func (r *VerifikasiRepository) DeleteVerifikasi(
	id uint64,
) error {

	var data models.Verifikasi

	if err := r.DB.
		Where("id_verifikasi = ?", id).
		First(&data).
		Error; err != nil {

		return err
	}

	return r.DB.Delete(&data).Error
}