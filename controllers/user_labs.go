package controllers

import "backend/models"

// labsIDValid memeriksa lab yang dikirim saat membuat atau mengubah user.
// Nilai kosong dianggap valid, artinya user belum punya lab.
func (c *UserController) labsIDValid(labsID *uint64) (bool, error) {
	if labsID == nil {
		return true, nil
	}

	var total int64

	err := c.Repository.DB.
		Model(&models.Labs{}).
		Where("id = ?", *labsID).
		Count(&total).
		Error
	if err != nil {
		return false, err
	}

	return total > 0, nil
}
