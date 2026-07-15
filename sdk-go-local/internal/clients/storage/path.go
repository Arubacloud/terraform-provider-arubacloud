package storage

// Path constants follow the server-canonical lowerCamelCase rule:
//   - Single-word collections stay lowercase: snapshots, backups, restores.
//   - Compound collections use lowerCamelCase: blockStorages.
//
// Do not flatten these to all-lowercase. Downstream provisioners store and re-emit
// request URIs verbatim, so a casing change causes silent provisioning failures.
// Verified via examples/all-resources/ create.log (2026-05-28, commit f548a4f alignment).
const (

	// Storage Bucket paths
	BlockStoragesPath = "/projects/%s/providers/Aruba.Storage/blockStorages"
	BlockStoragePath  = "/projects/%s/providers/Aruba.Storage/blockStorages/%s"

	//Snapshot paths
	SnapshotsPath = "/projects/%s/providers/Aruba.Storage/snapshots"
	SnapshotPath  = "/projects/%s/providers/Aruba.Storage/snapshots/%s"

	//Backup paths
	BackupsPath = "/projects/%s/providers/Aruba.Storage/backups"
	BackupPath  = "/projects/%s/providers/Aruba.Storage/backups/%s"

	//Restore paths
	RestoresPath = "/projects/%s/providers/Aruba.Storage/backups/%s/restores"
	RestorePath  = "/projects/%s/providers/Aruba.Storage/backups/%s/restores/%s"
)
