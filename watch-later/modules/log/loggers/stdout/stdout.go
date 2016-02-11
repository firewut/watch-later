package stdout

import (
	"io"
	"log"
)

type Log struct {
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
}

func NewLog(
	traceHandle io.Writer,
	infoHandle io.Writer,
	warningHandle io.Writer,
	errorHandle io.Writer,
) *Log {
	return &Log{
		Trace: log.New(traceHandle,
			"\nTRACE: ",
			log.Ldate|log.Ltime),

		Info: log.New(infoHandle,
			"\nINFO: ",
			log.Ldate|log.Ltime),

		Warning: log.New(warningHandle,
			"\nWARNING: ",
			log.Ldate|log.Ltime),

		Error: log.New(errorHandle,
			"\nERROR: ",
			log.Ldate|log.Ltime),
	}
}
