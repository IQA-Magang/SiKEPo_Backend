package controllers

import (
	"backend/repositories"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type NotificationController struct {
	Repo repositories.NotificationRepository
}

func (c *NotificationController) GetByUserID(ctx *fiber.Ctx) error {
	userID, err := strconv.ParseUint(ctx.Params("user_id"), 10, 64)
	if err != nil || userID == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "ID user tidak valid",
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

func (c *NotificationController) MarkRead(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil || id == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "ID notifikasi tidak valid",
		})
	}

	if err := c.Repo.MarkAsRead(id); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal menandai notifikasi selesai dibaca",
			"error":   err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"status":  "success",
		"message": "Notifikasi berhasil ditandai dibaca",
	})
}
