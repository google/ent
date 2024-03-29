package objectstore

import (
	"bytes"
	"context"
	"fmt"

	"github.com/google/ent/datastore"
	"github.com/google/ent/log"
	"github.com/google/ent/utils"
	"github.com/multiformats/go-multihash"
)

type Store struct {
	Inner datastore.DataStore
}

func (s Store) Get(ctx context.Context, digest utils.Digest) ([]byte, error) {
	b, err := s.Inner.Get(ctx, digest.String())
	if err != nil {
		decodedDigest, err := multihash.Decode(digest)
		if err == nil && decodedDigest.Code == multihash.SHA2_256 {
			oldDigest := utils.DigestToHumanString(digest)
			log.Infof(ctx, "decoded digest: %v", decodedDigest)
			log.Infof(ctx, "old digest: %v", oldDigest)
			b, err = s.Inner.Get(ctx, oldDigest)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	actualDigest := utils.ComputeDigest(b)
	if !bytes.Equal(actualDigest, digest) {
		return nil, fmt.Errorf("mismatching digest: wanted:%q got:%q", digest.String(), actualDigest.String())
	}
	return b, nil
}

func (s Store) Put(ctx context.Context, b []byte) (utils.Digest, error) {
	digest := utils.ComputeDigest(b)
	err := s.Inner.Put(ctx, digest.String(), b)
	if err != nil {
		// Return digest anyways, useful for logging errors.
		return digest, err
	}
	return digest, nil
}

func (s Store) Has(ctx context.Context, digest utils.Digest) (bool, error) {
	return s.Inner.Has(ctx, digest.String())
}
