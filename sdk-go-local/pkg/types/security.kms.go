package types

// KMS Properties
type KmsPropertiesRequest struct {
	BillingPeriod *BillingPeriod `json:"billingPeriod,omitempty"`
}

type KmsPropertiesResponse struct {
	BillingPeriod *BillingPeriod `json:"billingPeriod,omitempty"`
}

// KMS Request/Response
type KmsRequest struct {
	Metadata   RegionalResourceMetadataRequest `json:"metadata"`
	Properties KmsPropertiesRequest            `json:"properties"`
}

type KmsResponse struct {
	Metadata   ResourceMetadataResponse `json:"metadata"`
	Properties KmsPropertiesResponse    `json:"properties"`
	Status     ResourceStatusResponse   `json:"status,omitempty"`
}

type KmsListResponse struct {
	ListResponse
	Values []KmsResponse `json:"values"`
}

// Key Algorithm enum
type KeyAlgorithm string

const (
	KeyAlgorithmAes KeyAlgorithm = "Aes"
	KeyAlgorithmRsa KeyAlgorithm = "Rsa"
)

// Key Creation Source enum
type KeyCreationSource string

const (
	KeyCreationSourceCmp   KeyCreationSource = "Cmp"
	KeyCreationSourceOther KeyCreationSource = "Other"
)

// Key Type enum
type KeyType string

const (
	KeyTypeSymmetric  KeyType = "Symmetric"
	KeyTypeAsymmetric KeyType = "Asymmetric"
)

// Key Status enum
type KeyStatus string

const (
	KeyStatusActive     KeyStatus = "Active"
	KeyStatusInCreation KeyStatus = "InCreation"
	KeyStatusDeleting   KeyStatus = "Deleting"
	KeyStatusDeleted    KeyStatus = "Deleted"
	KeyStatusFailed     KeyStatus = "Failed"
)

// Key Request/Response
type KeyRequest struct {
	Name      string       `json:"name"`
	Algorithm KeyAlgorithm `json:"algorithm"`
}

type KeyResponse struct {
	KeyID          *string            `json:"keyId,omitempty"`
	PrivateKeyID   *string            `json:"privateKeyId,omitempty"`
	Name           *string            `json:"name,omitempty"`
	Algorithm      *KeyAlgorithm      `json:"algorithm,omitempty"`
	CreationSource *KeyCreationSource `json:"creationSource,omitempty"`
	Type           *KeyType           `json:"type,omitempty"`
	Status         *KeyStatus         `json:"status,omitempty"`
}

type KeyListResponse struct {
	ListResponse
	Values []KeyResponse `json:"values"`
}

// Service Status enum
type ServiceStatus string

const (
	ServiceStatusInCreation           ServiceStatus = "InCreation"
	ServiceStatusActive               ServiceStatus = "Active"
	ServiceStatusUpdating             ServiceStatus = "Updating"
	ServiceStatusDeleting             ServiceStatus = "Deleting"
	ServiceStatusDeleted              ServiceStatus = "Deleted"
	ServiceStatusFailed               ServiceStatus = "Failed"
	ServiceStatusCertificateAvailable ServiceStatus = "CertificateAvailable"
)

// KMIP Request/Response
type KmipRequest struct {
	Name string `json:"name"`
}

type KmipResponse struct {
	ID           *string        `json:"id,omitempty"`
	Name         *string        `json:"name,omitempty"`
	Type         *string        `json:"type,omitempty"`
	Status       *ServiceStatus `json:"status,omitempty"`
	CreationDate *string        `json:"creationDate,omitempty"` // date-time format
	DeletionDate *string        `json:"deletionDate,omitempty"` // date-time format, nullable
}

type KmipListResponse struct {
	ListResponse
	Values []KmipResponse `json:"values"`
}

// KMIP Certificate Download Response
type KmipCertificateResponse struct {
	Key  string `json:"key"`
	Cert string `json:"cert"`
}
