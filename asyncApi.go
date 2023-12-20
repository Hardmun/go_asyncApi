package main

import "os"

func main() {
	args := os.Args
	switch len(args) {
	case 2:
		arg := args[1]

		if arg == "-clearLogs" {
			println("clearLogs")
		} else {
			println(arg)
		}
	case 3:

	}
	//argsLen := len(args)
	//if argsLen == 2 {
	//	println(2)
	//}
}
