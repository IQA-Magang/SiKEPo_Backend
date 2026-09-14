package repositories

import (
	"backend/models"

	"gorm.io/gorm"
)

type DokumenPeralatanRepository struct {
	DB *gorm.DB
}

func NewDokumenPeralatanRepository(db *gorm.DB) *DokumenPeralatanRepository {
	return &DokumenPeralatanRepository{
		DB: db,
	}
}

// GetAll mengambil seluruh dokumen peralatan.
func (r *DokumenPeralatanRepository) GetAll() ([]models.DokumenPeralatan, error) {
	var dokumen []models.DokumenPeralatan

	err := r.DB.
		Preload("Peralatan").
		Order("id DESC").
		Find(&dokumen).Error

	if err != nil {
		return nil, err
	}

	return dokumen, nil
}

// GetByID mengambil dokumen berdasarkan ID.
func (r *DokumenPeralatanRepository) GetByID(
	id uint64,
) (*models.DokumenPeralatan, error) {
	var dokumen models.DokumenPeralatan

	err := r.DB.
		Preload("Peralatan").
		Where("id = ?", id).
		First(&dokumen).Error

	if err != nil {
		return nil, err
	}

	return &dokumen, nil
}

// GetByNamaDokumen mengambil dokumen berdasarkan nama.
func (r *DokumenPeralatanRepository) GetByNamaDokumen(
	namaDokumen string,
) (*models.DokumenPeralatan, error) {
	var dokumen models.DokumenPeralatan

	err := r.DB.
		Preload("Peralatan").
		Where("nama_dokumen = ?", namaDokumen).
		First(&dokumen).Error

	if err != nil {
		return nil, err
	}

	return &dokumen, nil
}

// GetByPeralatanID mengambil seluruh dokumen dari satu peralatan.
func (r *DokumenPeralatanRepository) GetByPeralatanID(
	peralatanID uint64,
) ([]models.DokumenPeralatan, error) {
	var dokumen []models.DokumenPeralatan

	err := r.DB.
		Preload("Peralatan").
		Where("peralatan_id = ?", peralatanID).
		Order("id DESC").
		Find(&dokumen).Error

	if err != nil {
		return nil, err
	}

	return dokumen, nil
}

// ExistsByNamaDokumen mengecek nama dokumen.
func (r *DokumenPeralatanRepository) ExistsByNamaDokumen(
	namaDokumen string,
) (bool, error) {
	var count int64

	err := r.DB.
		Model(&models.DokumenPeralatan{}).
		Where("nama_dokumen = ?", namaDokumen).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// Create membuat dokumen baru.
func (r *DokumenPeralatanRepository) Create(
	dokumen *models.DokumenPeralatan,
) error {
	return r.DB.Create(dokumen).Error
}

// Update mengubah data dokumen.
func (r *DokumenPeralatanRepository) Update(
	id uint64,
	updates map[string]interface{},
) error {
	return r.DB.
		Model(&models.DokumenPeralatan{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// Delete menghapus dokumen.
// Karena menggunakan gorm.DeletedAt,
// maka operasi ini adalah soft delete.
func (r *DokumenPeralatanRepository) Delete(id uint64) error {
	return r.DB.Delete(&models.DokumenPeralatan{}, id).Error
}
