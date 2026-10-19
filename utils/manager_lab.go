package utils

import (
	"errors"

	"backend/models"

	"gorm.io/gorm"
)

var (
	ErrRuanganPeralatanTidakDitemukan = errors.New("ruangan peralatan tidak ditemukan")
	ErrRuanganTanpaLab                = errors.New("ruangan peralatan belum terhubung ke lab")
	ErrLabTidakDitemukan              = errors.New("lab tidak ditemukan")
	ErrLabTanpaManager                = errors.New("lab belum memiliki manager")
)

// IsErrLab bernilai true untuk error data master, yaitu ruangan, lab, atau
// manager yang belum lengkap. Error seperti ini dikembalikan ke user sebagai 400.
func IsErrLab(err error) bool {
	return errors.Is(err, ErrRuanganPeralatanTidakDitemukan) ||
		errors.Is(err, ErrRuanganTanpaLab) ||
		errors.Is(err, ErrLabTidakDitemukan) ||
		errors.Is(err, ErrLabTanpaManager)
}

// GetLabDenganManager mengambil lab berdasarkan ID dan memastikan lab itu
// punya manager, sehingga pemanggil bisa langsung memakai *lab.ManagerID.
func GetLabDenganManager(db *gorm.DB, labID uint64) (*models.Labs, error) {
	var lab models.Labs

	if err := db.Where("id = ?", labID).First(&lab).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrLabTidakDitemukan
		}
		return nil, err
	}

	if lab.ManagerID == nil || *lab.ManagerID == 0 {
		return nil, ErrLabTanpaManager
	}

	return &lab, nil
}

// GetLabPemilikPeralatan mencari lab pemilik peralatan melalui ruangan tempat
// peralatan berada: peralatan, ruangan, lalu lab. Lab yang dikembalikan
// dipastikan punya manager.
func GetLabPemilikPeralatan(db *gorm.DB, peralatan *models.Peralatan) (*models.Labs, error) {
	if peralatan == nil {
		return nil, ErrRuanganPeralatanTidakDitemukan
	}

	var ruangan models.Ruangan

	if err := db.Where("id = ?", peralatan.RuanganID).First(&ruangan).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRuanganPeralatanTidakDitemukan
		}
		return nil, err
	}

	if ruangan.LabsID == nil || *ruangan.LabsID == 0 {
		return nil, ErrRuanganTanpaLab
	}

	return GetLabDenganManager(db, *ruangan.LabsID)
}
