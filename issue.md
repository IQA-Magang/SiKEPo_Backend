# [TASK-01] Implementasi CRUD Master Data Peralatan

| Field        | Detail                                       |
| ------------ | -------------------------------------------- |
| **Assignee** | Junior IT / Junior Backend Developer         |
| **Priority** | 🔴 High                                     |
| **Module**   | Master Data / Inventory                      |
| **Framework**| Go Fiber v2 + GORM + MySQL                   |
| **Deadline** | _Disesuaikan oleh Lead_                      |
| **Label**    | `backend`, `crud`, `master-data`, `peralatan`|

---

## 📋 Deskripsi Tugas

Membangun **CRUD lengkap** untuk modul **Master Data Peralatan** pada sistem SiKEPo Backend.
Modul ini mencakup pengelolaan data peralatan laboratorium dengan **5 kategori**:

1. **Alat Ukur (Piranti Lunak)**
2. **Alat Bantu**
3. **Referensi Uji**
4. **Golden Sample**
5. **Komponen Pendukung**

Setiap peralatan memiliki relasi ke **Ruangan Asal**, **Ruangan Sekarang**, **PIC Penanggung Jawab**, dan **User yang input data**.

> **⚠️ Penting:** Ikuti pola arsitektur yang sudah ada (`models → repositories → controllers → routes`). Lihat modul `users` sebagai referensi.

---

## 🗄️ Tabel Database & DDL

### 1. Tabel `kategori_peralatan`

```sql
CREATE TABLE kategori_peralatan (
    id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    nama_kategori VARCHAR(100) NOT NULL UNIQUE,
    deskripsi     TEXT NULL,
    created_at    TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at    TIMESTAMP NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

**Seed data:**

```sql
INSERT INTO kategori_peralatan (nama_kategori, deskripsi) VALUES
('Alat Ukur (Piranti Lunak)', 'Software dan alat ukur berbasis digital'),
('Alat Bantu',                'Peralatan penunjang kegiatan pengujian'),
('Referensi Uji',            'Material atau standar referensi untuk pengujian'),
('Golden Sample',            'Sampel acuan/benchmark untuk kalibrasi'),
('Komponen Pendukung',       'Komponen pelengkap peralatan utama');
```

### 2. Tabel `ruangan`

```sql
CREATE TABLE ruangan (
    id           BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    nama_ruangan VARCHAR(150) NOT NULL UNIQUE,
    gedung       VARCHAR(100) NULL,
    lantai       VARCHAR(20)  NULL,
    pic_user_id  BIGINT UNSIGNED NULL,
    created_at   TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at   TIMESTAMP NULL,
    CONSTRAINT fk_ruangan_pic FOREIGN KEY (pic_user_id) REFERENCES users(user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 3. Tabel `master_peralatan` ⭐

```sql
CREATE TABLE master_peralatan (
    id                      BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    kategori_id             BIGINT UNSIGNED NOT NULL,
    ruangan_asal_id         BIGINT UNSIGNED NOT NULL,
    ruangan_sekarang_id     BIGINT UNSIGNED NOT NULL,
    pic_penanggung_jawab_id BIGINT UNSIGNED NOT NULL,
    kode_aset               VARCHAR(50)  NOT NULL UNIQUE,
    nama_alat               VARCHAR(150) NOT NULL,
    merk                    VARCHAR(100) NULL,
    model                   VARCHAR(100) NULL,
    nomor_seri              VARCHAR(100) NULL,
    firmware_version        VARCHAR(50)  NULL,
    frekuensi_penggunaan    ENUM('Rendah','Sedang','Tinggi') DEFAULT 'Sedang',
    jumlah                  INT UNSIGNED NOT NULL DEFAULT 1,
    status_kelayakan        ENUM('LAYAK','TIDAK_LAYAK','KARANTINA') NOT NULL DEFAULT 'KARANTINA',
    status_penggunaan       ENUM('TERSEDIA','DIPINJAM','DIPINDAHKAN','PERBAIKAN','PEMELIHARAAN') NOT NULL DEFAULT 'TERSEDIA',
    no_sertifikat_kalibrasi VARCHAR(100) NULL,
    tgl_kalibrasi_terakhir  DATE NULL,
    tgl_expired_kalibrasi   DATE NULL,
    input_by                BIGINT UNSIGNED NOT NULL,
    created_at              TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at              TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at              TIMESTAMP NULL,
    CONSTRAINT fk_peralatan_kategori        FOREIGN KEY (kategori_id)             REFERENCES kategori_peralatan(id),
    CONSTRAINT fk_peralatan_ruangan_asal    FOREIGN KEY (ruangan_asal_id)         REFERENCES ruangan(id),
    CONSTRAINT fk_peralatan_ruangan_sekarang FOREIGN KEY (ruangan_sekarang_id)    REFERENCES ruangan(id),
    CONSTRAINT fk_peralatan_pic             FOREIGN KEY (pic_penanggung_jawab_id) REFERENCES users(user_id),
    CONSTRAINT fk_peralatan_input_by        FOREIGN KEY (input_by)                REFERENCES users(user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 📊 ER Diagram (Teks)

```
┌────────────────────┐         ┌─────────────────┐
│ kategori_peralatan │         │     users       │
│────────────────────│         │─────────────────│
│ id (PK)            │         │ user_id (PK)    │
│ nama_kategori      │         │ name, nip, role │
└────────┬───────────┘         └──┬──────┬───┬───┘
         │ 1:N                    │      │   │
         ▼                        │      │   │
┌──────────────────────────────┐  │      │   │
│      master_peralatan        │  │      │   │
│──────────────────────────────│  │      │   │
│ id (PK)                      │  │      │   │
│ kode_aset (UNIQUE)           │  │      │   │
│ nama_alat                    │  │      │   │
│ kategori_id (FK) ────────────┤  │      │   │
│ ruangan_asal_id (FK) ───────┤──┤──┐   │   │
│ ruangan_sekarang_id (FK) ───┤──┤──┤   │   │
│ pic_penanggung_jawab_id (FK)─┤◄─┘  │   │   │
│ input_by (FK) ───────────────┤◄────┘   │   │
│ status_kelayakan             │         │   │
│ status_penggunaan            │         │   │
│ frekuensi_penggunaan         │         │   │
└──────────────────────────────┘         │   │
                                         │   │
                              ┌──────────┘   │
                              ▼              ▼
                 ┌─────────────────────┐
                 │      ruangan        │
                 │─────────────────────│
                 │ id (PK)             │
                 │ nama_ruangan        │
                 │ pic_user_id (FK) ───┤→ users
                 └─────────────────────┘
```

---

## 📁 Struktur File

```
SiKEPo_Backend/
├── models/
│   ├── user.go                          # ✅ Sudah ada
│   ├── kategori_peralatan.go            # 🆕
│   ├── ruangan.go                       # 🆕
│   └── peralatan.go                     # 🆕
├── repositories/
│   ├── user_repository.go               # ✅ Sudah ada (referensi pola)
│   ├── kategori_peralatan_repository.go # 🆕
│   ├── ruangan_repository.go            # 🆕
│   └── peralatan_repository.go          # 🆕
├── controllers/
│   ├── user_controller.go               # ✅ Sudah ada (referensi pola)
│   ├── kategori_peralatan_controller.go # 🆕
│   ├── ruangan_controller.go            # 🆕
│   └── peralatan_controller.go          # 🆕
├── routes/
│   ├── user_routes.go                   # ✅ Sudah ada
│   └── peralatan_routes.go              # 🆕
├── config/database.go                   # ✅ Update: tambahkan AutoMigrate
└── main.go                              # ✅ Update: tambahkan wiring DI
```

| Layer          | Tanggung Jawab                                                          |
| -------------- | ----------------------------------------------------------------------- |
| **Model**      | Struct Go + tag GORM & JSON. Mapping langsung ke tabel MySQL.            |
| **Repository** | Query database (CRUD). Tidak ada logic HTTP.                            |
| **Controller** | Handle request/response, validasi input, panggil repository.            |
| **Routes**     | Registrasi endpoint ke Fiber router + middleware.                       |

---

## 📐 Spesifikasi Model Go

### `models/kategori_peralatan.go`

```go
package models

import (
    "time"
    "gorm.io/gorm"
)

type KategoriPeralatan struct {
    ID            uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
    NamaKategori  string         `gorm:"column:nama_kategori;size:100;not null;unique" json:"nama_kategori"`
    Deskripsi     string         `gorm:"column:deskripsi;type:text" json:"deskripsi"`
    CreatedAt     time.Time      `gorm:"column:created_at" json:"created_at"`
    UpdatedAt     time.Time      `gorm:"column:updated_at" json:"updated_at"`
    DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at" json:"-"`
}

func (KategoriPeralatan) TableName() string {
    return "kategori_peralatan"
}
```

### `models/ruangan.go`

```go
package models

import (
    "time"
    "gorm.io/gorm"
)

type Ruangan struct {
    ID          uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
    NamaRuangan string         `gorm:"column:nama_ruangan;size:150;not null;unique" json:"nama_ruangan"`
    Gedung      string         `gorm:"column:gedung;size:100" json:"gedung"`
    Lantai      string         `gorm:"column:lantai;size:20" json:"lantai"`
    PICUserID   *uint64        `gorm:"column:pic_user_id" json:"pic_user_id"`
    CreatedAt   time.Time      `gorm:"column:created_at" json:"created_at"`
    UpdatedAt   time.Time      `gorm:"column:updated_at" json:"updated_at"`
    DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at" json:"-"`

    // Relasi
    PICUser *User `gorm:"foreignKey:PICUserID;references:UserID" json:"pic_user,omitempty"`
}

func (Ruangan) TableName() string {
    return "ruangan"
}
```

### `models/peralatan.go`

```go
package models

import (
    "time"
    "gorm.io/gorm"
)

type MasterPeralatan struct {
    ID                    uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
    KategoriID            uint64         `gorm:"column:kategori_id;not null" json:"kategori_id"`
    RuanganAsalID         uint64         `gorm:"column:ruangan_asal_id;not null" json:"ruangan_asal_id"`
    RuanganSekarangID     uint64         `gorm:"column:ruangan_sekarang_id;not null" json:"ruangan_sekarang_id"`
    PICPenanggungJawabID  uint64         `gorm:"column:pic_penanggung_jawab_id;not null" json:"pic_penanggung_jawab_id"`
    KodeAset              string         `gorm:"column:kode_aset;size:50;not null;unique" json:"kode_aset"`
    NamaAlat              string         `gorm:"column:nama_alat;size:150;not null" json:"nama_alat"`
    Merk                  string         `gorm:"column:merk;size:100" json:"merk"`
    Model                 string         `gorm:"column:model;size:100" json:"model"`
    NomorSeri             string         `gorm:"column:nomor_seri;size:100" json:"nomor_seri"`
    FirmwareVersion       string         `gorm:"column:firmware_version;size:50" json:"firmware_version"`
    FrekuensiPenggunaan   string         `gorm:"column:frekuensi_penggunaan;type:enum('Rendah','Sedang','Tinggi');default:'Sedang'" json:"frekuensi_penggunaan"`
    Jumlah                uint           `gorm:"column:jumlah;not null;default:1" json:"jumlah"`
    StatusKelayakan       string         `gorm:"column:status_kelayakan;type:enum('LAYAK','TIDAK_LAYAK','KARANTINA');not null;default:'KARANTINA'" json:"status_kelayakan"`
    StatusPenggunaan      string         `gorm:"column:status_penggunaan;type:enum('TERSEDIA','DIPINJAM','DIPINDAHKAN','PERBAIKAN','PEMELIHARAAN');not null;default:'TERSEDIA'" json:"status_penggunaan"`
    NoSertifikatKalibrasi string         `gorm:"column:no_sertifikat_kalibrasi;size:100" json:"no_sertifikat_kalibrasi"`
    TglKalibrasiTerakhir  *time.Time     `gorm:"column:tgl_kalibrasi_terakhir;type:date" json:"tgl_kalibrasi_terakhir"`
    TglExpiredKalibrasi   *time.Time     `gorm:"column:tgl_expired_kalibrasi;type:date" json:"tgl_expired_kalibrasi"`
    InputBy               uint64         `gorm:"column:input_by;not null" json:"input_by"`
    CreatedAt             time.Time      `gorm:"column:created_at" json:"created_at"`
    UpdatedAt             time.Time      `gorm:"column:updated_at" json:"updated_at"`
    DeletedAt             gorm.DeletedAt `gorm:"column:deleted_at" json:"-"`

    // Relasi (Preload)
    Kategori          *KategoriPeralatan `gorm:"foreignKey:KategoriID;references:ID" json:"kategori,omitempty"`
    RuanganAsal       *Ruangan           `gorm:"foreignKey:RuanganAsalID;references:ID" json:"ruangan_asal,omitempty"`
    RuanganSekarang   *Ruangan           `gorm:"foreignKey:RuanganSekarangID;references:ID" json:"ruangan_sekarang,omitempty"`
    PICPenanggungJawab *User             `gorm:"foreignKey:PICPenanggungJawabID;references:UserID" json:"pic_penanggung_jawab,omitempty"`
    InputByUser       *User              `gorm:"foreignKey:InputBy;references:UserID" json:"input_by_user,omitempty"`
}

func (MasterPeralatan) TableName() string {
    return "peralatan"
}
```

---

## 🔌 Spesifikasi API Endpoints

**Base URL:** `/api/v1/peralatan`

| # | Method  | Endpoint                          | Deskripsi                    | Auth |
|---|---------|-----------------------------------|------------------------------|------|
| 1 | `POST`  | `/api/v1/peralatan`               | Create peralatan baru        | ✅   |
| 2 | `GET`   | `/api/v1/peralatan`               | List (paginasi+filter)       | ✅   |
| 3 | `GET`   | `/api/v1/peralatan/:id`           | Detail peralatan             | ✅   |
| 4 | `PUT`   | `/api/v1/peralatan/:id`           | Update peralatan             | ✅   |
| 5 | `PATCH` | `/api/v1/peralatan/:id/lokasi`    | Pemindahan ruangan           | ✅   |
| 6 | `DELETE`| `/api/v1/peralatan/:id`           | Soft delete                  | ✅   |

---

### 1. `POST /api/v1/peralatan` — Create

**Request Body:**

```json
{
    "kategori_id": 1,
    "ruangan_asal_id": 2,
    "ruangan_sekarang_id": 2,
    "pic_penanggung_jawab_id": null,
    "kode_aset": "ALT-UKR-001",
    "nama_alat": "Spectrum Analyzer",
    "merk": "Keysight",
    "model": "N9000B",
    "nomor_seri": "SN-2024-001",
    "firmware_version": "v3.2.1",
    "frekuensi_penggunaan": "Tinggi",
    "jumlah": 1,
    "no_sertifikat_kalibrasi": "CERT-KAL-2024-001",
    "tgl_kalibrasi_terakhir": "2024-06-15",
    "tgl_expired_kalibrasi": "2025-06-15"
}
```

**Business Logic:**

1. Validasi `kode_aset` wajib & **unik** (cek di database).
2. Validasi `nama_alat` wajib.
3. Validasi FK: `kategori_id` harus ada di `kategori_peralatan`, `ruangan_asal_id` & `ruangan_sekarang_id` harus ada di `ruangan`.
4. **`status_kelayakan` default = `KARANTINA`** — di-set server, bukan client.
5. **`status_penggunaan` default = `TERSEDIA`** — di-set server.
6. **Auto-assign PIC:** Jika `pic_penanggung_jawab_id` **kosong/null**, ambil `pic_user_id` dari tabel `ruangan` (berdasarkan `ruangan_sekarang_id`).
7. **`input_by`** diambil dari **JWT token** user yang login (bukan dari request body).
8. Validasi `frekuensi_penggunaan` ∈ `{Rendah, Sedang, Tinggi}`.
9. Jika `tgl_expired_kalibrasi < tgl_kalibrasi_terakhir` → return error.

**Response Sukses (201):**

```json
{
    "success": true,
    "message": "Peralatan berhasil ditambahkan",
    "data": {
        "id": 1,
        "kode_aset": "ALT-UKR-001",
        "nama_alat": "Spectrum Analyzer",
        "status_kelayakan": "KARANTINA",
        "status_penggunaan": "TERSEDIA",
        "kategori": { "id": 1, "nama_kategori": "Alat Ukur (Piranti Lunak)" },
        "ruangan_asal": { "id": 2, "nama_ruangan": "Lab Kalibrasi A" },
        "ruangan_sekarang": { "id": 2, "nama_ruangan": "Lab Kalibrasi A" },
        "pic_penanggung_jawab": { "user_id": 5, "name": "Budi Santoso" },
        "input_by_user": { "user_id": 3, "name": "Admin SiKEPo" }
    }
}
```

**Response Error (400):**

```json
{
    "success": false,
    "message": "Validasi gagal",
    "errors": { "kode_aset": "Kode aset sudah digunakan" }
}
```

---

### 2. `GET /api/v1/peralatan` — List All (Pagination + Filter)

**Query Parameters:**

| Parameter           | Tipe   | Default      | Keterangan                                      |
|---------------------|--------|--------------|--------------------------------------------------|
| `page`              | int    | `1`          | Halaman ke-                                      |
| `limit`             | int    | `10`         | Jumlah per halaman (max: 100)                    |
| `search`            | string | `""`         | Cari di `nama_alat`, `kode_aset`, `merk`         |
| `kategori_id`       | int    | -            | Filter kategori                                  |
| `ruangan_id`        | int    | -            | Filter ruangan sekarang                          |
| `status_kelayakan`  | string | -            | Filter: LAYAK / TIDAK_LAYAK / KARANTINA          |
| `status_penggunaan` | string | -            | Filter: TERSEDIA / DIPINJAM / dst                |
| `sort_by`           | string | `created_at` | Kolom sorting                                    |
| `sort_order`        | string | `desc`       | `asc` / `desc`                                   |

**Business Logic:**

1. Soft delete aware (GORM otomatis filter `deleted_at IS NULL`).
2. Search: `LIKE %keyword%` pada `nama_alat`, `kode_aset`, `merk`.
3. **Preload** relasi: `Kategori`, `RuanganAsal`, `RuanganSekarang`, `PICPenanggungJawab`, `InputByUser`.
4. Hitung `total_data` dan `total_pages`.
5. Jika `limit > 100`, paksa `limit = 100`.

**Response (200):**

```json
{
    "success": true,
    "message": "Data peralatan berhasil diambil",
    "data": [ ... ],
    "meta": { "page": 1, "limit": 10, "total_data": 57, "total_pages": 6 }
}
```

---

### 3. `GET /api/v1/peralatan/:id` — Detail

- Cari by `id`, **Preload** semua relasi. Return `404` jika tidak ditemukan.

---

### 4. `PUT /api/v1/peralatan/:id` — Update

**Request Body:** Sama seperti POST (tanpa `input_by`).

**Business Logic:**

1. Cari by ID → `404` jika tidak ada.
2. Validasi `kode_aset` unik (exclude ID saat ini).
3. Validasi semua FK valid.
4. Validasi `status_kelayakan` ∈ `{LAYAK, TIDAK_LAYAK, KARANTINA}`.
5. Validasi `status_penggunaan` ∈ `{TERSEDIA, DIPINJAM, DIPINDAHKAN, PERBAIKAN, PEMELIHARAAN}`.
6. Validasi `frekuensi_penggunaan` ∈ `{Rendah, Sedang, Tinggi}`.
7. Validasi tanggal kalibrasi (expired ≥ terakhir).

---

### 5. `PATCH /api/v1/peralatan/:id/lokasi` — Pemindahan Ruangan

**Request Body:**

```json
{
    "ruangan_sekarang_id": 5,
    "pic_penanggung_jawab_id": null
}
```

**Business Logic:**

1. Cari by ID → `404` jika tidak ada.
2. Validasi `ruangan_sekarang_id` valid.
3. **Auto-assign PIC:** Jika `pic_penanggung_jawab_id` kosong, set dari `ruangan.pic_user_id`.
4. Set `status_penggunaan = "DIPINDAHKAN"`.
5. Update: `ruangan_sekarang_id`, `pic_penanggung_jawab_id`, `status_penggunaan`.

---

### 6. `DELETE /api/v1/peralatan/:id` — Soft Delete

- Cari by ID → `404`, lalu GORM soft delete (`deleted_at`).

---

## 🔧 Panduan Implementasi

### Step 1: Buat Model Files

- [ ] `models/kategori_peralatan.go`, `models/ruangan.go`, `models/peralatan.go`
- [ ] Semua struct punya `TableName()`, tag `gorm:` & `json:`, pointer untuk nullable

### Step 2: Update `config/database.go`

```go
database.AutoMigrate(
    &models.KategoriPeralatan{},
    &models.Ruangan{},
    &models.MasterPeralatan{},
)
```

### Step 3: Buat Repository

**`repositories/peralatan_repository.go`:**

```go
type PeralatanRepository struct { DB *gorm.DB }

func NewPeralatanRepository(db *gorm.DB) *PeralatanRepository
func (r *PeralatanRepository) Create(p *models.MasterPeralatan) error
func (r *PeralatanRepository) FindAll(params PeralatanQueryParams) ([]models.MasterPeralatan, int64, error)
func (r *PeralatanRepository) FindByID(id uint64) (*models.MasterPeralatan, error)
func (r *PeralatanRepository) Update(id uint64, p *models.MasterPeralatan) error
func (r *PeralatanRepository) UpdateRuangan(id, ruanganID, picID uint64) error
func (r *PeralatanRepository) Delete(id uint64) error
func (r *PeralatanRepository) IsKodeAsetExists(kode string, excludeID ...uint64) (bool, error)
```

```go
type PeralatanQueryParams struct {
    Page, Limit                          int
    Search                               string
    KategoriID, RuanganID                uint64
    StatusKelayakan, StatusPenggunaan    string
    SortBy, SortOrder                    string
}
```

### Step 4: Buat Controller

```go
type PeralatanController struct {
    Repository        *repositories.PeralatanRepository
    RuanganRepository *repositories.RuanganRepository
    KategoriRepository *repositories.KategoriPeralatanRepository
}
```

**Validasi helper:**

```go
func isValidStatusKelayakan(s string) bool {
    for _, v := range []string{"LAYAK", "TIDAK_LAYAK", "KARANTINA"} {
        if v == s { return true }
    }
    return false
}

func isValidStatusPenggunaan(s string) bool {
    for _, v := range []string{"TERSEDIA", "DIPINJAM", "DIPINDAHKAN", "PERBAIKAN", "PEMELIHARAAN"} {
        if v == s { return true }
    }
    return false
}

func isValidFrekuensi(f string) bool {
    for _, v := range []string{"Rendah", "Sedang", "Tinggi"} {
        if v == f { return true }
    }
    return false
}
```

### Step 5: Registrasi Routes (`routes/peralatan_routes.go`)

```go
func PeralatanRoutes(app *fiber.App, ctrl *controllers.PeralatanController) {
    api := app.Group("/api/v1/peralatan", middleware.RequireAuth)
    api.Post("/", ctrl.Create)
    api.Get("/", ctrl.GetAll)
    api.Get("/:id", ctrl.GetByID)
    api.Put("/:id", ctrl.Update)
    api.Patch("/:id/lokasi", ctrl.UpdateRuangan)
    api.Delete("/:id", ctrl.Delete)
}
```

### Step 6: Wiring di `main.go`

```go
peralatanRepo := repositories.NewPeralatanRepository(config.DB)
ruanganRepo := repositories.NewRuanganRepository(config.DB)
kategoriRepo := repositories.NewKategoriPeralatanRepository(config.DB)

peralatanCtrl := &controllers.PeralatanController{
    Repository:         peralatanRepo,
    RuanganRepository:  ruanganRepo,
    KategoriRepository: kategoriRepo,
}

routes.PeralatanRoutes(app, peralatanCtrl)
```

---

## ✅ Acceptance Criteria

### Fungsional

- [ ] **AC-01:** POST berhasil create dengan `status_kelayakan=KARANTINA`, `status_penggunaan=TERSEDIA`.
- [ ] **AC-02:** POST tanpa `pic_penanggung_jawab_id` → PIC otomatis dari `ruangan.pic_user_id`.
- [ ] **AC-03:** POST dengan `kode_aset` duplikat → error `400`.
- [ ] **AC-04:** POST dengan `kategori_id` invalid → error `400`.
- [ ] **AC-05:** `input_by` otomatis dari JWT token user yang login.
- [ ] **AC-06:** GET list dengan pagination benar (`meta.total_data`, `meta.total_pages`).
- [ ] **AC-07:** GET list `?search=spectrum` → filter berdasarkan nama/kode/merk.
- [ ] **AC-08:** GET list `?kategori_id=1&status_kelayakan=LAYAK` → filter gabungan.
- [ ] **AC-09:** GET detail mengembalikan semua relasi (kategori, ruangan asal, ruangan sekarang, PIC, input_by).
- [ ] **AC-10:** GET dengan ID tidak ada → `404`.
- [ ] **AC-11:** PUT berhasil update, validasi kode_aset unik (exclude self).
- [ ] **AC-12:** PATCH lokasi → `ruangan_sekarang_id` berubah, `status_penggunaan=DIPINDAHKAN`, auto-assign PIC.
- [ ] **AC-13:** DELETE → soft delete, data hilang dari list tapi masih ada di DB.
- [ ] **AC-14:** Semua response format konsisten (`success`, `message`, `data`).

### Non-Fungsional

- [ ] **AC-15:** Semua endpoint dilindungi `RequireAuth` middleware.
- [ ] **AC-16:** Menggunakan GORM query builder, bukan raw SQL.
- [ ] **AC-17:** Mengikuti pola arsitektur modul `users`.
- [ ] **AC-18:** Pesan error dalam Bahasa Indonesia.

---

## 🧪 Panduan Testing

### A. Unit Test (`go test`)

```bash
go test ./... -v           # Test semua
go test ./controllers/... -v  # Test controller saja
go test ./... -v -cover    # Dengan coverage
```

**Test cases wajib:**

| Test Case                           | Yang Diverifikasi                                    |
|-------------------------------------|------------------------------------------------------|
| `TestCreate_Success`                | Status 201, status_kelayakan=KARANTINA               |
| `TestCreate_DuplicateKodeAset`      | Status 400, error kode_aset                          |
| `TestCreate_AutoAssignPIC`          | PIC terisi otomatis dari ruangan                     |
| `TestCreate_InvalidKategori`        | Status 400, kategori tidak ditemukan                  |
| `TestGetAll_Pagination`             | meta.total_data & total_pages benar                   |
| `TestGetAll_SearchFilter`           | Filter search + kategori + status                     |
| `TestGetByID_NotFound`              | Status 404                                            |
| `TestUpdate_Success`                | Data terupdate, validasi kode_aset unik               |
| `TestUpdateRuangan_AutoPIC`         | Ruangan berubah, PIC auto-assign, status=DIPINDAHKAN  |
| `TestDelete_SoftDelete`             | deleted_at terisi, data hilang dari list               |

### B. API Test via cURL

```bash
# 1. Create
curl -X POST http://localhost:5000/api/v1/peralatan \
  -H "Content-Type: application/json" -H "Authorization: Bearer <TOKEN>" \
  -d '{"kategori_id":1,"ruangan_asal_id":2,"ruangan_sekarang_id":2,"kode_aset":"ALT-UKR-001","nama_alat":"Spectrum Analyzer","merk":"Keysight","model":"N9000B","nomor_seri":"SN-001","jumlah":1}'

# 2. List + filter
curl "http://localhost:5000/api/v1/peralatan?page=1&limit=5&search=spectrum&status_kelayakan=KARANTINA" \
  -H "Authorization: Bearer <TOKEN>"

# 3. Detail
curl http://localhost:5000/api/v1/peralatan/1 -H "Authorization: Bearer <TOKEN>"

# 4. Update
curl -X PUT http://localhost:5000/api/v1/peralatan/1 \
  -H "Content-Type: application/json" -H "Authorization: Bearer <TOKEN>" \
  -d '{"kategori_id":1,"ruangan_asal_id":2,"ruangan_sekarang_id":2,"pic_penanggung_jawab_id":5,"kode_aset":"ALT-UKR-001","nama_alat":"Spectrum Analyzer (Updated)","status_kelayakan":"LAYAK","status_penggunaan":"TERSEDIA","jumlah":1}'

# 5. Pindah ruangan
curl -X PATCH http://localhost:5000/api/v1/peralatan/1/lokasi \
  -H "Content-Type: application/json" -H "Authorization: Bearer <TOKEN>" \
  -d '{"ruangan_sekarang_id":5}'

# 6. Soft delete
curl -X DELETE http://localhost:5000/api/v1/peralatan/1 -H "Authorization: Bearer <TOKEN>"
```

---

## 📌 Catatan untuk Developer

1. **Jangan ubah file existing** kecuali `config/database.go` dan `main.go` untuk wiring.
2. **Konvensi:** Pesan dalam Bahasa Indonesia, nama variabel/fungsi dalam Bahasa Inggris.
3. **Selalu Preload** relasi saat return data peralatan.
4. **Commit format:** `feat(peralatan): <deskripsi>` — contoh: `feat(peralatan): add CRUD endpoints`.


> 
