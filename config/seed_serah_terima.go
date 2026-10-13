package config

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
	"time"

	"backend/models"
	"backend/utils"

	"gorm.io/gorm"
)

// SeedSerahTerima menambah peralatan contoh dan pengajuan pada setiap tahap
// serah terima keluar: menunggu checklist, siap diserahkan, sedang dipinjam,
// dan dibatalkan karena butir TS. Berjalan bila SEED_SERAH_TERIMA=true di .env.
//
// Berbeda dengan SeedPeminjaman, seeder ini tidak peduli tabel peminjaman sudah
// berisi atau belum. Ia hanya dilewati bila alat contoh DEM-001 sudah ada,
// sehingga aman dijalankan pada database yang sudah dipakai menguji. Gunakan
// hanya di database lokal.
func SeedSerahTerima() {
	if os.Getenv("SEED_SERAH_TERIMA") != "true" || DB == nil {
		return
	}

	if err := seedSerahTerima(DB); err != nil {
		log.Printf("Seeder serah terima gagal: %v", err)
	}
}

type seedAlatDemo struct {
	nomorAset string
	nama      string
	merek     string
	tipe      string
	kategori  uint
	picStaff2 bool
}

// Dua belas alat contoh. Delapan pertama dipakai pengajuan di bawah, empat
// terakhir dibiarkan bebas supaya bisa dipakai menguji pengajuan baru.
var daftarAlatDemo = []seedAlatDemo{
	{"DEM-001", "Osiloskop Digital", "Keysight", "DSOX1204G", 1, false},
	{"DEM-002", "Optical Power Meter", "EXFO", "FPM-300", 1, true},
	{"DEM-003", "Spectrum Analyzer", "Rigol", "DSA815", 1, false},
	{"DEM-004", "Signal Generator", "Rigol", "DSG815", 1, true},
	{"DEM-005", "OTDR", "EXFO", "FTB-1v2", 1, false},
	{"DEM-006", "Clamp Meter", "Fluke", "376 FC", 1, false},
	{"DEM-007", "Visual Fault Locator", "Fluke", "VisiFault", 2, true},
	{"DEM-008", "Function Generator", "Rigol", "DG1022Z", 1, false},
	{"DEM-009", "Data Logger Suhu dan Kelembapan", "Testo", "175 H1", 2, false},
	{"DEM-010", "Power Supply DC", "Keysight", "E36312A", 2, false},
	{"DEM-011", "Optical Fiber Cleaver", "Fujikura", "CT-30", 2, true},
	{"DEM-012", "Network Analyzer", "Keysight", "E5063A", 1, true},
}

const (
	tahapMenungguChecklist = "disetujui"
	tahapSiapDiserahkan    = "siap"
	tahapSedangDipinjam    = "dipinjam"
	tahapDibatalkanTS      = "batal"
)

type seedKasusSerah struct {
	alat      int
	staff2    bool // true bila peminjamnya staff2, selain itu staff1
	tahap     string
	mulaiHari int
	lamaHari  int
	eksternal bool
	tujuan    string
	lokasi    string
}

var daftarKasusSerah = []seedKasusSerah{
	{0, true, tahapMenungguChecklist, 1, 3, false, "Pengujian respons filter frekuensi", "Lab Radio"},
	{1, false, tahapMenungguChecklist, 2, 3, true, "Pengukuran redaman fiber di lokasi pelanggan", "Site pelanggan, Bandung"},
	{2, true, tahapSiapDiserahkan, 0, 3, false, "Analisis spektrum pemancar", "Lab Access"},
	{3, false, tahapSiapDiserahkan, 0, 2, false, "Pengujian sensitivitas penerima", "Lab Device"},
	{4, true, tahapSedangDipinjam, -2, 6, true, "Survei jalur fiber optik antar-gedung", "Site pelanggan, Cimahi"},
	{5, true, tahapSedangDipinjam, -1, 4, false, "Pengukuran arus beban panel uji", "Lab Metro"},
	{6, false, tahapSedangDipinjam, -3, 5, false, "Pelacakan putus fiber pada jumper", "Lab Access"},
	{7, true, tahapDibatalkanTS, -1, 2, false, "Pembangkitan sinyal uji modem", "Lab Radio"},
}

func seedSerahTerima(db *gorm.DB) error {
	var ada int64
	if err := db.Model(&models.Peralatan{}).
		Where("nomor_aset = ?", daftarAlatDemo[0].nomorAset).
		Count(&ada).Error; err != nil {
		return err
	}

	if ada > 0 {
		log.Println("Seeder serah terima dilewati: alat contoh DEM-001 sudah ada")
		return nil
	}

	return db.Transaction(func(tx *gorm.DB) error {
		bahan, err := siapkanBahanSeed(tx)
		if err != nil {
			return err
		}

		alat, err := buatAlatDemo(tx, bahan)
		if err != nil {
			return err
		}

		for _, k := range daftarKasusSerah {
			peminjam := bahan.staff1
			if k.staff2 {
				peminjam = bahan.staff2
			}

			sp := seedPengajuan{
				peminjam:  peminjam,
				alat:      alat[k.alat],
				berhenti:  3,
				mulaiHari: k.mulaiHari,
				lamaHari:  k.lamaHari,
				umurJam:   200,
				tujuan:    k.tujuan,
				lokasi:    k.lokasi,
				eksternal: k.eksternal,
			}

			if err := buatSeedPengajuan(tx, sp); err != nil {
				return err
			}

			var p models.Peminjaman
			if err := tx.
				Where("peralatan_id = ? AND peminjam_id = ?", alat[k.alat].ID, peminjam.UserID).
				Order("id DESC").
				First(&p).Error; err != nil {
				return err
			}

			if err := lanjutkanSerahTerimaSeed(tx, &p, alat[k.alat], k); err != nil {
				return err
			}
		}

		log.Printf("Seeder serah terima selesai: %d alat contoh (DEM-001 sampai DEM-012) dan %d pengajuan dibuat",
			len(alat), len(daftarKasusSerah))
		log.Println("Gunakan staff@sikepo.local dan staff2@sikepo.local, password password123")

		return nil
	})
}

// buatAlatDemo membuat alat contoh beserta detailnya. Ruangan dan kelompok aset
// meminjam dari dua alat dummy bawaan supaya lab pemiliknya pasti punya manager.
func buatAlatDemo(tx *gorm.DB, bahan *seedBahan) ([]*models.Peralatan, error) {
	hasil := make([]*models.Peralatan, 0, len(daftarAlatDemo))

	kalibrasi := time.Now().AddDate(0, -2, 0)
	jatuhTempo := kalibrasi.AddDate(1, 0, 0)

	for i, d := range daftarAlatDemo {
		induk := bahan.alatA
		if i%2 == 1 {
			induk = bahan.alatB
		}

		pic := bahan.staff1
		if d.picStaff2 {
			pic = bahan.staff2
		}

		alat := models.Peralatan{
			NomorAset:           d.nomorAset,
			NamaPeralatan:       d.nama,
			KategoriID:          d.kategori,
			KategoriPeralatanID: d.kategori,
			KelompokAsetID:      induk.KelompokAsetID,
			RuanganID:           induk.RuanganID,
			PICID:               uint(pic.UserID),
			Merek:               d.merek,
			TipeModel:           d.tipe,
			NomorSeri:           fmt.Sprintf("%s-%03d", d.merek[:3], i+1),
			StatusAlat:          "Aktif",
			StatusVerifikasi:    "Disetujui",
			Keterangan:          "Alat contoh untuk uji serah terima",
		}

		if err := tx.Create(&alat).Error; err != nil {
			return nil, err
		}

		var err error

		if d.kategori == 1 {
			err = tx.Create(&models.DetailAlatUkur{
				PeralatanID:     alat.ID,
				MetodeKelayakan: "kalibrasi eksternal",
				NoSertifikat:    fmt.Sprintf("SER-DEM-%03d", i+1),
				TglKalibrasi:    &kalibrasi,
				TglJatuhTempo:   &jatuhTempo,
				IntervalBulan:   12,
				JenisLabel:      "calibration",
				StatusKelayakan: "Layak",
			}).Error
		} else {
			err = tx.Create(&models.DetailAlatBantu{
				PeralatanID:             alat.ID,
				FungsiKegunaan:          "Alat bantu pengujian",
				JenisPemeriksaanBerkala: "verifikasi fungsi",
				KriteriaPemeriksaan:     "Fungsi dasar normal",
				TglPemeriksaanTerakhir:  &kalibrasi,
				TglJatuhTempo:           &jatuhTempo,
				IntervalBulan:           12,
			}).Error
		}

		if err != nil {
			return nil, err
		}

		hasil = append(hasil, &alat)
	}

	return hasil, nil
}

// lanjutkanSerahTerimaSeed membawa pengajuan DISETUJUI ke tahap serah terima
// yang diminta kasus, lengkap dengan lembar F/010, logbook, dan notifikasi.
func lanjutkanSerahTerimaSeed(tx *gorm.DB, p *models.Peminjaman, alat *models.Peralatan, k seedKasusSerah) error {
	if k.tahap == tahapMenungguChecklist {
		return nil
	}

	var pengelola, peminjam models.User
	if err := tx.Where("user_id = ?", p.PengelolaID).First(&pengelola).Error; err != nil {
		return err
	}

	if err := tx.Where("user_id = ?", p.PeminjamID).First(&peminjam).Error; err != nil {
		return err
	}

	sekarang := time.Now()
	diperiksa := sekarang.Add(-2 * time.Hour)
	if k.tahap == tahapSedangDipinjam {
		diperiksa = sekarang.Add(-time.Duration(-k.mulaiHari*24-6) * time.Hour)
	}

	diterima := diperiksa.Add(30 * time.Minute)
	nama := fmt.Sprintf("%s (%s)", alat.NamaPeralatan, alat.NomorAset)

	// ---------- lembar F/010 ----------

	items := make([]models.SerahTerimaItem, 0, len(models.DaftarButirSerahTerima))
	butirTS := ""

	for _, b := range models.DaftarButirSerahTerima {
		hasil, keterangan := models.HasilSesuai, ""

		switch b.Nomor {
		case 4:
			keterangan = "Kabel daya, probe, adaptor, tas pembawa"
		case 6, 7:
			hasil = models.HasilNA
		}

		if k.tahap == tahapDibatalkanTS && b.Nomor == 3 {
			hasil = models.HasilTidakSesuai
			keterangan = "Casing retak dan konektor port longgar"
			butirTS = fmt.Sprintf("butir %d, %s", b.Nomor, keterangan)
		}

		items = append(items, models.SerahTerimaItem{
			NomorButir: b.Nomor,
			Butir:      b.Butir,
			Hasil:      hasil,
			Keterangan: keterangan,
		})
	}

	serah := models.SerahTerima{
		PeminjamanID:  p.ID,
		DiperiksaOleh: p.PengelolaID,
		DiperiksaAt:   diperiksa,
		PengelolaTTD:  ttdContoh(1),
		AdaTS:         k.tahap == tahapDibatalkanTS,
		Items:         items,
	}

	if k.tahap == tahapSedangDipinjam {
		serah.DiterimaOleh = &p.PeminjamID
		serah.DiterimaAt = &diterima
		serah.PeminjamTTD = ttdContoh(2)
	}

	if err := tx.Create(&serah).Error; err != nil {
		return err
	}

	catat := func(aksi, status string, userID uint64, keterangan string, waktu time.Time) error {
		return tx.Create(&models.LogbookPeralatan{
			PeralatanID: p.PeralatanID,
			Jenis:       models.LogbookJenisPeminjaman,
			ReferensiID: &p.ID,
			Aksi:        aksi,
			Status:      status,
			UserID:      &userID,
			Keterangan:  keterangan,
			CreatedAt:   waktu,
		}).Error
	}

	switch k.tahap {
	case tahapSiapDiserahkan:
		if err := tx.Model(&models.Peminjaman{}).Where("id = ?", p.ID).
			Update("status", models.StatusPeminjamanSiapDiserahkan).Error; err != nil {
			return err
		}

		if err := catat(models.LogbookAksiChecklistKeluar, models.StatusPeminjamanSiapDiserahkan, p.PengelolaID,
			fmt.Sprintf("Pemeriksaan serah terima keluar oleh Pengelola Peralatan (%s). Tidak ada butir tidak sesuai. Menunggu konfirmasi terima dari peminjam.",
				pengelola.Name), diperiksa); err != nil {
			return err
		}

		return utils.SendNotification(tx, p.PeminjamID, "serah_terima_menunggu_ttd", "Peralatan siap diserahkan",
			fmt.Sprintf("Pengelola sudah memeriksa %s untuk pengajuan %s. Silakan periksa checklist dan konfirmasi penerimaan.",
				nama, p.Kode))

	case tahapSedangDipinjam:
		if err := tx.Model(&models.Peminjaman{}).Where("id = ?", p.ID).
			Updates(map[string]interface{}{
				"status":            models.StatusPeminjamanSedangDipinjam,
				"tgl_keluar_aktual": diterima,
			}).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.Peralatan{}).Where("id = ?", alat.ID).
			Update("status_alat", "Dipinjam").Error; err != nil {
			return err
		}

		if err := catat(models.LogbookAksiChecklistKeluar, models.StatusPeminjamanSiapDiserahkan, p.PengelolaID,
			fmt.Sprintf("Pemeriksaan serah terima keluar oleh Pengelola Peralatan (%s). Tidak ada butir tidak sesuai. Menunggu konfirmasi terima dari peminjam.",
				pengelola.Name), diperiksa); err != nil {
			return err
		}

		if err := catat(models.LogbookAksiSerahTerimaKeluar, models.StatusPeminjamanSedangDipinjam, p.PeminjamID,
			fmt.Sprintf("Peralatan diterima oleh %s. Lokasi penggunaan %s, rencana kembali %s.",
				peminjam.Name, p.LokasiPenggunaan, formatTanggalSeed(p.RencanaTanggalKembali)), diterima); err != nil {
			return err
		}

		return utils.SendNotification(tx, p.PengelolaID, "peralatan_diserahkan", "Peralatan telah diterima peminjam",
			fmt.Sprintf("%s mengonfirmasi penerimaan %s untuk pengajuan %s. Rencana kembali %s.",
				peminjam.Name, nama, p.Kode, formatTanggalSeed(p.RencanaTanggalKembali)))

	case tahapDibatalkanTS:
		alasan := "Pemeriksaan serah terima keluar menemukan ketidaksesuaian: " + butirTS

		if err := tx.Model(&models.Peminjaman{}).Where("id = ?", p.ID).
			Updates(map[string]interface{}{
				"status":            models.StatusPeminjamanDibatalkanTS,
				"alasan_pembatalan": alasan,
			}).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.Peralatan{}).Where("id = ?", alat.ID).
			Update("status_alat", "Do Not Use").Error; err != nil {
			return err
		}

		if err := catat(models.LogbookAksiChecklistKeluar, models.StatusPeminjamanDibatalkanTS, p.PengelolaID,
			fmt.Sprintf("Pemeriksaan serah terima keluar oleh Pengelola Peralatan (%s). Ditemukan butir TS.", pengelola.Name),
			diperiksa); err != nil {
			return err
		}

		if err := catat(models.LogbookAksiDibatalkanChecklist, models.StatusPeminjamanDibatalkanTS, p.PengelolaID,
			alasan+". Peralatan ditandai Do Not Use dan perlu ditangani sesuai TLKM13/IK/012.",
			diperiksa.Add(time.Second)); err != nil {
			return err
		}

		if err := utils.SendNotification(tx, p.ManagerLabID, "peralatan_do_not_use", "Peralatan ditandai Do Not Use",
			fmt.Sprintf("%s ditandai Do Not Use saat serah terima pengajuan %s. %s. Peralatan perlu ditangani sesuai TLKM13/IK/012.",
				nama, p.Kode, alasan)); err != nil {
			return err
		}

		return utils.SendNotification(tx, p.PeminjamID, "peminjaman_dibatalkan", "Peminjaman dibatalkan",
			fmt.Sprintf("Pengajuan %s untuk %s dibatalkan karena peralatan tidak lolos pemeriksaan saat serah terima.",
				p.Kode, nama))
	}

	return nil
}

func formatTanggalSeed(t time.Time) string {
	return t.Format("2006-01-02")
}

// ttdContoh menggambar tanda tangan coretan sederhana sebagai PNG data URL,
// supaya gambar tanda tangan di lembar serah terima terlihat saat demo.
func ttdContoh(variasi int) string {
	img := image.NewNRGBA(image.Rect(0, 0, 240, 90))
	warna := color.NRGBA{R: 20, G: 30, B: 90, A: 255}

	titik := func(x, y int) {
		for dx := 0; dx < 2; dx++ {
			for dy := 0; dy < 2; dy++ {
				img.SetNRGBA(x+dx, y+dy, warna)
			}
		}
	}

	prevX, prevY := 10, 45

	for x := 10; x < 230; x++ {
		gelombang := math.Sin(float64(x) / float64(6+variasi*3))
		y := 45 + int(gelombang*float64(22-variasi*4)) + int(math.Sin(float64(x)/2.5)*4)

		langkah := int(math.Max(math.Abs(float64(x-prevX)), math.Abs(float64(y-prevY))))
		if langkah == 0 {
			langkah = 1
		}

		for s := 0; s <= langkah; s++ {
			titik(prevX+(x-prevX)*s/langkah, prevY+(y-prevY)*s/langkah)
		}

		prevX, prevY = x, y
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return ""
	}

	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}
