package utils

import (
	"fmt"
	"log"
	"os"
)

var logFile *os.File = nil

func InitLog(filePath string) {
	logFile, _ = os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
}

func LogLn(items ...any) {
	if logFile != nil {
		fmt.Fprintln(logFile, items...)
	}
	log.Println(items...)
}

func LogF(format string, items ...any) {
	if logFile != nil {
		fmt.Fprintf(logFile, format, items...)
	}
	log.Printf(format, items...)
}

func LogError(msg string, err error) {
	if logFile != nil {
		fmt.Fprintln(logFile, msg, err)
	}
	log.Println(msg, err)
}

func LogFatal(items ...any) {
	if logFile != nil {
		fmt.Fprintln(logFile, items...)
	}
	log.Fatal(items...)
}

func GetLogFile() *os.File {
	return logFile
}
