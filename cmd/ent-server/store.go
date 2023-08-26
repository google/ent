package main

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
)

type Store struct {
	c *firestore.Client
}

func NewStore(c *firestore.Client) *Store {
	return &Store{c: c}
}

const (
	MapEntryCollection = "map_entry"
)

type Digest struct {
	Code   int64  `firestore:"0"`
	Digest []byte `firestore:"1"`
}

type MapEntry struct {
	PublicKey []byte `firestore:"0"`

	Tag    string `firestore:"1"`
	Target Digest `firestore:"2"`

	EntrySignature []byte `firestore:"3"`

	CreationTime    time.Time `firestore:"4"`
	ClientIPAddress string    `firestore:"5"`
	RequestBytes    []byte    `firestore:"6"`
}

func (s *Store) GetMapEntry(ctx context.Context, publicKey []byte, label string) (*MapEntry, error) {
	doc, err := s.c.Collection(MapEntryCollection).Query.Where("0", "==", publicKey).Where("1", "==", label).Documents(ctx).Next()
	if err != nil {
		return nil, err
	}
	e := MapEntry{}
	if err := doc.DataTo(&e); err != nil {
		return nil, err
	}
	return &e, nil
}

func (s *Store) SetMapEntry(ctx context.Context, e *MapEntry) error {
	_, _, err := s.c.Collection(MapEntryCollection).Add(ctx, e)
	return err
}
