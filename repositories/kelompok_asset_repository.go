package repositories

import (
	"backend/models"

	"gorm.io/gorm"
)

type KelompokAssetRepository struct {
	DB *gorm.DB
}

func NewKelompokAssetRepository(db *gorm.DB) *KelompokAssetRepository {
	return &KelompokAssetRepository{
		DB: db,
	}
}

// =====================================================
// CREATE
// =====================================================

func (r *KelompokAssetRepository) Create(
	data *models.KelompokAsset,
) error {

	return r.DB.Create(data).Error
}

// =====================================================
// GET ALL
// =====================================================

func (r *KelompokAssetRepository) FindAll(
	search string,
	labID uint64,
	picID uint64,
) ([]models.KelompokAsset, error) {

	var list []models.KelompokAsset

	query := r.DB.
		Model(&models.KelompokAsset{}).
		Preload("Lab").
		Preload("PIC")

	// SEARCH
	if search != "" {

		searchValue := "%" + search + "%"

		query = query.Where(
			"kode LIKE ? OR nama LIKE ?",
			searchValue,
			searchValue,
		)
	}

	// FILTER LAB
	if labID > 0 {
		query = query.Where(
			"lab_id = ?",
			labID,
		)
	}

	// FILTER PIC
	if picID > 0 {
		query = query.Where(
			"pic_id = ?",
			picID,
		)
	}

	err := query.
		Order("created_at DESC").
		Find(&list).
		Error

	if err != nil {
		return nil, err
	}

	return list, nil
}

// =====================================================
// GET BY ID
// =====================================================

func (r *KelompokAssetRepository) FindByID(
	id uint64,
) (*models.KelompokAsset, error) {

	var data models.KelompokAsset

	err := r.DB.
		Preload("Lab").
		Preload("PIC").
		Where("id = ?", id).
		First(&data).
		Error

	if err != nil {
		return nil, err
	}

	return &data, nil
}

// =====================================================
// UPDATE
// =====================================================

func (r *KelompokAssetRepository) Update(
	id uint64,
	updates map[string]interface{},
) error {

	if len(updates) == 0 {
		return nil
	}

	result := r.DB.
		Model(&models.KelompokAsset{}).
		Where("id = ?", id).
		Updates(updates)

	return result.Error
}

// =====================================================
// DELETE
// =====================================================

func (r *KelompokAssetRepository) Delete(
	id uint64,
) error {

	result := r.DB.
		Where("id = ?", id).
		Delete(&models.KelompokAsset{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// =====================================================
// CHECK DUPLICATE
// =====================================================

func (r *KelompokAssetRepository) IsExists(
	labID uint64,
	kode string,
	excludeID ...uint64,
) (bool, error) {

	query := r.DB.
		Model(&models.KelompokAsset{}).
		Where("lab_id = ?", labID).
		Where("kode = ?", kode)

	// Untuk UPDATE
	// agar data dirinya sendiri tidak dianggap duplicate
	if len(excludeID) > 0 {
		query = query.Where(
			"id != ?",
			excludeID[0],
		)
	}

	var count int64

	err := query.Count(&count).Error

	return count > 0, err
}
