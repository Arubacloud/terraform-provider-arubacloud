package types

import "time"

type UserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserResponse struct {
	Username     string     `json:"username"`
	CreationDate *time.Time `json:"creationDate,omitempty"`
	CreatedBy    *string    `json:"createdBy,omitempty"`
}

type DatabaseUserListResponse struct {
	ListResponse
	Values []UserResponse `json:"values"`
}
