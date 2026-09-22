# Daftar API SiKEPo Backend

Dokumen ini berisi daftar seluruh endpoint yang tersedia di backend saat ini.

## Base URL

```text
http://localhost:5000
```

## 1. Public / Health

### GET /

Cek server berjalan.

```bash
curl http://localhost:5000/
```

### GET /recaptcha/sitekey

Ambil site key reCAPTCHA.

```bash
curl http://localhost:5000/recaptcha/sitekey
```

### GET /static/login.html

Halaman login statis.

```bash
curl http://localhost:5000/static/login.html
```

---

## 2. User / Auth

Base: `/api/users`

### POST /api/users/login

Login user.

```bash
curl -X POST http://localhost:5000/api/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@sikepo.local",
    "password": "password123"
  }'
```

### POST /api/users/

Create user. Role: admin.

```bash
curl -X POST http://localhost:5000/api/users/ \
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

### GET /api/users/

List semua user. Role: admin.

```bash
curl http://localhost:5000/api/users/ \
  -H "Authorization: Bearer <token>"
```

### GET /api/users/:id

Detail user by ID. Role: admin.

```bash
curl http://localhost:5000/api/users/1 \
  -H "Authorization: Bearer <token>"
```

### PUT /api/users/:id

Update user. Role: admin.

```bash
curl -X PUT http://localhost:5000/api/users/1 \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "User Updated"
  }'
```

### DELETE /api/users/:id

Delete user. Role: admin.

```bash
curl -X DELETE http://localhost:5000/api/users/1 \
  -H "Authorization: Bearer <token>"
```

---

## 3. Labs

Base: `/api/labs`

### GET /api/labs/

List semua lab.

```bash
curl http://localhost:5000/api/labs/ \
  -H "Authorization: Bearer <token>"
```

### GET /api/labs/:id

Detail lab by ID.

```bash
curl http://localhost:5000/api/labs/1 \
  -H "Authorization: Bearer <token>"
```

### POST /api/labs/

Create lab. Role: admin.

```bash
curl -X POST http://localhost:5000/api/labs/ \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "nama_labs": "Laboratorium Kimia",
    "kode_labs": "LAB-KIM-01",
    "manager_id": 2
  }'
```

### PUT /api/labs/:id

Update lab. Role: admin.

```bash
curl -X PUT http://localhost:5000/api/labs/1 \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "nama_labs": "Laboratorium Baru"
  }'
```

### DELETE /api/labs/:id

Delete lab. Role: admin.

```bash
curl -X DELETE http://localhost:5000/api/labs/1 \
  -H "Authorization: Bearer <token>"
```

---

## 4. Ruangan

Base: `/api/ruangan`

### GET /api/ruangan/

List semua ruangan.

```bash
curl http://localhost:5000/api/ruangan/ \
  -H "Authorization: Bearer <token>"
```

### GET /api/ruangan/:id

Detail ruangan by ID.

```bash
curl http://localhost:5000/api/ruangan/1 \
  -H "Authorization: Bearer <token>"
```

### GET /api/ruangan/labs/:labs_id

List ruangan berdasarkan lab.

```bash
curl http://localhost:5000/api/ruangan/labs/1 \
  -H "Authorization: Bearer <token>"
```

### GET /api/ruangan/pic/:pic_user_id

List ruangan berdasarkan PIC user.

```bash
curl http://localhost:5000/api/ruangan/pic/3 \
  -H "Authorization: Bearer <token>"
```

### POST /api/ruangan/

Create ruangan. Role: admin.

```bash
curl -X POST http://localhost:5000/api/ruangan/ \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "nama_ruangan": "Ruang Instrumen A",
    "kode_ruangan": "R-101",
    "lantai_ruangan": "1",
    "labs_id": 1,
    "pic_user_id": 3
  }'
```

### PUT /api/ruangan/:id

Update ruangan. Role: admin.

```bash
curl -X PUT http://localhost:5000/api/ruangan/1 \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "nama_ruangan": "Ruang Instrumen B"
  }'
```

### DELETE /api/ruangan/:id

Delete ruangan. Role: admin.

```bash
curl -X DELETE http://localhost:5000/api/ruangan/1 \
  -H "Authorization: Bearer <token>"
```

---

## 5. Kelompok Asset

Base: `/api/kelompok-asset`

### POST /api/kelompok-asset/

Create kelompok asset.

```bash
curl -X POST http://localhost:5000/api/kelompok-asset/ \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "lab_id": 1,
    "pic_id": 1,
    "kode": "KA-INS",
    "nama": "Kelompok Instrumentasi"
  }'
```

### GET /api/kelompok-asset/

List semua kelompok asset.

```bash
curl http://localhost:5000/api/kelompok-asset/ \
  -H "Authorization: Bearer <token>"
```

### GET /api/kelompok-asset/:id

Detail kelompok asset by ID.

```bash
curl http://localhost:5000/api/kelompok-asset/1 \
  -H "Authorization: Bearer <token>"
```

### PUT /api/kelompok-asset/:id

Update kelompok asset.

```bash
curl -X PUT http://localhost:5000/api/kelompok-asset/1 \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "nama": "Kelompok Instrumentasi Update"
  }'
```

### DELETE /api/kelompok-asset/:id

Delete kelompok asset.

```bash
curl -X DELETE http://localhost:5000/api/kelompok-asset/1 \
  -H "Authorization: Bearer <token>"
```

---

## 6. Dokumen Peralatan

Base: `/api/dokumen-peralatan`

### GET /api/dokumen-peralatan/

List semua dokumen peralatan.

```bash
curl http://localhost:5000/api/dokumen-peralatan/ \
  -H "Authorization: Bearer <token>"
```

### GET /api/dokumen-peralatan/peralatan/:peralatan_id

List dokumen berdasarkan peralatan.

```bash
curl http://localhost:5000/api/dokumen-peralatan/peralatan/1 \
  -H "Authorization: Bearer <token>"
```

### GET /api/dokumen-peralatan/:id

Detail dokumen peralatan by ID.

```bash
curl http://localhost:5000/api/dokumen-peralatan/1 \
  -H "Authorization: Bearer <token>"
```

### POST /api/dokumen-peralatan/

Create dokumen. Role: admin.

```bash
curl -X POST http://localhost:5000/api/dokumen-peralatan/ \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "peralatan_id": 1,
    "nama_dokumen": "Manual Book",
    "path_file": "/uploads/manual.pdf"
  }'
```

### PUT /api/dokumen-peralatan/:id

Update dokumen. Role: admin.

```bash
curl -X PUT http://localhost:5000/api/dokumen-peralatan/1 \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "nama_dokumen": "Manual Book Update"
  }'
```

### DELETE /api/dokumen-peralatan/:id

Delete dokumen. Role: admin.

```bash
curl -X DELETE http://localhost:5000/api/dokumen-peralatan/1 \
  -H "Authorization: Bearer <token>"
```

---

## 7. Kategori Peralatan

Base: `/api/kategori-peralatan`

### GET /api/kategori-peralatan/

List semua kategori peralatan.

```bash
curl http://localhost:5000/api/kategori-peralatan/ \
  -H "Authorization: Bearer <token>"
```

### GET /api/kategori-peralatan/:id

Detail kategori peralatan by ID.

```bash
curl http://localhost:5000/api/kategori-peralatan/1 \
  -H "Authorization: Bearer <token>"
```

### POST /api/kategori-peralatan/

Create kategori peralatan. Role: admin.

```bash
curl -X POST http://localhost:5000/api/kategori-peralatan/ \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "nama_kategori": "Alat Ukur",
    "description": "Peralatan untuk pengukuran dan kalibrasi"
  }'
```

### PUT /api/kategori-peralatan/:id

Update kategori peralatan. Role: admin.

```bash
curl -X PUT http://localhost:5000/api/kategori-peralatan/1 \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "nama_kategori": "Alat Ukur Baru",
    "description": "Deskripsi update"
  }'
```

### DELETE /api/kategori-peralatan/:id

Delete kategori peralatan. Role: admin.

```bash
curl -X DELETE http://localhost:5000/api/kategori-peralatan/1 \
  -H "Authorization: Bearer <token>"
```

---

## 8. Peralatan

Base: `/api/peralatan`

### GET /api/peralatan/

List semua peralatan.

```bash
curl http://localhost:5000/api/peralatan/ \
  -H "Authorization: Bearer <token>"
```

### POST /api/peralatan/

Create peralatan. Role: admin.

```bash
curl -X POST http://localhost:5000/api/peralatan/ \
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

### POST /api/peralatan/:id/foto

Upload foto peralatan.

```bash
curl -X POST http://localhost:5000/api/peralatan/1/foto \
  -H "Authorization: Bearer <token>" \
  -F "foto=@/path/to/foto.jpg"
```

### GET /api/peralatan/:id/qr

Generate QR code peralatan.

```bash
curl http://localhost:5000/api/peralatan/1/qr \
  -H "Authorization: Bearer <token>"
```

---

## 9. Notifikasi

Base: `/api/notifications`

### GET /api/notifications/user/:user_id

List notifikasi user. Role: manager.

```bash
curl http://localhost:5000/api/notifications/user/2 \
  -H "Authorization: Bearer <token>"
```

### PATCH /api/notifications/:id/read

Tandai notifikasi sudah dibaca. Role: manager.

```bash
curl -X PATCH http://localhost:5000/api/notifications/1/read \
  -H "Authorization: Bearer <token>"
```

---

