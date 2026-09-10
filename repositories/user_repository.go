package repositories

import (
	"errors"

	"backend/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrEmailAlreadyExists = errors.New("email sudah digunakan")
	ErrNIPAlreadyExists   = errors.New("NIP sudah digunakan")
)

type UserRepository struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		DB: db,
	}
}

// ==============================
// GET ALL
// ==============================

func (r *UserRepository) GetAllUsers() ([]models.User, error) {

	var users []models.User

	err := r.DB.
		Order("user_id DESC").
		Find(&users).
		Error

	return users, err
}

// ==============================
// GET BY ID
// ==============================

func (r *UserRepository) GetUserByID(id uint64) (*models.User, error) {

	var user models.User

	err := r.DB.
		Where("user_id = ?", id).
		First(&user).
		Error

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// ==============================
// GET BY EMAIL
// ==============================

func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {

	var user models.User

	err := r.DB.
		Where("email = ?", email).
		First(&user).
		Error

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// ==============================
// GET BY NIP
// ==============================

func (r *UserRepository) GetUserByNIP(nip string) (*models.User, error) {

	var user models.User

	err := r.DB.
		Where("nip = ?", nip).
		First(&user).
		Error

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// ==============================
// CREATE
// ==============================

func (r *UserRepository) CreateUser(
	user *models.User,
	password string,
) error {

	// ==============================
	// CEK NIP
	// ==============================

	var nipCount int64

	err := r.DB.
		Model(&models.User{}).
		Where("nip = ?", user.NIP).
		Count(&nipCount).
		Error

	if err != nil {
		return err
	}

	if nipCount > 0 {
		return ErrNIPAlreadyExists
	}

	// ==============================
	// CEK EMAIL
	// ==============================

	var emailCount int64

	err = r.DB.
		Model(&models.User{}).
		Where("email = ?", user.Email).
		Count(&emailCount).
		Error

	if err != nil {
		return err
	}

	if emailCount > 0 {
		return ErrEmailAlreadyExists
	}

	// ==============================
	// HASH PASSWORD
	// ==============================

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		return err
	}

	user.Password = string(passwordHash)

	// ==============================
	// SIMPAN USER
	// ==============================

	return r.DB.Create(user).Error
}

// ==============================
// UPDATE
// ==============================

func (r *UserRepository) UpdateUser(
	id uint64,
	user *models.User,
	password string,
) error {

	var existingUser models.User

	// ==============================
	// CARI USER
	// ==============================

	err := r.DB.
		Where("user_id = ?", id).
		First(&existingUser).
		Error

	if err != nil {
		return err
	}

	// ==============================
	// CEK NIP
	// ==============================

	var nipCount int64

	err = r.DB.
		Model(&models.User{}).
		Where(
			"nip = ? AND user_id != ?",
			user.NIP,
			id,
		).
		Count(&nipCount).
		Error

	if err != nil {
		return err
	}

	if nipCount > 0 {
		return ErrNIPAlreadyExists
	}

	// ==============================
	// CEK EMAIL
	// ==============================

	var emailCount int64

	err = r.DB.
		Model(&models.User{}).
		Where(
			"email = ? AND user_id != ?",
			user.Email,
			id,
		).
		Count(&emailCount).
		Error

	if err != nil {
		return err
	}

	if emailCount > 0 {
		return ErrEmailAlreadyExists
	}

	// ==============================
	// UPDATE DATA
	// ==============================

	existingUser.NIP = user.NIP
	existingUser.Name = user.Name
	existingUser.Email = user.Email
	existingUser.Role = user.Role
	existingUser.Position = user.Position
	existingUser.PIC = user.PIC

	// ==============================
	// UPDATE PASSWORD
	// ==============================

	if password != "" {

		passwordHash, err := bcrypt.GenerateFromPassword(
			[]byte(password),
			bcrypt.DefaultCost,
		)

		if err != nil {
			return err
		}

		existingUser.Password = string(passwordHash)
	}

	// ==============================
	// SIMPAN
	// ==============================

	return r.DB.Save(&existingUser).Error
}

// ==============================
// DELETE
// ==============================

func (r *UserRepository) DeleteUser(id uint64) error {

	var user models.User

	err := r.DB.
		Where("user_id = ?", id).
		First(&user).
		Error

	if err != nil {
		return err
	}

	return r.DB.Delete(&user).Error
}