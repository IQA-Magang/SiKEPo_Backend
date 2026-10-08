package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"backend/models"
	"backend/utils"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedPeminjaman mengisi data contoh peminjaman untuk pengembangan frontend
// dan demo. Hanya berjalan bila SEED_PEMINJAMAN=true di .env dan tabel
// peminjaman masih kosong.
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
			status:     &p.ManajerPeminjamStatus, catatan: &p.ManajerPeminjamCatatan, at: &p.ManajerPeminjamAt,
		},
		{
			label: "Pengelola Peralatan", kata: "Diperiksa dan diteruskan ke Manager Lab",
			aksiSetuju: models.LogbookAksiDiteruskanPengelola, aksiTolak: models.LogbookAksiDitolakPengelola,
			menunggu: models.StatusPeminjamanMenungguPengelola, statusLanjut: models.StatusPeminjamanMenungguManagerLab,
			approverID: p.PengelolaID,
			status:     &p.PengelolaStatus, catatan: &p.PengelolaCatatan, at: &p.PengelolaAt,
		},
		{
			label: "Manager Lab", kata: "Disetujui",
			aksiSetuju: models.LogbookAksiDisetujuiManagerLab, aksiTolak: models.LogbookAksiDitolakManagerLab,
			menunggu: models.StatusPeminjamanMenungguManagerLab, statusLanjut: models.StatusPeminjamanDisetujui,
			approverID: p.ManagerLabID,
			status:     &p.ManagerLabStatus, catatan: &p.ManagerLabCatatan, at: &p.ManagerLabAt,
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
			if i == 1 {
				*t.catatan = "Spesifikasi dan kelengkapan sesuai"
			}
			selesai[i] = waktu

		case i == berhenti && sp.ditolak:
			*t.status = models.TahapRejected
			*t.at = &waktu
			*t.catatan = alasanTolak[i]
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
