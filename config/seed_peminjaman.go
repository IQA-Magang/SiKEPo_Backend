package config

import (
	"bytes"
	"encoding/base64"
	"errors"
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

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedPeminjaman mengisi data contoh peminjaman untuk pengembangan frontend
// dan demo. Hanya berjalan bila SEED_PEMINJAMAN=true di .env. Pengajuan tahap
// persetujuan hanya dibuat bila tabel peminjaman masih kosong, sedangkan data
// serah terima dan pengembalian (alat DEM-001 sampai DEM-014) dibuat bila alat
// DEM-001 belum ada.
//
// Seeder ini juga menyiapkan data pendukung pada database dummy: lab untuk
// staff, satu manager dan satu staff tambahan di lab SSA, serta dua alat yang
// layak pinjam. Gunakan hanya di database lokal.
func SeedPeminjaman() {
	if os.Getenv("SEED_PEMINJAMAN") != "true" || DB == nil {
		return
	}

	if err := seedPeminjaman(DB); err != nil {
		log.Printf("Seeder peminjaman gagal: %v", err)
	}

	// Data serah terima dan pengembalian berjalan sesudahnya dengan flag yang sama.
	if err := seedSerahTerima(DB); err != nil {
		log.Printf("Seeder serah terima gagal: %v", err)
	}
}

type seedBahan struct {
	staff1 *models.User
	staff2 *models.User
	alatA  *models.Peralatan
	alatB  *models.Peralatan
}

// seedPengajuan menggambarkan satu pengajuan contoh.
type seedPengajuan struct {
	peminjam *models.User
	alat     *models.Peralatan

	// berhenti adalah tahap yang sedang menunggu atau yang menolak:
	// 0 Manajer Peminjam, 1 Pengelola, 2 Manager Lab, 3 sudah disetujui penuh.
	berhenti int
	ditolak  bool

	alternatif *models.Peralatan

	mulaiHari int
	lamaHari  int
	umurJam   int

	tujuan    string
	lokasi    string
	eksternal bool
}

func seedPeminjaman(db *gorm.DB) error {
	var total int64
	if err := db.Unscoped().Model(&models.Peminjaman{}).Count(&total).Error; err != nil {
		return err
	}

	if total > 0 {
		log.Println("Seeder peminjaman dilewati: tabel peminjaman sudah berisi data")
		return nil
	}

	return db.Transaction(func(tx *gorm.DB) error {
		bahan, err := siapkanBahanSeed(tx)
		if err != nil {
			return err
		}

		specs := []seedPengajuan{
			{peminjam: bahan.staff2, alat: bahan.alatA, berhenti: 0, mulaiHari: 10, lamaHari: 2, umurJam: 3,
				tujuan: "Pengujian respons filter frekuensi", lokasi: "Lab Radio"},
			{peminjam: bahan.staff2, alat: bahan.alatB, berhenti: 1, mulaiHari: 10, lamaHari: 2, umurJam: 20,
				tujuan: "Monitoring suhu dan kelembapan ruang uji", lokasi: "Lab Metro"},
			{peminjam: bahan.staff1, alat: bahan.alatA, berhenti: 2, mulaiHari: 20, lamaHari: 2, umurJam: 30,
				tujuan: "Kalibrasi internal perangkat uji", lokasi: "Lab Kalibrasi"},
			{peminjam: bahan.staff1, alat: bahan.alatB, berhenti: 3, mulaiHari: 20, lamaHari: 3, umurJam: 52,
				tujuan: "Pengukuran lapangan di lokasi pelanggan", lokasi: "Site pelanggan, Bandung", eksternal: true},
			{peminjam: bahan.staff2, alat: bahan.alatA, berhenti: 1, ditolak: true, alternatif: bahan.alatB,
				mulaiHari: 15, lamaHari: 2, umurJam: 75,
				tujuan: "Pengujian sinyal komunikasi", lokasi: "Lab Access"},
			{peminjam: bahan.staff1, alat: bahan.alatB, berhenti: 2, ditolak: true,
				mulaiHari: 25, lamaHari: 2, umurJam: 98,
				tujuan: "Verifikasi kondisi lingkungan", lokasi: "Lab Device"},
		}

		for _, spec := range specs {
			if err := buatSeedPengajuan(tx, spec); err != nil {
				return err
			}
		}

		log.Printf("Seeder peminjaman selesai: %d pengajuan contoh dibuat", len(specs))
		log.Println("Akun contoh tambahan: manager2@sikepo.local dan staff2@sikepo.local, password password123")

		return nil
	})
}

// =====================================================
// DATA PENDUKUNG
// =====================================================

func cariUserSeed(tx *gorm.DB, email string) (*models.User, error) {
	var user models.User

	if err := tx.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user %s tidak ditemukan, seeder dummy utama harus sudah berjalan: %w", email, err)
	}

	return &user, nil
}

func cariLabSeed(tx *gorm.DB, kode string) (*models.Labs, error) {
	var lab models.Labs

	if err := tx.Where("kode_labs = ?", kode).First(&lab).Error; err != nil {
		return nil, fmt.Errorf("lab %s tidak ditemukan: %w", kode, err)
	}

	return &lab, nil
}

// pastikanUserSeed mengambil user berdasarkan email, atau membuatnya bila belum ada.
func pastikanUserSeed(tx *gorm.DB, user models.User) (*models.User, error) {
	var existing models.User

	result := tx.Where("email = ?", user.Email).Limit(1).Find(&existing)
	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected > 0 {
		return &existing, nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user.Password = string(hash)

	if err := tx.Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func siapkanBahanSeed(tx *gorm.DB) (*seedBahan, error) {
	manager1, err := cariUserSeed(tx, "manager@sikepo.local")
	if err != nil {
		return nil, err
	}

	staff1, err := cariUserSeed(tx, "staff@sikepo.local")
	if err != nil {
		return nil, err
	}

	labIQA, err := cariLabSeed(tx, "IQA")
	if err != nil {
		return nil, err
	}

	labSSA, err := cariLabSeed(tx, "SSA")
	if err != nil {
		return nil, err
	}

	if staff1.LabsID == nil {
		if err := tx.Model(staff1).Update("labs_id", labIQA.ID).Error; err != nil {
			return nil, err
		}
		staff1.LabsID = &labIQA.ID
	}

	manager2, err := pastikanUserSeed(tx, models.User{
		NIP:      "1980010104",
		Name:     "Manager Lab SSA",
		Email:    "manager2@sikepo.local",
		Role:     "manager",
		Position: "Manager SSA",
	})
	if err != nil {
		return nil, err
	}

	// Lab SSA dipindahkan ke manager kedua hanya bila masih dipegang manager
	// dummy bawaan, agar data lab yang sudah diatur sendiri tidak tertimpa.
	if labSSA.ManagerID != nil && *labSSA.ManagerID == manager1.UserID {
		if err := tx.Model(labSSA).Update("manager_id", manager2.UserID).Error; err != nil {
			return nil, err
		}
	}

	staff2, err := pastikanUserSeed(tx, models.User{
		NIP:      "1980010105",
		Name:     "Staff SSA",
		Email:    "staff2@sikepo.local",
		Role:     "staff",
		Position: "Staff Laboratorium SSA",
		LabsID:   &labSSA.ID,
	})
	if err != nil {
		return nil, err
	}

	if staff2.LabsID == nil {
		if err := tx.Model(staff2).Update("labs_id", labSSA.ID).Error; err != nil {
			return nil, err
		}
		staff2.LabsID = &labSSA.ID
	}

	// Dua alat ukur atau alat bantu yang layak pinjam. Pengelolanya harus
	// berbeda dari staff kedua supaya peminjam dan Pengelola berbeda orang.
	var alat []models.Peralatan

	cariAlat := func() error {
		return tx.
			Where(
				"deleted_at IS NULL AND status_alat = ? AND status_verifikasi = ? AND kategori_peralatan_id IN ? AND pic_id <> 0 AND pic_id <> ?",
				"Aktif", "Disetujui", []uint{1, 2}, staff2.UserID,
			).
			Order("id").
			Limit(2).
			Find(&alat).
			Error
	}

	if err := cariAlat(); err != nil {
		return nil, err
	}

	if len(alat) < 2 {
		if err := tx.Model(&models.Peralatan{}).
			Where("nomor_aset IN ?", []string{"AST-001", "AST-002"}).
			Updates(map[string]interface{}{
				"status_alat":       "Aktif",
				"status_verifikasi": "Disetujui",
			}).Error; err != nil {
			return nil, err
		}

		if err := cariAlat(); err != nil {
			return nil, err
		}
	}

	if len(alat) < 2 {
		return nil, errors.New("butuh dua peralatan alat ukur atau alat bantu yang layak pinjam")
	}

	return &seedBahan{staff1: staff1, staff2: staff2, alatA: &alat[0], alatB: &alat[1]}, nil
}

// =====================================================
// PENGAJUAN CONTOH
// =====================================================

type seedTahap struct {
	label        string
	kata         string
	aksiSetuju   string
	aksiTolak    string
	menunggu     string
	statusLanjut string
	approverID   uint64
	status       *string
	catatan      *string
	at           **time.Time
	ttd          *string
}

func buatSeedPengajuan(tx *gorm.DB, sp seedPengajuan) error {
	labPeminjam, err := utils.GetLabDenganManager(tx, *sp.peminjam.LabsID)
	if err != nil {
		return err
	}

	labPemilik, err := utils.GetLabPemilikPeralatan(tx, sp.alat)
	if err != nil {
		return err
	}

	dilewati := *labPeminjam.ManagerID == *labPemilik.ManagerID

	berhenti := sp.berhenti
	if dilewati && berhenti == 0 {
		berhenti = 1
	}

	sekarang := time.Now()
	dibuat := sekarang.Add(-time.Duration(sp.umurJam) * time.Hour)
	keluar := time.Date(sekarang.Year(), sekarang.Month(), sekarang.Day()+sp.mulaiHari, 0, 0, 0, 0, time.Local)
	kembali := keluar.AddDate(0, 0, sp.lamaHari)

	p := models.Peminjaman{
		PeralatanID:           uint64(sp.alat.ID),
		PeminjamID:            sp.peminjam.UserID,
		TujuanPenggunaan:      sp.tujuan,
		LokasiPenggunaan:      sp.lokasi,
		IsEksternal:           sp.eksternal,
		RencanaTanggalKeluar:  keluar,
		RencanaTanggalKembali: kembali,
		KebutuhanKelengkapan:  "Probe dan kabel penghubung",
		KebutuhanAksesori:     "Adaptor daya",
		LabPeminjamID:         labPeminjam.ID,
		LabPemilikID:          labPemilik.ID,
		ManajerPeminjamID:     *labPeminjam.ManagerID,
		PengelolaID:           uint64(sp.alat.PICID),
		ManagerLabID:          *labPemilik.ManagerID,
		CreatedAt:             dibuat,
		UpdatedAt:             dibuat,
	}

	tahap := []seedTahap{
		{
			label: "Manajer Peminjam", kata: "Divalidasi",
			aksiSetuju: models.LogbookAksiDivalidasiManajerPinjam, aksiTolak: models.LogbookAksiDitolakManajerPinjam,
			menunggu: models.StatusPeminjamanMenungguManajerPeminjam, statusLanjut: models.StatusPeminjamanMenungguPengelola,
			approverID: p.ManajerPeminjamID,
			status:     &p.ManajerPeminjamStatus, catatan: &p.ManajerPeminjamCatatan, at: &p.ManajerPeminjamAt, ttd: &p.ManajerPeminjamTTD,
		},
		{
			label: "Pengelola Peralatan", kata: "Diperiksa dan diteruskan ke Manager Lab",
			aksiSetuju: models.LogbookAksiDiteruskanPengelola, aksiTolak: models.LogbookAksiDitolakPengelola,
			menunggu: models.StatusPeminjamanMenungguPengelola, statusLanjut: models.StatusPeminjamanMenungguManagerLab,
			approverID: p.PengelolaID,
			status:     &p.PengelolaStatus, catatan: &p.PengelolaCatatan, at: &p.PengelolaAt, ttd: &p.PengelolaTTD,
		},
		{
			label: "Manager Lab", kata: "Disetujui",
			aksiSetuju: models.LogbookAksiDisetujuiManagerLab, aksiTolak: models.LogbookAksiDitolakManagerLab,
			menunggu: models.StatusPeminjamanMenungguManagerLab, statusLanjut: models.StatusPeminjamanDisetujui,
			approverID: p.ManagerLabID,
			status:     &p.ManagerLabStatus, catatan: &p.ManagerLabCatatan, at: &p.ManagerLabAt, ttd: &p.ManagerLabTTD,
		},
	}

	alasanTolak := []string{
		"Jadwal bentrok dengan kegiatan lain",
		"Rentang ukur alat tidak sesuai kebutuhan pengujian",
		"Alat dibutuhkan untuk kegiatan laboratorium pada tanggal tersebut",
	}

	statusAwal := models.StatusPeminjamanMenungguManajerPeminjam
	if dilewati {
		statusAwal = models.StatusPeminjamanMenungguPengelola
	}

	p.Status = models.StatusPeminjamanDisetujui
	if berhenti < 3 {
		p.Status = tahap[berhenti].menunggu
	}
	if sp.ditolak {
		p.Status = models.StatusPeminjamanDitolak
	}

	// selesai menandai tahap yang sudah diputuskan beserta waktunya.
	selesai := make([]time.Time, len(tahap))

	for i := range tahap {
		t := &tahap[i]
		waktu := dibuat.Add(time.Duration(i+1) * 45 * time.Minute)

		switch {
		case i == 0 && dilewati && berhenti > 0:
			*t.status = models.TahapSkipped

		case i < berhenti:
			*t.status = models.TahapApproved
			*t.at = &waktu
			*t.ttd = ttdContoh(i + 1)
			if i == 1 {
				*t.catatan = "Spesifikasi dan kelengkapan sesuai"
				lolos := true
				p.PengelolaCekSpesifikasi = &lolos
				p.PengelolaCekOperator = &lolos
				if sp.eksternal {
					p.PengelolaCekLokasi = &lolos
				}
			}
			selesai[i] = waktu

		case i == berhenti && sp.ditolak:
			*t.status = models.TahapRejected
			*t.at = &waktu
			*t.catatan = alasanTolak[i]
			if i == 1 {
				lolos, gagal := true, false
				p.PengelolaCekSpesifikasi = &gagal
				p.PengelolaCekOperator = &lolos
			}
			selesai[i] = waktu

		default:
			*t.status = models.TahapPending
		}
	}

	if sp.ditolak && berhenti == 1 && sp.alternatif != nil {
		id := uint64(sp.alternatif.ID)
		p.AlatAlternatifID = &id
	}

	if err := tx.Create(&p).Error; err != nil {
		return err
	}

	kode := fmt.Sprintf("PJM-%d-%04d", p.CreatedAt.Year(), p.ID)
	if err := tx.Model(&models.Peminjaman{}).Where("id = ?", p.ID).Update("kode", kode).Error; err != nil {
		return err
	}
	p.Kode = kode

	// ---------- logbook ----------

	catat := func(aksi string, status string, userID *uint64, keterangan string, waktu time.Time) error {
		return tx.Create(&models.LogbookPeralatan{
			PeralatanID: p.PeralatanID,
			Jenis:       models.LogbookJenisPeminjaman,
			ReferensiID: &p.ID,
			Aksi:        aksi,
			Status:      status,
			UserID:      userID,
			Keterangan:  keterangan,
			CreatedAt:   waktu,
		}).Error
	}

	peminjamID := sp.peminjam.UserID
	jenis := "internal"
	if sp.eksternal {
		jenis = "eksternal"
	}

	if err := catat(models.LogbookAksiPengajuan, statusAwal, &peminjamID,
		fmt.Sprintf("Diajukan oleh %s untuk %s, peminjaman %s, lokasi penggunaan %s, rencana %s sampai %s. Tujuan: %s.",
			sp.peminjam.Name, sp.alat.NamaPeralatan, jenis, sp.lokasi,
			keluar.Format("2006-01-02"), kembali.Format("2006-01-02"), sp.tujuan),
		dibuat); err != nil {
		return err
	}

	if dilewati {
		if err := catat(models.LogbookAksiManajerPinjamDilewati, statusAwal, nil,
			"Tahap Manajer Peminjam dilewati karena Manajer Peminjam sama dengan Manager Lab pemilik peralatan.",
			dibuat); err != nil {
			return err
		}
	}

	for i := range tahap {
		t := tahap[i]

		if *t.status != models.TahapApproved && *t.status != models.TahapRejected {
			continue
		}

		var approver models.User
		if err := tx.Where("user_id = ?", t.approverID).First(&approver).Error; err != nil {
			return err
		}

		uid := t.approverID

		if *t.status == models.TahapApproved {
			ket := fmt.Sprintf("%s oleh %s (%s).", t.kata, t.label, approver.Name)
			if *t.catatan != "" {
				ket += fmt.Sprintf(" Catatan: %s.", *t.catatan)
			}

			if err := catat(t.aksiSetuju, t.statusLanjut, &uid, ket, selesai[i]); err != nil {
				return err
			}
			continue
		}

		ket := fmt.Sprintf("Ditolak oleh %s (%s). Alasan: %s.", t.label, approver.Name, *t.catatan)
		if sp.alternatif != nil {
			ket += fmt.Sprintf(" Usulan alat lain: %s (%s).", sp.alternatif.NamaPeralatan, sp.alternatif.NomorAset)
		}

		if err := catat(t.aksiTolak, models.StatusPeminjamanDitolak, &uid, ket, selesai[i]); err != nil {
			return err
		}
	}

	// ---------- notifikasi: satu untuk pihak yang perlu bertindak atau tahu ----------

	nama := fmt.Sprintf("%s (%s)", sp.alat.NamaPeralatan, sp.alat.NomorAset)

	switch {
	case sp.ditolak:
		return utils.SendNotification(tx, p.PeminjamID, "peminjaman_ditolak", "Pengajuan peminjaman ditolak",
			fmt.Sprintf("Pengajuan %s untuk %s ditolak oleh %s. Alasan: %s.",
				p.Kode, nama, tahap[berhenti].label, *tahap[berhenti].catatan))

	case p.Status == models.StatusPeminjamanDisetujui:
		return utils.SendNotification(tx, p.PeminjamID, "peminjaman_disetujui", "Peminjaman disetujui",
			fmt.Sprintf("Pengajuan %s untuk %s disetujui Manager Lab.", p.Kode, nama))
	}

	jenisNotifikasi := []string{"peminjaman_diajukan", "peminjaman_divalidasi", "peminjaman_diteruskan"}

	return utils.SendNotification(tx, tahap[berhenti].approverID, jenisNotifikasi[berhenti],
		"Pengajuan peminjaman menunggu keputusan Anda",
		fmt.Sprintf("Pengajuan %s oleh %s untuk %s menunggu keputusan Anda sebagai %s.",
			p.Kode, sp.peminjam.Name, nama, tahap[berhenti].label))
}

// =====================================================
// SERAH TERIMA DAN PENGEMBALIAN
// =====================================================

// seedAlatDemo adalah alat contoh DEM-001 sampai DEM-014. Delapan pertama
// dipakai pengajuan serah terima, dua terakhir pengajuan yang sudah
// dikembalikan, dan sisanya dibiarkan bebas untuk menguji pengajuan baru.
type seedAlatDemo struct {
	nomorAset string
	nama      string
	merek     string
	tipe      string
	kategori  uint
	picStaff2 bool
}

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
	{"DEM-013", "LAN Cable Tester", "Fluke", "MS2-100", 2, false},
	{"DEM-014", "Digital Caliper", "Mitutoyo", "500-196-30", 1, false},
}

const (
	tahapMenungguChecklist = "disetujui"
	tahapSiapDiserahkan    = "siap"
	tahapSedangDipinjam    = "dipinjam"
	tahapDibatalkanTS      = "batal"

	// Sudah dikembalikan. Dua hasil tindak lanjut: menunggu verifikasi ulang,
	// atau Do Not Use karena ada kejadian.
	tahapDikembalikanVerifikasi = "kembali_verifikasi"
	tahapDikembalikanDNU        = "kembali_dnu"
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
	{12, true, tahapDikembalikanVerifikasi, -5, 2, false, "Pengujian kabel jaringan gedung", "Lab Access"},
	{13, true, tahapDikembalikanDNU, -4, 2, false, "Pengukuran dimensi komponen uji", "Lab Uji Mekanik"},
}

// seedSerahTerima menambah peralatan contoh dan pengajuan pada setiap tahap
// serah terima dan pengembalian: menunggu checklist, siap diserahkan, sedang
// dipinjam, dibatalkan karena butir TS, dan sudah dikembalikan.
//
// Tidak peduli tabel peminjaman sudah berisi atau belum. Hanya dilewati bila
// alat contoh DEM-001 sudah ada, sehingga aman dijalankan pada database yang
// sudah dipakai menguji.
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

		log.Printf("Seeder serah terima selesai: %d alat contoh (DEM-001 sampai DEM-014) dan %d pengajuan dibuat",
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

	// Alat yang sudah dikembalikan juga pernah melewati tahap sedang dipinjam.
	pernahDipinjam := k.tahap == tahapSedangDipinjam ||
		k.tahap == tahapDikembalikanVerifikasi ||
		k.tahap == tahapDikembalikanDNU

	sekarang := time.Now()
	diperiksa := sekarang.Add(-2 * time.Hour)
	if pernahDipinjam {
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

	if pernahDipinjam {
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

	case tahapSedangDipinjam, tahapDikembalikanVerifikasi, tahapDikembalikanDNU:
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

		if err := utils.SendNotification(tx, p.PengelolaID, "peralatan_diserahkan", "Peralatan telah diterima peminjam",
			fmt.Sprintf("%s mengonfirmasi penerimaan %s untuk pengajuan %s. Rencana kembali %s.",
				peminjam.Name, nama, p.Kode, formatTanggalSeed(p.RencanaTanggalKembali))); err != nil {
			return err
		}

		if k.tahap == tahapSedangDipinjam {
			return nil
		}

		return kembalikanSeed(tx, p, alat, k, &serah, pengelola, peminjam, diterima.Add(48*time.Hour))

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

// kembalikanSeed mencatat pengembalian untuk alat contoh: kolom Kembali pada
// lembar F/010, penutupan peminjaman, status alat, logbook, dan notifikasi.
// Hasilnya sama dengan yang dibuat endpoint pengembalian.
func kembalikanSeed(
	tx *gorm.DB,
	p *models.Peminjaman,
	alat *models.Peralatan,
	k seedKasusSerah,
	serah *models.SerahTerima,
	pengelola, peminjam models.User,
	kembaliAt time.Time,
) error {
	dnu := k.tahap == tahapDikembalikanDNU

	uraian := ""
	alasan := ""
	tindakLanjut := models.TindakLanjutPerluVerifikasi

	if dnu {
		uraian = "Alat terjatuh saat diangkut, casing retak"
		alasan = "terjadi kejadian selama peminjaman: " + uraian +
			"; kondisi tidak sesuai saat kembali (butir 3, Casing retak dan konektor port longgar)"
		tindakLanjut = models.TindakLanjutDoNotUse
	}

	for _, item := range serah.Items {
		hasil, ket := item.Hasil, item.Keterangan

		if dnu && item.NomorButir == 3 {
			hasil, ket = models.HasilTidakSesuai, "Casing retak dan konektor port longgar"
		}

		if err := tx.Model(&models.SerahTerimaItem{}).Where("id = ?", item.ID).
			Updates(map[string]interface{}{
				"hasil_kembali":      hasil,
				"keterangan_kembali": ket,
			}).Error; err != nil {
			return err
		}
	}

	if err := tx.Model(&models.SerahTerima{}).Where("id = ?", serah.ID).
		Updates(map[string]interface{}{
			"kembali_diterima_oleh": p.PengelolaID,
			"kembali_at":            kembaliAt,
			"kembali_pengelola_ttd": ttdContoh(1),
			"kembali_peminjam_ttd":  ttdContoh(2),
			"ada_kejadian":          dnu,
			"uraian_kejadian":       uraian,
			"kembali_ada_ts":        dnu,
			"tindak_lanjut":         tindakLanjut,
			"alasan_tindak_lanjut":  alasan,
		}).Error; err != nil {
		return err
	}

	if err := tx.Model(&models.Peminjaman{}).Where("id = ?", p.ID).
		Updates(map[string]interface{}{
			"status":             models.StatusPeminjamanDikembalikan,
			"tgl_kembali_aktual": kembaliAt,
		}).Error; err != nil {
		return err
	}

	perubahanAlat := map[string]interface{}{
		"status_alat":       "Karantina",
		"status_verifikasi": "Belum Diverifikasi",
	}
	if dnu {
		perubahanAlat = map[string]interface{}{"status_alat": "Do Not Use"}
	}

	if err := tx.Model(&models.Peralatan{}).Where("id = ?", alat.ID).
		Updates(perubahanAlat).Error; err != nil {
		return err
	}

	nama := fmt.Sprintf("%s (%s)", alat.NamaPeralatan, alat.NomorAset)

	catat := func(aksi string, keterangan string, waktu time.Time) error {
		uid := p.PengelolaID

		return tx.Create(&models.LogbookPeralatan{
			PeralatanID: p.PeralatanID,
			Jenis:       models.LogbookJenisPeminjaman,
			ReferensiID: &p.ID,
			Aksi:        aksi,
			Status:      models.StatusPeminjamanDikembalikan,
			UserID:      &uid,
			Keterangan:  keterangan,
			CreatedAt:   waktu,
		}).Error
	}

	ket := fmt.Sprintf("Peralatan dikembalikan oleh %s dan diterima Pengelola Peralatan (%s).",
		peminjam.Name, pengelola.Name)
	if dnu {
		ket += " Kejadian: " + uraian + "."
	}

	if err := catat(models.LogbookAksiPengembalian, ket, kembaliAt); err != nil {
		return err
	}

	if dnu {
		if err := catat(models.LogbookAksiDihentikanPengembalian,
			"Peralatan ditandai Do Not Use karena "+alasan+". Perlu ditangani sesuai TLKM13/IK/012.",
			kembaliAt.Add(time.Second)); err != nil {
			return err
		}

		if err := utils.SendNotification(tx, p.ManagerLabID, "peralatan_do_not_use", "Peralatan ditandai Do Not Use",
			fmt.Sprintf("%s ditandai Do Not Use saat pengembalian pengajuan %s karena %s. Peralatan perlu ditangani sesuai TLKM13/IK/012.",
				nama, p.Kode, alasan)); err != nil {
			return err
		}

		return utils.SendNotification(tx, p.PeminjamID, "peralatan_dikembalikan", "Pengembalian peralatan dicatat",
			fmt.Sprintf("Pengembalian %s untuk pengajuan %s sudah dicatat. Peralatan dihentikan penggunaannya karena %s.",
				nama, p.Kode, alasan))
	}

	if err := catat(models.LogbookAksiPerluVerifikasi,
		"Peralatan masuk Karantina dan perlu diverifikasi ulang sesuai TLKM13/IK/003 sebelum dipakai kembali.",
		kembaliAt.Add(time.Second)); err != nil {
		return err
	}

	return utils.SendNotification(tx, p.PeminjamID, "peralatan_dikembalikan", "Pengembalian peralatan dicatat",
		fmt.Sprintf("Pengembalian %s untuk pengajuan %s sudah dicatat. Peralatan akan diverifikasi ulang oleh Pengelola.",
			nama, p.Kode))
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
