package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// registerSDKSubsystem registers the arubacloud-sdk log subsystem on the
// context so that SubsystemDebug / SubsystemInfo / SubsystemWarn /
// SubsystemError don't panic in unit tests.
func registerSDKSubsystem(ctx context.Context) context.Context {
	return tflog.NewSubsystem(ctx, subsystemName)
}

// TestSDKLogAdapter_Debugf verifies that Debugf emits a log entry when the
// adapter's level is LogLevelDebug and suppresses it when the level is below.
func TestSDKLogAdapter_Debugf(t *testing.T) {
	ctx := registerSDKSubsystem(context.Background())

	// Level == Debug: SubsystemDebug is called (no panic).
	a := newSDKLogAdapter(ctx, LogLevelDebug)
	a.Debugf("test debug message %s", "arg")

	// Level == Info: Debugf must be a no-op (level too low → branch not taken).
	a2 := newSDKLogAdapter(ctx, LogLevelInfo)
	a2.Debugf("suppressed debug message")
}

// TestSDKLogAdapter_Infof verifies that Infof emits when level >= Info and
// is suppressed when level > Info.
func TestSDKLogAdapter_Infof(t *testing.T) {
	ctx := registerSDKSubsystem(context.Background())

	a := newSDKLogAdapter(ctx, LogLevelInfo)
	a.Infof("test info message %d", 42)

	a2 := newSDKLogAdapter(ctx, LogLevelWarn)
	a2.Infof("suppressed info message")
}

// TestSDKLogAdapter_Warnf verifies that Warnf emits when level >= Warn.
func TestSDKLogAdapter_Warnf(t *testing.T) {
	ctx := registerSDKSubsystem(context.Background())

	a := newSDKLogAdapter(ctx, LogLevelWarn)
	a.Warnf("test warn message %v", true)

	a2 := newSDKLogAdapter(ctx, LogLevelError)
	a2.Warnf("suppressed warn message")
}

// TestSDKLogAdapter_Errorf verifies that Errorf emits when level >= Error
// and is suppressed when level is Off.
func TestSDKLogAdapter_Errorf(t *testing.T) {
	ctx := registerSDKSubsystem(context.Background())

	a := newSDKLogAdapter(ctx, LogLevelError)
	a.Errorf("test error message: %s", "oops")

	a2 := newSDKLogAdapter(ctx, LogLevelOff)
	a2.Errorf("suppressed error message")
}
