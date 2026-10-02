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

// =====================================================
// GET ALL
// =====================================================

func (r *UserRepository) GetAllUsers() ([]models.User, error) {

	var users []models.User

	err := r.DB.
		Order("user_id DESC").
		Find(&users).
		Error

	return users, err
}

func (r *UserRepository) GetUsersByLabsID(labsID uint64) ([]models.User, error) {
	var users []models.User

	err := r.DB.
		Where("labs_id = ?", labsID).
		Order("user_id DESC").
		Find(&users).
		Error

	return users, err
}

func (r *UserRepository) GetUsersByLabsIDs(labsIDs []uint64) ([]models.User, error) {
	var users []models.User
	if len(labsIDs) == 0 {
		return users, nil
	}

	err := r.DB.
		Where("labs_id IN ?", labsIDs).
		Order("user_id DESC").
		Find(&users).
		Error
	return users, err
}

// =====================================================
// GET BY ID
// =====================================================

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

func (r *UserRepository) GetUserByIDAndLabsID(id, labsID uint64) (*models.User, error) {
	var user models.User

	err := r.DB.
		Where("user_id = ? AND labs_id = ?", id, labsID).
		First(&user).
		Error

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetUserByIDAndLabsIDs(id uint64, labsIDs []uint64) (*models.User, error) {
	var user models.User
	if len(labsIDs) == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	err := r.DB.
		Where("user_id = ? AND labs_id IN ?", id, labsIDs).
		First(&user).
		Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// =====================================================
// GET BY EMAIL
// =====================================================

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

// =====================================================
// GET BY NIP
// =====================================================

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

// =====================================================
// CREATE
// =====================================================

func (r *UserRepository) CreateUser(
	user *models.User,
	password string,
) error {

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

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		return err
	}

	user.Password = string(passwordHash)

	return r.DB.Create(user).Error
}

// =====================================================
// UPDATE
// =====================================================

func (r *UserRepository) UpdateUser(
	id uint64,
	user *models.User,
	password string,
) error {

	var existingUser models.User

	err := r.DB.
		Where("user_id = ?", id).
		First(&existingUser).
		Error

	if err != nil {
		return err
	}

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

	existingUser.NIP = user.NIP
	existingUser.Name = user.Name
	existingUser.Email = user.Email
	existingUser.Role = user.Role
	existingUser.Position = user.Position
	existingUser.Pengelola = user.Pengelola

	// =====================================================
	// UPDATE LAB SCOPE
	// =====================================================

	existingUser.LabsID = user.LabsID

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

	return r.DB.Save(&existingUser).Error
}

func (r *UserRepository) SetStaffPengelolaByLabsIDs(
	id uint64,
	labsIDs []uint64,
	pengelola bool,
) (*models.User, error) {
	if len(labsIDs) == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	var staff models.User
	if err := r.DB.
		Where("user_id = ? AND labs_id IN ? AND role = ?", id, labsIDs, "staff").
		First(&staff).
		Error; err != nil {
		return nil, err
	}

	result := r.DB.Model(&models.User{}).
		Where("user_id = ? AND labs_id IN ? AND role = ?", id, labsIDs, "staff").
		Update("pengelola", pengelola)
	if result.Error != nil {
		return nil, result.Error
	}

	var updatedUser models.User
	if err := r.DB.
		Where("user_id = ? AND labs_id IN ? AND role = ?", id, labsIDs, "staff").
		First(&updatedUser).
		Error; err != nil {
		return nil, err
	}
	return &updatedUser, nil
}

// =====================================================
// DELETE
// =====================================================

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
