package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"backend/models"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	// Load .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, using default environment variables")
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser,
		dbPassword,
		dbHost,
		dbPort,
		dbName,
	)

	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to database: %v", err))
	}

	// Cek apakah tabel-tabel penting sudah ada
	hasUserTable := database.Migrator().HasTable(&models.User{})
	hasRuanganTable := database.Migrator().HasTable(&models.Ruangan{})
	hasLabsTable := database.Migrator().HasTable(&models.Labs{})
	hasPeralatanTable := database.Migrator().HasTable(&models.Peralatan{})
	hasDokumenPeralatanTable := database.Migrator().HasTable(&models.DokumenPeralatan{})
	hasDetailAlatUkurTable := database.Migrator().HasTable(&models.DetailAlatUkur{})
	hasDetailAlatBantuTable := database.Migrator().HasTable(&models.DetailAlatBantu{})
	hasDetailArtefakAcuanTable := database.Migrator().HasTable(&models.DetailArtefakAcuan{})
	hasDetailKomponenPendukungTable := database.Migrator().HasTable(&models.DetailKomponenPendukung{})
	hasKategoriPeralatanTable := database.Migrator().HasTable(&models.KategoriPeralatan{})
	hasKelompokAssetTable := database.Migrator().HasTable(&models.KelompokAsset{})

	if !hasUserTable || !hasRuanganTable || !hasLabsTable || !hasPeralatanTable || !hasDokumenPeralatanTable || !hasDetailAlatUkurTable || !hasDetailAlatBantuTable || !hasDetailArtefakAcuanTable || !hasDetailKomponenPendukungTable || !hasKategoriPeralatanTable || !hasKelompokAssetTable {
		log.Println("Beberapa tabel belum ada. Membuat tabel...")

		err := database.AutoMigrate(
			&models.User{},
			&models.Ruangan{},
			&models.Labs{},
			&models.Peralatan{},
			&models.DetailAlatUkur{},
			&models.DetailAlatBantu{},
			&models.DetailArtefakAcuan{},
			&models.DetailKomponenPendukung{},
			&models.DokumenPeralatan{},
			&models.KategoriPeralatan{},
			&models.KelompokAsset{},
			&models.Notification{},
		)
		if err != nil {
			panic(fmt.Sprintf("Failed to migrate database tables: %v", err))
		}

		log.Println("Semua tabel berhasil dibuat!")
	} else {
		log.Println("Semua tabel sudah ada. AutoMigrate dilewati.")
	}

	DB = database
	SeedDummyData()

	log.Println("Database connected successfully!")
}

func SeedDummyData() {
	if DB == nil {
		return
	}

	var userCount int64
	if err := DB.Model(&models.User{}).Count(&userCount).Error; err != nil {
		log.Printf("Gagal mengecek data user: %v", err)
		return
	}

	if userCount == 0 {
		users := []models.User{
			{
				NIP:      "1980010101",
				Name:     "Admin Utama",
				Email:    "admin@sikepo.local",
				Role:     "admin",
				Position: "Administrator",
				PIC:      true,
			},
			{
				NIP:      "1980010102",
				Name:     "Manager Lab",
				Email:    "manager@sikepo.local",
				Role:     "manager",
				Position: "Manager Laboratorium",
				PIC:      true,
			},
			{
				NIP:      "1980010103",
				Name:     "Staff Lab",
				Email:    "staff@sikepo.local",
				Role:     "staff",
				Position: "Staff Laboratorium",
				PIC:      false,
			},
		}

		for i := range users {
			hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
			if err != nil {
				log.Printf("Gagal hash password user %s: %v", users[i].Email, err)
				return
			}
			users[i].Password = string(hash)
		}

		if err := DB.Create(&users).Error; err != nil {
			log.Printf("Gagal membuat data dummy users: %v", err)
			return
		}

		log.Println("Data dummy users berhasil dibuat")
	}

	var kategoriCount int64
	if err := DB.Model(&models.KategoriPeralatan{}).Count(&kategoriCount).Error; err != nil {
		log.Printf("Gagal mengecek data kategori peralatan: %v", err)
		return
	}

	if kategoriCount == 0 {
		kategoris := []models.KategoriPeralatan{
			{NamaKategori: "Alat Ukur", Description: "Peralatan untuk pengukuran dan kalibrasi"},
			{NamaKategori: "Alat Bantu", Description: "Peralatan pendukung operasional laboratorium"},
			{NamaKategori: "Artefak Acuan", Description: "Standar acuan dan referensi pengukuran"},
			{NamaKategori: "Komponen Pendukung", Description: "Komponen pendukung dan bahan konsumsi"},
		}

		if err := DB.Create(&kategoris).Error; err != nil {
			log.Printf("Gagal membuat data dummy kategori peralatan: %v", err)
			return
		}

		log.Println("Data dummy kategori peralatan berhasil dibuat")
	}

	var labCount int64
	if err := DB.Model(&models.Labs{}).Count(&labCount).Error; err != nil {
		log.Printf("Gagal mengecek data lab: %v", err)
		return
	}

	if labCount == 0 {
		var manager models.User
		if err := DB.Where("role = ?", "manager").First(&manager).Error; err != nil {
			log.Printf("Gagal mengambil user manager untuk seed lab: %v", err)
			return
		}

		labs := []models.Labs{
			{NamaLabs: "Lab IQA", KodeLabs: "IQA", ManagerID: &manager.UserID},
			{NamaLabs: "Lab DES", KodeLabs: "DES", ManagerID: &manager.UserID},
			{NamaLabs: "Lab SSA", KodeLabs: "SSA", ManagerID: &manager.UserID},
			{NamaLabs: "TIM", KodeLabs: "TIM", ManagerID: &manager.UserID},
		}

		if err := DB.Create(&labs).Error; err != nil {
			log.Printf("Gagal membuat data dummy labs: %v", err)
			return
		}

		log.Println("Data dummy labs berhasil dibuat")
	}

	var assetCount int64
	if err := DB.Model(&models.KelompokAsset{}).Count(&assetCount).Error; err != nil {
		log.Printf("Gagal mengecek data kelompok asset: %v", err)
		return
	}

	if assetCount == 0 {
		var admin models.User
		if err := DB.Where("role = ?", "admin").First(&admin).Error; err != nil {
			log.Printf("Gagal mengambil user admin untuk seed kelompok asset: %v", err)
			return
		}

		var labs []models.Labs
		if err := DB.Order("id ASC").Find(&labs).Error; err != nil {
			log.Printf("Gagal mengambil data labs untuk seed asset: %v", err)
			return
		}
		if len(labs) == 0 {
			log.Println("Lab belum tersedia, seed kelompok asset dilewati")
			return
		}

		if len(labs) < 4 {
			log.Println("Data lab belum lengkap, seed kelompok asset dilewati")
			return
		}

		assets := []models.KelompokAsset{
			{LabID: labs[0].ID, PICID: admin.UserID, Kode: "FBA", Nama: "FBA"},
			{LabID: labs[0].ID, PICID: admin.UserID, Kode: "SFT", Nama: "SFT"},
			{LabID: labs[0].ID, PICID: admin.UserID, Kode: "ENE", Nama: "ENE"},
			{LabID: labs[1].ID, PICID: admin.UserID, Kode: "DEV", Nama: "DEV"},
			{LabID: labs[1].ID, PICID: admin.UserID, Kode: "TRA", Nama: "TRA"},
			{LabID: labs[2].ID, PICID: admin.UserID, Kode: "KAL", Nama: "KAL"},
		}

		if err := DB.Create(&assets).Error; err != nil {
			log.Printf("Gagal membuat data dummy kelompok asset: %v", err)
			return
		}

		log.Println("Data dummy kelompok asset berhasil dibuat")
	}

	var roomCount int64
	if err := DB.Model(&models.Ruangan{}).Count(&roomCount).Error; err != nil {
		log.Printf("Gagal mengecek data ruangan: %v", err)
		return
	}

	if roomCount == 0 {
		var staff models.User
		if err := DB.Where("role = ?", "staff").First(&staff).Error; err != nil {
			log.Printf("Gagal mengambil user staff untuk seed ruangan: %v", err)
			return
		}

		var labs []models.Labs
		if err := DB.Order("id ASC").Find(&labs).Error; err != nil {
			log.Printf("Gagal mengambil data labs untuk seed ruangan: %v", err)
			return
		}
		if len(labs) == 0 {
			log.Println("Lab belum tersedia, seed ruangan dilewati")
			return
		}

		if len(labs) < 4 {
			log.Println("Data lab belum lengkap, seed ruangan dilewati")
			return
		}

		ruanganList := []models.Ruangan{
			{NamaRuangan: "Lab Uji Optik", KodeRuangan: "IQA-01", LantaiRuangan: "1", LabsID: &labs[0].ID, PICUserID: &staff.UserID},
			{NamaRuangan: "Lab Uji Mekanik", KodeRuangan: "IQA-02", LantaiRuangan: "1", LabsID: &labs[0].ID, PICUserID: &staff.UserID},
			{NamaRuangan: "Lab Uji Lingkungan", KodeRuangan: "IQA-03", LantaiRuangan: "1", LabsID: &labs[0].ID, PICUserID: &staff.UserID},
			{NamaRuangan: "Lab Uji Material", KodeRuangan: "IQA-04", LantaiRuangan: "1", LabsID: &labs[0].ID, PICUserID: &staff.UserID},
			{NamaRuangan: "Lab Electrical Safety", KodeRuangan: "IQA-05", LantaiRuangan: "1", LabsID: &labs[0].ID, PICUserID: &staff.UserID},
			{NamaRuangan: "Lab Laser Safety", KodeRuangan: "IQA-06", LantaiRuangan: "1", LabsID: &labs[0].ID, PICUserID: &staff.UserID},
			{NamaRuangan: "Lab Energy", KodeRuangan: "IQA-07", LantaiRuangan: "1", LabsID: &labs[0].ID, PICUserID: &staff.UserID},
			{NamaRuangan: "Lab Battery", KodeRuangan: "IQA-08", LantaiRuangan: "1", LabsID: &labs[0].ID, PICUserID: &staff.UserID},
			{NamaRuangan: "Gedung Anechoic Chamber", KodeRuangan: "IQA-09", LantaiRuangan: "1", LabsID: &labs[0].ID, PICUserID: &staff.UserID},
			{NamaRuangan: "Lab Radio", KodeRuangan: "DES-01", LantaiRuangan: "1", LabsID: &labs[1].ID, PICUserID: &staff.UserID},
			{NamaRuangan: "Lab Device", KodeRuangan: "DES-02", LantaiRuangan: "1", LabsID: &labs[1].ID, PICUserID: &staff.UserID},
			{NamaRuangan: "Lab Metro", KodeRuangan: "DES-03", LantaiRuangan: "1", LabsID: &labs[1].ID, PICUserID: &staff.UserID},
			{NamaRuangan: "Lab Transport", KodeRuangan: "DES-04", LantaiRuangan: "1", LabsID: &labs[1].ID, PICUserID: &staff.UserID},
			{NamaRuangan: "Lab Access", KodeRuangan: "DES-05", LantaiRuangan: "1", LabsID: &labs[1].ID, PICUserID: &staff.UserID},
			{NamaRuangan: "Lab Kalibrasi", KodeRuangan: "SSA-01", LantaiRuangan: "1", LabsID: &labs[2].ID, PICUserID: &staff.UserID},
			{NamaRuangan: "Gudang", KodeRuangan: "TIM-01", LantaiRuangan: "1", LabsID: &labs[3].ID, PICUserID: &staff.UserID},
		}

		if err := DB.Create(&ruanganList).Error; err != nil {
			log.Printf("Gagal membuat data dummy ruangan: %v", err)
			return
		}

		log.Println("Data dummy ruangan berhasil dibuat")
	}

	var peralatanCount int64
	if err := DB.Model(&models.Peralatan{}).Count(&peralatanCount).Error; err != nil {
		log.Printf("Gagal mengecek data peralatan: %v", err)
		return
	}

	if peralatanCount == 0 {
		var kategori []models.KategoriPeralatan
		if err := DB.Order("id ASC").Find(&kategori).Error; err != nil {
			log.Printf("Gagal mengambil data kategori peralatan: %v", err)
			return
		}
		if len(kategori) == 0 {
			log.Println("Kategori peralatan belum tersedia, seed peralatan dilewati")
			return
		}

		var room []models.Ruangan
		if err := DB.Order("id ASC").Find(&room).Error; err != nil {
			log.Printf("Gagal mengambil data ruangan: %v", err)
			return
		}
		if len(room) == 0 {
			log.Println("Ruangan belum tersedia, seed peralatan dilewati")
			return
		}

		var asset []models.KelompokAsset
		if err := DB.Order("id ASC").Find(&asset).Error; err != nil {
			log.Printf("Gagal mengambil data kelompok asset: %v", err)
			return
		}
		if len(asset) == 0 {
			log.Println("Kelompok asset belum tersedia, seed peralatan dilewati")
			return
		}

		var staff models.User
		if err := DB.Where("role = ?", "staff").First(&staff).Error; err != nil {
			log.Printf("Gagal mengambil user staff untuk seed peralatan: %v", err)
			return
		}

		peralatanList := []models.Peralatan{
			{
				NomorAset:           "AST-001",
				NamaPeralatan:       "Multimeter Digital",
				KategoriID:          1,
				KategoriPeralatanID: 1,
				KelompokAsetID:      uint(asset[0].ID),
				RuanganID:           uint(room[0].ID),
				PICID:               uint(staff.UserID),
				Merek:               "Fluke",
				TipeModel:           "87V",
				NomorSeri:           "FLU-001",
				StatusAlat:          "Aktif",
				Keterangan:          "Alat ukur multimeter digital",
			},
			{
				NomorAset:           "AST-002",
				NamaPeralatan:       "Thermohygrometer",
				KategoriID:          2,
				KategoriPeralatanID: 2,
				KelompokAsetID:      uint(asset[1].ID),
				RuanganID:           uint(room[1].ID),
				PICID:               uint(staff.UserID),
				Merek:               "Extech",
				TipeModel:           "RHT10",
				NomorSeri:           "EXT-002",
				StatusAlat:          "Aktif",
				Keterangan:          "Alat bantu monitoring lingkungan",
			},
			{
				NomorAset:           "AST-003",
				NamaPeralatan:       "Blok Kalibrasi Standar",
				KategoriID:          3,
				KategoriPeralatanID: 3,
				KelompokAsetID:      uint(asset[0].ID),
				RuanganID:           uint(room[0].ID),
				PICID:               uint(staff.UserID),
				Merek:               "Ressolar",
				TipeModel:           "BK-01",
				NomorSeri:           "RES-003",
				StatusAlat:          "Aktif",
				Keterangan:          "Artefak acuan untuk kalibrasi",
			},
			{
				NomorAset:           "AST-004",
				NamaPeralatan:       "Timbangan Analitik",
				KategoriID:          4,
				KategoriPeralatanID: 4,
				KelompokAsetID:      uint(asset[2].ID),
				RuanganID:           uint(room[1].ID),
				PICID:               uint(staff.UserID),
				Merek:               "Mettler",
				TipeModel:           "MS-200",
				NomorSeri:           "MET-004",
				StatusAlat:          "Aktif",
				Keterangan:          "Komponen pendukung analitik",
			},
		}

		if err := DB.Create(&peralatanList).Error; err != nil {
			log.Printf("Gagal membuat data dummy peralatan: %v", err)
			return
		}

		log.Println("Data dummy peralatan berhasil dibuat")

		var created []models.Peralatan
		if err := DB.Order("id ASC").Find(&created).Error; err != nil {
			log.Printf("Gagal mengambil data peralatan terbuat: %v", err)
			return
		}

		if len(created) >= 4 {
			kalibrasi := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
			pemeriksaan := time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC)
			karakterisasi := time.Date(2024, 3, 5, 0, 0, 0, 0, time.UTC)
			terima := time.Date(2024, 4, 12, 0, 0, 0, 0, time.UTC)
			kedaluwarsa := time.Date(2026, 4, 12, 0, 0, 0, 0, time.UTC)

			if err := DB.Create(&models.DetailAlatUkur{
				PeralatanID:          created[0].ID,
				ParameterRentangUkur: "Tegangan, Arus, Resistansi",
				Resolusi:             "0.1 mV",
				AkurasiSpesifikasi:   "±0.05%",
				Satuan:               "V",
				PerantiLunakVersi:    "1.2.0",
				MetodeKelayakan:      "Kalibrasi referensi",
				NoSertifikat:         "SER-UK-001",
				TglKalibrasi:         &kalibrasi,
				IntervalBulan:        12,
				NilaiKoreksi:         "0.02",
				Ketidakpastian:       "0.01%",
				JenisLabel:           "Digital",
				StatusKelayakan:      "Layak",
			}).Error; err != nil {
				log.Printf("Gagal membuat detail alat ukur: %v", err)
			}

			if err := DB.Create(&models.DetailAlatBantu{
				PeralatanID:              created[1].ID,
				FungsiKegunaan:           "Monitoring kelembaban dan suhu",
				PerantiLunakVersi:        "2.5.0",
				JenisPemeriksaanBerkala:  "Pemeriksaan lingkungan",
				KriteriaPemeriksaan:      "Validasi sensor setiap bulan",
				TglPemeriksaanTerakhir:   &pemeriksaan,
				IntervalBulan:            6,
				KarakteristikAcuan:       "Sensor 0-100% RH",
				JadwalKarakterisasiUlang: "6 bulan",
			}).Error; err != nil {
				log.Printf("Gagal membuat detail alat bantu: %v", err)
			}

			if err := DB.Create(&models.DetailArtefakAcuan{
				PeralatanID:                   created[2].ID,
				JenisDeskripsi:                "Blok kalibrasi referensi",
				KarakteristikYangDiacu:        "Tegangan dan arus stabil",
				NilaiSpesifikasiKarakterisasi: "0.05%",
				MetodeKarakterisasi:           "Standar nasional",
				NoLaporanKarakterisasi:        "LPK-003",
				TglKarakterisasiTerakhir:      &karakterisasi,
				IntervalBulan:                 12,
				KondisiPenyimpanan:            "Rak tertutup, suhu terkontrol",
			}).Error; err != nil {
				log.Printf("Gagal membuat detail artefak acuan: %v", err)
			}

			if err := DB.Create(&models.DetailKomponenPendukung{
				PeralatanID:               created[3].ID,
				SubKategori:               "Komponen bobot presisi",
				DeskripsiSpesifikasi:      "Timbangan presisi untuk sampel analitik",
				SumberPemasok:             "PT Metrikindo",
				NoLotBatchEdisi:           "LOT-001",
				GradeMutu:                 "A",
				SatuanKemasan:             "Unit",
				TglTerimaTerbit:           &terima,
				TglKedaluwarsa:            &kedaluwarsa,
				KondisiPenyimpanan:        "Kering dan bersih",
				PengaruhThdKeabsahanHasil: true,
				StatusKetersediaan:        "Tersedia",
			}).Error; err != nil {
				log.Printf("Gagal membuat detail komponen pendukung: %v", err)
			}
		}
	}
}
