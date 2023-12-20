package main

import (
	"fmt"
	"os"
)

var logFile *os.File

type msgError struct {
	Error struct {
		Status int    `json:"status"`
		Reason string `json:"reason"`
	} `json:"error"`
}

func openLogFile(path string) error {
	var err error
	logFile, err = os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	return nil
}

func errMessage(err error) struct{} {
	return struct{}{}
}

func callAsyncApi(uuid *string) {
	println(*uuid)
}

func main() {
	err := openLogFile("defo/error.log")
	if err != nil {
		newMessage := msgError{}

		os.Exit(0)
	}
	fmt.Printf("error: %v, logfile^ %v", err, *logFile)

	args := os.Args
	switch len(args) {
	case 2:
		arg := args[1]

		if arg == "-clearLogs" {
			println("clearLogs")
		} else {
			callAsyncApi(&arg)
		}
	case 3:
		//TODO: Clear log with parameter - folder
	default:
		os.Exit(0)
	}
}
