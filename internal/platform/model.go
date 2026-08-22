package platform

import "time"

const maxNameLength = 100

type Platform struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	CredentialCount int       `json:"credential_count"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type CreateInput struct {
	Name string `json:"name"`
}

type UpdateInput = CreateInput
