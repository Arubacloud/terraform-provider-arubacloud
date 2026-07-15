package noop

import "github.com/Arubacloud/sdk-go/internal/ports/logger"

// Compile-time assertion that *NoOpLogger satisfies logger.Logger.
var _ logger.Logger = (*NoOpLogger)(nil)

// NoOpLogger is a logger that does nothing
type NoOpLogger struct{}

func (l *NoOpLogger) Debugf(format string, args ...interface{}) {}
func (l *NoOpLogger) Infof(format string, args ...interface{})  {}
func (l *NoOpLogger) Warnf(format string, args ...interface{})  {}
func (l *NoOpLogger) Errorf(format string, args ...interface{}) {}
