package main

import (
	"asyncApi/internal/logs"
	"asyncApi/utils"
	"log"
	"os"
)

func main() {
	errLog, err := logs.GetErrorLog()
	if err != nil {
		log.Fatal(err)
	}
	defer errLog.Close()

	args := os.Args

	switch len(args) {
	case 2:
		arg := args[1]
		if arg == "-clearLogs" {
			errLog.ClearLogs()
		} else {
			//err = callAsyncApi(&arg)
			//if err != nil {
			//	loggErrorMessage(errWrap(&err, "main", "err = callAsyncApi(&arg)"))
			//	fmt.Println(err.Error())
			//}
		}
	case 3:
		if args[1] == "-clear" {
			if err = utils.ClearTempFiles(args[2]); err != nil {
				errLog.Write(err)
			}
		}
	default:
	}
	os.Exit(0)
}
