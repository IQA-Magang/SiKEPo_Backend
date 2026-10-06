package config

import (
	"fmt"

	"backend/models"

	"gorm.io/gorm"
)

// MigratePeminjaman menyiapkan struktur database untuk modul peminjaman.
//
// Sengaja dibuat terpisah dari AutoMigrate utama, yang hanya berjalan bila ada
// tabel yang belum ada. Dengan begitu tabel lain pada database yang sudah
// berjalan tidak ikut disentuh. Fungsi ini aman dijalankan berulang.
func MigratePeminjaman(db *gorm.DB) error {
	if err := ensureUserLabsColumn(db); err != nil {
		return err
	}

	if err := db.AutoMigrate(&models.Peminjaman{}); err != nil {
		return fmt.Errorf("failed to migrate peminjaman table: %w", err)
	}

	return nil
}

// ensureUserLabsColumn menambahkan kolom users.labs_id bila belum ada.
// Hanya kolom itu yang ditambahkan, tabel users tidak diubah dengan cara lain.
func ensureUserLabsColumn(db *gorm.DB) error {
	if db.Migrator().HasColumn(&models.User{}, "LabsID") {
		return nil
	}

	if err := db.Migrator().AddColumn(&models.User{}, "LabsID"); err != nil {
		return fmt.Errorf("failed to add users.labs_id column: %w", err)
	}

	return nil
}
