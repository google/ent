package objectstore

import (
	"bytes"
	"context"
	"fmt"

	"github.com/google/ent/datastore"
	"github.com/multiformats/go-multihash"
)

const hashType = multihash.SHA2_256

type Store struct {
	Inner datastore.DataStore
}

func (s Store) Get(ctx context.Context, h multihash.Multihash) ([]byte, error) {
	b, err := s.Inner.Get(ctx, h.HexString())
	if err != nil {
		return nil, err
	}
	actualHash, err := multihash.Sum(b, hashType, -1)
	if err != nil {
		return nil, err
	}
	if bytes.Compare(actualHash, h) != 0 {
		return nil, fmt.Errorf("mismatching hashes: wanted:%q got:%q", h.String(), actualHash.String())
	}
	return b, nil
}

func (s Store) Add(ctx context.Context, b []byte) (multihash.Multihash, error) {
	h, err := multihash.Sum(b, hashType, -1)
	if err != nil {
		return nil, err
	}
	err = s.Inner.Set(ctx, h.HexString(), b)
	if err != nil {
		return nil, err
	}
	return h, nil
}
