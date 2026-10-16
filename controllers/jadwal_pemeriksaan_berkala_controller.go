package controllers

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"backend/repositories"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type JadwalPemeriksaanBerkalaController struct {
	repo repositories.JadwalPemeriksaanBerkalaRepository
}

func NewJadwalPemeriksaanBerkalaController(
	repo repositories.JadwalPemeriksaanBerkalaRepository,
) *JadwalPemeriksaanBerkalaController {
	return &JadwalPemeriksaanBerkalaController{repo: repo}
}

type CreateJadwalRequest struct {
	PeralatanID         uint   `json:"peralatan_id"`
	JenisUtama          string `json:"jenis_utama"`
	FrekuensiPenggunaan string `json:"frekuensi_penggunaan"`
	IntervalUtama       int    `json:"interval_utama"`
	IntervalTambahan    int    `json:"interval_tambahan"`
	TanggalTerakhir     string `json:"tanggal_terakhir"`
	Metode              string `json:"metode"`
	NomorDokumen        string `json:"nomor_dokumen"`
	AlatStandar         string `json:"alat_standar_benda_uji"`
	JenisPemeliharaan   string `json:"jenis_pemeliharaan"`
	PIC                 string `json:"pic"`
	Keterangan          string `json:"keterangan"`
}

func (c *JadwalPemeriksaanBerkalaController) Create(
	ctx *fiber.Ctx,
) error {
	var req CreateJadwalRequest

	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format request tidak valid",
		})
	}

	if req.PeralatanID == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Peralatan wajib dipilih",
		})
	}

	if req.IntervalUtama <= 0 || req.IntervalTambahan <= 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Interval pemeriksaan harus lebih dari 0",
		})
	}

	tanggal, err := time.Parse("2006-01-02", req.TanggalTerakhir)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Tanggal terakhir harus berformat YYYY-MM-DD",
		})
	}

	err = c.repo.CreatePair(
		req.PeralatanID,
		req.JenisUtama,
		req.FrekuensiPenggunaan,
		req.IntervalUtama,
		req.IntervalTambahan,
		tanggal,
		req.Metode,
		req.NomorDokumen,
		req.AlatStandar,
		req.JenisPemeliharaan,
		req.PIC,
		req.Keterangan,
	)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Jadwal utama dan tambahan berhasil dibuat",
	})
}

func (c *JadwalPemeriksaanBerkalaController) FindAll(
	ctx *fiber.Ctx,
) error {
	data, err := c.repo.FindAll()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data jadwal",
		})
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}

func (c *JadwalPemeriksaanBerkalaController) FindByID(
	ctx *fiber.Ctx,
) error {
	id64, err := strconv.ParseUint(ctx.Params("id"), 10, 32)
	if err != nil || id64 == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "ID jadwal tidak valid",
		})
	}

	data, err := c.repo.FindByID(uint(id64))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Jadwal tidak ditemukan",
			})
		}

		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil detail jadwal",
		})
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}

type RealisasiRequest struct {
	TanggalRealisasi string `json:"tanggal_realisasi"`
}

func (c *JadwalPemeriksaanBerkalaController) UpdateRealisasi(
	ctx *fiber.Ctx,
) error {
	id64, err := strconv.ParseUint(ctx.Params("id"), 10, 32)
	if err != nil || id64 == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "ID jadwal tidak valid",
		})
	}

	var req RealisasiRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format request tidak valid",
		})
	}

	tanggal, err := time.Parse(
		"2006-01-02",
		strings.TrimSpace(req.TanggalRealisasi),
	)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Tanggal realisasi harus berformat YYYY-MM-DD",
		})
	}

	err = c.repo.UpdateRealisasi(uint(id64), tanggal)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Jadwal tidak ditemukan",
			})
		}

		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal memperbarui realisasi jadwal",
		})
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"message": "Realisasi dan jadwal berikutnya berhasil diperbarui",
	})
}
