package logger

import (
	"log"
)

const (
	DEBUG_LEVEL = 4
	INFO_LEVEL  = 3
)

var Level int = 4

func Info(format string, args ...interface{}) {
	log.Printf("INFO: "+format, args...)
}

func Warn(format string, args ...interface{}) {
	log.Printf("WARN: "+format, args...)
}

func Error(format string, args ...interface{}) {
	log.Printf("Error: "+format, args...)
}

func Debug(format string, args ...interface{}) {
	if Level < DEBUG_LEVEL {
		return
	}
	log.Printf("DEBUG: "+format, args...)
}
