package input

import (
	"asyncApi/utils"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var (
	WorkDir   string
	InpParams *InpParamsStruct
)

type ModeType int

const (
	Ord          ModeType = 1
	DiadocUpload          = 3
)

type InpParamsStruct struct {
	Server      string            `json:"server"`
	EndPoint    string            `json:"endPoint"`
	Ssl         bool              `json:"ssl"`
	Mode        ModeType          `json:"mode"`
	OrigResp    bool              `json:"origResp"`
	Login       string            `json:"login"`
	Password    string            `json:"password"`
	Method      string            `json:"method"`
	ConnPool    int               `json:"connPool"`
	Errlist     []string          `json:"errlist"`
	Headers     map[string]string `json:"headers"`
	GetParams   map[string]string `json:"params"`
	DownloadDir string            `json:"downloadDir"`
	Body        any               `json:"body"`
}

func getInputParams(uuidDir string) (*InpParamsStruct, error) {
	var (
		jsonFile *os.File
		err      error
		byteJSON []byte
	)
	WorkDir = filepath.Join(utils.GetDataPath(), uuidDir)

	jsonFile, err = os.Open(filepath.Join(WorkDir, "data.json"))
	if err != nil {
		return nil, err
	}

	byteJSON, err = io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	err = jsonFile.Close()
	if err != nil {
		return nil, err
	}

	if !json.Valid(byteJSON) {
		return nil, fmt.Errorf("invalid JSON string: %v", string(byteJSON))
	}

	var data InpParamsStruct
	err = json.Unmarshal(byteJSON, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func InitializeInputParams(uuidDir string) (*InpParamsStruct, error) {
	var err error

	InpParams, err = getInputParams(uuidDir)
	if err != nil {
		return nil, err
	}
	return InpParams, nil
}
