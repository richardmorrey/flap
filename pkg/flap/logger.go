package flap

import (
	"log"
	"path/filepath"
	"os"
	"fmt"
)

type LogLevel uint8

const (
	llNone LogLevel  = iota
	llError
	llInfo
	llDebug
)

var glogger *log.Logger
var gLogLevel = llNone
func _log(ll LogLevel,a ...interface{}) {
	if glogger != nil && ll <= gLogLevel {
		glogger.Output(3,fmt.Sprint(a...))
	}
}

func logError(e error, a ...interface{}) error {
	_log(llError,e,a)
	return e
}

func logInfo(a ...interface{}) {
	_log(llInfo,a)
}

func logDebug(a ...interface{}) {
	_log(llDebug,a)
}

func NewLogger(ll LogLevel, logFolder string) {
	logpath := filepath.Join(logFolder,"flap.log")
	f, _ := os.OpenFile(logpath,os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	glogger = log.New(f, "flap ", log.LstdFlags | log.Lshortfile)
	gLogLevel = ll
}

