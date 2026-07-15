package aruba

import "context"

type StorageClient interface {
	Snapshots() SnapshotsClient
	Volumes() VolumesClient
	Backups() StorageBackupsClient
	Restores() StorageRestoreClient
}

type storageClientImpl struct {
	snapshotsClient SnapshotsClient
	volumesClient   VolumesClient
	backupsClient   StorageBackupsClient
	restoresClient  StorageRestoreClient
}

var _ StorageClient = (*storageClientImpl)(nil)

func (c *storageClientImpl) Snapshots() SnapshotsClient {
	return c.snapshotsClient
}

func (c *storageClientImpl) Volumes() VolumesClient {
	return c.volumesClient
}

func (c *storageClientImpl) Backups() StorageBackupsClient {
	return c.backupsClient
}

func (c *storageClientImpl) Restores() StorageRestoreClient {
	return c.restoresClient
}

type SnapshotsClient interface {
	List(ctx context.Context, project Ref, opts ...CallOption) (*List[*Snapshot], error)
	Get(ctx context.Context, ref Ref, opts ...CallOption) (*Snapshot, error)
	Create(ctx context.Context, snap *Snapshot, opts ...CallOption) (*Snapshot, error)
	Update(ctx context.Context, snap *Snapshot, opts ...CallOption) (*Snapshot, error)
	Delete(ctx context.Context, ref Ref, opts ...CallOption) error
}

type VolumesClient interface {
	List(ctx context.Context, project Ref, opts ...CallOption) (*List[*BlockStorage], error)
	Get(ctx context.Context, ref Ref, opts ...CallOption) (*BlockStorage, error)
	Create(ctx context.Context, vol *BlockStorage, opts ...CallOption) (*BlockStorage, error)
	Update(ctx context.Context, vol *BlockStorage, opts ...CallOption) (*BlockStorage, error)
	Delete(ctx context.Context, ref Ref, opts ...CallOption) error
}

type StorageBackupsClient interface {
	List(ctx context.Context, project Ref, opts ...CallOption) (*List[*StorageBackup], error)
	Get(ctx context.Context, ref Ref, opts ...CallOption) (*StorageBackup, error)
	Create(ctx context.Context, b *StorageBackup, opts ...CallOption) (*StorageBackup, error)
	Update(ctx context.Context, b *StorageBackup, opts ...CallOption) (*StorageBackup, error)
	Delete(ctx context.Context, ref Ref, opts ...CallOption) error
}

type StorageRestoreClient interface {
	List(ctx context.Context, backup Ref, opts ...CallOption) (*List[*StorageRestore], error)
	Get(ctx context.Context, ref Ref, opts ...CallOption) (*StorageRestore, error)
	Create(ctx context.Context, r *StorageRestore, opts ...CallOption) (*StorageRestore, error)
	// Update modifies a StorageRestore resource. NOTE: Aruba Cloud platform
	// support for PUT on restore resources is not currently documented; this
	// method may return a 4xx error in practice. Prefer Create+Delete workflows.
	// See https://github.com/Arubacloud/sdk-go/issues/273
	Update(ctx context.Context, r *StorageRestore, opts ...CallOption) (*StorageRestore, error)
	Delete(ctx context.Context, ref Ref, opts ...CallOption) error
}
