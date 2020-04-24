package internal

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
	"sync"
)

// package logger
var thePackageLogger *logrus.Entry
var logOnce sync.Once

func getLog() *logrus.Entry {
	logOnce.Do(func() {
		thePackageLogger = logrus.WithFields(logrus.Fields{"package": "internal"})
	})
	return thePackageLogger
}


var logLevelByString = map[string]logrus.Level{
	"PANIC": 	logrus.PanicLevel,
	"FATAL": 	logrus.FatalLevel,
	"ERROR": 	logrus.ErrorLevel,
	"WARN":  	logrus.WarnLevel,
	"WARNING":	logrus.WarnLevel,
	"INFO":  	logrus.InfoLevel,
	"DEBUG": 	logrus.DebugLevel,
	"TRACE": 	logrus.TraceLevel,
}

//var LogLevel logrus.Level


func ConfigureLogger(ll string, json bool) {
	LogLevel := logLevelByString[strings.ToUpper(ll)]
	if LogLevel == 0 {
		_, _ = fmt.Fprintf(os.Stderr, "\nInvalid --logLevel value '%s'. Must be one of PANIC, FATAL, WARNING, WARN, INFO, DEBUG or TRACE\n", ll)
		os.Exit(2)
	}
	logrus.SetLevel(LogLevel)
	if json {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}
}
