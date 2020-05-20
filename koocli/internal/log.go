/*
  Copyright (C) 2020 Serge ALEXANDRE

  This file is part of koobind project

  koobind is free software: you can redistribute it and/or modify
  it under the terms of the GNU General Public License as published by
  the Free Software Foundation, either version 3 of the License, or
  (at your option) any later version.

  koobind is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
  GNU General Public License for more details.

  You should have received a copy of the GNU General Public License
  along with koobind.  If not, see <http://www.gnu.org/licenses/>.
*/
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
