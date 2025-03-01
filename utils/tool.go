package utils

import (
	"blblcd/model"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

func FileOrPathExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

func ExcutePath() string {
	excutePath, err := os.Executable()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	return filepath.Dir(excutePath)
}

func ReadTextFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func PresetPath(path string) {
	if !FileOrPathExists(path) {
		os.MkdirAll(path, os.ModePerm)
	}
}

func EncodePaginationOffset(pagination model.PaginationOffset) string {
	paginationJSON, err := json.Marshal(pagination)
	if err != nil {
		fmt.Println("Error marshaling pagination:", err)
		return ""
	}

	return string(paginationJSON)
}

func DecodePaginationOffset(paginationStr string) (*model.PaginationOffset, error) {
	var paginationOffset model.PaginationOffset
	err := json.Unmarshal([]byte(paginationStr), &paginationOffset)
	if err != nil {
		fmt.Println("Error unmarshaling inner JSON:", err)
		return nil, err
	}
	return &paginationOffset, nil
}
