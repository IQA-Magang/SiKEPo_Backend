package repositories

import (
	"backend/models"

	"gorm.io/gorm"
)

type LogbookRepository struct {
	DB *gorm.DB
}

func NewLogbookRepository(db *gorm.DB) *LogbookRepository {
	return &LogbookRepository{DB: db}
}

// =====================================================
// CREATE
// =====================================================

// Create menambah satu baris logbook. Kirim transaksi sebagai tx agar baris
// ini ikut dibatalkan bila aksi utamanya gagal.
func (r *LogbookRepository) Create(
	tx *gorm.DB,
	data *models.LogbookPeralatan,
) error {

	return tx.Create(data).Error
}

// =====================================================
// GET BY PERALATAN
// =====================================================

// FindByPeralatan mengambil logbook satu peralatan, terbaru di atas.
// Jenis boleh kosong untuk mengambil semua jenis.
func (r *LogbookRepository) FindByPeralatan(
	peralatanID uint64,
	jenis string,
) ([]models.LogbookPeralatan, error) {

	query := r.DB.
		// Unscoped agar nama pelaku tetap tampil walau user sudah dihapus.
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Unscoped().Select("user_id", "name")
		}).
		Where("peralatan_id = ?", peralatanID)

	if jenis != "" {
		query = query.Where("jenis = ?", jenis)
	}

	var data []models.LogbookPeralatan

	err := query.
		Order("created_at DESC, id DESC").
		Find(&data).
		Error

	return data, err
}
