package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

var logFile *os.File

type jsonStruct struct {
	BaseURL  string `json:"base_Url"`
	Url      string `json:"url"`
	Ssl      bool   `json:"ssl"`
	Login    string `json:"login"`
	Password string `json:"password"`
	Headers  struct {
		ContentType string `json:"Content-type"`
	}
	Method string      `json:"method"`
	Data   interface{} `json:"data"`
}

func openFile(path string) (*os.File, error) {
	lFile, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	return lFile, nil
}

//	func ErrorMsgToJSON(status int, err error) {
//		msgStruct := struct {
//			Error struct {
//				Status int    `json:"status"`
//				Reason string `json:"reason"`
//			} `json:"error"`
//		}{
//			Error: struct {
//				Status int    `json:"status"`
//				Reason string `json:"reason"`
//			}(struct {
//				Status int
//				Reason string
//			}{Status: status, Reason: err.Error()}),
//		}
//
// }
func systemError(errMsg error) {
	sysFile, errSys := openFile("sys.log")
	if errSys != nil {
		log.Fatal(errSys)
	}

	defer func(sysFile *os.File) {
		errClose := sysFile.Close()
		if errClose != nil {
			log.Fatal(errClose)
		}
	}(sysFile)

	errorLog := log.New(sysFile, "[error]", log.LstdFlags|log.Lshortfile)
	errorLog.Println(errMsg.Error())
}

func loggErrorMessage(err error) {
	errorLog := log.New(logFile, "[error]", log.LstdFlags|log.Lshortfile)
	errorLog.Println(err.Error())
}

func openJSON(path string) (jsonStruct, error) {
	var (
		jsonFile *os.File
		err      error
		byteJSON []byte
	)

	jsonFile, err = os.Open(path)
	if err != nil {
		return jsonStruct{}, err
	}

	byteJSON, err = io.ReadAll(jsonFile)
	if err != nil {
		return jsonStruct{}, err
	}

	err = jsonFile.Close()
	if err != nil {
		return jsonStruct{}, err
	}

	var data jsonStruct
	err = json.Unmarshal(byteJSON, &data)
	if err != nil {
		return jsonStruct{}, err
	}

	return data, nil
}

func callAsyncApi(uuid *string) error {
	data, err := openJSON(filepath.Join(*uuid, "data.json"))
	if err != nil {
		return err
	}

	fmt.Println(data)

	return nil
}

func main() {
	var err error
	logFile, err = openFile("error.log")
	if err != nil {
		systemError(err)
	}
	//Closing the logFile
	defer func(logFile *os.File) {
		err := logFile.Close()
		if err != nil {
			systemError(err)
		}
	}(logFile)

	args := os.Args

	switch len(args) {
	case 2:
		arg := args[1]

		if arg == "-clearLogs" {
			//TODO: Clear log with parameter
		} else {
			err = callAsyncApi(&arg)
			if err != nil {
				loggErrorMessage(err)
				fmt.Println(err.Error())
			}
		}
	case 3:
		//TODO: Clear log with parameter - folder
	default:
		os.Exit(0)
	}
}
