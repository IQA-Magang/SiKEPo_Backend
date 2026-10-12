package repositories

import (
	"time"

	"backend/models"
)

// pemakaiRow adalah satu baris hasil query pemakai terakhir.
type pemakaiRow struct {
	PeralatanID   uint64
	UserID        uint64
	Nama          string
	TanggalKeluar time.Time
	PeminjamanID  uint64
	Kode          string
	Status        string
}

// isiPemakaiTerakhir mengisi field PemakaiTerakhir pada semua alat di daftar
// dengan satu query, bukan satu query per alat.
//
// Pemakai terakhir adalah peminjam dari serah terima keluar terbaru, yaitu
// pengajuan yang tgl_keluar_aktual-nya terisi. Alat yang belum pernah
// diserahkan dibiarkan null.
func (r *peralatanRepository) isiPemakaiTerakhir(data []models.Peralatan) error {
	if len(data) == 0 {
		return nil
	}

	ids := make([]uint64, 0, len(data))
	for _, alat := range data {
		ids = append(ids, uint64(alat.ID))
	}

	var rows []pemakaiRow

	err := r.db.
		Table("peminjaman AS p").
		Select(`p.peralatan_id AS peralatan_id,
			p.peminjam_id AS user_id,
			u.name AS nama,
			p.tgl_keluar_aktual AS tanggal_keluar,
			p.id AS peminjaman_id,
			p.kode AS kode,
			p.status AS status`).
		Joins("JOIN users u ON u.user_id = p.peminjam_id").
		Where("p.deleted_at IS NULL").
		Where("p.tgl_keluar_aktual IS NOT NULL").
		Where("p.peralatan_id IN ?", ids).
		Order("p.tgl_keluar_aktual DESC, p.id DESC").
		Scan(&rows).
		Error
	if err != nil {
		return err
	}

	// Baris sudah terurut dari yang terbaru, jadi yang pertama untuk tiap alat menang.
	terbaru := make(map[uint64]*models.PemakaiTerakhir, len(data))

	for _, row := range rows {
		if _, sudahAda := terbaru[row.PeralatanID]; sudahAda {
			continue
		}

		terbaru[row.PeralatanID] = &models.PemakaiTerakhir{
			UserID:          row.UserID,
			Nama:            row.Nama,
			TanggalKeluar:   row.TanggalKeluar,
			PeminjamanID:    row.PeminjamanID,
			Kode:            row.Kode,
			Status:          row.Status,
			SedangDigunakan: row.Status == models.StatusPeminjamanSedangDipinjam,
		}
	}

	for i := range data {
		data[i].PemakaiTerakhir = terbaru[uint64(data[i].ID)]
	}

	return nil
}
