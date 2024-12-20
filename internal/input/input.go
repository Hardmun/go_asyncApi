package input

import (
	"asyncApi/utils"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type InpParams struct {
	Server    string            `json:"server"`
	EndPoint  string            `json:"endPoint"`
	Ssl       bool              `json:"ssl"`
	IsORD     bool              `json:"isORD"`
	OrigResp  bool              `json:"origResp"`
	Login     string            `json:"login"`
	Password  string            `json:"password"`
	Method    string            `json:"method"`
	ConnPool  int               `json:"connPool"`
	Errlist   []string          `json:"errlist"`
	Headers   map[string]string `json:"headers"`
	Data      any               `json:"data"`
	Directory string
}

func GetInputParams(uuidDir string) (InpParams, error) {
	var (
		jsonFile *os.File
		err      error
		byteJSON []byte
	)

	workDir := filepath.Join(utils.GetDataPath(), uuidDir)

	jsonFile, err = os.Open(filepath.Join(workDir, "data.json"))
	if err != nil {
		return InpParams{}, err
	}

	byteJSON, err = io.ReadAll(jsonFile)
	if err != nil {
		return InpParams{}, err
	}

	err = jsonFile.Close()
	if err != nil {
		return InpParams{}, err
	}

	if !json.Valid(byteJSON) {
		return InpParams{}, fmt.Errorf("invalid JSON string: %v", string(byteJSON))
	}

	var data InpParams
	err = json.Unmarshal(byteJSON, &data)
	if err != nil {
		return InpParams{}, err
	}

	data.Directory = workDir

	return data, nil
}
