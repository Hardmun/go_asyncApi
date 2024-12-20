package logs

import (
	"asyncApi/utils"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

var (
	onceErrorLogger     sync.Once
	ErrorLoggerInstance Logger
)

type Logger interface {
	Write(errMsg ...any)
	Fatal(errMsg ...any)
	ClearLogs()
	Close()
}

type LogStruct struct {
	logType string
	logFile *os.File
}

func (l *LogStruct) Write(errMsg ...any) {
	newLine := log.New(l.logFile, fmt.Sprintf("[%s]", l.logType), log.LstdFlags)
	newLine.Println(errMsg...)
}

func (l *LogStruct) Fatal(errMsg ...any) {
	l.Write(errMsg...)
	log.Fatal(errMsg...)
}

func (l *LogStruct) ClearLogs() {
	if err := os.Truncate(l.logFile.Name(), 0); err != nil {
		l.Fatal(err)
	}
}

func (l *LogStruct) Close() {
	err := l.logFile.Close()
	if err != nil {
		l.Write(err)
	}
}

func newLog(logType string) (Logger, error) {
	logDir, err := utils.DirPath(utils.GetAbsPath(), "logs")
	if err != nil {
		return nil, err
	}

	filename := "error.log"

	var lFile *os.File
	lFile, err = os.OpenFile(filepath.Join(logDir, filename), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	return &LogStruct{
		logType: logType,
		logFile: lFile,
	}, nil
}

func GetErrorLog() (Logger, error) {
	var err error
	onceErrorLogger.Do(func() {
		ErrorLoggerInstance, err = newLog("ERR")
	})
	if err != nil {
		return nil, err
	}

	return ErrorLoggerInstance, nil
}
