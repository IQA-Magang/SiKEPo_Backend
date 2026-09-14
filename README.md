# SiKEPo_Backend

Backend API untuk sistem SiKEPo yang dibangun dengan Go, Fiber, GORM, dan MySQL. Proyek ini menyediakan API untuk autentikasi user, pengelolaan lab, ruangan, serta data peralatan.

## Teknologi yang digunakan
- Go
- Fiber v2
- GORM
- MySQL
- JWT
- bcrypt
- dotenv
- CORS

## Persyaratan

- Go 1.22+ / sesuai versi yang tersedia di environment Anda
- MySQL Server
- .env configuration

## Struktur project

```bash
SiKEPo_Backend/
├── config/
│   └── database.go
├── controllers/
│   ├── labs_controller.go
│   ├── peralatan_controller.go
│   └── user_controller.go
├── middleware/
│   └── auth_middleware.go
├── models/
│   ├── labs.go
│   ├── peralatan.go
│   ├── ruangan.go
│   └── user.go
├── public/
│   └── login.html
├── repositories/
│   ├── labs_repository.go
│   ├── peralatan_repository.go
│   ├── ruangan_repository.go
│   └── user_repository.go
├── routes/
│   ├── labs_routes.go
│   ├── peralatan_routes.go
│   └── user_routes.go
├── utils/
│   └── jwt.go
├── .env
├── go.mod
├── main.go
├── README.md
└── .gitignore
```

## Konfigurasi environment

Buat file `.env` di root project dengan contoh berikut:

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

Keterangan:

- `DB_HOST` : host MySQL
- `DB_PORT` : port MySQL
- `DB_USER` : username MySQL
- `DB_PASSWORD` : password MySQL
- `DB_NAME` : nama database
- `JWT_SECRET` : kunci untuk menandatangani token JWT
- `SITE_KEY` : site key untuk reCAPTCHA
- `APP_PORT` : port aplikasi, default jika tidak diisi adalah `5000`

## Cara menjalankan project

1. Pastikan MySQL sudah berjalan dan database tersedia.
2. Buat file `.env` sesuai contoh.
3. Jalankan dependency download:

```bash
go mod download
```

4. Jalankan aplikasi:

```bash
go run .
```

5. Aplikasi akan berjalan di port yang ditentukan melalui `APP_PORT` atau default `5000`.

6. Cek kesehatan API:

```bash
curl http://localhost:5000/
```

Contoh response:

```json
{
  "success": true,
  "message": "Backend API Running"
}
```

## Endpoint API yang tersedia

### 1. Health check dan public

#### GET /
Menampilkan status server.

```bash
GET http://localhost:5000/
```

#### GET /recaptcha/sitekey
Mengembalikan site key reCAPTCHA.

```bash
GET http://localhost:5000/recaptcha/sitekey
```

#### GET /static/{file}
Menyediakan file statis dari folder `public` untuk keperluan testing atau frontend sederhana.

```bash
GET http://localhost:5000/static/login.html
```

---

### 2. User API
Base route: `/api/users`

#### Public

##### POST /api/users/login
Login user.

Request body contoh:

```json
{
  "email": "admin@example.com",
  "password": "password123"
}
```

Contoh response sukses:

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

Contoh response error:

```json
{
  "success": false,
  "message": "Email atau password salah"
}
```

##### GET /api/users
Menampilkan seluruh data user.

Contoh response sukses:

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

##### GET /api/users/:id
Menampilkan detail user berdasarkan ID.

Contoh response sukses:

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

Contoh response 404:

```json
{
  "success": false,
  "message": "User tidak ditemukan"
}
```

#### Admin only

##### POST /api/users
Membuat user baru.

Request body contoh:

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

Contoh response sukses:

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

Contoh response conflict:

```json
{
  "success": false,
  "message": "Email sudah digunakan"
}
```

Role yang diterima biasanya:

- `admin`
- `staff`
- `manager`

##### PUT /api/users/:id
Update data user.

Request body contoh:

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

Contoh response sukses:

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

##### DELETE /api/users/:id
Hapus user.

Contoh response sukses:

```json
{
  "success": true,
  "message": "User berhasil dihapus"
}
```

Catatan: route user memakai middleware `RequireAuth` dan `RequireRoles("admin")` hanya untuk operasi create/update/delete.

---

### 3. Peralatan API
Base route: `/api/v1/peralatan`

Semua endpoint di bawah ini membutuhkan autentikasi JWT (`Authorization: Bearer <token>`).

#### POST /api/v1/peralatan/
Membuat data peralatan baru.

Request body contoh:

```json
{
  "ruangan_id": 1,
  "pic_id": 2,
  "nomor_aset": "AST-001",
  "nama_peralatan": "Laptop Lenovo",
  "merk": "Lenovo",
  "model": "ThinkPad T14",
  "nomor_seri": "SN-123",
  "jumlah": 1,
  "kategori_peralatan": "peralatan",
  "kondisi": "sesuai",
  "status_kelayakan": "pending",
  "metode": "internal",
  "jenis_pakai": "tidak_habis_pakai",
  "verified_at": null,
  "verification_note": "",
  "verified_by": null
}
```

Contoh response sukses:

```json
{
  "success": true,
  "data": {
    "id": 1,
    "ruangan_id": 1,
    "pic_id": 2,
    "nomor_aset": "AST-001",
    "nama_peralatan": "Laptop Lenovo",
    "merk": "Lenovo",
    "model": "ThinkPad T14",
    "nomor_seri": "SN-123",
    "jumlah": 1,
    "kategori_peralatan": "peralatan",
    "kondisi": "sesuai",
    "status_kelayakan": "pending",
    "metode": "internal",
    "jenis_pakai": "tidak_habis_pakai",
    "input_by": 1,
    "verified_by": null,
    "verified_at": null,
    "verification_note": ""
  }
}
```

Contoh response validasi gagal:

```json
{
  "success": false,
  "message": "Validasi gagal",
  "errors": {
    "nomor_aset": "Nomor aset wajib diisi",
    "nama_peralatan": "Nama peralatan wajib diisi"
  }
}
```

#### GET /api/v1/peralatan/
Menampilkan daftar peralatan dengan pagination dan filter.

Query params yang didukung:

- `page` (default `1`)
- `limit` (default `10`, max `100`)
- `search`
- `ruangan_id`
- `pic_id`
- `status_kelayakan`

Contoh:

```bash
GET http://localhost:5000/api/v1/peralatan?page=1&limit=10&search=laptop
```

Contoh response sukses:

```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "ruangan_id": 1,
      "pic_id": 2,
      "nomor_aset": "AST-001",
      "nama_peralatan": "Laptop Lenovo",
      "merk": "Lenovo",
      "model": "ThinkPad T14",
      "jumlah": 1,
      "kategori_peralatan": "peralatan",
      "kondisi": "sesuai",
      "status_kelayakan": "pending"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 10,
    "total_data": 1
  }
}
```

#### GET /api/v1/peralatan/:id
Menampilkan detail peralatan berdasarkan ID.

Contoh response sukses:

```json
{
  "success": true,
  "data": {
    "id": 1,
    "ruangan_id": 1,
    "pic_id": 2,
    "nomor_aset": "AST-001",
    "nama_peralatan": "Laptop Lenovo",
    "merk": "Lenovo",
    "model": "ThinkPad T14",
    "jumlah": 1,
    "kategori_peralatan": "peralatan",
    "kondisi": "sesuai",
    "status_kelayakan": "pending",
    "ruangan": {
      "id": 1,
      "nama_ruangan": "Lab Komputer"
    },
    "pic": {
      "user_id": 2,
      "name": "User PIC",
      "email": "pic@example.com"
    }
  }
}
```

Contoh response 404:

```json
{
  "success": false,
  "message": "Data peralatan tidak ditemukan"
}
```

#### PUT /api/v1/peralatan/:id
Update data peralatan.

Request body contoh:

```json
{
  "nama_peralatan": "Laptop ASUS",
  "status_kelayakan": "aktif",
  "kondisi": "sesuai"
}
```

Contoh response sukses:

```json
{
  "success": true,
  "data": {
    "id": 1,
    "nama_peralatan": "Laptop ASUS",
    "status_kelayakan": "aktif",
    "kondisi": "sesuai"
  }
}
```

#### DELETE /api/v1/peralatan/:id
Hapus data peralatan.

Contoh response sukses:

```json
{
  "success": true,
  "message": "Peralatan berhasil dihapus"
}
```

---

### 4. Labs API
File route labs sudah tersedia di `routes/labs_routes.go`, namun pada saat ini route tersebut belum dipanggil di `main.go`. Jadi endpoint labs belum aktif secara default sampai ditambahkan di bootstrap aplikasi.

Struktur route yang ada di source code:

```go
labs.Get("/", labsController.GetAll)
labs.Get("/:id", labsController.GetByID)
labs.Post("/", labsController.Create)
labs.Put("/:id", labsController.Update)
labs.Delete("/:id", labsController.Delete)
```

Jika nanti di-mount ke aplikasi, maka biasanya akan tersedia dalam bentuk:

- `GET /api/labs`
- `GET /api/labs/:id`
- `POST /api/labs`
- `PUT /api/labs/:id`
- `DELETE /api/labs/:id`

#### POST /api/labs
Membuat data lab baru.

Request body contoh:

```json
{
  "nama_labs": "Laboratorium Kimia",
  "kode_labs": "LAB-KIM-01",
  "manager_id": 1
}
```

Contoh response sukses:

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

Contoh response conflict:

```json
{
  "success": false,
  "message": "Kode lab sudah digunakan"
}
```

#### GET /api/labs
Menampilkan seluruh data lab.

Contoh response sukses:

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

#### GET /api/labs/:id
Menampilkan detail lab berdasarkan ID.

#### PUT /api/labs/:id
Update data lab.

Request body contoh:

```json
{
  "nama_labs": "Laboratorium Kimia Baru",
  "kode_labs": "LAB-KIM-02",
  "manager_id": 2
}
```

#### DELETE /api/labs/:id
Hapus lab.

Contoh response sukses:

```json
{
  "success": true,
  "message": "Lab berhasil dihapus"
}
```

---

### 5. Ruangan API
Model ruangan sudah tersedia di `models/ruangan.go`, tetapi belum ada router khusus yang terdaftar di `main.go`. Artinya endpoint ruangan juga belum aktif secara default sampai dibuat route yang menghubungkan controller dan repository.

#### Struktur data ruangan

```json
{
  "id": 1,
  "nama_ruangan": "Lab Komputer",
  "kode_ruangan": "R-101",
  "labs_id": 1,
  "pic_user_id": 2,
  "created_at": "2026-09-10T00:00:00Z",
  "updated_at": "2026-09-10T00:00:00Z"
}
```

#### Field utama ruangan

- `id` : ID ruangan
- `nama_ruangan` : nama ruangan
- `kode_ruangan` : kode ruangan unik
- `labs_id` : ID lab yang terkait
- `pic_user_id` : ID user PIC ruangan

#### Endpoint yang kemungkinan akan dibuat

- `GET /api/ruangan`
- `GET /api/ruangan/:id`
- `POST /api/ruangan`
- `PUT /api/ruangan/:id`
- `DELETE /api/ruangan/:id`

#### Contoh request create ruangan

```json
{
  "nama_ruangan": "Lab Komputer",
  "kode_ruangan": "R-101",
  "labs_id": 1,
  "pic_user_id": 2
}
```

#### Contoh response sukses

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

#### Contoh response error jika kode ruangan duplikat

```json
{
  "success": false,
  "message": "Kode ruangan sudah digunakan"
}
```

---

## Middleware dan autentikasi

### JWT middleware
File: `middleware/auth_middleware.go`

Middleware ini akan:

- membaca header `Authorization`
- memvalidasi format `Bearer <token>`
- memverifikasi token JWT dengan `JWT_SECRET`
- menyimpan `user_id`, `email`, dan `role` ke `c.Locals(...)`
- meneruskan request jika token valid

### Role middleware
`RequireRoles(allowedRoles ...string)` digunakan untuk membatasi akses endpoint berdasarkan role user.

---

## Database behavior

Saat aplikasi dijalankan, `config.ConnectDatabase()` akan:

- memuat `.env`
- membuka koneksi MySQL
- mengecek apakah tabel `users` sudah ada
- menjalankan `AutoMigrate` jika belum ada

Catatan: model lain seperti `labs`, `ruangan`, dan `peralatan` membutuhkan migrasi/manual creation sesuai kebutuhan.

---

## Catatan penting

- Route `PeralatanRoutes` diproteksi dengan `RequireAuth`.
- Route `UserRoutes` hanya untuk create/update/delete user yang dibatasi admin.
- Endpoint `/api/v1/peralatan` merupakan API utama untuk pengelolaan aset.
- Aplikasi masih menggunakan struktur backend sederhana tanpa arsitektur yang sangat kompleks, tetapi sudah mengikuti pola umum: model -> repository -> controller -> route.

## Troubleshooting umum

### 1. Error koneksi database
Pastikan:

- MySQL server aktif
- env DB benar
- database sudah dibuat
- user MySQL punya hak akses

### 2. JWT error
Pastikan:

- file `.env` sudah punya `JWT_SECRET`
- token dikirim dalam format `Bearer <token>`

### 3. Port sudah dipakai
Ubah nilai `APP_PORT` atau hentikan service yang memakai port tersebut.

---

## Developer notes

Project ini masih bisa dikembangkan lebih lanjut dengan:

- menambahkan dokumentasi Swagger/OpenAPI
- membuat route labs aktif di `main.go`
- menambahkan migrasi semua model secara otomatis
- menambah unit/integration test
- menambahkan soft delete dan audit log

## Lisensi

Project ini dibuat untuk kebutuhan internal pengembangan backend SiKEPo. Sesuaikan lisensi jika digunakan di lingkungan produksi atau tim lain.
