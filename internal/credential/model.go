package credential

import "time"

const (
	maxProviderLength = 100
	maxTextLength     = 4096
	maxPasswordLength = 16384
)

type Credential struct {
	ID        int64     `json:"id"`
	Provider  string    `json:"provider"`
	Account   string    `json:"account"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateInput struct {
	Provider string `json:"provider"`
	Account  string `json:"account"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type UpdateInput = CreateInput

type Filter struct {
	Provider string
	Query    string
}
