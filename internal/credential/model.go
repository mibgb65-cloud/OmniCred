package credential

import "time"

const (
	maxProviderLength = 100
	maxTextLength     = 4096
	maxPasswordLength = 16384
)

type Credential struct {
	ID         int64     `json:"id"`
	Provider   string    `json:"provider"`
	Account    string    `json:"account"`
	Username   string    `json:"username"`
	Password   string    `json:"password"`
	TOTPSecret string    `json:"totp_secret"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type CreateInput struct {
	Provider   string `json:"provider"`
	Account    string `json:"account"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	TOTPSecret string `json:"totp_secret"`
}

type UpdateInput = CreateInput

type Filter struct {
	Provider string
	Query    string
}

type TOTPCode struct {
	CredentialID int64  `json:"credential_id"`
	Code         string `json:"code"`
}

type TOTPCodeList struct {
	Items            []TOTPCode `json:"items"`
	SecondsRemaining int        `json:"seconds_remaining"`
	Period           int        `json:"period"`
	GeneratedAt      time.Time  `json:"generated_at"`
}
