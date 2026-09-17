package repositories

import (
	"backend/models"

	"gorm.io/gorm"
)

type RuanganRepository struct {
	DB *gorm.DB
}

func NewRuanganRepository(db *gorm.DB) *RuanganRepository {
	return &RuanganRepository{
		DB: db,
	}
}

// GetAll mengambil seluruh data ruangan.
// Relasi Labs dan PICUser ikut di-load.
func (r *RuanganRepository) GetAll() ([]models.Ruangan, error) {
	var ruangan []models.Ruangan

	err := r.DB.
		Preload("Labs").
		Preload("PICUser").
		Order("id DESC").
		Find(&ruangan).Error

	if err != nil {
		return nil, err
	}

	return ruangan, nil
}

// GetByID mengambil ruangan berdasarkan ID.
func (r *RuanganRepository) GetByID(id uint64) (*models.Ruangan, error) {
	var ruangan models.Ruangan

	err := r.DB.
		Preload("Labs").
		Preload("PICUser").
		Where("id = ?", id).
		First(&ruangan).Error

	if err != nil {
		return nil, err
	}

	return &ruangan, nil
}

// GetByKodeRuangan mengambil ruangan berdasarkan kode ruangan.
func (r *RuanganRepository) GetByKodeRuangan(
	kodeRuangan string,
) (*models.Ruangan, error) {
	var ruangan models.Ruangan

	err := r.DB.
		Preload("Labs").
		Preload("PICUser").
		Where("kode_ruangan = ?", kodeRuangan).
		First(&ruangan).Error

	if err != nil {
		return nil, err
	}

	return &ruangan, nil
}

// GetByNamaRuangan mengambil ruangan berdasarkan nama ruangan.
func (r *RuanganRepository) GetByNamaRuangan(
	namaRuangan string,
) (*models.Ruangan, error) {
	var ruangan models.Ruangan

	err := r.DB.
		Preload("Labs").
		Preload("PICUser").
		Where("nama_ruangan = ?", namaRuangan).
		First(&ruangan).Error

	if err != nil {
		return nil, err
	}

	return &ruangan, nil
}

// GetByLabsID mengambil seluruh ruangan dalam satu lab.
func (r *RuanganRepository) GetByLabsID(
	labsID uint64,
) ([]models.Ruangan, error) {
	var ruangan []models.Ruangan

	err := r.DB.
		Preload("Labs").
		Preload("PICUser").
		Where("labs_id = ?", labsID).
		Order("id DESC").
		Find(&ruangan).Error

	if err != nil {
		return nil, err
	}

	return ruangan, nil
}

// GetByPICUserID mengambil seluruh ruangan yang ditangani
// oleh seorang PIC.
func (r *RuanganRepository) GetByPICUserID(
	picUserID uint64,
) ([]models.Ruangan, error) {
	var ruangan []models.Ruangan

	err := r.DB.
		Preload("Labs").
		Preload("PICUser").
		Where("pic_user_id = ?", picUserID).
		Order("id DESC").
		Find(&ruangan).Error

	if err != nil {
		return nil, err
	}

	return ruangan, nil
}

// ExistsByKodeRuangan mengecek apakah kode ruangan sudah digunakan.
func (r *RuanganRepository) ExistsByKodeRuangan(
	kodeRuangan string,
) (bool, error) {
	var count int64

	err := r.DB.
		Model(&models.Ruangan{}).
		Where("kode_ruangan = ?", kodeRuangan).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// ExistsByNamaRuangan mengecek apakah nama ruangan sudah digunakan.
func (r *RuanganRepository) ExistsByNamaRuangan(
	namaRuangan string,
) (bool, error) {
	var count int64

	err := r.DB.
		Model(&models.Ruangan{}).
		Where("nama_ruangan = ?", namaRuangan).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// Create membuat ruangan baru.
func (r *RuanganRepository) Create(
	ruangan *models.Ruangan,
) error {
	return r.DB.Create(ruangan).Error
}

// Update mengubah data ruangan.
func (r *RuanganRepository) Update(
	id uint64,
	updates map[string]interface{},
) error {
	return r.DB.
		Model(&models.Ruangan{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// Delete menghapus ruangan berdasarkan ID.
// Karena menggunakan gorm.DeletedAt,
// operasi ini merupakan soft delete.
func (r *RuanganRepository) Delete(id uint64) error {
	return r.DB.Delete(&models.Ruangan{}, id).Error
}
