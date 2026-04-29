package logger

import (
	"awesomeProject/internal/constants"
	"context"
	"os"
	"time"

	joonix "github.com/joonix/log"
	"github.com/sirupsen/logrus"

	"awesomeProject/config"
)

type Logger struct {
	*logrus.Logger
}

var Log *Logger

const (
	AppType = "appType"
)

type LoggerError struct {
	Error error
}

func panicIfError(err error) {
	if err != nil {
		panic(LoggerError{err})
	}
}

func SetupLogger() {
	level, err := logrus.ParseLevel(config.Load().LogLevel)
	panicIfError(err)

	var formatter = joonix.NewFormatter()
	formatter.TimestampFormat = func(fields logrus.Fields, now time.Time) error {
		fields["timeStamp"] = now.Format(time.RFC3339)
		return nil
	}
	logrusVar := &logrus.Logger{
		Out:       os.Stdout,
		Formatter: formatter,
		Hooks:     make(logrus.LevelHooks),
		Level:     level,
	}
	Log = &Logger{logrusVar}
}

func Get(ctx context.Context) *logrus.Entry {
	entry := Log.WithField("serviceName", "awesome-service")
	if appType := ctx.Value(AppType); appType != nil {
		entry = entry.WithField(AppType, appType)
	}
	if correlationId := ctx.Value(constants.CorrelationId); correlationId != nil {
		entry = entry.WithField(constants.CorrelationId, correlationId)
	}
	return entry
}
