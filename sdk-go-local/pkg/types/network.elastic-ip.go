package types

type ElasticIPPropertiesRequest struct {
	BillingPlanCommon *BillingPlanCommon `json:"billingPlan,omitempty"`
}

type ElasticIPPropertiesResponse struct {
	LinkedResources []LinkedResourceCommon `json:"linkedResources,omitempty"`

	Address           *string            `json:"address,omitempty"`
	BillingPlanCommon *BillingPlanCommon `json:"billingPlan,omitempty"`
}

type ElasticIPRequest struct {
	Metadata   RegionalResourceMetadataRequest `json:"metadata"`
	Properties ElasticIPPropertiesRequest      `json:"properties"`
}

type ElasticIPResponse struct {
	Metadata   ResourceMetadataResponse    `json:"metadata"`
	Properties ElasticIPPropertiesResponse `json:"properties"`

	Status ResourceStatusResponse `json:"status,omitempty"`
}

type ElasticIPListResponse struct {
	ListResponse
	Values []ElasticIPResponse `json:"values"`
}
