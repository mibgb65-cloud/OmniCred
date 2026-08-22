package platform

import (
	"context"
	"time"
)

type Store interface {
	CreatePlatform(context.Context, Platform) (Platform, error)
	ListPlatforms(context.Context) ([]Platform, error)
	UpdatePlatform(context.Context, int64, string, time.Time) (Platform, error)
	DeletePlatform(context.Context, int64) error
}
