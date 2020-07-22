package log

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

func init() {
	devMode := os.Getenv("DEV")
	if len(devMode) > 0 {
		appName := os.Getenv("APP_NAME")
		var fileName string
		if len(appName) > 0 {
			fileName = appName + ".log"
		} else {
			executableName, err := os.Executable()
			if err == nil {
				fileName, _ = filepath.Abs(executableName)
				fileName = fileName + ".log"
			}
		}

		f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("Error using log file %s \n", fileName)
		} else {
			fmt.Printf("Using log file %s \n", fileName)
			logrus.SetOutput(f)
		}
	} else {
		logrus.SetOutput(os.Stdout)
	}
	lvl := logrus.InfoLevel
	logLevel := os.Getenv("LOG_LEVEL")
	if len(logLevel) > 0 {
		level, err := logrus.ParseLevel(logLevel)
		if err == nil {
			lvl = level
		}
	}
	logrus.SetLevel(lvl)

	logFormat := os.Getenv("LOG_JSON")
	if len(logFormat) > 0 {
		jsFormatter := &logrus.JSONFormatter{
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "time",
				logrus.FieldKeyLevel: "severity",
				logrus.FieldKeyMsg:   "message",
			},
		}
		jsFormatter.TimestampFormat = time.RFC3339Nano
		logrus.SetFormatter(jsFormatter)
	} else {
		formatter := new(logrus.TextFormatter)
		formatter.TimestampFormat = time.RFC3339Nano
		formatter.FullTimestamp = true
		formatter.ForceColors = false
		logrus.SetFormatter(formatter)
	}
}

// GetLogger returns new logger
func GetLogger() *logrus.Logger {
	return logrus.StandardLogger()
}
