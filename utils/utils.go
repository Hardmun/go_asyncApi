package utils

import (
	"log"
	"os"
	"path/filepath"
	"sync"
)

var (
	absPathOnce  sync.Once
	AbsPath      string
	dataPathOnce sync.Once
	DataPath     string
)

var serviceMode bool

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
		if serviceMode {
			var exePath string
			exePath, err = os.Executable()
			if err != nil {
				log.Fatal(err)
			}
			AbsPath = filepath.Dir(exePath)
		} else {
			AbsPath, err = filepath.Abs("./")
			if err != nil {
				log.Fatal(err)
			}
		}

	})
	return AbsPath
}

func GetDataPath() string {
	dataPathOnce.Do(func() {
		var err error
		DataPath, err = DirPath(GetAbsPath(), "data")
		if err != nil {
			log.Fatal(err)
		}
	})
	return DataPath
}

func ClearTempFiles(uuid string) error {
	dp := GetDataPath()

	err := os.RemoveAll(filepath.Join(dp, uuid))
	if err != nil {
		return err
	}

	return nil
}

// - true: file runs under the service
// - false: file runs under the user
func SetServiceMode(mode bool) {
	serviceMode = mode
}
