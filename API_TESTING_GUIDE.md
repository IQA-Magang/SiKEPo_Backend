# Panduan Testing API SiKEPo Backend

Dokumen ini berisi panduan testing endpoint yang benar-benar tersedia di proyek saat ini.

## 1. Base URL

```text
http://localhost:5000
```

## 2. Header umum

Untuk endpoint yang membutuhkan login:

```http
Authorization: Bearer <token>
Content-Type: application/json
```

## 3. Data dummy yang dibuat otomatis

Saat aplikasi pertama kali dijalankan dan database masih kosong, sistem akan otomatis membuat:

- User:
  - admin@sikepo.local / password123
  - manager@sikepo.local / password123
  - staff@sikepo.local / password123
- 2 lab
- 2 ruangan
- 4 kategori peralatan
- 3 kelompok asset
- 4 peralatan + detail spesifikasi

---

## 4. Public / Health

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

### GET /static/login.html

```bash
curl http://localhost:5000/static/login.html
```

---

## 5. User Auth

Base route: `/api/users`

### POST /api/users/login

Request:

```json
{
  "email": "admin@sikepo.local",
  "password": "password123"
}
```

Curl:

```bash
curl -X POST http://localhost:5000/api/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@sikepo.local",
    "password": "password123"
  }'
```

Response sukses:

```json
{
  "success": true,
  "message": "Login berhasil",
  "data": {
    "token": "<jwt_token>",
    "user": {
      "user_id": 1,
      "nip": "1980010101",
      "name": "Admin Utama",
      "email": "admin@sikepo.local",
      "role": "admin",
      "position": "Administrator",
      "pic": true
    }
  }
}
```

### GET /api/users

```bash
curl http://localhost:5000/api/users \
  -H "Authorization: Bearer <token>"
```

### GET /api/users/:id

```bash
curl http://localhost:5000/api/users/1 \
  -H "Authorization: Bearer <token>"
```

### POST /api/users

Hanya admin.

```bash
curl -X POST http://localhost:5000/api/users \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "nip": "2000010101",
    "name": "User Baru",
    "email": "baru@example.com",
    "password": "password123",
    "role": "staff",
    "position": "Staff Laboratorium",
    "pic": false
  }'
```

### PUT /api/users/:id

Hanya admin.

### DELETE /api/users/:id

Hanya admin.

---

## 6. Labs

Base route: `/api/labs`

### GET /api/labs

```bash
curl http://localhost:5000/api/labs \
  -H "Authorization: Bearer <token>"
```

### GET /api/labs/:id

```bash
curl http://localhost:5000/api/labs/1 \
  -H "Authorization: Bearer <token>"
```

### POST /api/labs

Admin only.

```json
{
  "nama_labs": "Laboratorium Kimia",
  "kode_labs": "LAB-KIM-01",
  "manager_id": 2
}
```

### PUT /api/labs/:id

Admin only.

### DELETE /api/labs/:id

Admin only.

---

## 7. Ruangan

Base route: `/api/ruangan`

### GET /api/ruangan

```bash
curl http://localhost:5000/api/ruangan \
  -H "Authorization: Bearer <token>"
```

### GET /api/ruangan/:id

```bash
curl http://localhost:5000/api/ruangan/1 \
  -H "Authorization: Bearer <token>"
```

### GET /api/ruangan/labs/:labs_id

```bash
curl http://localhost:5000/api/ruangan/labs/1 \
  -H "Authorization: Bearer <token>"
```

### GET /api/ruangan/pic/:pic_user_id

```bash
curl http://localhost:5000/api/ruangan/pic/3 \
  -H "Authorization: Bearer <token>"
```

### POST /api/ruangan

Admin only.

```json
{
  "nama_ruangan": "Ruang Instrumen A",
  "kode_ruangan": "R-101",
  "lantai_ruangan": "1",
  "labs_id": 1,
  "pic_user_id": 3
}
```

---

## 8. Kelompok Asset

Base route: `/api/kelompok-asset`

### GET /api/kelompok-asset

```bash
curl http://localhost:5000/api/kelompok-asset \
  -H "Authorization: Bearer <token>"
```

### GET /api/kelompok-asset/:id

```bash
curl http://localhost:5000/api/kelompok-asset/1 \
  -H "Authorization: Bearer <token>"
```

### POST /api/kelompok-asset

```json
{
  "lab_id": 1,
  "pic_id": 1,
  "kode": "KA-INS",
  "nama": "Kelompok Instrumentasi"
}
```

---

## 9. Dokumen Peralatan

Base route: `/api/dokumen-peralatan`

### GET /api/dokumen-peralatan

```bash
curl http://localhost:5000/api/dokumen-peralatan \
  -H "Authorization: Bearer <token>"
```

### GET /api/dokumen-peralatan/peralatan/:peralatan_id

```bash
curl http://localhost:5000/api/dokumen-peralatan/peralatan/1 \
  -H "Authorization: Bearer <token>"
```

### POST /api/dokumen-peralatan

Admin only.

---

## 10. Peralatan

Base route: `/api/peralatan`

### POST /api/peralatan/

Request body contoh:

```json
{
  "nomor_aset": "AST-005",
  "nama_peralatan": "Multimeter Digital",
  "kategori_id": 1,
  "kategori_peralatan_id": 1,
  "kelompok_aset_id": 1,
  "ruangan_id": 1,
  "pic_id": 3,
  "merek": "Fluke",
  "tipe_model": "87V",
  "nomor_seri": "FLU-005",
  "foto": "/uploads/peralatan/AST-005.jpg",
  "status_alat": "Aktif",
  "keterangan": "Alat ukur digital",
  "detail": {
    "parameter_rentang_ukur": "Tegangan, Arus, Resistansi",
    "resolusi": "0.1 mV",
    "akurasi_spesifikasi": "±0.05%",
    "satuan": "V",
    "peranti_lunak_versi": "1.2.0",
    "metode_kelayakan": "Kalibrasi referensi",
    "no_sertifikat": "SER-UK-005",
    "tgl_kalibrasi": "2024-01-15T00:00:00Z",
    "interval_bulan": 12,
    "nilai_koreksi": "0.02",
    "ketidakpastian": "0.01%",
    "jenis_label": "Digital",
    "status_kelayakan": "Layak"
  }
}
```

Curl:

```bash
curl -X POST http://localhost:5000/api/peralatan/ \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "nomor_aset": "AST-005",
    "nama_peralatan": "Multimeter Digital",
    "kategori_id": 1,
    "kategori_peralatan_id": 1,
    "kelompok_aset_id": 1,
    "ruangan_id": 1,
    "pic_id": 3,
    "merek": "Fluke",
    "tipe_model": "87V",
    "nomor_seri": "FLU-005",
    "foto": "/uploads/peralatan/AST-005.jpg",
    "status_alat": "Aktif",
    "keterangan": "Alat ukur digital",
    "detail": {
      "parameter_rentang_ukur": "Tegangan, Arus, Resistansi",
      "resolusi": "0.1 mV",
      "akurasi_spesifikasi": "±0.05%",
      "satuan": "V",
      "peranti_lunak_versi": "1.2.0",
      "metode_kelayakan": "Kalibrasi referensi",
      "no_sertifikat": "SER-UK-005",
      "tgl_kalibrasi": "2024-01-15T00:00:00Z",
      "interval_bulan": 12,
      "nilai_koreksi": "0.02",
      "ketidakpastian": "0.01%",
      "jenis_label": "Digital",
      "status_kelayakan": "Layak"
    }
  }'
```

Response sukses:

```json
{
  "status": "success",
  "message": "Peralatan beserta detail spesifikasinya berhasil ditambahkan"
}
```

### GET /api/peralatan/:id/qr

```bash
curl http://localhost:5000/api/peralatan/1/qr \
  -H "Authorization: Bearer <token>" \
  -o peralatan-1.png
```

---

## 11. Ringkasan endpoint utama

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
GET    /api/labs
GET    /api/labs/:id
POST   /api/labs
PUT    /api/labs/:id
DELETE /api/labs/:id
GET    /api/ruangan
GET    /api/ruangan/:id
GET    /api/ruangan/labs/:labs_id
GET    /api/ruangan/pic/:pic_user_id
POST   /api/ruangan
PUT    /api/ruangan/:id
DELETE /api/ruangan/:id
GET    /api/kelompok-asset
GET    /api/kelompok-asset/:id
POST   /api/kelompok-asset
PUT    /api/kelompok-asset/:id
DELETE /api/kelompok-asset/:id
GET    /api/dokumen-peralatan
GET    /api/dokumen-peralatan/peralatan/:peralatan_id
GET    /api/dokumen-peralatan/:id
POST   /api/dokumen-peralatan
PUT    /api/dokumen-peralatan/:id
DELETE /api/dokumen-peralatan/:id
POST   /api/peralatan/
GET    /api/peralatan/:id/qr
```
