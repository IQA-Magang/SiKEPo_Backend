package repositories

import "backend/models"

// FindPemakaiTerakhir mengambil pemakai terakhir untuk satu alat. Mengembalikan
// nil bila alat itu belum pernah diserahkan kepada siapa pun.
func (r *peralatanRepository) FindPemakaiTerakhir(id uint) (*models.PemakaiTerakhir, error) {
	data := []models.Peralatan{{ID: id}}

	if err := r.isiPemakaiTerakhir(data); err != nil {
		return nil, err
	}

	return data[0].PemakaiTerakhir, nil
}
