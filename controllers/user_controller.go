package controllers

import (
	"errors"
	"strconv"
	"strings"

	"backend/config"
	"backend/models"
	"backend/repositories"
	"backend/utils"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserController struct {
	Repository *repositories.UserRepository
}

// ========================================
// GET ALL USERS
// ========================================

func (c *UserController) GetUsers(ctx *fiber.Ctx) error {

	users, err := c.Repository.GetAllUsers()

	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data user",
			"error":   err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"message": "Data user berhasil diambil",
		"data":    users,
	})
}

// ========================================
// GET USER BY ID
// ========================================

func (c *UserController) GetUserByID(ctx *fiber.Ctx) error {

	id, err := strconv.ParseUint(
		ctx.Params("id"),
		10,
		64,
	)

	if err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "ID user tidak valid",
		})
	}

	user, err := c.Repository.GetUserByID(id)

	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.Status(404).JSON(fiber.Map{
				"success": false,
				"message": "User tidak ditemukan",
			})
		}

		return ctx.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data user",
			"error":   err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"message": "Data user berhasil ditemukan",
		"data":    user,
	})
}

// ========================================
// CREATE USER
// ========================================

func (c *UserController) CreateUser(ctx *fiber.Ctx) error {

	type CreateUserRequest struct {
		NIP      string `json:"nip"`
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
		Position string `json:"position"`
	}

	var request CreateUserRequest

	if err := ctx.BodyParser(&request); err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Format JSON tidak valid",
		})
	}

	// Normalisasi
	request.NIP = strings.TrimSpace(request.NIP)
	request.Name = strings.TrimSpace(request.Name)
	request.Email = strings.TrimSpace(request.Email)
	request.Role = strings.TrimSpace(request.Role)
	request.Position = strings.TrimSpace(request.Position)

	// Validasi
	if request.NIP == "" {
		return ctx.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "NIP wajib diisi",
		})
	}

	if request.Name == "" {
		return ctx.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Nama wajib diisi",
		})
	}

	if request.Email == "" {
		return ctx.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Email wajib diisi",
		})
	}

	if request.Password == "" {
		return ctx.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Password wajib diisi",
		})
	}

	if len(request.Password) < 6 {
		return ctx.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Password minimal 6 karakter",
		})
	}

	if !isValidRole(request.Role) {
		return ctx.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Role harus admin, staff, atau manager",
		})
	}

	if request.Position == "" {
		return ctx.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Position wajib diisi",
		})
	}

	// Model
	user := models.User{
		NIP:      request.NIP,
		Name:     request.Name,
		Email:    request.Email,
		Role:     request.Role,
		Position: request.Position,
	}

	// Create
	err := c.Repository.CreateUser(
		&user,
		request.Password,
	)

	if err != nil {

		if errors.Is(err, repositories.ErrNIPAlreadyExists) {
			return ctx.Status(409).JSON(fiber.Map{
				"success": false,
				"message": "NIP sudah digunakan",
			})
		}

		if errors.Is(err, repositories.ErrEmailAlreadyExists) {
			return ctx.Status(409).JSON(fiber.Map{
				"success": false,
				"message": "Email sudah digunakan",
			})
		}

		return ctx.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Gagal membuat user",
			"error":   err.Error(),
		})
	}

	return ctx.Status(201).JSON(fiber.Map{
		"success": true,
		"message": "User berhasil dibuat",
		"data":    user,
	})
}

// ========================================
// UPDATE USER
// ========================================

func (c *UserController) UpdateUser(ctx *fiber.Ctx) error {

	id, err := strconv.ParseUint(
		ctx.Params("id"),
		10,
		64,
	)

	if err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "ID user tidak valid",
		})
	}

	type UpdateUserRequest struct {
		NIP      string `json:"nip"`
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
		Position string `json:"position"`
	}

	var request UpdateUserRequest

	if err := ctx.BodyParser(&request); err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Format JSON tidak valid",
		})
	}

	request.NIP = strings.TrimSpace(request.NIP)
	request.Name = strings.TrimSpace(request.Name)
	request.Email = strings.TrimSpace(request.Email)
	request.Role = strings.TrimSpace(request.Role)
	request.Position = strings.TrimSpace(request.Position)

	if request.NIP == "" ||
		request.Name == "" ||
		request.Email == "" ||
		request.Role == "" ||
		request.Position == "" {

		return ctx.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Semua field wajib diisi kecuali password",
		})
	}

	if !isValidRole(request.Role) {
		return ctx.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Role harus admin, staff, atau manager",
		})
	}

	user := models.User{
		NIP:      request.NIP,
		Name:     request.Name,
		Email:    request.Email,
		Role:     request.Role,
		Position: request.Position,
	}

	err = c.Repository.UpdateUser(
		id,
		&user,
		request.Password,
	)

	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.Status(404).JSON(fiber.Map{
				"success": false,
				"message": "User tidak ditemukan",
			})
		}

		if errors.Is(err, repositories.ErrNIPAlreadyExists) {
			return ctx.Status(409).JSON(fiber.Map{
				"success": false,
				"message": "NIP sudah digunakan",
			})
		}

		if errors.Is(err, repositories.ErrEmailAlreadyExists) {
			return ctx.Status(409).JSON(fiber.Map{
				"success": false,
				"message": "Email sudah digunakan",
			})
		}

		return ctx.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengubah user",
			"error":   err.Error(),
		})
	}

	updatedUser, err := c.Repository.GetUserByID(id)

	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "User berhasil diubah tetapi gagal mengambil data terbaru",
		})
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"message": "User berhasil diubah",
		"data":    updatedUser,
	})
}

// ========================================
// DELETE USER
// ========================================

func (c *UserController) DeleteUser(ctx *fiber.Ctx) error {

	id, err := strconv.ParseUint(
		ctx.Params("id"),
		10,
		64,
	)

	if err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "ID user tidak valid",
		})
	}

	err = c.Repository.DeleteUser(id)

	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.Status(404).JSON(fiber.Map{
				"success": false,
				"message": "User tidak ditemukan",
			})
		}

		return ctx.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menghapus user",
			"error":   err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"message": "User berhasil dihapus",
	})
}

// ========================================
// VALIDASI ROLE
// ========================================

func isValidRole(role string) bool {

	switch role {
	case "admin", "staff", "manager":
		return true
	default:
		return false
	}
}

// ========================================
// LOGIN
// ========================================

func (c *UserController) Login(ctx *fiber.Ctx) error {

	type LoginRequest struct {
		Email          string `json:"email"`
		Password       string `json:"password"`
		RecaptchaToken string `json:"recaptcha_token"`
	}

	var req LoginRequest

	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Format JSON tidak valid",
		})
	}

	req.Email = strings.TrimSpace(req.Email)

	if req.Email == "" || req.Password == "" || req.RecaptchaToken == "" {
		return ctx.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Email, password, dan recaptcha_token wajib diisi",
		})
	}

	// Verify recaptcha
	ok, err := config.VerifyRecaptcha(req.RecaptchaToken)
	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Gagal memverifikasi recaptcha",
			"error":   err.Error(),
		})
	}

	if !ok {
		return ctx.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Recaptcha tidak valid",
		})
	}

	user, err := c.Repository.GetUserByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.Status(401).JSON(fiber.Map{
				"success": false,
				"message": "Email atau password salah",
			})
		}

		return ctx.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data user",
			"error":   err.Error(),
		})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return ctx.Status(401).JSON(fiber.Map{
			"success": false,
			"message": "Email atau password salah",
		})
	}

	token, err := utils.CreateToken(user)
	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Gagal membuat token",
			"error":   err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"message": "Login berhasil",
		"data": fiber.Map{
			"token": token,
			"user":  user,
		},
	})
}
