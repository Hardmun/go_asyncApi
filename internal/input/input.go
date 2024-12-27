package input

import (
	"asyncApi/utils"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type ModuleType int

const (
	Dflt ModuleType = iota
	Ord
	Diadoc
)

type InpParams struct {
	Server    string            `json:"server"`
	EndPoint  string            `json:"endPoint"`
	Ssl       bool              `json:"ssl"`
	Module    ModuleType        `json:"module"`
	OrigResp  bool              `json:"origResp"`
	Login     string            `json:"login"`
	Password  string            `json:"password"`
	Method    string            `json:"method"`
	ConnPool  int               `json:"connPool"`
	Errlist   []string          `json:"errlist"`
	Headers   map[string]string `json:"headers"`
	Params    map[string]string `json:"params"`
	Body      any               `json:"body"`
	Directory string
}

func GetInputParams(uuidDir string) (*InpParams, error) {
	var (
		jsonFile *os.File
		err      error
		byteJSON []byte
	)
	workDir := filepath.Join(utils.GetDataPath(), uuidDir)

	jsonFile, err = os.Open(filepath.Join(workDir, "data.json"))
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

	var data InpParams
	err = json.Unmarshal(byteJSON, &data)
	if err != nil {
		return nil, err
	}

	data.Directory = workDir

	return &data, nil
}
