package types

import "time"

type GrantUserCommon struct {
	Username string `json:"username"`
}

type GrantRoleCommon struct {
	Name string `json:"name"`
}

type GrantDatabaseResponse struct {
	Name string `json:"name"`
}
type GrantRequest struct {
	User GrantUserCommon `json:"user"`
	Role GrantRoleCommon `json:"role"`
}

type GrantResponse struct {
	User         GrantUserCommon       `json:"user"`
	Role         GrantRoleCommon       `json:"role"`
	Database     GrantDatabaseResponse `json:"database"`
	CreationDate *time.Time            `json:"creationDate,omitempty"`
	CreatedBy    *string               `json:"createdBy,omitempty"`
}

type GrantListResponse struct {
	ListResponse
	Values []GrantResponse `json:"values"`
}
