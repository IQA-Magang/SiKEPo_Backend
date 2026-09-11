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

// ========================================
// GET ALL
// ========================================

func (r *VerifikasiRepository) GetAllVerifikasi() ([]models.Verifikasi, error) {

	var data []models.Verifikasi

	err := r.DB.
		Preload("Peralatan").
		Preload("VerifiedByUser").
		Preload("HasilVerifikasi").
		Where("verifikasi.deleted_at IS NULL").
		Order("id_verifikasi DESC").
		Find(&data).
		Error

	return data, err
}

// ========================================
// GET BY ID
// ========================================

func (r *VerifikasiRepository) GetVerifikasiByID(
	id uint64,
) (*models.Verifikasi, error) {

	var data models.Verifikasi

	err := r.DB.
		Preload("Peralatan").
		Preload("VerifiedByUser").
		Preload("HasilVerifikasi").
		Where("id_verifikasi = ?", id).
		First(&data).
		Error

	if err != nil {
		return nil, err
	}

	return &data, nil
}

// ========================================
// GET BY PERALATAN
// ========================================

func (r *VerifikasiRepository) GetByPeralatanID(
	peralatanID uint64,
) ([]models.Verifikasi, error) {

	var data []models.Verifikasi

	err := r.DB.
		Preload("Peralatan").
		Preload("VerifiedByUser").
		Preload("HasilVerifikasi").
		Where("id_peralatan = ?", peralatanID).
		Order("tanggal_verifikasi DESC").
		Find(&data).
		Error

	return data, err
}

// ========================================
// CREATE
// ========================================

func (r *VerifikasiRepository) CreateVerifikasi(
	data *models.Verifikasi,
) error {

	return r.DB.Create(data).Error
}

// ========================================
// APPROVE
// ========================================

func (r *VerifikasiRepository) ApproveVerifikasi(
	id uint64,
	verifiedBy uint64,
) error {

	return r.DB.
		Model(&models.Verifikasi{}).
		Where("id_verifikasi = ?", id).
		Updates(map[string]interface{}{
			"keputusan":   "Layak",
			"verified_by": verifiedBy,
			"verified_at": gorm.Expr("NOW()"),
		}).
		Error
}

// ========================================
// REJECT
// ========================================

func (r *VerifikasiRepository) RejectVerifikasi(
	id uint64,
	verifiedBy uint64,
	tindakLanjut string,
	catatan string,
) error {

	return r.DB.
		Model(&models.Verifikasi{}).
		Where("id_verifikasi = ?", id).
		Updates(map[string]interface{}{
			"keputusan":     "Tidak Layak",
			"verified_by":   verifiedBy,
			"verified_at":   gorm.Expr("NOW()"),
			"tindak_lanjut": tindakLanjut,
			"catatan":       catatan,
		}).
		Error
}

// ========================================
// CREATE HASIL VERIFIKASI
// ========================================

func (r *VerifikasiRepository) CreateHasilVerifikasi(
	data *models.HasilVerifikasi,
) error {

	return r.DB.Create(data).Error
}

// ========================================
// GET HASIL BY VERIFIKASI
// ========================================

func (r *VerifikasiRepository) GetHasilByVerifikasiID(
	id uint64,
) (*models.HasilVerifikasi, error) {

	var data models.HasilVerifikasi

	err := r.DB.
		Where("id_verifikasi = ?", id).
		First(&data).
		Error

	if err != nil {
		return nil, err
	}

	return &data, nil
}

// ========================================
// DELETE
// ========================================

func (r *VerifikasiRepository) DeleteVerifikasi(
	id uint64,
) error {

	var data models.Verifikasi

	err := r.DB.
		Where("id_verifikasi = ?", id).
		First(&data).
		Error

	if err != nil {
		return err
	}

	return r.DB.Delete(&data).Error
}
