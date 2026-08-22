package credential

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"
)

type Service struct {
	store Store
	now   func() time.Time
}

func NewService(store Store) *Service {
	return &Service{store: store, now: time.Now}
}

func (service *Service) Create(ctx context.Context, input CreateInput) (Credential, error) {
	normalized, err := validateInput(input)
	if err != nil {
		return Credential{}, err
	}

	now := service.now().UTC()
	return service.store.Create(ctx, Credential{
		Provider:  normalized.Provider,
		Account:   normalized.Account,
		Username:  normalized.Username,
		Password:  normalized.Password,
		CreatedAt: now,
		UpdatedAt: now,
	})
}

func (service *Service) Get(ctx context.Context, id int64) (Credential, error) {
	if err := validateID(id); err != nil {
		return Credential{}, err
	}
	return service.store.Get(ctx, id)
}

func (service *Service) List(ctx context.Context, filter Filter) ([]Credential, error) {
	filter.Provider = strings.ToLower(strings.TrimSpace(filter.Provider))
	filter.Query = strings.TrimSpace(filter.Query)
	return service.store.List(ctx, filter)
}

func (service *Service) Update(ctx context.Context, id int64, input UpdateInput) (Credential, error) {
	if err := validateID(id); err != nil {
		return Credential{}, err
	}
	normalized, err := validateInput(input)
	if err != nil {
		return Credential{}, err
	}

	existing, err := service.store.Get(ctx, id)
	if err != nil {
		return Credential{}, err
	}
	existing.Provider = normalized.Provider
	existing.Account = normalized.Account
	existing.Username = normalized.Username
	existing.Password = normalized.Password
	existing.UpdatedAt = service.now().UTC()
	return service.store.Update(ctx, existing)
}

func (service *Service) Delete(ctx context.Context, id int64) error {
	if err := validateID(id); err != nil {
		return err
	}
	return service.store.Delete(ctx, id)
}

func validateID(id int64) error {
	if id <= 0 {
		return &ValidationError{Field: "id", Message: "must be a positive integer"}
	}
	return nil
}

func validateInput(input CreateInput) (CreateInput, error) {
	input.Provider = strings.ToLower(strings.TrimSpace(input.Provider))
	input.Account = strings.TrimSpace(input.Account)
	input.Username = strings.TrimSpace(input.Username)

	checks := []struct {
		field    string
		value    string
		required bool
		maximum  int
	}{
		{"provider", input.Provider, true, maxProviderLength},
		{"account", input.Account, true, maxTextLength},
		{"username", input.Username, false, maxTextLength},
		{"password", input.Password, true, maxPasswordLength},
	}
	for _, check := range checks {
		if check.required && check.value == "" {
			return CreateInput{}, &ValidationError{Field: check.field, Message: "is required"}
		}
		if utf8.RuneCountInString(check.value) > check.maximum {
			return CreateInput{}, &ValidationError{Field: check.field, Message: "is too long"}
		}
	}
	return input, nil
}
