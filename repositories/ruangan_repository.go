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

func (r *RuanganRepository) GetByID(id uint64) (*models.Ruangan, error) {
	var ru models.Ruangan
	err := r.DB.Where("id = ?", id).First(&ru).Error
	if err != nil {
		return nil, err
	}
	return &ru, nil
}
