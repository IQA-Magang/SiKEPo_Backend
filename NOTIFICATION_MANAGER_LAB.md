# Notifikasi Manager Lab saat Peralatan Ditambahkan

Dokumen ini menjelaskan cara penggunaan dan implementasi notifikasi yang dikirim ke manager lab ketika data peralatan baru berhasil dibuat.

## Tujuan

Ketika staff atau admin menambahkan data peralatan baru, manager lab yang terkait dengan lab tersebut akan menerima notifikasi. Tujuannya agar manager lab segera mengetahui ada asset baru yang masuk ke ruangan/lab miliknya.

---

## Alur kerja

1. User mengirim request create peralatan ke endpoint `/api/peralatan/`.
2. Sistem menyimpan data peralatan ke database.
3. Sistem mengambil ruangan berdasarkan `ruangan_id`.
4. Sistem mengambil lab dari ruangan tersebut.
5. Sistem membaca `manager_id` dari lab.
6. Sistem membuat entri notifikasi untuk user manager tersebut.
7. Manager lab dapat melihat notifikasi dari endpoint yang tersedia.

---

## Struktur data notifikasi

Model notifikasi berada di `models/notification.go`.

```go
type Notification struct {
    ID        uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
    UserID    uint64    `gorm:"column:user_id;not null" json:"user_id"`
    Type      string    `gorm:"column:type;size:50;not null;default:'equipment_added'" json:"type"`
    Title     string    `gorm:"column:title;size:150;not null" json:"title"`
    Message   string    `gorm:"column:message;type:text;not null" json:"message"`
    IsRead    bool      `gorm:"column:is_read;default:false" json:"is_read"`
    CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
    UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"column:deleted_at" json:"-"`
}
```

Field penting:
- `user_id`: penerima notifikasi
- `type`: tipe notifikasi, default `equipment_added`
- `title`: judul notifikasi
- `message`: isi detail notifikasi
- `is_read`: status sudah dibaca atau belum

---

## Endpoint notifikasi

Route notifikasi sudah di-register di [routes/notification_routes.go](../routes/notification_routes.go).

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
      "type": "equipment_added",
      "title": "Peralatan baru ditambahkan",
      "message": "Peralatan Multimeter Digital (AST-001) telah ditambahkan ke ruangan Ruang Instrumen A.",
      "is_read": false,
      "created_at": "2026-09-15T12:00:00Z"
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

---

## Implementasi backend

### Trigger notifikasi
Trigger dilakukan di controller peralatan setelah data berhasil dibuat.

File yang relevan:
- [controllers/peralatan_controller.go](../controllers/peralatan_controller.go)
- [main.go](../main.go)
- [config/database.go](../config/database.go)

Bagian inti:

```go
peralatan, err := c.Repo.FindByNomorAset(nomorAset)
if err != nil {
    return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
        "status":  "error",
        "message": "Peralatan berhasil dibuat tetapi gagal mengambil ID",
        "error":   err.Error(),
    })
}

c.sendManagerLabNotification(peralatan)
```

Fungsi `sendManagerLabNotification` bekerja seperti ini:

```go
func (c *PeralatanController) sendManagerLabNotification(peralatan *models.Peralatan) {
    if c.DB == nil || c.NotificationRepo == nil || peralatan == nil {
        return
    }

    var room models.Ruangan
    if err := c.DB.Preload("Labs").First(&room, peralatan.RuanganID).Error; err != nil {
        log.Printf("Gagal mengambil ruangan untuk notifikasi manager lab: %v", err)
        return
    }

    if room.Labs == nil || room.Labs.ManagerID == nil {
        return
    }

    notification := &models.Notification{
        UserID:  *room.Labs.ManagerID,
        Type:    "equipment_added",
        Title:   "Peralatan baru ditambahkan",
        Message: fmt.Sprintf("Peralatan %s (%s) telah ditambahkan ke ruangan %s.", peralatan.NamaPeralatan, peralatan.NomorAset, room.NamaRuangan),
    }

    if err := c.NotificationRepo.Create(notification); err != nil {
        log.Printf("Gagal mengirim notifikasi ke manager lab: %v", err)
    }
}
```

---

## Catatan penting

- Notifikasi dikirim berdasarkan `ruangan_id` yang terhubung ke lab tertentu.
- Manager yang menerima notifikasi harus memiliki `user_id` yang tersimpan di `labs.manager_id`.
- Tabel notifikasi otomatis dibuat saat aplikasi menjalankan AutoMigrate di [config/database.go](../config/database.go).
- Endpoint diberi middleware `RequireRoles("manager")`, jadi hanya role manager yang bisa melihat dan membaca notifikasi pada route ini.

---

## Contoh skenario

Skenario:

- Lab A memiliki manager_id = 2
- Staff menambahkan peralatan baru dengan `ruangan_id = 5`
- Ruangan 5 terhubung ke Lab A
- Sistem akan membuat notifikasi untuk user dengan ID 2

Hasil:

Manager lab menerima notifikasi bahwa peralatan baru sudah ditambahkan ke ruangan terkait.

---

## Kesimpulan

Notifikasi manager lab saat penambahan peralatan sudah diimplementasikan dengan pola sederhana dan aman:

- tidak mengganggu flow utama create peralatan
- data notifikasi tersimpan di tabel terpisah
- mudah di-query dan ditandai sudah dibaca
- dapat dikembangkan ke push notification, websocket, atau email di masa depan
