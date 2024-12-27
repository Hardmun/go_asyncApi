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
	utils.SetServiceMode(false)

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
		} else if args[1] == "-clear" {
			_ = utils.ClearTempFiles("")
		} else {
			var iParam *input.InpParams
			iParam, err = input.GetInputParams(arg)
			if err != nil {
				errLog.Fatal(err)
			}

			switch iParam.Module {
			case input.Ord:
				err = ord.CallOrdApi(iParam)
			case input.Diadoc:
				err = common.CallAsyncApi(iParam)
			default:
				err = common.CallAsyncApi(iParam)
			}
			if err != nil {
				errLog.Fatal(err)
			}
		}
	case 3:
		if args[1] == "-clear" {
			_ = utils.ClearTempFiles(args[2])
		}
	default:
	}
}
