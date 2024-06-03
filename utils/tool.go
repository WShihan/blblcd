package utils

import (
	"os"
)

func FileOrPathExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}
