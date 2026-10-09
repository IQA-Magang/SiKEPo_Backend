package repositories

import (
	"time"

	"backend/models"

	"gorm.io/gorm"
)

type PeminjamanRepository struct {
	DB *gorm.DB
}

func NewPeminjamanRepository(db *gorm.DB) *PeminjamanRepository {
	return &PeminjamanRepository{DB: db}
}

// PeminjamanFilter menentukan data yang tampil pada daftar peminjaman.
type PeminjamanFilter struct {
	UserID  uint64
	IsAdmin bool

	// Peran: kosong, peminjam, manajer_peminjam, pengelola, atau manager_lab.
	Peran string

	// Status: kosong berarti semua status.
	Status string

	// MenungguSaya hanya menampilkan pengajuan yang sedang menunggu
	// keputusan dari user ini.
	MenungguSaya bool

	PeralatanID uint64
}

func (r *PeminjamanRepository) withRelations(query *gorm.DB) *gorm.DB {
	return query.
		Preload("Peralatan").
		Preload("Peminjam").
		Preload("LabPeminjam").
		Preload("LabPemilik").
		Preload("ManajerPeminjam").
		Preload("Pengelola").
		Preload("ManagerLab").
		Preload("AlatAlternatif").
		Preload("SerahTerima").
		Preload("SerahTerima.Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("nomor_butir ASC")
		})
}

// =====================================================
// FIND BY ID
// =====================================================

func (r *PeminjamanRepository) FindByID(id uint64) (*models.Peminjaman, error) {
	var data models.Peminjaman

	err := r.withRelations(r.DB).
		Where("id = ?", id).
		First(&data).
		Error
	if err != nil {
		return nil, err
	}

	return &data, nil
}

// =====================================================
// FIND ALL
// =====================================================

// FindAll mengambil daftar peminjaman, terbaru di atas. Tanpa filter peran,
// user hanya melihat pengajuan yang melibatkan dirinya, sedangkan admin
// melihat semuanya.
func (r *PeminjamanRepository) FindAll(filter PeminjamanFilter) ([]models.Peminjaman, error) {
	query := r.withRelations(r.DB.Model(&models.Peminjaman{}))

	uid := filter.UserID

	switch filter.Peran {
	case "peminjam":
		query = query.Where("peminjam_id = ?", uid)

	case "manajer_peminjam":
		// Pengajuan yang tahap Manajer Peminjam-nya dilewati tidak ditampilkan.
		query = query.
			Where("manajer_peminjam_id = ?", uid).
			Where("manajer_peminjam_status <> ?", models.TahapSkipped)

	case "pengelola":
		query = query.Where("pengelola_id = ?", uid)

	case "manager_lab":
		query = query.Where("manager_lab_id = ?", uid)

	default:
		if !filter.IsAdmin {
			query = query.Where(
				"(peminjam_id = ? OR manajer_peminjam_id = ? OR pengelola_id = ? OR manager_lab_id = ?)",
				uid, uid, uid, uid,
			)
		}
	}

	if filter.MenungguSaya {
		query = query.Where(
			"((status = ? AND manajer_peminjam_id = ?) OR (status = ? AND pengelola_id = ?) OR (status = ? AND manager_lab_id = ?) OR (status = ? AND pengelola_id = ?) OR (status = ? AND peminjam_id = ?))",
			models.StatusPeminjamanMenungguManajerPeminjam, uid,
			models.StatusPeminjamanMenungguPengelola, uid,
			models.StatusPeminjamanMenungguManagerLab, uid,
			// Disetujui: Pengelola perlu mengisi checklist serah terima.
			models.StatusPeminjamanDisetujui, uid,
			// Siap diserahkan: peminjam perlu mengonfirmasi terima.
			models.StatusPeminjamanSiapDiserahkan, uid,
		)
	}

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	if filter.PeralatanID != 0 {
		query = query.Where("peralatan_id = ?", filter.PeralatanID)
	}

	var data []models.Peminjaman

	err := query.
		Order("id DESC").
		Find(&data).
		Error

	return data, err
}

// =====================================================
// JADWAL BENTROK
// =====================================================

// HasJadwalBentrok mengecek apakah peralatan sudah dipegang pengajuan lain
// yang masih berjalan pada rentang tanggal yang beririsan. excludeID dipakai
// untuk mengabaikan pengajuan yang sedang diproses, atau 0 bila tidak ada.
func (r *PeminjamanRepository) HasJadwalBentrok(
	tx *gorm.DB,
	peralatanID uint64,
	keluar time.Time,
	kembali time.Time,
	excludeID uint64,
) (bool, error) {

	query := tx.Model(&models.Peminjaman{}).
		Where("peralatan_id = ?", peralatanID).
		Where("status IN ?", models.StatusPeminjamanAktif).
		Where("rencana_tanggal_keluar <= ? AND rencana_tanggal_kembali >= ?", kembali, keluar)

	if excludeID != 0 {
		query = query.Where("id <> ?", excludeID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return false, err
	}

	return total > 0, nil
}
