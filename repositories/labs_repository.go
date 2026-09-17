package repositories

import (
	"backend/models"

	"gorm.io/gorm"
)

type LabsRepository struct {
	DB *gorm.DB
}

func NewLabsRepository(db *gorm.DB) *LabsRepository {
	return &LabsRepository{
		DB: db,
	}
}

// GetAll mengambil seluruh data labs.
// Manager juga ikut di-load melalui Preload.
func (r *LabsRepository) GetAll() ([]models.Labs, error) {
	var labs []models.Labs

	err := r.DB.
		Preload("Manager").
		Order("id DESC").
		Find(&labs).Error

	if err != nil {
		return nil, err
	}

	return labs, nil
}

// GetByID mengambil lab berdasarkan ID.
func (r *LabsRepository) GetByID(id uint64) (*models.Labs, error) {
	var labs models.Labs

	err := r.DB.
		Preload("Manager").
		Where("id = ?", id).
		First(&labs).Error

	if err != nil {
		return nil, err
	}

	return &labs, nil
}

// GetByKodeLabs mengambil lab berdasarkan kode lab.
func (r *LabsRepository) GetByKodeLabs(kodeLabs string) (*models.Labs, error) {
	var labs models.Labs

	err := r.DB.
		Preload("Manager").
		Where("kode_labs = ?", kodeLabs).
		First(&labs).Error

	if err != nil {
		return nil, err
	}

	return &labs, nil
}

// GetByNamaLabs mengambil lab berdasarkan nama lab.
func (r *LabsRepository) GetByNamaLabs(namaLabs string) (*models.Labs, error) {
	var labs models.Labs

	err := r.DB.
		Preload("Manager").
		Where("nama_labs = ?", namaLabs).
		First(&labs).Error

	if err != nil {
		return nil, err
	}

	return &labs, nil
}

// ExistsByKodeLabs mengecek apakah kode lab sudah digunakan.
func (r *LabsRepository) ExistsByKodeLabs(kodeLabs string) (bool, error) {
	var count int64

	err := r.DB.
		Model(&models.Labs{}).
		Where("kode_labs = ?", kodeLabs).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// ExistsByNamaLabs mengecek apakah nama lab sudah digunakan.
func (r *LabsRepository) ExistsByNamaLabs(namaLabs string) (bool, error) {
	var count int64

	err := r.DB.
		Model(&models.Labs{}).
		Where("nama_labs = ?", namaLabs).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// Create membuat data lab baru.
func (r *LabsRepository) Create(labs *models.Labs) error {
	return r.DB.Create(labs).Error
}

// Update mengubah data lab.
func (r *LabsRepository) Update(
	id uint64,
	updates map[string]interface{},
) error {
	return r.DB.
		Model(&models.Labs{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// Delete menghapus lab berdasarkan ID.
// Karena model menggunakan gorm.DeletedAt,
// operasi ini akan menjadi soft delete.
func (r *LabsRepository) Delete(id uint64) error {
	return r.DB.Delete(&models.Labs{}, id).Error
}
