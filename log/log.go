// Package log wraps stdlib logging utilities
package log

import (
	"log"
	"os"
)

// Level denotes logging levels that are used
// to contextualize logging information.
type Level int

const (
	DebugLevel Level = 1 << iota
	InfoLevel
	ErrorLevel
)

var (
	outputFd        = os.Stdout
	timestampFormat = "Jan 2 15:04:05 2006"

	// CurrentLevel contains the current logging level
	CurrentLevel = InfoLevel
)

const (
	debugAbbrev = "DBUG"

	infoAbbrev = "INFO"

	errorAbbrev = "ERRR"

	fatalAbbrev = "FATL"
)

func logMessage(abbrev, msg string) {
	log.Printf("[%v]: %v\n", abbrev, msg)
}

// Debug logs the given message as a debug message.
func Debug(msg string) {
	if CurrentLevel <= DebugLevel {
		logMessage(debugAbbrev, msg)
	}
}

// Info logs the given message as a info message.
func Info(msg string) {
	if CurrentLevel <= InfoLevel {
		logMessage(infoAbbrev, msg)
	}
}

// Error logs the given message as a error message.
func Error(msg string) {
	if CurrentLevel <= ErrorLevel {
		logMessage(errorAbbrev, msg)
	}
}

// Fatal logs the given message as a error message and calls os.Exit(1).
func Fatal(err error) {
	logMessage(fatalAbbrev, err.Error())

	os.Exit(1)
}
