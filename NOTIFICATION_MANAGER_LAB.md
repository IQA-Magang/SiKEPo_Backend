# Notifikasi Umum untuk PIC dan Manager Lab

Dokumen ini menjelaskan alur notifikasi yang dipakai saat ini pada backend SiKEPo.

## Tujuan

Notifikasi dibuat secara umum agar reusable untuk berbagai event, bukan hanya untuk manager lab. Pada flow saat ini:

1. Saat peralatan baru dibuat, sistem mengirim notifikasi ke PIC peralatan.
2. Saat status verifikasi peralatan berubah, sistem mengirim notifikasi ke manager lab yang terkait dengan ruangan/lab.

Dengan pola ini, notifikasi tidak lagi hanya dibuat khusus untuk satu role saja.

---

## Alur kerja saat ini

### 1. Saat membuat peralatan

1. User mengirim request create peralatan ke endpoint `/api/peralatan`.
2. Sistem menyimpan data peralatan ke database.
3. Sistem mengambil `PICID` dari peralatan.
4. Sistem memanggil utility notifikasi umum `NotifyNewPeralatanCreated`.
5. Notifikasi masuk ke user PIC terkait dengan tipe `peralatan_created`.

### 2. Saat status verifikasi berubah

1. Setelah verifikasi diproses atau hasil verifikasi diperbarui, handler verifikasi memanggil helper `NotifyManagerOnVerificationStatusChange`.
2. Sistem mengambil ruangan dari `ruangan_id` pada peralatan.
3. Sistem mencari `manager_id` dari lab terkait.
4. Sistem mengirim notifikasi ke manager dengan tipe `peralatan_verification_updated`.

---

## Utility notifikasi umum

Utility dibuat di [utils/notifications.go](utils/notifications.go).

Fungsi utama:

```go
func SendNotification(db *gorm.DB, userID uint64, notificationType, title, message string) error
func SendBulkNotification(db *gorm.DB, userIDs []uint64, notificationType, title, message string) error
func NotifyNewPeralatanCreated(db *gorm.DB, peralatan *models.Peralatan) error
func NotifyManagerOnVerificationStatusChange(db *gorm.DB, peralatan *models.Peralatan, status string) error
```

Contoh notifikasi untuk PIC:

```go
if err := utils.NotifyNewPeralatanCreated(db, peralatan); err != nil {
    log.Printf("Gagal mengirim notifikasi ke PIC: %v", err)
}
```

Contoh notifikasi untuk manager:

```go
if err := utils.NotifyManagerOnVerificationStatusChange(db, peralatan, "Layak"); err != nil {
    log.Printf("Gagal mengirim notifikasi ke manager: %v", err)
}
```

---

## Struktur data notifikasi

Model notifikasi masih menggunakan struktur berikut di [models/notification.go](models/notification.go):

```go
type Notification struct {
    ID        uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
    UserID    uint64         `gorm:"column:user_id;not null;index" json:"user_id"`
    Type      string         `gorm:"column:type;size:50;not null;default:'equipment_added'" json:"type"`
    Title     string         `gorm:"column:title;size:150;not null" json:"title"`
    Message   string         `gorm:"column:message;type:text;not null" json:"message"`
    IsRead    bool           `gorm:"column:is_read;default:false" json:"is_read"`
    CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
    UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"column:deleted_at" json:"-"`
}
```

Field penting:
- `user_id`: user penerima notifikasi
- `type`: tipe event, contoh `peralatan_created` atau `peralatan_verification_updated`
- `title`: judul ringkas notifikasi
- `message`: isi detail notifikasi
- `is_read`: status sudah dibaca atau belum

---

## Endpoint notifikasi yang tersedia saat ini

Endpoint untuk membaca notifikasi berada di [routes/notification_routes.go](routes/notification_routes.go).

### 1. Ambil daftar notifikasi user

```http
GET /api/notifications/user/:user_id
```

Header:

```http
Authorization: Bearer <token>
Content-Type: application/json
```

Contoh:

```bash
curl http://localhost:5000/api/notifications/user/2 \
  -H "Authorization: Bearer <token>"
```

Contoh response:

```json
{
  "status": "success",
  "data": [
    {
      "id": 1,
      "user_id": 2,
      "type": "peralatan_created",
      "title": "Peralatan baru ditambahkan",
      "message": "Peralatan Multimeter Digital (AST-001) telah ditambahkan dan menunggu proses verifikasi.",
      "is_read": false,
      "created_at": "2026-09-18T12:00:00Z"
    }
  ],
  "count": 1
}
```

### 2. Tandai notifikasi sudah dibaca

```http
PATCH /api/notifications/:id/read
```

Contoh:

```bash
curl -X PATCH http://localhost:5000/api/notifications/1/read \
  -H "Authorization: Bearer <token>"
```

Response:

```json
{
  "status": "success",
  "message": "Notifikasi berhasil ditandai dibaca"
}
```

> Catatan: route saat ini menggunakan middleware `RequireRoles("manager")`, sehingga endpoint baca notifikasi sedang difokuskan untuk role manager. Untuk PIC, utility tetap bisa dipakai dan record notifikasi akan tersimpan di database; jika nanti diperlukan, route dapat dibuka untuk semua role dengan validasi `user_id`.

---

## Test API flow

Berikut contoh pengujian end-to-end untuk flow notifikasi.

### A. Login sebagai PIC atau staff

```bash
curl -X POST http://localhost:5000/api/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "staff@sikepo.local",
    "password": "password123"
  }'
```

Ambil token dari response lalu gunakan untuk membuat peralatan:

```bash
curl -X POST http://localhost:5000/api/peralatan \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "nama_peralatan": "Multimeter Digital",
    "kategori_id": 1,
    "kelompok_aset_id": 1,
    "ruangan_id": 1,
    "pic_id": 3,
    "merek": "Fluke",
    "tipe_model": "87V",
    "nomor_seri": "FLK-001",
    "status_alat": "Aktif",
    "keterangan": "Untuk pengukuran listrik",
    "detail": {
      "peranti_lunak_versi": "1.0.0",
      "metode_kelayakan": "kalibrasi internal",
      "no_sertifikat": "SRT-001",
      "tgl_kalibrasi": "2026-09-01T00:00:00Z",
      "interval_bulan": 6,
      "fungsi_sbg_alat_standar": false,
      "jenis_label": "calibration",
      "status_kelayakan": "Layak"
    }
  }'
```

Kondisi yang diharapkan:
- record peralatan baru berhasil dibuat
- notifikasi dengan tipe `peralatan_created` masuk ke user PIC (`pic_id`)

### B. Cek notifikasi PIC

```bash
curl http://localhost:5000/api/notifications/user/3 \
  -H "Authorization: Bearer <token_manager>"
```

Jika route diubah menjadi user-aware nanti, hasilnya akan terlihat sesuai `user_id` PIC. Saat ini, endpoint default mengikuti konfigurasi manager-only.

### C. Simulasi perubahan status verifikasi

Pada handler verifikasi, panggil helper seperti ini:

```go
c.NotifyVerificationStatusChanged(peralatan, "Layak")
```

Atau lewat utility langsung:

```go
if err := utils.NotifyManagerOnVerificationStatusChange(db, peralatan, "Layak"); err != nil {
    log.Printf("Error: %v", err)
}
```

Kondisi yang diharapkan:
- notifikasi baru muncul untuk manager lab dari ruangan tersebut
- tipe notifikasi: `peralatan_verification_updated`
- message berisi perubahan status yang terjadi

### D. Cek notifikasi manager

```bash
curl http://localhost:5000/api/notifications/user/2 \
  -H "Authorization: Bearer <token_manager>"
```

Response contoh:

```json
{
  "status": "success",
  "data": [
    {
      "id": 10,
      "user_id": 2,
      "type": "peralatan_verification_updated",
      "title": "Status verifikasi peralatan berubah",
      "message": "Status verifikasi peralatan Multimeter Digital (AST-001) berubah menjadi Layak.",
      "is_read": false
    }
  ],
  "count": 1
}
```

---

## Contoh skenario nyata

Skenario:

- User PIC dengan ID 3 menambah peralatan baru
- Manager lab dengan ID 2 ada di lab yang sama
- Setelah status verifikasi di-update menjadi `Layak`, manager menerima notifikasi

Hasil yang diinginkan:

- PIC mengetahui peralatan sudah ditambahkan dan menunggu verifikasi
- Manager tahu proses verifikasi sudah berubah dan bisa mengambil tindakan lebih lanjut

---

## Kesimpulan

Notifikasi saat ini dibuat memakai utility umum di [utils/notifications.go](utils/notifications.go) agar dapat dipakai di banyak event. Pembuatan peralatan memicu notifikasi ke PIC, sementara perubahan status verifikasi memicu notifikasi ke manager lab.
