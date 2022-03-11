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

func (s Store) Get(ctx context.Context, digest utils.Digest) ([]byte, error) {
	b, err := s.Inner.Get(ctx, string(digest))
	if err != nil {
		return nil, err
	}
	actualDigest := utils.ComputeDigest(b)
	if actualDigest != digest {
		return nil, fmt.Errorf("mismatching digest: wanted:%q got:%q", string(digest), string(actualDigest))
	}
	return b, nil
}

func (s Store) Put(ctx context.Context, b []byte) (utils.Digest, error) {
	h := utils.ComputeDigest(b)
	err := s.Inner.Put(ctx, string(h), b)
	if err != nil {
		// Return digest anyways, useful for logging errors.
		return h, err
	}
	return h, nil
}

func (s Store) Has(ctx context.Context, digest utils.Digest) (bool, error) {
	return s.Inner.Has(ctx, string(digest))
}
