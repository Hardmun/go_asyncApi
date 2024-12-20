package main

import (
	"asyncApi/internal/api/common"
	"asyncApi/internal/api/ord"
	"asyncApi/internal/input"
	"asyncApi/internal/logs"
	"asyncApi/utils"
	"log"
	"os"
)

func main() {
	utils.SetServiceMode(true)

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
			var iParam input.InpParams
			iParam, err = input.GetInputParams(arg)
			if err != nil {
				errLog.Fatal(err)
			}

			if iParam.IsORD {
				err = ord.CallOrdApi(iParam)
				if err != nil {
					errLog.Fatal(err)
				}
				return
			}

			err = common.CallAsyncApi(iParam)
			if err != nil {
				errLog.Fatal(err)
			}
		}
	case 3:
		if args[1] == "-clear" {
			if err = utils.ClearTempFiles(args[2]); err != nil {
				errLog.Fatal(err)
			}
		}
	default:
	}
}
