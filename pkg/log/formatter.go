package log

import (
	"bytes"
	"fmt"
	"runtime"

	"github.com/sirupsen/logrus"
)

// PrettifiedFormatter populates mandatory log fields
func PrettifiedFormatter() logrus.Formatter {
	return &logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.999999",
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "@timestamp",
			logrus.FieldKeyLevel: "@level",
			logrus.FieldKeyMsg:   "@message",
			logrus.FieldKeyFunc:  "@caller",
			logrus.FieldKeyFile:  "@file",
		},
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			pc, file, line, _ := runtime.Caller(9)
			name := runtime.FuncForPC(pc).Name()
			if i := bytes.LastIndexAny([]byte(name), "."); i != -1 {
				name = name[i+1:]
			}

			if i := bytes.LastIndexAny([]byte(file), "/"); i != -1 {
				file = file[i+1:]
			}

			return name, fmt.Sprintf("%s:%d", file, line)
		},
		PrettyPrint: false,
	}
}
