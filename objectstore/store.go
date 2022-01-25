package objectstore

import (
	"context"
	"fmt"

	"github.com/google/ent/datastore"
	"github.com/google/ent/utils"
	"github.com/multiformats/go-multihash"
)

const hashType = multihash.SHA2_256

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

func (s Store) Add(ctx context.Context, b []byte) (utils.Hash, error) {
	h := utils.ComputeHash(b)
	err := s.Inner.Set(ctx, string(h), b)
	if err != nil {
		return "", err
	}
	return h, nil
}
