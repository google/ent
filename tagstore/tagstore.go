package tagstore

import "context"

type TagStore interface {
	Set(ctx context.Context, name string, value []byte) error
	Get(ctx context.Context, name string) ([]byte, error)
	// TODO: Support prefix.
	List(ctx context.Context) ([]string, error)
}
