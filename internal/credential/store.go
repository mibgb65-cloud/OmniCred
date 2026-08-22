package credential

import "context"

type Store interface {
	Create(context.Context, Credential) (Credential, error)
	Get(context.Context, int64) (Credential, error)
	List(context.Context, Filter) ([]Credential, error)
	Update(context.Context, Credential) (Credential, error)
	Delete(context.Context, int64) error
}
