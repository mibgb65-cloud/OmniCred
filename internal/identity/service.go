package identity

import (
	"context"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	maxNameLength     = 256
	maxTextLength     = 2048
	maxEmailLength    = 320
	maxPasswordLength = 16384
)

type Service struct {
	store Store
	now   func() time.Time
}

func NewService(store Store) *Service {
	return &Service{store: store, now: time.Now}
}

func (service *Service) Create(ctx context.Context, input CreateInput) (Profile, error) {
	normalized, err := validateInput(input)
	if err != nil {
		return Profile{}, err
	}
	now := service.now().UTC()
	return service.store.CreateIdentity(ctx, profileFromInput(normalized, now, now))
}

func (service *Service) Get(ctx context.Context, id int64) (Profile, error) {
	if err := validateID(id); err != nil {
		return Profile{}, err
	}
	return service.store.GetIdentity(ctx, id)
}

func (service *Service) List(ctx context.Context, filter Filter) ([]Profile, error) {
	filter.Country = strings.TrimSpace(filter.Country)
	filter.Query = strings.TrimSpace(filter.Query)
	return service.store.ListIdentities(ctx, filter)
}

func (service *Service) Update(ctx context.Context, id int64, input UpdateInput) (Profile, error) {
	if err := validateID(id); err != nil {
		return Profile{}, err
	}
	normalized, err := validateInput(input)
	if err != nil {
		return Profile{}, err
	}
	existing, err := service.store.GetIdentity(ctx, id)
	if err != nil {
		return Profile{}, err
	}
	updated := profileFromInput(normalized, existing.CreatedAt, service.now().UTC())
	updated.ID = existing.ID
	return service.store.UpdateIdentity(ctx, updated)
}

func (service *Service) Delete(ctx context.Context, id int64) error {
	if err := validateID(id); err != nil {
		return err
	}
	return service.store.DeleteIdentity(ctx, id)
}

func profileFromInput(input CreateInput, createdAt, updatedAt time.Time) Profile {
	return Profile{
		Country: input.Country, FullName: input.FullName, LocalizedName: input.LocalizedName,
		FirstName: input.FirstName, MiddleName: input.MiddleName, LastName: input.LastName,
		Gender: input.Gender, BirthDate: input.BirthDate, StreetAddress: input.StreetAddress,
		City: input.City, Region: input.Region, PostalCode: input.PostalCode, Phone: input.Phone,
		Email: input.Email, Password: input.Password, CreatedAt: createdAt, UpdatedAt: updatedAt,
	}
}

func validateID(id int64) error {
	if id <= 0 {
		return &ValidationError{Field: "id", Message: "must be a positive integer"}
	}
	return nil
}

func validateInput(input CreateInput) (CreateInput, error) {
	input.Country = strings.TrimSpace(input.Country)
	input.FullName = strings.TrimSpace(input.FullName)
	input.LocalizedName = strings.TrimSpace(input.LocalizedName)
	input.FirstName = strings.TrimSpace(input.FirstName)
	input.MiddleName = strings.TrimSpace(input.MiddleName)
	input.LastName = strings.TrimSpace(input.LastName)
	input.Gender = strings.ToLower(strings.TrimSpace(input.Gender))
	input.BirthDate = strings.TrimSpace(input.BirthDate)
	input.StreetAddress = strings.TrimSpace(input.StreetAddress)
	input.City = strings.TrimSpace(input.City)
	input.Region = strings.TrimSpace(input.Region)
	input.PostalCode = strings.TrimSpace(input.PostalCode)
	input.Phone = strings.TrimSpace(input.Phone)
	input.Email = strings.TrimSpace(input.Email)

	checks := []struct {
		field    string
		value    string
		required bool
		maximum  int
	}{
		{"country", input.Country, true, maxNameLength},
		{"full_name", input.FullName, true, maxNameLength},
		{"localized_name", input.LocalizedName, false, maxNameLength},
		{"first_name", input.FirstName, false, maxNameLength},
		{"middle_name", input.MiddleName, false, maxNameLength},
		{"last_name", input.LastName, false, maxNameLength},
		{"birth_date", input.BirthDate, false, 10},
		{"street_address", input.StreetAddress, false, maxTextLength},
		{"city", input.City, false, maxNameLength},
		{"region", input.Region, false, maxNameLength},
		{"postal_code", input.PostalCode, false, 32},
		{"phone", input.Phone, false, 64},
		{"email", input.Email, false, maxEmailLength},
		{"password", input.Password, false, maxPasswordLength},
	}
	for _, check := range checks {
		if check.required && check.value == "" {
			return CreateInput{}, &ValidationError{Field: check.field, Message: "is required"}
		}
		if utf8.RuneCountInString(check.value) > check.maximum {
			return CreateInput{}, &ValidationError{Field: check.field, Message: "is too long"}
		}
	}

	if input.Gender != "" && input.Gender != "male" && input.Gender != "female" && input.Gender != "other" {
		return CreateInput{}, &ValidationError{Field: "gender", Message: "must be male, female, or other"}
	}
	if input.BirthDate != "" {
		date, err := time.Parse("2006-01-02", input.BirthDate)
		if err != nil || date.Format("2006-01-02") != input.BirthDate {
			return CreateInput{}, &ValidationError{Field: "birth_date", Message: "must use YYYY-MM-DD format"}
		}
	}
	if input.Email != "" {
		address, err := mail.ParseAddress(input.Email)
		if err != nil || address.Address != input.Email {
			return CreateInput{}, &ValidationError{Field: "email", Message: "must be a valid email address"}
		}
	}
	return input, nil
}
