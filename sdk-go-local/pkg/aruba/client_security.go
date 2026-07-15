package aruba

import "context"

type SecurityClient interface {
	KMS() KMSClient
	Keys() KeysClient
	Kmips() KmipsClient
}

type securityClientImpl struct {
	kmsClient   KMSClient
	keysClient  KeysClient
	kmipsClient KmipsClient
}

var _ SecurityClient = (*securityClientImpl)(nil)

func (c *securityClientImpl) KMS() KMSClient     { return c.kmsClient }
func (c *securityClientImpl) Keys() KeysClient   { return c.keysClient }
func (c *securityClientImpl) Kmips() KmipsClient { return c.kmipsClient }

type KMSClient interface {
	List(ctx context.Context, project Ref, opts ...CallOption) (*List[*KMS], error)
	Get(ctx context.Context, ref Ref, opts ...CallOption) (*KMS, error)
	Create(ctx context.Context, k *KMS, opts ...CallOption) (*KMS, error)
	Update(ctx context.Context, k *KMS, opts ...CallOption) (*KMS, error)
	Delete(ctx context.Context, ref Ref, opts ...CallOption) error
}

// KeysClient is the wrapper-level interface for Key CRUD operations.
// No Update — Family B resource with no update operation.
type KeysClient interface {
	List(ctx context.Context, kms Ref, opts ...CallOption) (*List[*Key], error)
	Get(ctx context.Context, ref Ref, opts ...CallOption) (*Key, error)
	Create(ctx context.Context, k *Key, opts ...CallOption) (*Key, error)
	Delete(ctx context.Context, ref Ref, opts ...CallOption) error
}

// KmipsClient is the wrapper-level interface for Kmip CRUD operations.
// No Update — Family B resource with no update operation.
// Download retrieves the KMIP certificate (key+cert pair) for a service instance.
type KmipsClient interface {
	List(ctx context.Context, kms Ref, opts ...CallOption) (*List[*Kmip], error)
	Get(ctx context.Context, ref Ref, opts ...CallOption) (*Kmip, error)
	Create(ctx context.Context, km *Kmip, opts ...CallOption) (*Kmip, error)
	Delete(ctx context.Context, ref Ref, opts ...CallOption) error
	Download(ctx context.Context, ref Ref, opts ...CallOption) (*KmipCertificate, error)
}
