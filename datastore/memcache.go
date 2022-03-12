package datastore

import (
	"context"

	"github.com/google/ent/log"
	"google.golang.org/appengine/v2/memcache"
)

type Memcache struct {
	Inner DataStore
}

func (s Memcache) Get(ctx context.Context, name string) ([]byte, error) {
	item, err := memcache.Get(ctx, name)
	if err != nil {
		if err != memcache.ErrCacheMiss {
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
	return item.Value, nil
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
	err := memcache.Set(ctx, &memcache.Item{
		Key:   name,
		Value: value,
	})
	if err != nil {
		log.Errorf(ctx, "error adding %q to memcache: %v", name, err)
	} else {
		log.Infof(ctx, "added %q to memcache", name)
	}
}
