package identity

import "context"

type Store interface {
	CreateIdentity(context.Context, Profile) (Profile, error)
	GetIdentity(context.Context, int64) (Profile, error)
	ListIdentities(context.Context, Filter) ([]Profile, error)
	UpdateIdentity(context.Context, Profile) (Profile, error)
	DeleteIdentity(context.Context, int64) error
}
