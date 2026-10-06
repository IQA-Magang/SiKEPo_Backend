package controllers

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"backend/models"
	"backend/repositories"
	"backend/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PeminjamanController menangani alur peminjaman peralatan sesuai TLKM13/IK/005.
//
// Kode dipecah ke tiga file:
//   - peminjaman_controller.go            : helper, pengajuan, daftar, detail
//   - peminjaman_keputusan_controller.go  : keputusan tiga tahap persetujuan
//   - peminjaman_response.go              : bentuk response untuk frontend
type PeminjamanController struct {
	Repository        *repositories.PeminjamanRepository
	LogbookRepository *repositories.LogbookRepository
}

func NewPeminjamanController(
	repository *repositories.PeminjamanRepository,
	logbookRepository *repositories.LogbookRepository,
) *PeminjamanController {

	return &PeminjamanController{
		Repository:        repository,
		LogbookRepository: logbookRepository,
	}
}

type CreatePeminjamanRequest struct {
	PeralatanID           uint64 `json:"peralatan_id"`
	NamaOperator          string `json:"nama_operator"`
	TujuanPenggunaan      string `json:"tujuan_penggunaan"`
	NomorSPK              string `json:"nomor_spk"`
	LokasiPenggunaan      string `json:"lokasi_penggunaan"`
	IsEksternal           bool   `json:"is_eksternal"`
	RencanaTanggalKeluar  string `json:"rencana_tanggal_keluar"`
	RencanaTanggalKembali string `json:"rencana_tanggal_kembali"`
	KebutuhanKelengkapan  string `json:"kebutuhan_kelengkapan"`
	KebutuhanAksesori     string `json:"kebutuhan_aksesori"`
	Catatan               string `json:"catatan"`
}

// =====================================================
// ERROR HELPER
// =====================================================

// peminjamanError membawa status HTTP, sehingga error validasi yang terjadi di
// dalam transaksi tetap dikembalikan ke user dengan status yang tepat.
type peminjamanError struct {
	status  int
	message string
}

func (e *peminjamanError) Error() string {
	return e.message
}

func errPeminjaman(status int, format string, args ...interface{}) error {
	return &peminjamanError{
		status:  status,
		message: fmt.Sprintf(format, args...),
	}
}

func respondPeminjamanError(ctx *fiber.Ctx, err error, fallback string) error {
	var pe *peminjamanError
	if errors.As(err, &pe) {
		return ctx.Status(pe.status).JSON(fiber.Map{
			"success": false,
			"message": pe.message,
		})
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Data tidak ditemukan",
		})
	}

	return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"success": false,
		"message": fallback,
		"error":   err.Error(),
	})
}

// =====================================================
// HELPER UMUM
// =====================================================

// currentUser mengambil user yang sedang login langsung dari database,
// supaya role-nya selalu yang terbaru dan bukan dari token lama.
func (c *PeminjamanController) currentUser(ctx *fiber.Ctx) (*models.User, error) {
	userID, err := getAuthenticatedUserID(ctx)
	if err != nil {
		return nil, errPeminjaman(fiber.StatusUnauthorized, "%s", err.Error())
	}

	var user models.User
	if err := c.Repository.DB.Where("user_id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errPeminjaman(fiber.StatusUnauthorized, "User tidak ditemukan")
		}
		return nil, err
	}

	return &user, nil
}

func parsePeminjamanID(ctx *fiber.Ctx) (uint64, error) {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil || id == 0 {
		return 0, errPeminjaman(fiber.StatusBadRequest, "ID peminjaman tidak valid")
	}
	return id, nil
}

// tanggalSaja membuang jam, menit, dan detik agar perbandingan tanggal adil.
func tanggalSaja(t time.Time) time.Time {
	t = t.In(time.Local)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
}

func formatTanggalID(t time.Time) string {
	return t.In(time.Local).Format("2006-01-02")
}

// catatLogbook menulis satu baris logbook jenis peminjaman. Kirim transaksi
// sebagai tx supaya baris ini ikut batal bila aksi utamanya gagal. Status yang
// dicatat adalah status peminjaman pada saat fungsi ini dipanggil.
func (c *PeminjamanController) catatLogbook(
	tx *gorm.DB,
	p *models.Peminjaman,
	aksi string,
	userID uint64,
	keterangan string,
) error {

	entry := &models.LogbookPeralatan{
		PeralatanID: p.PeralatanID,
		Jenis:       models.LogbookJenisPeminjaman,
		ReferensiID: &p.ID,
		Aksi:        aksi,
		Status:      p.Status,
		Keterangan:  keterangan,
	}

	if userID != 0 {
		entry.UserID = &userID
	}

	return c.LogbookRepository.Create(tx, entry)
}

// canView: admin, peminjam, dan ketiga penyetuju boleh melihat pengajuan.
func canViewPeminjaman(user *models.User, p *models.Peminjaman) bool {
	return user.Role == "admin" ||
		user.UserID == p.PeminjamID ||
		user.UserID == p.ManajerPeminjamID ||
		user.UserID == p.PengelolaID ||
		user.UserID == p.ManagerLabID
}

// =====================================================
// KELAYAKAN PERALATAN
// =====================================================

// detailKelayakan membaca detail peralatan sesuai kategorinya. Hasilnya adalah
// tanggal jatuh tempo pemeriksaan berikutnya, nama pemeriksaannya, dan alasan
// bila detail itu sendiri menyatakan alat tidak boleh dipinjam.
func detailKelayakan(db *gorm.DB, p *models.Peralatan) (jatuhTempo *time.Time, label string, alasan string, err error) {
	switch p.KategoriPeralatanID {

	case 1:
		var detail models.DetailAlatUkur
		result := db.Where("peralatan_id = ?", p.ID).Limit(1).Find(&detail)
		if result.Error != nil {
			return nil, "", "", result.Error
		}
		if result.RowsAffected == 0 {
			return nil, "", "", nil
		}
		if strings.EqualFold(detail.JenisLabel, "do not use") {
			alasan = "Peralatan berlabel Do Not Use"
		} else if strings.EqualFold(detail.StatusKelayakan, "Tidak layak") {
			alasan = "Peralatan berstatus kelayakan Tidak layak"
		}
		return detail.TglJatuhTempo, "kalibrasi", alasan, nil

	case 2:
		var detail models.DetailAlatBantu
		result := db.Where("peralatan_id = ?", p.ID).Limit(1).Find(&detail)
		if result.Error != nil {
			return nil, "", "", result.Error
		}
		if result.RowsAffected == 0 {
			return nil, "", "", nil
		}
		return detail.TglJatuhTempo, "pemeriksaan berkala", "", nil

	case 3:
		var detail models.DetailArtefakAcuan
		result := db.Where("peralatan_id = ?", p.ID).Limit(1).Find(&detail)
		if result.Error != nil {
			return nil, "", "", result.Error
		}
		if result.RowsAffected == 0 {
			return nil, "", "", nil
		}
		if detail.Status != "" && !strings.EqualFold(detail.Status, "aktif") {
			alasan = fmt.Sprintf("Artefak acuan berstatus %s sehingga tidak dapat dipinjam", detail.Status)
		}
		return detail.TglKarakterisasiUlang, "karakterisasi ulang", alasan, nil
	}

	return nil, "", "", nil
}

// cekKelayakanPeminjaman memastikan peralatan boleh dipinjam sampai rencana
// tanggal kembali, sesuai TLKM13/IK/005 butir 7a dan 7d.
func cekKelayakanPeminjaman(db *gorm.DB, p *models.Peralatan, rencanaKembali time.Time) error {
	if p.KategoriPeralatanID == 4 {
		return errPeminjaman(fiber.StatusBadRequest,
			"Komponen pendukung tidak dapat dipinjam")
	}

	if p.StatusAlat != "Aktif" {
		return errPeminjaman(fiber.StatusBadRequest,
			"Peralatan berstatus %s sehingga tidak dapat dipinjam", p.StatusAlat)
	}

	if p.StatusVerifikasi != "Disetujui" {
		return errPeminjaman(fiber.StatusBadRequest,
			"Peralatan belum lolos verifikasi, status verifikasi: %s", p.StatusVerifikasi)
	}

	jatuhTempo, label, alasan, err := detailKelayakan(db, p)
	if err != nil {
		return err
	}

	if alasan != "" {
		return errPeminjaman(fiber.StatusBadRequest, "%s", alasan)
	}

	if jatuhTempo == nil {
		return nil
	}

	batas := tanggalSaja(*jatuhTempo)

	if batas.Before(tanggalSaja(time.Now())) {
		return errPeminjaman(fiber.StatusBadRequest,
			"Peralatan telah melewati tanggal jatuh tempo %s, yaitu %s", label, formatTanggalID(batas))
	}

	if tanggalSaja(rencanaKembali).After(batas) {
		return errPeminjaman(fiber.StatusBadRequest,
			"Rencana tanggal kembali tidak boleh melewati jatuh tempo %s, yaitu %s", label, formatTanggalID(batas))
	}

	return nil
}

// jatuhTempoPeralatan dipakai untuk menampilkan jatuh tempo pada response.
func (c *PeminjamanController) jatuhTempoPeralatan(p *models.Peralatan) *time.Time {
	jatuhTempo, _, _, err := detailKelayakan(c.Repository.DB, p)
	if err != nil {
		return nil
	}
	return jatuhTempo
}

// =====================================================
// VALIDASI INPUT PENGAJUAN
// =====================================================

func validasiPengajuan(req *CreatePeminjamanRequest) (keluar time.Time, kembali time.Time, err error) {
	if req.PeralatanID == 0 {
		return keluar, kembali, errPeminjaman(fiber.StatusBadRequest, "peralatan_id wajib diisi")
	}

	if req.TujuanPenggunaan == "" {
		return keluar, kembali, errPeminjaman(fiber.StatusBadRequest, "Tujuan penggunaan wajib diisi")
	}

	if req.LokasiPenggunaan == "" {
		return keluar, kembali, errPeminjaman(fiber.StatusBadRequest, "Lokasi penggunaan wajib diisi")
	}

	if req.RencanaTanggalKeluar == "" || req.RencanaTanggalKembali == "" {
		return keluar, kembali, errPeminjaman(fiber.StatusBadRequest,
			"Rencana tanggal keluar dan kembali wajib diisi")
	}

	keluar, err = parseTanggal(req.RencanaTanggalKeluar)
	if err != nil {
		return keluar, kembali, errPeminjaman(fiber.StatusBadRequest,
			"rencana_tanggal_keluar tidak valid, gunakan format YYYY-MM-DD")
	}

	kembali, err = parseTanggal(req.RencanaTanggalKembali)
	if err != nil {
		return keluar, kembali, errPeminjaman(fiber.StatusBadRequest,
			"rencana_tanggal_kembali tidak valid, gunakan format YYYY-MM-DD")
	}

	keluar = tanggalSaja(keluar)
	kembali = tanggalSaja(kembali)

	if keluar.Before(tanggalSaja(time.Now())) {
		return keluar, kembali, errPeminjaman(fiber.StatusBadRequest,
			"Rencana tanggal keluar tidak boleh sebelum hari ini")
	}

	if kembali.Before(keluar) {
		return keluar, kembali, errPeminjaman(fiber.StatusBadRequest,
			"Rencana tanggal kembali tidak boleh sebelum tanggal keluar")
	}

	return keluar, kembali, nil
}

// =====================================================
// CREATE PENGAJUAN
// POST /api/peminjaman
// =====================================================

func (c *PeminjamanController) Create(ctx *fiber.Ctx) error {
	user, err := c.currentUser(ctx)
	if err != nil {
		return respondPeminjamanError(ctx, err, "Gagal mengambil data user")
	}

	if user.Role != "staff" {
		return respondPeminjamanError(ctx,
			errPeminjaman(fiber.StatusForbidden, "Hanya staff yang dapat mengajukan peminjaman"), "")
	}

	var req CreatePeminjamanRequest
	if err := ctx.BodyParser(&req); err != nil {
		return respondPeminjamanError(ctx,
			errPeminjaman(fiber.StatusBadRequest, "Format JSON tidak valid"), "")
	}

	req.NamaOperator = strings.TrimSpace(req.NamaOperator)
	req.TujuanPenggunaan = strings.TrimSpace(req.TujuanPenggunaan)
	req.NomorSPK = strings.TrimSpace(req.NomorSPK)
	req.LokasiPenggunaan = strings.TrimSpace(req.LokasiPenggunaan)
	req.RencanaTanggalKeluar = strings.TrimSpace(req.RencanaTanggalKeluar)
	req.RencanaTanggalKembali = strings.TrimSpace(req.RencanaTanggalKembali)
	req.KebutuhanKelengkapan = strings.TrimSpace(req.KebutuhanKelengkapan)
	req.KebutuhanAksesori = strings.TrimSpace(req.KebutuhanAksesori)
	req.Catatan = strings.TrimSpace(req.Catatan)

	keluar, kembali, err := validasiPengajuan(&req)
	if err != nil {
		return respondPeminjamanError(ctx, err, "")
	}

	var created models.Peminjaman

	err = c.Repository.DB.Transaction(func(tx *gorm.DB) error {
		// Kunci baris peralatan agar dua pengajuan yang bersamaan tidak
		// sama-sama lolos pengecekan jadwal.
		var peralatan models.Peralatan
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND deleted_at IS NULL", req.PeralatanID).
			First(&peralatan).
			Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errPeminjaman(fiber.StatusNotFound, "Peralatan tidak ditemukan")
			}
			return err
		}

		if peralatan.PICID == 0 {
			return errPeminjaman(fiber.StatusBadRequest, "Peralatan belum memiliki Pengelola")
		}

		if err := cekKelayakanPeminjaman(tx, &peralatan, kembali); err != nil {
			return err
		}

		if user.LabsID == nil || *user.LabsID == 0 {
			return errPeminjaman(fiber.StatusBadRequest,
				"Lab asal Anda belum diatur. Hubungi admin")
		}

		labPeminjam, err := utils.GetLabDenganManager(tx, *user.LabsID)
		if err != nil {
			if utils.IsErrLab(err) {
				return errPeminjaman(fiber.StatusBadRequest,
					"Lab asal Anda tidak dapat dipakai: %s. Hubungi admin", err.Error())
			}
			return err
		}

		labPemilik, err := utils.GetLabPemilikPeralatan(tx, &peralatan)
		if err != nil {
			if utils.IsErrLab(err) {
				return errPeminjaman(fiber.StatusBadRequest,
					"Data lab pemilik peralatan belum lengkap: %s. Hubungi admin", err.Error())
			}
			return err
		}

		bentrok, err := c.Repository.HasJadwalBentrok(tx, uint64(peralatan.ID), keluar, kembali, 0)
		if err != nil {
			return err
		}
		if bentrok {
			return errPeminjaman(fiber.StatusConflict,
				"Peralatan sedang dipinjam atau sudah diajukan pada rentang tanggal tersebut")
		}

		// Tahap Manajer Peminjam dilewati bila orangnya sama dengan Manager Lab
		// pemilik, supaya orang yang sama tidak menyetujui dua kali.
		status := models.StatusPeminjamanMenungguManajerPeminjam
		statusManajerPeminjam := models.TahapPending
		dilewati := *labPeminjam.ManagerID == *labPemilik.ManagerID

		if dilewati {
			status = models.StatusPeminjamanMenungguPengelola
			statusManajerPeminjam = models.TahapSkipped
		}

		created = models.Peminjaman{
			PeralatanID:           uint64(peralatan.ID),
			PeminjamID:            user.UserID,
			NamaOperator:          req.NamaOperator,
			TujuanPenggunaan:      req.TujuanPenggunaan,
			NomorSPK:              req.NomorSPK,
			LokasiPenggunaan:      req.LokasiPenggunaan,
			IsEksternal:           req.IsEksternal,
			RencanaTanggalKeluar:  keluar,
			RencanaTanggalKembali: kembali,
			KebutuhanKelengkapan:  req.KebutuhanKelengkapan,
			KebutuhanAksesori:     req.KebutuhanAksesori,
			Catatan:               req.Catatan,
			Status:                status,
			LabPeminjamID:         labPeminjam.ID,
			LabPemilikID:          labPemilik.ID,
			ManajerPeminjamID:     *labPeminjam.ManagerID,
			ManajerPeminjamStatus: statusManajerPeminjam,
			PengelolaID:           uint64(peralatan.PICID),
			PengelolaStatus:       models.TahapPending,
			ManagerLabID:          *labPemilik.ManagerID,
			ManagerLabStatus:      models.TahapPending,
		}

		if err := tx.Create(&created).Error; err != nil {
			return err
		}

		// Kode dibuat dari ID agar pasti unik tanpa perlu penghitung terpisah.
		created.Kode = fmt.Sprintf("PJM-%d-%04d", created.CreatedAt.Year(), created.ID)
		if err := tx.Model(&models.Peminjaman{}).
			Where("id = ?", created.ID).
			Update("kode", created.Kode).Error; err != nil {
			return err
		}

		jenis := "internal"
		if created.IsEksternal {
			jenis = "eksternal"
		}

		if err := c.catatLogbook(tx, &created, models.LogbookAksiPengajuan, user.UserID,
			fmt.Sprintf("Diajukan oleh %s untuk %s, peminjaman %s, lokasi penggunaan %s, rencana %s sampai %s. Tujuan: %s.",
				user.Name, peralatan.NamaPeralatan, jenis, created.LokasiPenggunaan,
				formatTanggalID(keluar), formatTanggalID(kembali), created.TujuanPenggunaan),
		); err != nil {
			return err
		}

		penerimaNotifikasi := created.ManajerPeminjamID
		pesanNotifikasi := "menunggu validasi Anda sebagai Manajer Peminjam"

		if dilewati {
			if err := c.catatLogbook(tx, &created, models.LogbookAksiManajerPinjamDilewati, 0,
				"Tahap Manajer Peminjam dilewati karena Manajer Peminjam sama dengan Manager Lab pemilik peralatan."); err != nil {
				return err
			}

			penerimaNotifikasi = created.PengelolaID
			pesanNotifikasi = "menunggu pemeriksaan Anda sebagai Pengelola Peralatan"
		}

		return utils.SendNotification(tx, penerimaNotifikasi,
			"peminjaman_diajukan",
			"Pengajuan peminjaman peralatan",
			fmt.Sprintf("Pengajuan %s oleh %s untuk %s (%s) %s.",
				created.Kode, user.Name, peralatan.NamaPeralatan, peralatan.NomorAset, pesanNotifikasi),
		)
	})
	if err != nil {
		return respondPeminjamanError(ctx, err, "Gagal menyimpan pengajuan peminjaman")
	}

	result, err := c.Repository.FindByID(created.ID)
	if err != nil {
		return respondPeminjamanError(ctx, err, "Pengajuan tersimpan tetapi gagal mengambil data")
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Pengajuan peminjaman berhasil dikirim",
		"data":    c.buildResponse(result),
	})
}

// =====================================================
// GET ALL
// GET /api/peminjaman?peran=...&status=...&menunggu_saya=true&peralatan_id=...
// =====================================================

func (c *PeminjamanController) GetAll(ctx *fiber.Ctx) error {
	user, err := c.currentUser(ctx)
	if err != nil {
		return respondPeminjamanError(ctx, err, "Gagal mengambil data user")
	}

	peran := strings.ToLower(strings.TrimSpace(ctx.Query("peran")))
	switch peran {
	case "", "peminjam", "manajer_peminjam", "pengelola", "manager_lab":
	default:
		return respondPeminjamanError(ctx,
			errPeminjaman(fiber.StatusBadRequest,
				"peran harus peminjam, manajer_peminjam, pengelola, atau manager_lab"), "")
	}

	filter := repositories.PeminjamanFilter{
		UserID:       user.UserID,
		IsAdmin:      user.Role == "admin",
		Peran:        peran,
		Status:       strings.ToUpper(strings.TrimSpace(ctx.Query("status"))),
		MenungguSaya: ctx.Query("menunggu_saya") == "true",
	}

	if value := ctx.Query("peralatan_id"); value != "" {
		id, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return respondPeminjamanError(ctx,
				errPeminjaman(fiber.StatusBadRequest, "peralatan_id tidak valid"), "")
		}
		filter.PeralatanID = id
	}

	data, err := c.Repository.FindAll(filter)
	if err != nil {
		return respondPeminjamanError(ctx, err, "Gagal mengambil daftar peminjaman")
	}

	// make agar hasil kosong menjadi [] dan bukan null.
	items := make([]peminjamanResp, 0, len(data))
	for i := range data {
		items = append(items, c.buildResponse(&data[i]))
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"data":    items,
	})
}

// =====================================================
// GET BY ID
// GET /api/peminjaman/:id
// =====================================================

func (c *PeminjamanController) GetByID(ctx *fiber.Ctx) error {
	user, err := c.currentUser(ctx)
	if err != nil {
		return respondPeminjamanError(ctx, err, "Gagal mengambil data user")
	}

	id, err := parsePeminjamanID(ctx)
	if err != nil {
		return respondPeminjamanError(ctx, err, "")
	}

	data, err := c.Repository.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return respondPeminjamanError(ctx,
				errPeminjaman(fiber.StatusNotFound, "Peminjaman tidak ditemukan"), "")
		}
		return respondPeminjamanError(ctx, err, "Gagal mengambil data peminjaman")
	}

	if !canViewPeminjaman(user, data) {
		return respondPeminjamanError(ctx,
			errPeminjaman(fiber.StatusForbidden, "Anda tidak memiliki akses ke peminjaman ini"), "")
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"data":    c.buildResponse(data),
	})
}
