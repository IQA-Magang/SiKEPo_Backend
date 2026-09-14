# SiKEPo_Backend

Backend API SiKEPo dibangun dengan Go, Fiber, GORM, dan MySQL. Proyek ini menangani autentikasi user, lab, ruangan, kelompok asset, kategori peralatan, dokumen peralatan, serta peralatan dan detail spesifikasinya.

## Teknologi

- Go
- Fiber v2
- GORM
- MySQL
- JWT
- bcrypt
- dotenv
- CORS

## Persyaratan

- Go 1.22+
- MySQL Server aktif
- File `.env` di root project

## Konfigurasi environment

```env
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=sikepo
JWT_SECRET=your_super_secret_key
SITE_KEY=your_recaptcha_site_key
APP_PORT=5000
```

## Menjalankan project

```bash
go mod download
go run .
```

Aplikasi berjalan di port `APP_PORT`, atau default `5000` jika tidak diisi.

## Data dummy otomatis

Saat database baru dibuat dan masih kosong, aplikasi akan otomatis membuat data dummy saat startup. Data yang dibuat meliputi:

- 3 user:
  - admin@sikepo.local / password123
  - manager@sikepo.local / password123
  - staff@sikepo.local / password123
- 2 lab
- 2 ruangan
- 4 kategori peralatan
- 3 kelompok asset
- 4 peralatan dengan detail spesifikasi sesuai kategori

> Seed hanya dibuat jika tabel masih kosong. Jika data sudah ada, seed tidak akan overwrite data yang sudah tersimpan.

## Endpoint utama

### Public / health

- GET /
- GET /recaptcha/sitekey
- GET /static/login.html

### User

- POST /api/users/login
- GET /api/users
- GET /api/users/:id
- POST /api/users (admin only)
- PUT /api/users/:id (admin only)
- DELETE /api/users/:id (admin only)

### Labs

- GET /api/labs
- GET /api/labs/:id
- POST /api/labs (admin only)
- PUT /api/labs/:id (admin only)
- DELETE /api/labs/:id (admin only)

### Ruangan

- GET /api/ruangan
- GET /api/ruangan/:id
- GET /api/ruangan/labs/:labs_id
- GET /api/ruangan/pic/:pic_user_id
- POST /api/ruangan (admin only)
- PUT /api/ruangan/:id (admin only)
- DELETE /api/ruangan/:id (admin only)

### Kelompok Asset

- GET /api/kelompok-asset
- GET /api/kelompok-asset/:id
- POST /api/kelompok-asset
- PUT /api/kelompok-asset/:id
- DELETE /api/kelompok-asset/:id

### Dokumen Peralatan

- GET /api/dokumen-peralatan
- GET /api/dokumen-peralatan/peralatan/:peralatan_id
- GET /api/dokumen-peralatan/:id
- POST /api/dokumen-peralatan (admin only)
- PUT /api/dokumen-peralatan/:id (admin only)
- DELETE /api/dokumen-peralatan/:id (admin only)

### Peralatan

- POST /api/peralatan/
- GET /api/peralatan/:id/qr

## Contoh login cepat

```bash
curl -X POST http://localhost:5000/api/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@sikepo.local",
    "password": "password123"
  }'
```

## Catatan penting

- Semua endpoint autentikasi memakai header `Authorization: Bearer <token>`.
- Create/update/delete user, lab, ruangan, dan dokumen hanya untuk role `admin`.
- Dokumentasi request/response yang lebih lengkap ada di [API_TESTING_GUIDE.md](API_TESTING_GUIDE.md).
