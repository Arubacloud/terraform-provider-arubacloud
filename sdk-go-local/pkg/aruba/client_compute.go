package aruba

import "context"

type ComputeClient interface {
	CloudServers() CloudServersClient
	KeyPairs() KeyPairsClient
}

type computeClientImpl struct {
	cloudServerClient CloudServersClient
	keyPairClient     KeyPairsClient
}

var _ ComputeClient = (*computeClientImpl)(nil)

func (c *computeClientImpl) CloudServers() CloudServersClient {
	return c.cloudServerClient
}

func (c *computeClientImpl) KeyPairs() KeyPairsClient {
	return c.keyPairClient
}

type CloudServersClient interface {
	List(ctx context.Context, project Ref, opts ...CallOption) (*List[*CloudServer], error)
	Get(ctx context.Context, ref Ref, opts ...CallOption) (*CloudServer, error)
	Create(ctx context.Context, server *CloudServer, opts ...CallOption) (*CloudServer, error)
	Update(ctx context.Context, server *CloudServer, opts ...CallOption) (*CloudServer, error)
	Delete(ctx context.Context, ref Ref, opts ...CallOption) error
}

type KeyPairsClient interface {
	List(ctx context.Context, project Ref, opts ...CallOption) (*List[*KeyPair], error)
	Get(ctx context.Context, ref Ref, opts ...CallOption) (*KeyPair, error)
	Create(ctx context.Context, kp *KeyPair, opts ...CallOption) (*KeyPair, error)
	Delete(ctx context.Context, ref Ref, opts ...CallOption) error
}
