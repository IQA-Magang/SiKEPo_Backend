package config

import (
	"fmt"
	"log"
	"strings"

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

	if err := db.AutoMigrate(
		&models.Peminjaman{},
		&models.SerahTerima{},
		&models.SerahTerimaItem{},
	); err != nil {
		return fmt.Errorf("failed to migrate peminjaman tables: %w", err)
	}

	return ensureStatusAlatDoNotUse(db)
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

// ensureStatusAlatDoNotUse menambahkan nilai "Do Not Use" ke enum
// peralatan.status_alat pada database yang tabelnya sudah terlanjur dibuat.
// Nilai itu dipakai saat serah terima menemukan butir TS.
func ensureStatusAlatDoNotUse(db *gorm.DB) error {
	var columnType string

	err := db.Raw(
		`SELECT COLUMN_TYPE FROM information_schema.COLUMNS
		 WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?`,
		"peralatan",
		"status_alat",
	).Scan(&columnType).Error
	if err != nil {
		return fmt.Errorf("failed to read peralatan.status_alat type: %w", err)
	}

	if columnType == "" || strings.Contains(columnType, "'Do Not Use'") {
		return nil
	}

	if err := db.Migrator().AlterColumn(&models.Peralatan{}, "StatusAlat"); err != nil {
		return fmt.Errorf("failed to add 'Do Not Use' to peralatan.status_alat: %w", err)
	}

	log.Println("Enum peralatan.status_alat diperbarui dengan nilai 'Do Not Use'")

	return nil
}
