package objectstore

import (
	"context"
	"fmt"

	"github.com/google/ent/datastore"
	"github.com/google/ent/utils"
)

type Store struct {
	Inner datastore.DataStore
}

func (s Store) Get(ctx context.Context, h utils.Hash) ([]byte, error) {
	b, err := s.Inner.Get(ctx, string(h))
	if err != nil {
		return nil, err
	}
	actualHash := utils.ComputeHash(b)
	if actualHash != h {
		return nil, fmt.Errorf("mismatching hashes: wanted:%q got:%q", string(h), string(actualHash))
	}
	return b, nil
}

func (s Store) Put(ctx context.Context, b []byte) (utils.Hash, error) {
	h := utils.ComputeHash(b)
	err := s.Inner.Put(ctx, string(h), b)
	if err != nil {
		// Return digest anyways, useful for logging errors.
		return h, err
	}
	return h, nil
}

func (s Store) Has(ctx context.Context, h utils.Hash) (bool, error) {
	return s.Inner.Has(ctx, string(h))
}
