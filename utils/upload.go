package utils

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// UploadRoot returns the filesystem root shared by upload and static serving.
func UploadRoot() string {
	if configured := strings.TrimSpace(os.Getenv("UPLOAD_DIR")); configured != "" {
		return configured
	}

	if workingDirectory, err := os.Getwd(); err == nil {
		candidate := filepath.Join(workingDirectory, "docs")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	_, sourceFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(sourceFile), "..", "docs")
}

func SaveUploadedFile(file *multipart.FileHeader, folder, prefix string) (string, error) {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext == "" {
		return "", fmt.Errorf("file harus memiliki ekstensi")
	}

	directory := filepath.Join(UploadRoot(), "uploads", folder)
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

	return "/docs/uploads/" + folder + "/" + filename, nil
}
