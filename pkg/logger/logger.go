package logger

import (
	"log"
)

const (
	DEBUG_LEVEL = 4
	INFO_LEVEL  = 3
	WARN_LEVEL  = 2
	ERROR_LEVEL = 1
)

var Level int = INFO_LEVEL

func Info(format string, args ...interface{}) {
	if Level < INFO_LEVEL {
		return
	}
	log.Printf("INFO: "+format, args...)
}

func Warn(format string, args ...interface{}) {
	if Level < WARN_LEVEL {
		return
	}
	log.Printf("WARN: "+format, args...)
}

func Error(format string, args ...interface{}) {
	if Level < ERROR_LEVEL {
		return
	}
	log.Printf("Error: "+format, args...)
}

func Debug(format string, args ...interface{}) {
	if Level < DEBUG_LEVEL {
		return
	}
	log.Printf("DEBUG: "+format, args...)
}
