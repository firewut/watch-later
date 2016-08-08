package log

import (
	"fmt"
	"io"
	"project/modules/config"
	"project/modules/log/loggers/stdout"
	"strings"
)

type Log struct {
	HostName string
	StdOut   *stdout.Log
}

func NewLog(cfg *config.Config) (l *Log) {
	l = &Log{}

	l.HostName = cfg.HostName
	return
}

func InitLogggers(
	cfg *config.Config,
	traceHandle io.Writer,
	infoHandle io.Writer,
	warningHandle io.Writer,
	errorHandle io.Writer,
) (l *Log) {

	l = NewLog(cfg)
	l.StdOut = stdout.NewLog(traceHandle, infoHandle, warningHandle, errorHandle)

	return
}

func make_string_from_interface_slice(interfaced_slice ...interface{}) (s string) {
	s_slice := make([]string, len(interfaced_slice))
	for i, v := range interfaced_slice {
		s_slice[i] = fmt.Sprintf("%s", v)
	}

	s = strings.Join(s_slice, "\n")
	return s
}

func (l *Log) Trace(messages ...interface{}) {
	l.StdOut.Trace.Println(messages)
}

func (l *Log) Info(messages ...interface{}) {
	l.StdOut.Info.Println(messages)
}

func (l *Log) Warning(messages ...interface{}) {
	l.StdOut.Warning.Println(messages)
}

func (l *Log) Error(messages ...interface{}) {
	l.StdOut.Error.Println(messages)
}

func (l *Log) Success(messages ...interface{}) {
	l.StdOut.Info.Println(messages)
}
