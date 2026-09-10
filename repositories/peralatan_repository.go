package repositories

import (
	"backend/models"

	"gorm.io/gorm"
)

type PeralatanRepository struct {
	DB *gorm.DB
}

func NewPeralatanRepository(db *gorm.DB) *PeralatanRepository {
	return &PeralatanRepository{DB: db}
}

func (r *PeralatanRepository) Create(p *models.Peralatan) error {
	return r.DB.Create(p).Error
}

type PeralatanQueryParams struct {
	Page            int
	Limit           int
	Search          string
	RuanganID       uint64
	PICID           uint64
	StatusKelayakan string
}

func (r *PeralatanRepository) FindAll(params PeralatanQueryParams) ([]models.Peralatan, int64, error) {
	var list []models.Peralatan
	var total int64
	query := r.DB.Model(&models.Peralatan{})
	if params.Search != "" {
		search := "%" + params.Search + "%"
		query = query.Where("nama_peralatan LIKE ? OR nomor_aset LIKE ? OR merk LIKE ?", search, search, search)
	}
	if params.RuanganID > 0 {
		query = query.Where("ruangan_id = ?", params.RuanganID)
	}
	if params.PICID > 0 {
		query = query.Where("pic_id = ?", params.PICID)
	}
	if params.StatusKelayakan != "" {
		query = query.Where("status_kelayakan = ?", params.StatusKelayakan)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Preload("Ruangan").Preload("PIC").Preload("InputByUser").Preload("VerifiedByUser").
		Order("created_at DESC").Offset((params.Page - 1) * params.Limit).Limit(params.Limit).Find(&list).Error
	return list, total, err
}

func (r *PeralatanRepository) FindByID(id uint64) (*models.Peralatan, error) {
	var p models.Peralatan
	err := r.DB.Preload("Ruangan").Preload("PIC").Preload("InputByUser").Preload("VerifiedByUser").
		Where("id = ?", id).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PeralatanRepository) Update(id uint64, updates map[string]interface{}) error {
	return r.DB.Model(&models.Peralatan{}).Where("id = ?", id).Updates(updates).Error
}

func (r *PeralatanRepository) Delete(id uint64) error {
	return r.DB.Delete(&models.Peralatan{}, id).Error
}

func (r *PeralatanRepository) IsNomorAsetExists(nomorAset string, excludeID ...uint64) (bool, error) {
	query := r.DB.Model(&models.Peralatan{}).Where("nomor_aset = ?", nomorAset)
	if len(excludeID) > 0 {
		query = query.Where("id != ?", excludeID[0])
	}
	var count int64
	err := query.Count(&count).Error
	return count > 0, err
}
