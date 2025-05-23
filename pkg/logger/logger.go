package logger

import (
	"github.com/sirupsen/logrus"
	"os"
)

var Log = logrus.New()

func Init() {
	Log.SetOutput(os.Stdout)
	Log.SetLevel(logrus.DebugLevel)
	Log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
}
