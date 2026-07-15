package aruba

import "context"

type ContainerClient interface {
	KaaS() KaaSClient
	ContainerRegistry() ContainerRegistryClient
}

type containerClientImpl struct {
	kaasClient              KaaSClient
	containerRegistryClient ContainerRegistryClient
}

// ContainerRegistry implements ContainerClient.
func (c *containerClientImpl) ContainerRegistry() ContainerRegistryClient {
	return c.containerRegistryClient
}

var _ ContainerClient = (*containerClientImpl)(nil)

func (c *containerClientImpl) KaaS() KaaSClient {
	return c.kaasClient
}

type KaaSClient interface {
	List(ctx context.Context, project Ref, opts ...CallOption) (*List[*KaaS], error)
	Get(ctx context.Context, ref Ref, opts ...CallOption) (*KaaS, error)
	Create(ctx context.Context, k *KaaS, opts ...CallOption) (*KaaS, error)
	Update(ctx context.Context, k *KaaS, opts ...CallOption) (*KaaS, error)
	Delete(ctx context.Context, ref Ref, opts ...CallOption) error
}

type ContainerRegistryClient interface {
	List(ctx context.Context, project Ref, opts ...CallOption) (*List[*ContainerRegistry], error)
	Get(ctx context.Context, ref Ref, opts ...CallOption) (*ContainerRegistry, error)
	Create(ctx context.Context, r *ContainerRegistry, opts ...CallOption) (*ContainerRegistry, error)
	Update(ctx context.Context, r *ContainerRegistry, opts ...CallOption) (*ContainerRegistry, error)
	Delete(ctx context.Context, ref Ref, opts ...CallOption) error
}
