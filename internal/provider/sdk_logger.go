package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const subsystemName = "arubacloud-sdk"

// LogLevel controls which SDK log messages are forwarded to tflog.
type LogLevel int

const (
	LogLevelOff   LogLevel = iota // no SDK logging (default — preserves existing behavior)
	LogLevelError                 // errors only
	LogLevelWarn                  // warnings and errors
	LogLevelInfo                  // info, warnings, and errors
	LogLevelDebug                 // all messages including HTTP request/response detail
	LogLevelTrace                 // alias for Debug (SDK has no Trace method)
)

func (l LogLevel) String() string {
	switch l {
	case LogLevelOff:
		return "OFF"
	case LogLevelError:
		return "ERROR"
	case LogLevelWarn:
		return "WARN"
	case LogLevelInfo:
		return "INFO"
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelTrace:
		return "TRACE"
	default:
		return "UNKNOWN"
	}
}

// ParseLogLevel converts a string to a LogLevel. Empty string returns LogLevelOff.
// Returns an error for unrecognised values; callers should fall back to LogLevelOff.
func ParseLogLevel(s string) (LogLevel, error) {
	if s == "" {
		return LogLevelOff, nil
	}
	switch strings.ToUpper(s) {
	case "OFF":
		return LogLevelOff, nil
	case "ERROR":
		return LogLevelError, nil
	case "WARN":
		return LogLevelWarn, nil
	case "INFO":
		return LogLevelInfo, nil
	case "DEBUG":
		return LogLevelDebug, nil
	case "TRACE":
		return LogLevelTrace, nil
	default:
		return LogLevelOff, fmt.Errorf("invalid log_level %q: must be one of OFF, ERROR, WARN, INFO, DEBUG, TRACE", s)
	}
}

// sdkLogAdapter bridges the ArubaCloud SDK logger interface to the tflog subsystem
// "arubacloud-sdk". It satisfies the SDK's internal logger.Logger interface via
// Go structural typing — no import of the internal SDK package is needed.
//
// Messages are only forwarded when the configured level permits them, giving the
// user a second filter on top of TF_LOG / TF_LOG_PROVIDER.
type sdkLogAdapter struct {
	ctx   context.Context
	level LogLevel
}

func newSDKLogAdapter(ctx context.Context, level LogLevel) *sdkLogAdapter {
	return &sdkLogAdapter{ctx: ctx, level: level}
}

func (a *sdkLogAdapter) Debugf(format string, args ...interface{}) {
	if a.level >= LogLevelDebug {
		tflog.SubsystemDebug(a.ctx, subsystemName, fmt.Sprintf(format, args...))
	}
}

func (a *sdkLogAdapter) Infof(format string, args ...interface{}) {
	if a.level >= LogLevelInfo {
		tflog.SubsystemInfo(a.ctx, subsystemName, fmt.Sprintf(format, args...))
	}
}

func (a *sdkLogAdapter) Warnf(format string, args ...interface{}) {
	if a.level >= LogLevelWarn {
		tflog.SubsystemWarn(a.ctx, subsystemName, fmt.Sprintf(format, args...))
	}
}

func (a *sdkLogAdapter) Errorf(format string, args ...interface{}) {
	if a.level >= LogLevelError {
		tflog.SubsystemError(a.ctx, subsystemName, fmt.Sprintf(format, args...))
	}
}
