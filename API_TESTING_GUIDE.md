# Panduan Testing API SiKEPo Backend

Dokumen ini berisi checklist dan contoh payload JSON untuk melakukan testing API manual pada backend SiKEPo.

## 1. Persiapan

### Base URL
```text
http://localhost:5000
```

Jika menggunakan port lain, sesuaikan dengan nilai `APP_PORT` di file `.env`.

### Header umum

Untuk endpoint yang membutuhkan autentikasi:

```http
Authorization: Bearer <token>
Content-Type: application/json
```

### Contoh token

```text
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

---

## 2. Checklist Testing

### A. Health Check

- [ ] GET /
- [ ] GET /recaptcha/sitekey
- [ ] GET /static/login.html

### B. User Auth

- [ ] POST /api/users/login
- [ ] GET /api/users
- [ ] GET /api/users/:id
- [ ] POST /api/users
- [ ] PUT /api/users/:id
- [ ] DELETE /api/users/:id

### C. Peralatan

- [ ] POST /api/peralatan/
- [ ] GET /api/peralatan/:id/qr

### D. Labs dan Ruangan

- [ ] POST /api/labs
- [ ] GET /api/labs
- [ ] GET /api/labs/:id
- [ ] PUT /api/labs/:id
- [ ] DELETE /api/labs/:id
- [ ] POST /api/ruangan
- [ ] GET /api/ruangan
- [ ] GET /api/ruangan/:id
- [ ] PUT /api/ruangan/:id
- [ ] DELETE /api/ruangan/:id

> Catatan: route Labs dan Ruangan belum aktif secara default di `main.go` sampai di-register ke aplikasi.

---

## 3. Contoh Request dan JSON

## 3.1 Health Check

### GET /

```bash
curl http://localhost:5000/
```

Response:

```json
{
  "success": true,
  "message": "Backend API Running"
}
```

### GET /recaptcha/sitekey

```bash
curl http://localhost:5000/recaptcha/sitekey
```

Response:

```json
{
  "site_key": "your_site_key_here"
}
```

---

## 3.2 Login User

### POST /api/users/login

Request:

```bash
curl -X POST http://localhost:5000/api/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "password123"
  }'
```

Request JSON:

```json
{
  "email": "admin@example.com",
  "password": "password123"
}
```

Response sukses:

```json
{
  "success": true,
  "message": "Login berhasil",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "user_id": 1,
      "nip": "1234567890",
      "name": "Admin",
      "email": "admin@example.com",
      "role": "admin",
      "position": "Manager",
      "pic": true
    }
  }
}
```

Response error:

```json
{
  "success": false,
  "message": "Email atau password salah"
}
```

---

## 3.3 User API

### GET /api/users

```bash
curl http://localhost:5000/api/users \
  -H "Authorization: Bearer <token>"
```

Response:

```json
{
  "success": true,
  "message": "Data user berhasil diambil",
  "data": [
    {
      "user_id": 1,
      "nip": "1234567890",
      "name": "Admin",
      "email": "admin@example.com",
      "role": "admin",
      "position": "Manager",
      "pic": true
    }
  ]
}
```

### GET /api/users/:id

```bash
curl http://localhost:5000/api/users/1 \
  -H "Authorization: Bearer <token>"
```

Response:

```json
{
  "success": true,
  "message": "Data user berhasil ditemukan",
  "data": {
    "user_id": 1,
    "nip": "1234567890",
    "name": "Admin",
    "email": "admin@example.com",
    "role": "admin",
    "position": "Manager",
    "pic": true
  }
}
```

### POST /api/users

Request JSON:

```json
{
  "nip": "1234567890",
  "name": "Admin",
  "email": "admin@example.com",
  "password": "password123",
  "role": "admin",
  "position": "Manager",
  "pic": true
}
```

```bash
curl -X POST http://localhost:5000/api/users \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "nip": "1234567890",
    "name": "Admin",
    "email": "admin@example.com",
    "password": "password123",
    "role": "admin",
    "position": "Manager",
    "pic": true
  }'
```

Response sukses:

```json
{
  "success": true,
  "message": "User berhasil dibuat",
  "data": {
    "user_id": 1,
    "nip": "1234567890",
    "name": "Admin",
    "email": "admin@example.com",
    "role": "admin",
    "position": "Manager",
    "pic": true
  }
}
```

Response conflict:

```json
{
  "success": false,
  "message": "Email sudah digunakan"
}
```

### PUT /api/users/:id

Request JSON:

```json
{
  "nip": "1234567890",
  "name": "Admin Update",
  "email": "admin.update@example.com",
  "password": "newpassword123",
  "role": "admin",
  "position": "Supervisor",
  "pic": false
}
```

```bash
curl -X PUT http://localhost:5000/api/users/1 \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "nip": "1234567890",
    "name": "Admin Update",
    "email": "admin.update@example.com",
    "password": "newpassword123",
    "role": "admin",
    "position": "Supervisor",
    "pic": false
  }'
```

Response sukses:

```json
{
  "success": true,
  "message": "User berhasil diperbarui",
  "data": {
    "user_id": 1,
    "nip": "1234567890",
    "name": "Admin Update",
    "email": "admin.update@example.com",
    "role": "admin",
    "position": "Supervisor",
    "pic": false
  }
}
```

### DELETE /api/users/:id

```bash
curl -X DELETE http://localhost:5000/api/users/1 \
  -H "Authorization: Bearer <token>"
```

Response sukses:

```json
{
  "success": true,
  "message": "User berhasil dihapus"
}
```

---

## 3.4 Peralatan API

Semua endpoint peralatan membutuhkan header `Authorization`. Route yang tersedia saat ini adalah create peralatan dan generate QR code.

> Catatan: `foto` disimpan sebagai string URL atau path file. Endpoint create saat ini menerima JSON, bukan upload file multipart.

### POST /api/peralatan/

Request JSON:

```json
{
  "nomor_aset": "AST-001",
  "nama_peralatan": "Laptop Lenovo",
  "kategori_id": 2,
  "kelompok_aset_id": 1,
  "ruangan_id": 1,
  "pic_id": 2,
  "merek": "Lenovo",
  "tipe_model": "ThinkPad T14",
  "nomor_seri": "SN-123",
  "foto": "/uploads/peralatan/AST-001.jpg",
  "status_alat": "Aktif",
  "keterangan": "Laptop operasional",
  "detail": {
    "fungsi_kegunaan": "Pengolahan data",
    "peranti_lunak_versi": "Windows 11",
    "jenis_pemeriksaan_berkala": "Pemeriksaan visual",
    "kriteria_pemeriksaan": "Menyala dan tidak rusak",
    "interval_bulan": 6
  }
}
```

PowerShell:

```powershell
curl.exe -X POST "http://localhost:5000/api/peralatan/" `
  -H "Authorization: Bearer <token>" `
  -H "Content-Type: application/json" `
  --data-raw '{
    "nomor_aset": "AST-001",
    "nama_peralatan": "Laptop Lenovo",
    "kategori_id": 2,
    "kelompok_aset_id": 1,
    "ruangan_id": 1,
    "pic_id": 2,
    "merek": "Lenovo",
    "tipe_model": "ThinkPad T14",
    "nomor_seri": "SN-123",
    "foto": "/uploads/peralatan/AST-001.jpg",
    "status_alat": "Aktif",
    "keterangan": "Laptop operasional",
    "detail": {
      "fungsi_kegunaan": "Pengolahan data",
      "peranti_lunak_versi": "Windows 11",
      "jenis_pemeriksaan_berkala": "Pemeriksaan visual",
      "kriteria_pemeriksaan": "Menyala dan tidak rusak",
      "interval_bulan": 6
    }
  }'
```

Response sukses memiliki HTTP status `201 Created`:

```json
{
  "status": "success",
  "message": "Peralatan beserta detail spesifikasinya berhasil ditambahkan"
}
```

### GET /api/peralatan/:id/qr

Endpoint ini mengambil `nomor_aset` dari peralatan berdasarkan ID dan mengembalikan QR code dalam format PNG berukuran 256x256 piksel. Simpan hasil response sebagai file gambar:

```powershell
curl.exe "http://localhost:5000/api/peralatan/1/qr" `
  -H "Authorization: Bearer <token>" `
  -o "peralatan-1.png"
```

Buka file `peralatan-1.png` dengan QR scanner. Isi QR code adalah `nomor_aset`, misalnya `AST-001`.

Kemungkinan response error:

| HTTP status | Kondisi                                               |
| ----------- | ----------------------------------------------------- |
| `400`       | ID peralatan bukan angka atau bernilai 0              |
| `401`       | Header Authorization tidak ada atau token tidak valid |
| `404`       | Peralatan dengan ID tersebut tidak ditemukan          |
| `500`       | Gagal mengambil data atau membuat QR code             |

---

## 3.5 Labs API

> Route labs belum aktif di `main.go` sampai dideklarasikan.

### POST /api/labs

Request JSON:

```json
{
  "nama_labs": "Laboratorium Kimia",
  "kode_labs": "LAB-KIM-01",
  "manager_id": 1
}
```

```bash
curl -X POST http://localhost:5000/api/labs \
  -H "Content-Type: application/json" \
  -d '{
    "nama_labs": "Laboratorium Kimia",
    "kode_labs": "LAB-KIM-01",
    "manager_id": 1
  }'
```

Response sukses:

```json
{
  "success": true,
  "message": "Lab berhasil dibuat",
  "data": {
    "id": 1,
    "nama_labs": "Laboratorium Kimia",
    "kode_labs": "LAB-KIM-01",
    "manager_id": 1,
    "manager": {
      "user_id": 1,
      "name": "Admin",
      "email": "admin@example.com"
    }
  }
}
```

### GET /api/labs

```bash
curl http://localhost:5000/api/labs
```

Response:

```json
{
  "success": true,
  "message": "Data labs berhasil diambil",
  "data": [
    {
      "id": 1,
      "nama_labs": "Laboratorium Kimia",
      "kode_labs": "LAB-KIM-01",
      "manager_id": 1,
      "manager": {
        "user_id": 1,
        "name": "Admin",
        "email": "admin@example.com"
      }
    }
  ]
}
```

### GET /api/labs/:id

```bash
curl http://localhost:5000/api/labs/1
```

### PUT /api/labs/:id

Request JSON:

```json
{
  "nama_labs": "Laboratorium Kimia Baru",
  "kode_labs": "LAB-KIM-02",
  "manager_id": 2
}
```

### DELETE /api/labs/:id

```bash
curl -X DELETE http://localhost:5000/api/labs/1
```

---

## 3.6 Ruangan API

> Route ruangan belum dipasang di `main.go` saat ini.

### POST /api/ruangan

Request JSON:

```json
{
  "nama_ruangan": "Lab Komputer",
  "kode_ruangan": "R-101",
  "labs_id": 1,
  "pic_user_id": 2
}
```

```bash
curl -X POST http://localhost:5000/api/ruangan \
  -H "Content-Type: application/json" \
  -d '{
    "nama_ruangan": "Lab Komputer",
    "kode_ruangan": "R-101",
    "labs_id": 1,
    "pic_user_id": 2
  }'
```

Response sukses:

```json
{
  "success": true,
  "message": "Ruangan berhasil dibuat",
  "data": {
    "id": 1,
    "nama_ruangan": "Lab Komputer",
    "kode_ruangan": "R-101",
    "labs_id": 1,
    "pic_user_id": 2
  }
}
```

### GET /api/ruangan

```bash
curl http://localhost:5000/api/ruangan
```

Response:

```json
{
  "success": true,
  "message": "Data ruangan berhasil diambil",
  "data": [
    {
      "id": 1,
      "nama_ruangan": "Lab Komputer",
      "kode_ruangan": "R-101",
      "labs_id": 1,
      "pic_user_id": 2
    }
  ]
}
```

### GET /api/ruangan/:id

```bash
curl http://localhost:5000/api/ruangan/1
```

### PUT /api/ruangan/:id

Request JSON:

```json
{
  "nama_ruangan": "Lab Komputer Baru",
  "kode_ruangan": "R-102",
  "labs_id": 1,
  "pic_user_id": 3
}
```

### DELETE /api/ruangan/:id

```bash
curl -X DELETE http://localhost:5000/api/ruangan/1
```

---

## 4. Catatan Penting Saat Testing

- Pastikan server sudah berjalan dengan `go run .`
- Pastikan file `.env` sudah dibuat dan database MySQL aktif
- Gunakan token JWT yang valid untuk endpoint yang membutuhkan auth
- Untuk endpoint admin, pastikan role user yang login adalah `admin`
- Jika response 401, cek format Authorization header
- Jika response 400, cek validasi field dan tipe data JSON

---

## 5. Contoh Header Postman

### Authorization

```http
Authorization: Bearer <token>
```

### Body

```http
Content-Type: application/json
```

---

## 6. Skenario Testing Sederhana

1. Login via `/api/users/login`
2. Ambil token dari response
3. Gunakan token di semua endpoint protected
4. Buat user baru
5. Buat lab baru
6. Buat ruangan baru
7. Buat peralatan baru
8. Cek list data peralatan dan detailnya
9. Update data peralatan
10. Hapus data uji

---

## 7. Troubleshooting

### Error 401 Unauthorized

- Token belum dikirim
- Token sudah expired
- Header tidak sesuai format `Bearer <token>`

### Error 400 Bad Request

- JSON tidak valid
- field required tidak ada
- enum value salah

### Error 500 Internal Server Error

- database tidak terhubung
- setting `.env` salah
- tabel belum dibuat

---

## 8. Ringkasan Endpoint Utama

```text
GET    /
GET    /recaptcha/sitekey
GET    /static/login.html

POST   /api/users/login
GET    /api/users
GET    /api/users/:id
POST   /api/users
PUT    /api/users/:id
DELETE /api/users/:id

POST   /api/peralatan/
GET    /api/peralatan/:id/qr

POST   /api/labs
GET    /api/labs
GET    /api/labs/:id
PUT    /api/labs/:id
DELETE /api/labs/:id

POST   /api/ruangan
GET    /api/ruangan
GET    /api/ruangan/:id
PUT    /api/ruangan/:id
DELETE /api/ruangan/:id
```

Semua endpoint yang membutuhkan autentikasi harus dikirimkan header `Authorization` dengan format Bearer token.
