package logger

// Logger is the interface for logging within the SDK
type Logger interface {
	// Debugf logs a debug message with formatting
	Debugf(format string, args ...interface{})
	// Infof logs an info message with formatting
	Infof(format string, args ...interface{})
	// Warnf logs a warning message with formatting
	Warnf(format string, args ...interface{})
	// Errorf logs an error message with formatting
	Errorf(format string, args ...interface{})
}
