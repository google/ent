package datastore

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/google/ent/log"
)

type Memcache struct {
	Inner DataStore
	RDB   redis.Client
}

func (s Memcache) Get(ctx context.Context, name string) ([]byte, error) {
	cmd := s.RDB.Get(ctx, name)
	item, err := cmd.Bytes()
	if err != nil {
		if err != redis.Nil {
			log.Errorf(ctx, "error getting %q from memcache: %v", name, err)
		}
		b, err := s.Inner.Get(ctx, name)
		if err != nil {
			return nil, err
		}
		go s.TrySet(ctx, name, b)
		return b, nil
	}
	log.Infof(ctx, "got %q from memcache", name)
	return item, nil
}

func (s Memcache) Put(ctx context.Context, name string, value []byte) error {
	err := s.Inner.Put(ctx, name, value)
	if err != nil {
		return err
	}
	go s.TrySet(ctx, name, value)
	return nil
}

func (s Memcache) Has(ctx context.Context, name string) (bool, error) {
	return s.Inner.Has(ctx, name)
}

func (s Memcache) TrySet(ctx context.Context, name string, value []byte) {
	err := s.RDB.Set(ctx, name, value, 0)
	if err != nil {
		log.Errorf(ctx, "error adding %q to memcache: %v", name, err)
	} else {
		log.Infof(ctx, "added %q to memcache", name)
	}
}
