package platform

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

func (service *Service) Create(ctx context.Context, input CreateInput) (Platform, error) {
	name, err := validateName(input.Name)
	if err != nil {
		return Platform{}, err
	}
	now := service.now().UTC()
	return service.store.CreatePlatform(ctx, Platform{Name: name, CreatedAt: now, UpdatedAt: now})
}

func (service *Service) List(ctx context.Context) ([]Platform, error) {
	return service.store.ListPlatforms(ctx)
}

func (service *Service) Update(ctx context.Context, id int64, input UpdateInput) (Platform, error) {
	if err := validateID(id); err != nil {
		return Platform{}, err
	}
	name, err := validateName(input.Name)
	if err != nil {
		return Platform{}, err
	}
	return service.store.UpdatePlatform(ctx, id, name, service.now().UTC())
}

func (service *Service) Delete(ctx context.Context, id int64) error {
	if err := validateID(id); err != nil {
		return err
	}
	return service.store.DeletePlatform(ctx, id)
}

func validateID(id int64) error {
	if id <= 0 {
		return &ValidationError{Field: "id", Message: "must be a positive integer"}
	}
	return nil
}

func validateName(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return "", &ValidationError{Field: "name", Message: "is required"}
	}
	if utf8.RuneCountInString(value) > maxNameLength {
		return "", &ValidationError{Field: "name", Message: "is too long"}
	}
	return value, nil
}
