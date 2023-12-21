package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var logFile *os.File

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

func callAsyncApi(uuid *string) {
	jsonFile, err := os.Open(filepath.Join(*uuid, "da1ta.json"))
	if err != nil {
		loggErrorMessage(err)
	}

	println(*uuid, jsonFile)
}

func main() {
	logFile, err := openFile("error.log")
	if err != nil {
		systemError(err)
	}

	fmt.Println(*logFile)

	//defer func(logFile *os.File) {
	//	err := logFile.Close()
	//	if err != nil {
	//		systemError(err)
	//	}
	//}(logFile)

	args := os.Args

	switch len(args) {
	case 2:
		arg := args[1]

		if arg == "-clearLogs" {
			//TODO: Clear log with parameter
		} else {
			callAsyncApi(&arg)
		}
	case 3:
		//TODO: Clear log with parameter - folder
	default:
		os.Exit(0)
	}
}
