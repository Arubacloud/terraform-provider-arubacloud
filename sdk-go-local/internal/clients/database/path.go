package database

// Path constants follow the server-canonical lowerCamelCase rule:
//   - Single-word / acronym collections stay lowercase: dbaas, databases, backups,
//     grants, users.
//
// Do not flatten these to all-lowercase. Downstream provisioners store and re-emit
// request URIs verbatim, so a casing change causes silent provisioning failures.
// Verified via examples/all-resources/ create.log (2026-05-28, commit f548a4f alignment).
const (
	// DBaaS paths
	DBaaSPath     = "/projects/%s/providers/Aruba.Database/dbaas"
	DBaaSItemPath = "/projects/%s/providers/Aruba.Database/dbaas/%s"

	// Database paths
	DatabaseInstancesPath = "/projects/%s/providers/Aruba.Database/dbaas/%s/databases"
	DatabaseInstancePath  = "/projects/%s/providers/Aruba.Database/dbaas/%s/databases/%s"

	// Backup paths
	BackupsPath = "/projects/%s/providers/Aruba.Database/backups"
	BackupPath  = "/projects/%s/providers/Aruba.Database/backups/%s"

	// GrantDatabase Paths
	GrantsPath    = "/projects/%s/providers/Aruba.Database/dbaas/%s/databases/%s/grants"
	GrantItemPath = "/projects/%s/providers/Aruba.Database/dbaas/%s/databases/%s/grants/%s"

	// User paths
	UsersPath    = "/projects/%s/providers/Aruba.Database/dbaas/%s/users"
	UserItemPath = "/projects/%s/providers/Aruba.Database/dbaas/%s/users/%s"
)
