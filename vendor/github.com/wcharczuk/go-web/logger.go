package web

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/blendlabs/go-util"
)

// Logger is a type that implements basic logging methods.
type Logger interface {
	Write(args ...interface{})
	Log(args ...interface{})
	Logf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
}

// NewStandardOutputLogger returns a new logger to stdout and stderr.
func NewStandardOutputLogger() Logger {
	return &logger{
		log:      log.New(os.Stdout, "", 0),
		errorLog: log.New(os.Stderr, "", 0),
	}
}

// NewStandardOutputErrorLogger returns a new logger just to stderr.
func NewStandardOutputErrorLogger() Logger {
	return &logger{
		errorLog: log.New(os.Stderr, "", 0),
	}
}

// NewLogger returns a Logger writing to the given io.Writers.
func NewLogger(output io.Writer, errorOutput io.Writer) Logger {
	if errorOutput != nil {
		return &logger{
			log:      log.New(output, "", 0),
			errorLog: log.New(errorOutput, "", 0),
		}
	}
	return &logger{
		log: log.New(output, "", 0),
	}
}

type logger struct {
	log      *log.Logger
	errorLog *log.Logger
}

func (l *logger) Write(args ...interface{}) {
	if l.log != nil {
		output := fmt.Sprint(args...)
		l.log.Printf("%s\n", output)
	}
}

func (l *logger) Log(args ...interface{}) {
	if l.log != nil {
		timestamp := getLoggingTimestamp()
		output := fmt.Sprint(args...)
		l.log.Printf("%s %s\n", timestamp, output)
	}
}

func (l *logger) Logf(format string, args ...interface{}) {
	if l.log != nil {
		timestamp := getLoggingTimestamp()
		output := fmt.Sprintf(format, args...)
		l.log.Printf("%s %s\n", timestamp, output)
	}
}

func (l *logger) Error(args ...interface{}) {
	if l.errorLog != nil {
		timestamp := getLoggingTimestamp()
		output := fmt.Sprint(args...)
		l.errorLog.Printf("%s %s\n", timestamp, output)
	}
}

func (l *logger) Errorf(format string, args ...interface{}) {
	if l.errorLog != nil {
		timestamp := getLoggingTimestamp()
		output := fmt.Sprintf(format, args...)
		l.errorLog.Printf("%s %s\n", timestamp, output)
	}
}

func getLoggingTimestamp() string {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	return util.Color(timestamp, util.ColorGray)
}
