package repositories

import (
	"backend/models"

	"gorm.io/gorm"
)

type KategoriPeralatanRepository struct {
	DB *gorm.DB
}

func NewKategoriPeralatanRepository(db *gorm.DB) *KategoriPeralatanRepository {
	return &KategoriPeralatanRepository{DB: db}
}

func (r *KategoriPeralatanRepository) GetAll() ([]models.KategoriPeralatan, error) {
	var data []models.KategoriPeralatan
	if err := r.DB.Order("id DESC").Find(&data).Error; err != nil {
		return nil, err
	}
	return data, nil
}

func (r *KategoriPeralatanRepository) GetByID(id uint64) (*models.KategoriPeralatan, error) {
	var data models.KategoriPeralatan
	if err := r.DB.Where("id = ?", id).First(&data).Error; err != nil {
		return nil, err
	}
	return &data, nil
}

func (r *KategoriPeralatanRepository) GetByNama(nama string) (*models.KategoriPeralatan, error) {
	var data models.KategoriPeralatan
	if err := r.DB.Where("nama_kategori = ?", nama).First(&data).Error; err != nil {
		return nil, err
	}
	return &data, nil
}

func (r *KategoriPeralatanRepository) ExistsByNama(nama string) (bool, error) {
	var count int64
	if err := r.DB.Model(&models.KategoriPeralatan{}).Where("nama_kategori = ?", nama).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *KategoriPeralatanRepository) Create(data *models.KategoriPeralatan) error {
	return r.DB.Create(data).Error
}

func (r *KategoriPeralatanRepository) Update(id uint64, updates map[string]interface{}) error {
	return r.DB.Model(&models.KategoriPeralatan{}).Where("id = ?", id).Updates(updates).Error
}

func (r *KategoriPeralatanRepository) Delete(id uint64) error {
	return r.DB.Delete(&models.KategoriPeralatan{}, id).Error
}
