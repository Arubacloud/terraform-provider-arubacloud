package aruba

// Version is the current SDK release version.
const Version = "1.0.6"

// defaultUserAgent is injected as the User-Agent header on every request
// unless overridden via Options.WithUserAgent.
const defaultUserAgent = "sdk-go@" + Version
