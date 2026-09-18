package models

import "time"

type LogPeninjauanPeralatan struct {
	IDLog uint64 `gorm:"column:id_log;primaryKey;autoIncrement" json:"id_log"`

	IDPeralatan uint64 `gorm:"column:id_peralatan;not null;index" json:"id_peralatan"`

	IDVerifikasi uint64 `gorm:"column:id_verifikasi;not null;index" json:"id_verifikasi"`

	IDManager uint64 `gorm:"column:id_manager;not null;index" json:"id_manager"`

	Status string `gorm:"column:status;type:varchar(50);not null" json:"status"`

	Alasan string `gorm:"column:alasan;type:text;not null" json:"alasan"`

	Catatan string `gorm:"column:catatan;type:text" json:"catatan"`

	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`

	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`

	DeletedAt *time.Time `gorm:"column:deleted_at" json:"deleted_at,omitempty"`

	Peralatan *Peralatan `gorm:"foreignKey:IDPeralatan;references:ID" json:"peralatan,omitempty"`

	Verifikasi *Verifikasi `gorm:"foreignKey:IDVerifikasi;references:IDVerifikasi" json:"verifikasi,omitempty"`

	Manager *User `gorm:"foreignKey:IDManager;references:UserID" json:"manager,omitempty"`
}

func (LogPeninjauanPeralatan) TableName() string {
	return "log_peninjauan_peralatan"
}
