package schedule

// Path constants follow the server-canonical lowerCamelCase rule:
//   - Single-word collections stay lowercase: jobs.
//
// Do not flatten these to all-lowercase. Downstream provisioners store and re-emit
// request URIs verbatim, so a casing change causes silent provisioning failures.
// Verified via examples/all-resources/ create.log (2026-05-28, commit f548a4f alignment).
const (
	// Schedule Jobs paths
	JobsPath = "/projects/%s/providers/Aruba.Schedule/jobs"
	JobPath  = "/projects/%s/providers/Aruba.Schedule/jobs/%s"
)
