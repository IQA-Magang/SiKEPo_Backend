package repositories

import (
	"time"

	"backend/models"

	"gorm.io/gorm"
)

type DetailPeminjamanRepository struct {
	DB *gorm.DB
}

func NewDetailPeminjamanRepository(db *gorm.DB) *DetailPeminjamanRepository {
	return &DetailPeminjamanRepository{
		DB: db,
	}
}

// ========================================
// GET ALL
// ========================================

func (r *DetailPeminjamanRepository) GetAll() ([]models.DetailPeminjaman, error) {

	var data []models.DetailPeminjaman

	err := r.DB.
		Preload("Peralatan").
		Preload("VerifiedByUser").
		Order("id DESC").
		Find(&data).Error

	return data, err
}

// ========================================
// GET BY ID
// ========================================

func (r *DetailPeminjamanRepository) GetByID(
	id uint64,
) (*models.DetailPeminjaman, error) {

	var data models.DetailPeminjaman

	err := r.DB.
		Preload("Peralatan").
		Preload("VerifiedByUser").
		Where("id = ?", id).
		First(&data).Error

	if err != nil {
		return nil, err
	}

	return &data, nil
}

// ========================================
// GET BY PEMINJAMAN ID
// ========================================

func (r *DetailPeminjamanRepository) GetByPeminjamanID(
	peminjamanID uint64,
) ([]models.DetailPeminjaman, error) {

	var data []models.DetailPeminjaman

	err := r.DB.
		Preload("Peralatan").
		Preload("VerifiedByUser").
		Where("peminjaman_id = ?", peminjamanID).
		Order("id ASC").
		Find(&data).Error

	return data, err
}

// ========================================
// CREATE
// ========================================

func (r *DetailPeminjamanRepository) Create(
	data *models.DetailPeminjaman,
) error {

	return r.DB.Create(data).Error
}

// ========================================
// UPDATE DATA DETAIL
// ========================================

func (r *DetailPeminjamanRepository) Update(
	id uint64,
	data *models.DetailPeminjaman,
) error {

	var existing models.DetailPeminjaman

	err := r.DB.
		Where("id = ?", id).
		First(&existing).Error

	if err != nil {
		return err
	}

	existing.PeminjamanID = data.PeminjamanID
	existing.PeralatanID = data.PeralatanID
	existing.Jumlah = data.Jumlah
	existing.Catatan = data.Catatan

	return r.DB.Save(&existing).Error
}

// ========================================
// APPROVE
// ========================================

func (r *DetailPeminjamanRepository) Approve(
	id uint64,
	verifiedBy uint64,
	note string,
) error {

	now := time.Now()

	return r.DB.
		Model(&models.DetailPeminjaman{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":            "approved",
			"verified_by":       verifiedBy,
			"verified_at":       now,
			"verification_note": note,
		}).Error
}

// ========================================
// REJECT
// ========================================

func (r *DetailPeminjamanRepository) Reject(
	id uint64,
	verifiedBy uint64,
	note string,
) error {

	now := time.Now()

	return r.DB.
		Model(&models.DetailPeminjaman{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":            "rejected",
			"verified_by":       verifiedBy,
			"verified_at":       now,
			"verification_note": note,
		}).Error
}

// ========================================
// KONDISI SAAT PINJAM
// ========================================

func (r *DetailPeminjamanRepository) SetKondisiPinjam(
	id uint64,
	kondisi string,
	catatan string,
) error {

	return r.DB.
		Model(&models.DetailPeminjaman{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"kondisi_saat_pinjam": kondisi,
			"catatan":             catatan,
		}).Error
}

// ========================================
// KONDISI SAAT KEMBALI
// ========================================

func (r *DetailPeminjamanRepository) SetKondisiKembali(
	id uint64,
	kondisi string,
	catatan string,
) error {

	return r.DB.
		Model(&models.DetailPeminjaman{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"kondisi_saat_kembali": kondisi,
			"catatan":              catatan,
		}).Error
}

// ========================================
// DELETE / SOFT DELETE
// ========================================

func (r *DetailPeminjamanRepository) Delete(
	id uint64,
) error {

	var data models.DetailPeminjaman

	err := r.DB.
		Where("id = ?", id).
		First(&data).Error

	if err != nil {
		return err
	}

	return r.DB.Delete(&data).Error
}
