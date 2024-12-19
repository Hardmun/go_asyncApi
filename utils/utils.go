package utils

import (
	"log"
	"os"
	"path/filepath"
	"sync"
)

var (
	absPathOnce sync.Once
	AbsPath     string
)

func DirPath(path ...string) (string, error) {
	pathDir := filepath.Join(path...)
	if info, errDir := os.Stat(pathDir); errDir != nil || !info.IsDir() {
		if errDir = os.Mkdir(pathDir, os.ModePerm); errDir != nil {
			return "", errDir
		}
	}
	return pathDir, nil
}

func GetAbsPath() string {
	absPathOnce.Do(func() {
		var err error
		AbsPath, err = filepath.Abs("./")
		if err != nil {
			log.Fatal(err)
		}
	})
	return AbsPath
}

func ClearTempFiles(uuid string) error {
	err := os.RemoveAll(filepath.Join(GetAbsPath(), "data", uuid))
	if err != nil {
		return err
	}

	return nil
}
