package types

type ProjectPropertiesRequest struct {

	// Optional description of the project
	Description *string `json:"description,omitempty"`

	//Indicates if it's the default project
	Default bool `json:"default"`
}

type ProjectPropertiesResponse struct {

	// Optional description of the project
	Description *string `json:"description,omitempty"`

	//Indicates if it's the default project
	Default bool `json:"default"`

	ResourcesNumber int `json:"resourcesNumber,omitempty"`
}

type ProjectRequest struct {
	// Metadata of the Project
	Metadata ResourceMetadataRequest `json:"metadata"`

	// Spec contains the Project specification
	Properties ProjectPropertiesRequest `json:"properties"`
}

type ProjectResponse struct {
	// Metadata of the Project
	Metadata ResourceMetadataResponse `json:"metadata"`
	// Spec contains the Project specification
	Properties ProjectPropertiesResponse `json:"properties"`

	//Status ResourceStatusResponse `json:"status,omitempty"`
}

type ProjectListResponse struct {
	ListResponse
	Values []ProjectResponse `json:"values"`
}
