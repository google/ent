package datastore

import (
	"context"
	"fmt"
)

type InMemory struct {
	Inner map[string][]byte
}

func (s InMemory) Set(ctx context.Context, name string, value []byte) error {
	s.Inner[name] = value
	return nil
}

func (s InMemory) Get(ctx context.Context, name string) ([]byte, error) {
	b, ok := s.Inner[name]
	if ok {
		return b, nil
	} else {
		return nil, fmt.Errorf("not found")
	}
}

func (s InMemory) Has(ctx context.Context, name string) (bool, error) {
	_, ok := s.Inner[name]
	return ok, nil
}
