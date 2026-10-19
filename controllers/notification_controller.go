package controllers

import (
	"backend/repositories"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type NotificationController struct {
	Repo repositories.NotificationRepository
}

// GetByUserID handles GET /api/notifications/user/:user_id.
// Notifikasi bersifat pribadi, jadi user_id harus sama dengan user yang login.
func (c *NotificationController) GetByUserID(ctx *fiber.Ctx) error {
	userID, err := strconv.ParseUint(ctx.Params("user_id"), 10, 64)
	if err != nil || userID == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "ID user tidak valid",
		})
	}

	loginID, err := getAuthenticatedUserID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	if loginID != userID {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status":  "error",
			"message": "Anda hanya dapat melihat notifikasi milik sendiri",
		})
	}

	notifications, err := c.Repo.GetByUserID(userID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengambil notifikasi",
			"error":   err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"status": "success",
		"data":   notifications,
		"count":  len(notifications),
	})
}

// MarkRead handles PATCH /api/notifications/:id/read.
// Hanya pemilik notifikasi yang bisa menandainya dibaca.
func (c *NotificationController) MarkRead(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil || id == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "ID notifikasi tidak valid",
		})
	}

	loginID, err := getAuthenticatedUserID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	found, err := c.Repo.MarkAsReadByOwner(id, loginID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal menandai notifikasi selesai dibaca",
			"error":   err.Error(),
		})
	}

	if !found {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Notifikasi tidak ditemukan",
		})
	}

	return ctx.JSON(fiber.Map{
		"status":  "success",
		"message": "Notifikasi berhasil ditandai dibaca",
	})
}
