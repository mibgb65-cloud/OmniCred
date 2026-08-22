package credential

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"omnicred/internal/totp"
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
		Provider:      normalized.Provider,
		Account:       normalized.Account,
		Username:      normalized.Username,
		Password:      normalized.Password,
		TOTPSecret:    normalized.TOTPSecret,
		RecoveryCodes: normalized.RecoveryCodes,
		CreatedAt:     now,
		UpdatedAt:     now,
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
	existing.TOTPSecret = normalized.TOTPSecret
	existing.RecoveryCodes = normalized.RecoveryCodes
	existing.UpdatedAt = service.now().UTC()
	return service.store.Update(ctx, existing)
}

func (service *Service) TOTPCodes(ctx context.Context) (TOTPCodeList, error) {
	items, err := service.store.List(ctx, Filter{})
	if err != nil {
		return TOTPCodeList{}, err
	}
	now := service.now().UTC()
	codes := make([]TOTPCode, 0)
	for _, item := range items {
		if item.TOTPSecret == "" {
			continue
		}
		code, err := totp.Generate(item.TOTPSecret, now)
		if err != nil {
			return TOTPCodeList{}, fmt.Errorf("generate TOTP for credential %d: %w", item.ID, err)
		}
		codes = append(codes, TOTPCode{CredentialID: item.ID, Code: code})
	}
	return TOTPCodeList{
		Items: codes, SecondsRemaining: totp.SecondsRemaining(now), Period: totp.Period, GeneratedAt: now,
	}, nil
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
	input.TOTPSecret = strings.TrimSpace(input.TOTPSecret)

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
		{"totp_secret", input.TOTPSecret, false, 512},
	}
	for _, check := range checks {
		if check.required && check.value == "" {
			return CreateInput{}, &ValidationError{Field: check.field, Message: "is required"}
		}
		if utf8.RuneCountInString(check.value) > check.maximum {
			return CreateInput{}, &ValidationError{Field: check.field, Message: "is too long"}
		}
	}
	if input.TOTPSecret != "" {
		normalized, err := totp.NormalizeSecret(input.TOTPSecret)
		if err != nil {
			return CreateInput{}, &ValidationError{Field: "totp_secret", Message: "must be a valid Base32 secret"}
		}
		input.TOTPSecret = normalized
	}
	recoveryCodes, err := normalizeRecoveryCodes(input.RecoveryCodes)
	if err != nil {
		return CreateInput{}, err
	}
	input.RecoveryCodes = recoveryCodes
	return input, nil
}

func normalizeRecoveryCodes(values []string) ([]string, error) {
	if len(values) > 100 {
		return nil, &ValidationError{Field: "recovery_codes", Message: "must contain at most 100 codes"}
	}
	normalized := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if utf8.RuneCountInString(value) > 256 {
			return nil, &ValidationError{Field: "recovery_codes", Message: "contains a code that is too long"}
		}
		if _, exists := seen[value]; exists {
			return nil, &ValidationError{Field: "recovery_codes", Message: "must not contain duplicate codes"}
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}
	return normalized, nil
}
