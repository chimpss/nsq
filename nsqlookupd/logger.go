package nsqlookupd

import (
	"github.com/nsqio/nsq/internal/lg"
)

type Logger lg.Logger

const (
	LOG_DEBUG = lg.DEBUG
	LOG_INFO  = lg.INFO
	LOG_WARN  = lg.WARN
	LOG_ERROR = lg.ERROR
	LOG_FATAL = lg.FATAL
)

func (l *NSQLookupd) logf(level lg.LogLevel, f string, args ...interface{}) {
	lg.Logf(l.opts.Logger, l.opts.LogLevel, level, f, args...)
}
