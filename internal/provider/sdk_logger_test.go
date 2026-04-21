package provider

import (
	"context"
	"testing"
)

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input   string
		want    LogLevel
		wantErr bool
	}{
		{"", LogLevelOff, false},
		{"OFF", LogLevelOff, false},
		{"off", LogLevelOff, false},
		{"Off", LogLevelOff, false},
		{"ERROR", LogLevelError, false},
		{"error", LogLevelError, false},
		{"WARN", LogLevelWarn, false},
		{"warn", LogLevelWarn, false},
		{"INFO", LogLevelInfo, false},
		{"info", LogLevelInfo, false},
		{"DEBUG", LogLevelDebug, false},
		{"debug", LogLevelDebug, false},
		{"TRACE", LogLevelTrace, false},
		{"trace", LogLevelTrace, false},
		{"Debug", LogLevelDebug, false},
		{"INVALID", LogLevelOff, true},
		{"verbose", LogLevelOff, true},
		{"1", LogLevelOff, true},
	}

	for _, tc := range tests {
		got, err := ParseLogLevel(tc.input)
		if (err != nil) != tc.wantErr {
			t.Errorf("ParseLogLevel(%q): wantErr=%v, got err=%v", tc.input, tc.wantErr, err)
			continue
		}
		if !tc.wantErr && got != tc.want {
			t.Errorf("ParseLogLevel(%q): want %v, got %v", tc.input, tc.want, got)
		}
		if tc.wantErr && got != LogLevelOff {
			t.Errorf("ParseLogLevel(%q) error case: want fallback LogLevelOff, got %v", tc.input, got)
		}
	}
}

func TestLogLevelString(t *testing.T) {
	tests := []struct {
		level LogLevel
		want  string
	}{
		{LogLevelOff, "OFF"},
		{LogLevelError, "ERROR"},
		{LogLevelWarn, "WARN"},
		{LogLevelInfo, "INFO"},
		{LogLevelDebug, "DEBUG"},
		{LogLevelTrace, "TRACE"},
	}
	for _, tc := range tests {
		if got := tc.level.String(); got != tc.want {
			t.Errorf("LogLevel(%d).String(): want %q, got %q", tc.level, tc.want, got)
		}
	}
}

func TestLogLevelOrdering(t *testing.T) {
	// Verify that the level constants are ordered from least to most verbose,
	// which is the invariant the adapter's >= comparisons depend on.
	if LogLevelOff >= LogLevelError ||
		LogLevelError >= LogLevelWarn ||
		LogLevelWarn >= LogLevelInfo ||
		LogLevelInfo >= LogLevelDebug ||
		LogLevelDebug >= LogLevelTrace {
		t.Error("LogLevel constants are not in ascending verbosity order")
	}
}

func TestSDKLogAdapterLevel(t *testing.T) {
	// Verify that newSDKLogAdapter stores the level correctly.
	for _, level := range []LogLevel{LogLevelOff, LogLevelError, LogLevelWarn, LogLevelInfo, LogLevelDebug, LogLevelTrace} {
		a := newSDKLogAdapter(context.TODO(), level)
		if a.level != level {
			t.Errorf("newSDKLogAdapter(level=%v): adapter.level = %v", level, a.level)
		}
	}
}

func TestSDKLogAdapterGating(t *testing.T) {
	// For each configured level, verify which SDK log methods would fire.
	// We do this by inspecting the adapter's level field and the >= invariant
	// rather than invoking tflog (which requires a wired-up context).
	tests := []struct {
		level      LogLevel
		debugFires bool
		infoFires  bool
		warnFires  bool
		errorFires bool
	}{
		{LogLevelOff, false, false, false, false},
		{LogLevelError, false, false, false, true},
		{LogLevelWarn, false, false, true, true},
		{LogLevelInfo, false, true, true, true},
		{LogLevelDebug, true, true, true, true},
		{LogLevelTrace, true, true, true, true}, // Trace >= Debug, so all fire
	}

	for _, tc := range tests {
		a := newSDKLogAdapter(nil, tc.level)
		if got := a.level >= LogLevelDebug; got != tc.debugFires {
			t.Errorf("level=%v: Debugf fires=%v, want %v", tc.level, got, tc.debugFires)
		}
		if got := a.level >= LogLevelInfo; got != tc.infoFires {
			t.Errorf("level=%v: Infof fires=%v, want %v", tc.level, got, tc.infoFires)
		}
		if got := a.level >= LogLevelWarn; got != tc.warnFires {
			t.Errorf("level=%v: Warnf fires=%v, want %v", tc.level, got, tc.warnFires)
		}
		if got := a.level >= LogLevelError; got != tc.errorFires {
			t.Errorf("level=%v: Errorf fires=%v, want %v", tc.level, got, tc.errorFires)
		}
	}
}
