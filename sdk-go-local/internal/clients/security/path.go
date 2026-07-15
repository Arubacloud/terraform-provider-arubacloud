package security

// Path constants follow the server-canonical lowerCamelCase rule:
//   - Single-word / acronym collections stay lowercase: kms, kmip, keys.
//
// Do not flatten these to all-lowercase. Downstream provisioners store and re-emit
// request URIs verbatim, so a casing change causes silent provisioning failures.
// Verified via examples/all-resources/ create.log (2026-05-28, commit f548a4f alignment).
const (
	// KMS paths
	KMSsPath = "/projects/%s/providers/Aruba.Security/kms"
	KMSPath  = "/projects/%s/providers/Aruba.Security/kms/%s"

	// KMIP paths (nested under KMS)
	KmipsPath        = "/projects/%s/providers/Aruba.Security/kms/%s/kmip"
	KmipPath         = "/projects/%s/providers/Aruba.Security/kms/%s/kmip/%s"
	KmipDownloadPath = "/projects/%s/providers/Aruba.Security/kms/%s/kmip/%s/download"

	// Key paths (nested under KMS)
	KeysPath = "/projects/%s/providers/Aruba.Security/kms/%s/keys"
	KeyPath  = "/projects/%s/providers/Aruba.Security/kms/%s/keys/%s"
)
