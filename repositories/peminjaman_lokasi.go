package repositories

import (
	"strings"

	"backend/models"
)

// isiLokasiAsal mengisi lokasi asal peralatan pada setiap peminjaman, yaitu
// ruangan tempat alat berada dengan format yang sama seperti pilihan ruangan di
// form pengajuan: kode ruangan, nama ruangan, lalu nama lab. Peralatan harus
// sudah dimuat. Semua ruangan diambil dengan satu query.
func (r *PeminjamanRepository) isiLokasiAsal(data []*models.Peminjaman) error {
	ids := make([]uint64, 0, len(data))
	dilihat := map[uint64]bool{}

	for _, p := range data {
		if p.Peralatan == nil || p.Peralatan.RuanganID == 0 {
			continue
		}

		id := uint64(p.Peralatan.RuanganID)
		if !dilihat[id] {
			dilihat[id] = true
			ids = append(ids, id)
		}
	}

	if len(ids) == 0 {
		return nil
	}

	var ruangan []models.Ruangan
	if err := r.DB.Preload("Labs").Where("id IN ?", ids).Find(&ruangan).Error; err != nil {
		return err
	}

	lokasi := make(map[uint64]string, len(ruangan))

	for _, rg := range ruangan {
		bagian := []string{}

		for _, b := range []string{rg.KodeRuangan, rg.NamaRuangan} {
			if b != "" {
				bagian = append(bagian, b)
			}
		}

		if rg.Labs != nil && rg.Labs.NamaLabs != "" {
			bagian = append(bagian, rg.Labs.NamaLabs)
		}

		lokasi[rg.ID] = strings.Join(bagian, " — ")
	}

	for _, p := range data {
		if p.Peralatan != nil {
			p.LokasiAsalPeralatan = lokasi[uint64(p.Peralatan.RuanganID)]
		}
	}

	return nil
}
