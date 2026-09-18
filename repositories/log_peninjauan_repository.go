package repositories

import (
	"backend/models"

	"gorm.io/gorm"
)

type LogPeninjauanRepository struct {
	DB *gorm.DB
}

func NewLogPeninjauanRepository(
	db *gorm.DB,
) *LogPeninjauanRepository {

	return &LogPeninjauanRepository{
		DB: db,
	}
}

// =====================================================
// CREATE
// =====================================================

func (r *LogPeninjauanRepository) Create(
	tx *gorm.DB,
	data *models.LogPeninjauanPeralatan,
) error {

	return tx.Create(data).Error
}

// =====================================================
// GET ALL
// =====================================================

func (r *LogPeninjauanRepository) GetAll() (
	[]models.LogPeninjauanPeralatan,
	error,
) {

	var data []models.LogPeninjauanPeralatan

	err := r.DB.
		Preload("Peralatan").
		Preload("Verifikasi").
		Preload("Manager").
		Where("deleted_at IS NULL").
		Order("id_log DESC").
		Find(&data).
		Error

	return data, err
}

// =====================================================
// GET BY PERALATAN
// =====================================================

func (r *LogPeninjauanRepository) GetByPeralatan(
	peralatanID uint64,
) ([]models.LogPeninjauanPeralatan, error) {

	var data []models.LogPeninjauanPeralatan

	err := r.DB.
		Preload("Peralatan").
		Preload("Verifikasi").
		Preload("Manager").
		Where("id_peralatan = ?", peralatanID).
		Where("deleted_at IS NULL").
		Order("id_log DESC").
		Find(&data).
		Error

	return data, err
}
