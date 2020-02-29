package model

import (
	"log"
	"path/filepath"
	"os"
	"fmt"
)

type logLevel uint8

const (
	llNone logLevel  = iota
	llError
	llInfo
	llDebug
)

var glogger *log.Logger
var gLogLevel = llNone
func _log(ll logLevel,a ...interface{}) {
	if glogger != nil && ll <= gLogLevel {
		glogger.Output(3,fmt.Sprint(a...))
	}
}

func logError(e error, a ...interface{}) error {
	if a==nil {
		_log(llError,e)
	} else {
		_log(llError,e,a)
	}
	return e
}

func logInfo(a ...interface{}) {
	_log(llInfo,a...)
}

func logDebug(a ...interface{}) {
	_log(llDebug,a...)
}

func NewLogger(ll logLevel, logFolder string) {
	logpath := filepath.Join(logFolder,"flap.log")
	f, _ := os.OpenFile(logpath,os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	glogger = log.New(f, "model ", log.LstdFlags | log.Lshortfile)
	gLogLevel = ll
}

