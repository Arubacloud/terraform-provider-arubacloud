package types

import "time"

type DatabaseRequest struct {
	Name string `json:"name"`
}

type DatabaseResponse struct {
	Name         string     `json:"name"`
	CreationDate *time.Time `json:"creationDate,omitempty"`
	CreatedBy    *string    `json:"createdBy,omitempty"`
}

type DatabaseListResponse struct {
	ListResponse
	Values []DatabaseResponse `json:"values"`
}
