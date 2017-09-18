/*
M-GO Logger, based on Logrus.

It will setup a logrus instance and log things as JSON.
A configurable LOGLEVEL can be passed in as an environment variable, i.e.

  export LOGLEVEL="INFO"
*/
package log

import (
	"github.com/Sirupsen/logrus"
	"os"
)

var logger = logrus.New()

func init() {
	// Log as JSON instead of the default ASCII formatter.
	logger.Formatter = new(logrus.JSONFormatter)

	// Setup log leveling
	switch os.Getenv("LOGLEVEL") {
	default:
		logger.Level = logrus.InfoLevel
	case "DEBUG":
		logger.Level = logrus.DebugLevel
	case "INFO":
		logger.Level = logrus.InfoLevel
	case "WARN":
		logger.Level = logrus.WarnLevel
	case "ERROR":
		logger.Level = logrus.ErrorLevel
	case "FATAL":
		logger.Level = logrus.FatalLevel
	}
}

func WithField(key string, value interface{}) *logrus.Entry {
	return logger.WithField(key, value)
}

// Log a Debug level message
func Debug(args ...interface{}) {
	logger.Debug(args...)
}

// Log an Info level message
func Info(args ...interface{}) {
	logger.Info(args...)
}

// Log a Warn level message
func Warn(args ...interface{}) {
	logger.Warn(args...)
}

// Log an Error level message
func Error(args ...interface{}) {
	logger.Error(args...)
}

// Log a Fatal level message
func Fatal(args ...interface{}) {
	logger.Fatal(args...)
}

// Panic!
func Panic(args ...interface{}) {
	logger.Panic(args...)
}

func WithFields(fields logrus.Fields) *logrus.Entry {
	return logger.WithFields(fields)
}

func WithError(err error) *logrus.Entry {
	return logger.WithField("error", err)
}
