package utils

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
)

func SaveUploadedFile(file *multipart.FileHeader, destPath string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Ensure directory exists
	// Using .gemini scratch path implicitly or relative to execution?
	// We'll use relative "uploads" folder
	if err := os.MkdirAll(destPath, os.ModePerm); err != nil {
		return "", err
	}

	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
	filepath := filepath.Join(destPath, filename)

	dst, err := os.Create(filepath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return "", err
	}

	return filepath, nil
}
