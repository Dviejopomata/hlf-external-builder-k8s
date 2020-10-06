package utils

import (
	"github.com/otiai10/copy"
	"log"
	"os"
)

func Copy(src, dst string) error {
	return copy.Copy(src, dst)
}

func HandleErr(err error, msg string) {
	if err != nil {
		log.Printf("Error %s: %s", msg, err.Error())
		os.Exit(1)
	}
}

func Exists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}

	return true
}
