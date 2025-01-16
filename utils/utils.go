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
		if errDir = os.MkdirAll(pathDir, os.ModePerm); errDir != nil {
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

func ClearTempFiles(dr string) error {
	dp := GetDataPath()

	if dr == "" {
		items, err := os.ReadDir(dp)
		if err != nil {
			return err
		}

		for _, itm := range items {
			fullPath := filepath.Join(dp, itm.Name())
			err = os.RemoveAll(fullPath)
			if err != nil {
				return err
			}
		}
	} else {
		err := os.RemoveAll(filepath.Join(dp, dr))
		if err != nil {
			return err
		}
	}

	return nil
}

// - true: exe mode
// - false: ide mode
func SetServiceMode(mode bool) {
	serviceMode = mode
}
