package log

import (
	"os"

	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func init() {
	var log = logrus.New()
	log.Formatter = new(logrus.JSONFormatter)
	log.Formatter = new(logrus.TextFormatter)
	log.Formatter.(*logrus.TextFormatter).DisableColors = false
	log.Formatter.(*logrus.TextFormatter).DisableTimestamp = false
	log.Level = logrus.TraceLevel
	log.Out = os.Stdout
}
func Info(fmt string, args ...interface{}) {
	log.Infof(fmt, args...)
}

func Warn(fmt string, args ...interface{}) {
	log.Warnf(fmt, args...)
}

func Error(fmt string, args ...interface{}) {
	log.Errorf(fmt, args...)
}
