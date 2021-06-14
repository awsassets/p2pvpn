// Package log provides simply wrapped log for p2pvpn.
package log

import (
	logging "github.com/ipfs/go-log/v2"
	"go.uber.org/zap"
)

var (
	// _logger is the default logger for p2pvpn.
	_logger = Logger("p2pvpn")
)

func init() {
	SetAllLoggers(logging.LevelDebug)
}

func Logger(system string) *logging.ZapEventLogger {
	logger := logging.Logger(system)
	tmpLogger := logger.SugaredLogger.Desugar().WithOptions(zap.AddCallerSkip(1)).Sugar()
	logger.SugaredLogger = *tmpLogger
	return logger
}

func SetAllLoggers(lvl logging.LogLevel) {
	logging.SetAllLoggers(lvl)
}

func Debugf(template string, args ...interface{}) {
	_logger.Debugf(template, args...)
}

func Infof(template string, args ...interface{}) {
	_logger.Infof(template, args...)
}

func Warnf(template string, args ...interface{}) {
	_logger.Warnf(template, args...)
}

func Errorf(template string, args ...interface{}) {
	_logger.Errorf(template, args...)
}

func Fatalf(template string, args ...interface{}) {
	_logger.Fatalf(template, args...)
}
