package logger

import (
	"container/ring"
	"expvar"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/sirupsen/logrus"
)

type lastLogsHook struct {
	r *ring.Ring
}

func (l lastLogsHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (l *lastLogsHook) Fire(entry *logrus.Entry) error {
	l.r.Value, _ = entry.String()
	l.r = l.r.Next()
	return nil
}

var Log = func() *logrus.Logger {
	l := logrus.New()
	l.SetLevel(logrus.DebugLevel)

	lastLogs := &lastLogsHook{
		r: ring.New(50),
	}
	l.Hooks.Add(lastLogs)
	expvar.Publish("logs", expvar.Func(func() interface{} {
		var lines []string
		lastLogs.r.Do(func(i interface{}) {
			if i != nil {
				lines = append(lines, i.(string))
			}
		})
		return lines
	}))

	return l
}()

func WithFunction() *logrus.Entry {
	pc, file, line, _ := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	return Log.WithFields(logrus.Fields{
		"function": fmt.Sprintf("%s()", filepath.Base(f.Name())),
		"file":     fmt.Sprintf("%s:%d", file, line),
	})
}
