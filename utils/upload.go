package utils

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func SaveUploadedFile(file *multipart.FileHeader, folder, prefix string) (string, error) {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext == "" {
		return "", fmt.Errorf("file harus memiliki ekstensi")
	}

	directory := filepath.Join("public", "uploads", folder)
	if err := os.MkdirAll(directory, 0755); err != nil {
		return "", err
	}

	filename := fmt.Sprintf("%s-%d%s", prefix, time.Now().UnixNano(), ext)
	destination := filepath.Join(directory, filename)

	source, err := file.Open()
	if err != nil {
		return "", err
	}
	defer source.Close()

	target, err := os.Create(destination)
	if err != nil {
		return "", err
	}
	defer target.Close()

	if _, err := io.Copy(target, source); err != nil {
		return "", err
	}

	return "/static/uploads/" + folder + "/" + filename, nil
}
