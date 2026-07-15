package native

import (
	"log"
	"os"

	"github.com/Arubacloud/sdk-go/internal/ports/logger"
)

// Compile-time assertion that *DefaultLogger satisfies logger.Logger.
var _ logger.Logger = (*DefaultLogger)(nil)

// DefaultLogger is a simple logger implementation using standard log package
type DefaultLogger struct {
	debug *log.Logger
	info  *log.Logger
	warn  *log.Logger
	err   *log.Logger
}

// NewDefaultLogger creates a new default logger
func NewDefaultLogger() *DefaultLogger {
	return &DefaultLogger{
		debug: log.New(os.Stdout, "[DEBUG] ", log.LstdFlags),
		info:  log.New(os.Stdout, "[INFO] ", log.LstdFlags),
		warn:  log.New(os.Stderr, "[WARN] ", log.LstdFlags),
		err:   log.New(os.Stderr, "[ERROR] ", log.LstdFlags),
	}
}

func (l *DefaultLogger) Debugf(format string, args ...interface{}) {
	l.debug.Printf(format, args...)
}

func (l *DefaultLogger) Infof(format string, args ...interface{}) {
	l.info.Printf(format, args...)
}

func (l *DefaultLogger) Warnf(format string, args ...interface{}) {
	l.warn.Printf(format, args...)
}

func (l *DefaultLogger) Errorf(format string, args ...interface{}) {
	l.err.Printf(format, args...)
}
