# ERD Penjadwalan Pemeriksaan Peralatan SiKEPo

Dokumen ini mengusulkan struktur database untuk fitur jadwal kalibrasi,
verifikasi fungsi, pengecekan antara, dan pemeliharaan. Jenis tindakan
disatukan dalam satu tabel jadwal; satu baris jadwal mewakili satu jenis
tindakan pada satu peralatan. Setiap pelaksanaan disimpan sebagai riwayat
tersendiri.

> Diagram ini adalah rancangan target, bukan gambaran bahwa semua tabel dan
> relasi sudah diterapkan di database saat ini. `jadwal_pemeriksaan_berkala`
> sudah ada di backend, tetapi model/endpoint yang ada masih perlu disesuaikan
> agar mendukung banyak jadwal peralatan dan histori pelaksanaan.

## ERD

```mermaid
erDiagram
    LABS ||--o{ RUANGAN : memiliki
    RUANGAN ||--o{ PERALATAN : menempatkan
    PERALATAN ||--o{ JADWAL_PEMERIKSAAN_BERKALA : memiliki
    JENIS_PEMERIKSAAN ||--o{ JADWAL_PEMERIKSAAN_BERKALA : mengategorikan
    USERS o|--o{ JADWAL_PEMERIKSAAN_BERKALA : menjadi_pic
    JADWAL_PEMERIKSAAN_BERKALA ||--o{ RIWAYAT_PEMERIKSAAN : mencatat
    USERS o|--o{ RIWAYAT_PEMERIKSAAN : melaksanakan

    LABS {
        bigint id PK
        varchar kode_labs UK
        varchar nama_labs
    }

    RUANGAN {
        bigint id PK
        bigint lab_id FK
        varchar kode_ruangan
        varchar nama_ruangan
    }

    PERALATAN {
        bigint id PK
        varchar nomor_aset UK
        varchar nama_peralatan
        bigint ruangan_id FK
        varchar status_alat
    }

    JENIS_PEMERIKSAAN {
        bigint id PK
        varchar kode UK
        varchar nama
        boolean aktif
    }

    JADWAL_PEMERIKSAAN_BERKALA {
        bigint id PK
        bigint peralatan_id FK
        bigint jenis_pemeriksaan_id FK
        bigint pic_user_id FK
        varchar metode
        varchar dasar_pemeriksaan
        varchar frekuensi_penggunaan
        varchar status_proses
        int interval_bulan
        date tanggal_terakhir
        date tanggal_jatuh_tempo
        boolean aktif
        text keterangan
        datetime created_at
        datetime updated_at
    }

    RIWAYAT_PEMERIKSAAN {
        bigint id PK
        bigint jadwal_id FK
        bigint pelaksana_user_id FK
        date tanggal_pelaksanaan
        varchar hasil
        varchar nomor_dokumen
        varchar dokumen_path
        text catatan
        datetime created_at
    }

    USERS {
        bigint user_id PK
        varchar name
        varchar email
    }
```

### Kardinalitas dan aturan data

- Satu **lab** memiliki banyak ruangan; satu ruangan berada di satu lab.
- Satu **ruangan** menampung banyak peralatan; satu peralatan berada di satu
  ruangan sesuai struktur SiKEPo saat ini.
- Satu **peralatan** dapat memiliki banyak jadwal. Contohnya, satu alat dapat
  memiliki jadwal kalibrasi dan pengecekan antara sekaligus.
- Satu **jenis pemeriksaan** dapat digunakan oleh banyak jadwal.
- Satu **jadwal** dapat memiliki banyak riwayat pelaksanaan. Riwayat lama tidak
  ditimpa ketika pemeriksaan berikutnya dicatat.
- PIC pada jadwal dan pelaksana pada riwayat merujuk ke `users.user_id`.
  Keduanya boleh kosong jika penugasan/pelaksana belum ditentukan.
- Kombinasi peralatan dan jenis pemeriksaan sebaiknya hanya memiliki satu
  jadwal aktif. Tegakkan aturan ini dengan validasi backend atau constraint
  database yang sesuai dengan strategi pengarsipan jadwal.
- `interval_bulan`, `tanggal_terakhir`, dan `tanggal_jatuh_tempo` boleh kosong
  saat jadwal belum dikonfigurasi. Kondisi ini ditampilkan sebagai
  **Belum Diatur**, bukan sebagai tanggal atau interval buatan.
- Status urgensi seperti **Terlambat**, **Mendesak**, **Mendekati Jatuh Tempo**,
  dan **Aman** dihitung ketika data dibaca dari tanggal jatuh tempo dan ambang
  hari yang disepakati. Status urgensi tidak disimpan sebagai nilai tetap.
- Status proses jadwal, misalnya `terjadwal`, `dalam_proses`, atau `dijeda`,
  disimpan pada jadwal karena mencerminkan keadaan proses, bukan hitungan waktu.
- Status kondisi/penggunaan alat, seperti **Dipinjam**, **Rusak**, atau
  **Karantina**, tetap merupakan status peralatan, bukan status jadwal.
  Tampilan dapat menggabungkan status proses jadwal dengan urgensi dan status
  alat tanpa menyatukan ketiganya menjadi satu nilai status.

### Indeks yang disarankan

- `jadwal_pemeriksaan_berkala(peralatan_id, jenis_pemeriksaan_id)` untuk
  pencarian jadwal alat dan pemeriksaan duplikat.
- `jadwal_pemeriksaan_berkala(aktif, tanggal_jatuh_tempo)` untuk daftar tindakan
  dan pengurutan berdasarkan jatuh tempo.
- `jadwal_pemeriksaan_berkala(pic_user_id, aktif)` untuk daftar tugas PIC.
- `riwayat_pemeriksaan(jadwal_id, tanggal_pelaksanaan)` untuk riwayat per
  jadwal.

## Alur data penjadwalan dan pencatatan hasil

```mermaid
flowchart TD
    A[Pengguna memilih peralatan] --> B{Peralatan terdaftar?}
    B -- Tidak --> X[Tampilkan alasan dan hentikan]
    B -- Ya --> C[Pilih jenis pemeriksaan]
    C --> D{Jadwal aktif untuk jenis ini sudah ada?}
    D -- Ya --> E[Ubah jadwal yang ada]
    D -- Tidak --> F[Buat jadwal baru]
    E --> G[Validasi interval, metode, PIC, dan tanggal]
    F --> G
    G --> H{Data valid?}
    H -- Tidak --> Y[Tampilkan kesalahan validasi]
    H -- Ya --> I[Hitung tanggal jatuh tempo jika tanggal dan interval tersedia]
    I --> J[Simpan jadwal]
    J --> K[Tampilkan daftar tindakan]
    K --> L[Hitung status urgensi dari tanggal jatuh tempo]
    L --> M{Pengguna mencatat hasil?}
    M -- Tidak --> K
    M -- Ya --> N[Validasi hasil, tanggal, dan dokumen]
    N --> O{Data hasil valid?}
    O -- Tidak --> Y
    O -- Ya --> P[Mulai transaksi database]
    P --> Q[Tambah baris riwayat pemeriksaan]
    Q --> R[Perbarui tanggal terakhir dan jatuh tempo jadwal]
    R --> S{Semua perubahan berhasil?}
    S -- Tidak --> T[Rollback dan laporkan kegagalan]
    S -- Ya --> U[Commit transaksi]
    U --> K
```

## Makna data yang ditampilkan pada daftar tindakan

| Data pada layar | Sumber yang disarankan |
|---|---|
| Nomor aset, nama alat, lab | `peralatan` → `ruangan` → `labs` |
| Jenis dan metode pemeriksaan | `jadwal_pemeriksaan_berkala` → `jenis_pemeriksaan` |
| Tanggal terakhir dan hasil terakhir | Jadwal untuk tanggal ringkas; riwayat terbaru untuk hasil |
| Jatuh tempo dan sisa hari | `tanggal_jatuh_tempo`, dihitung relatif ke tanggal saat ini |
| Status urgensi | Hasil perhitungan backend dari jatuh tempo dan ambang hari |
| Status alat | `peralatan.status_alat` |
| PIC | `jadwal_pemeriksaan_berkala.pic_user_id` → `users` |
| Sertifikat/laporan dan catatan historis | `riwayat_pemeriksaan` |

## Hubungan dengan model backend saat ini

Model [JadwalPemeriksaanBerkala](../models/jadwal_pemeriksaan_berkala.go)
saat ini sudah memuat jenis jadwal, interval, tanggal, dan beberapa detail
metode/dokumen. Untuk mengikuti rancangan ini:

1. Pertahankan satu tabel jadwal, tetapi gunakan identitas jenis pemeriksaan
   yang konsisten. `jenis_pemeriksaan_id` pada diagram adalah usulan lookup
   table; alternatif yang lebih sederhana adalah kolom kode terkontrol
   `jenis_jadwal` pada tabel jadwal.
2. Ubah PIC dari teks bebas menjadi foreign key ke `users.user_id`.
3. Jangan menolak jadwal kedua hanya karena alat yang sama sudah memiliki
   jadwal; validasi duplikasi berdasarkan peralatan dan jenis pemeriksaan.
4. Tambahkan tabel riwayat pelaksanaan. `tanggal_realisasi` saja tidak cukup
   untuk menyimpan hasil dan beberapa kali pelaksanaan tanpa menimpa data.
5. Catat hasil dan perbarui ringkasan jadwal dalam satu transaksi agar tanggal
   jatuh tempo tidak berubah tanpa riwayat yang berhasil tersimpan.
6. Sebelum migrasi, verifikasi tipe ID dan constraint pada database aktual;
   konfigurasi migrasi backend saat ini menonaktifkan pembuatan foreign key
   otomatis oleh GORM.
