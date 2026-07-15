package aruba

import "context"

type DatabaseClient interface {
	DBaaS() DBaaSClient
	Databases() DatabasesClient
	Backups() BackupsClient
	Users() UsersClient
	Grants() GrantsClient
}

type databaseClientImpl struct {
	dbaasClient     DBaaSClient
	databasesClient DatabasesClient
	backupsClient   BackupsClient
	usersClient     UsersClient
	grantsClient    GrantsClient
}

var _ DatabaseClient = (*databaseClientImpl)(nil)

func (c databaseClientImpl) DBaaS() DBaaSClient {
	return c.dbaasClient
}

func (c databaseClientImpl) Databases() DatabasesClient {
	return c.databasesClient
}

func (c databaseClientImpl) Backups() BackupsClient {
	return c.backupsClient
}

func (c databaseClientImpl) Users() UsersClient {
	return c.usersClient
}

func (c databaseClientImpl) Grants() GrantsClient {
	return c.grantsClient
}

type DBaaSClient interface {
	List(ctx context.Context, project Ref, opts ...CallOption) (*List[*DBaaS], error)
	Get(ctx context.Context, ref Ref, opts ...CallOption) (*DBaaS, error)
	Create(ctx context.Context, dbaas *DBaaS, opts ...CallOption) (*DBaaS, error)
	Update(ctx context.Context, dbaas *DBaaS, opts ...CallOption) (*DBaaS, error)
	Delete(ctx context.Context, ref Ref, opts ...CallOption) error
}

type DatabasesClient interface {
	List(ctx context.Context, dbaas Ref, opts ...CallOption) (*List[*Database], error)
	Get(ctx context.Context, ref Ref, opts ...CallOption) (*Database, error)
	Create(ctx context.Context, db *Database, opts ...CallOption) (*Database, error)
	Update(ctx context.Context, db *Database, opts ...CallOption) (*Database, error)
	Delete(ctx context.Context, ref Ref, opts ...CallOption) error
}

type BackupsClient interface {
	List(ctx context.Context, project Ref, opts ...CallOption) (*List[*DBaaSBackup], error)
	Get(ctx context.Context, ref Ref, opts ...CallOption) (*DBaaSBackup, error)
	Create(ctx context.Context, b *DBaaSBackup, opts ...CallOption) (*DBaaSBackup, error)
	Delete(ctx context.Context, ref Ref, opts ...CallOption) error
}

type UsersClient interface {
	List(ctx context.Context, dbaas Ref, opts ...CallOption) (*List[*User], error)
	Get(ctx context.Context, ref Ref, opts ...CallOption) (*User, error)
	Create(ctx context.Context, u *User, opts ...CallOption) (*User, error)
	Update(ctx context.Context, u *User, opts ...CallOption) (*User, error)
	Delete(ctx context.Context, ref Ref, opts ...CallOption) error
}

type GrantsClient interface {
	List(ctx context.Context, database Ref, opts ...CallOption) (*List[*Grant], error)
	Get(ctx context.Context, ref Ref, opts ...CallOption) (*Grant, error)
	Create(ctx context.Context, g *Grant, opts ...CallOption) (*Grant, error)
	Update(ctx context.Context, g *Grant, opts ...CallOption) (*Grant, error)
	Delete(ctx context.Context, ref Ref, opts ...CallOption) error
}
