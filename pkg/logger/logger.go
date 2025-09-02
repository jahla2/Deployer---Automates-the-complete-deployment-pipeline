package logger

import (
	"fmt"
	"log"
)

// ANSI color codes
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	White  = "\033[37m"
	Bold   = "\033[1m"
)

type Logger struct {
	prefix string
}

func New(prefix string) *Logger {
	return &Logger{
		prefix: prefix,
	}
}

func (l *Logger) Info(msg string, args ...interface{}) {
	formatted := fmt.Sprintf(msg, args...)
	log.Printf("%s[INFO]%s %s", Cyan, Reset, formatted)
}

func (l *Logger) Error(msg string, args ...interface{}) {
	formatted := fmt.Sprintf(msg, args...)
	log.Printf("%s[ERROR]%s %s", Red, Reset, formatted)
}

func (l *Logger) Warning(msg string, args ...interface{}) {
	formatted := fmt.Sprintf(msg, args...)
	log.Printf("%s[WARNING]%s %s", Yellow, Reset, formatted)
}

func (l *Logger) Success(msg string, args ...interface{}) {
	formatted := fmt.Sprintf(msg, args...)
	log.Printf("%s[SUCCESS]%s %s", Green, Reset, formatted)
}