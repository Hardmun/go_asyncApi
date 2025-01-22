package main

import (
	"asyncApi/internal/api/common"
	"asyncApi/internal/api/diadoc"
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
		} else if args[1] == "-clear" {
			err = utils.ClearTempFiles("")
			errLog.Write(err)
		} else {
			var iParam *input.InpParamsStruct
			iParam, err = input.InitializeInputParams(arg)

			if err != nil {
				errLog.Fatal(err)
			}

			switch iParam.Mode {
			case input.Ord:
				err = ord.CallOrdApi(iParam)
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
			return
		}

		var iParam *input.InpParamsStruct
		iParam, err = input.InitializeInputParams(args[1])
		if err != nil {
			errLog.Fatal(err)
		}

		if iParam.Mode == input.DiadocUpload {
			err = diadoc.UploadFiles(args[2])
			if err != nil {
				errLog.Fatal(err)
			}
		}
	default:
	}
}
