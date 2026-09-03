package identity

import "time"

type Profile struct {
	ID            int64     `json:"id"`
	Country       string    `json:"country"`
	FullName      string    `json:"full_name"`
	LocalizedName string    `json:"localized_name"`
	FirstName     string    `json:"first_name"`
	MiddleName    string    `json:"middle_name"`
	LastName      string    `json:"last_name"`
	Gender        string    `json:"gender"`
	BirthDate     string    `json:"birth_date"`
	StreetAddress string    `json:"street_address"`
	City          string    `json:"city"`
	Region        string    `json:"region"`
	PostalCode    string    `json:"postal_code"`
	Phone         string    `json:"phone"`
	Email         string    `json:"email"`
	Password      string    `json:"password"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CreateInput struct {
	Country       string `json:"country"`
	FullName      string `json:"full_name"`
	LocalizedName string `json:"localized_name"`
	FirstName     string `json:"first_name"`
	MiddleName    string `json:"middle_name"`
	LastName      string `json:"last_name"`
	Gender        string `json:"gender"`
	BirthDate     string `json:"birth_date"`
	StreetAddress string `json:"street_address"`
	City          string `json:"city"`
	Region        string `json:"region"`
	PostalCode    string `json:"postal_code"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	Password      string `json:"password"`
}

type UpdateInput = CreateInput

type Filter struct {
	Country string
	Query   string
}
