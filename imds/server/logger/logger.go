package logger

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

const (
	DebugLevel = "debug"
	InfoLevel  = "info"
	WarnLevel  = "warn"
	ErrorLevel = "error"
)

type Fields map[string]interface{}

type Entry struct {
	Level   string
	Message string
	Time    time.Time
	Data    Fields
}

func (e *Entry) Debug(args ...interface{}) {
	e.log(DebugLevel, args...)
}

func (e *Entry) Info(args ...interface{}) {
	e.log(InfoLevel, args...)
}

func (e *Entry) Warn(args ...interface{}) {
	e.log(WarnLevel, args...)
}

func (e *Entry) Error(args ...interface{}) {
	e.log(ErrorLevel, args...)
}

func (e *Entry) log(level string, args ...interface{}) {
	e.Level = level
	e.Message = fmt.Sprint(args...)
	e.Time = time.Now()
	e.output()
}

func (e *Entry) output() {
	msg := fmt.Sprintf("%s%s ▶ %s",
		e.Time.Format("[2006-01-02 15:04:05]"),
		fmt.Sprintf("[%s]", strings.ToUpper(e.Level)),
		e.Message,
	)

	keys := []string{}

	var errStr string
	for key, val := range e.Data {
		if key == "error" {
			errStr = fmt.Sprintf("%s", val)
			continue
		}

		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		msg += fmt.Sprintf(" ◆ %s=%v", key,
			fmt.Sprintf("%#v", e.Data[key]))
	}

	if errStr != "" {
		msg += "\n" + errStr
	}

	if string(msg[len(msg)-1]) != "\n" {
		msg += "\n"
	}

	fmt.Print(msg)
}

func WithFields(fields Fields) *Entry {
	return &Entry{
		Data: fields,
	}
}

func Debug(args ...interface{}) {
	entry := &Entry{}
	entry.Debug(args...)
}

func Info(args ...interface{}) {
	entry := &Entry{}
	entry.Info(args...)
}

func Warn(args ...interface{}) {
	entry := &Entry{}
	entry.Warn(args...)
}

func Error(args ...interface{}) {
	entry := &Entry{}
	entry.Error(args...)
}
