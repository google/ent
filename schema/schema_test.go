//
// Copyright 2022 The Ent Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package schema

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/ent/datastore"
	"github.com/google/ent/objectstore"
	"github.com/google/ent/utils"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
)

func TestGetString(t *testing.T) {
	o := objectstore.Store{
		Inner: datastore.InMemory{
			Inner: make(map[string][]byte),
		},
	}

	stringFieldDigest, err := o.Put(context.Background(), []byte("hello"))
	if err != nil {
		t.Fatalf("failed to put string: %v", err)
	}

	s, err := GetString(o, stringFieldDigest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s != "hello" {
		t.Fatalf("unexpected string: %v", s)
	}
}

func TestGetUint32(t *testing.T) {
	o := objectstore.Store{
		Inner: datastore.InMemory{
			Inner: make(map[string][]byte),
		},
	}

	uintFieldDigest, err := o.Put(context.Background(), []byte{0x00, 0x00, 0x07, 0xC4})
	if err != nil {
		t.Fatalf("failed to put uint32: %v", err)
	}

	s, err := GetUint32(o, uintFieldDigest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s != 1988 {
		t.Fatalf("unexpected value: %v", s)
	}
}

func TestGetStruct(t *testing.T) {
	o := objectstore.Store{
		Inner: datastore.InMemory{
			Inner: make(map[string][]byte),
		},
	}

	stringFieldDigest, err := o.Put(context.Background(), []byte("hello"))
	if err != nil {
		t.Fatalf("failed to put string: %v", err)
	}
	uintFieldDigest, err := o.Put(context.Background(), []byte{0x00, 0x00, 0x07, 0xC4})
	if err != nil {
		t.Fatalf("failed to put string: %v", err)
	}
	node := &utils.DAGNode{
		Links: []cid.Cid{
			cid.NewCidV1(cid.Raw, multihash.Multihash(uintFieldDigest)),
			cid.NewCidV1(cid.Raw, multihash.Multihash(stringFieldDigest)),
			cid.NewCidV1(cid.Raw, multihash.Multihash(uintFieldDigest)),
			cid.NewCidV1(cid.Raw, multihash.Multihash(uintFieldDigest)),
		},
	}
	nodeBytes, err := utils.SerializeDAGNode(node)
	if err != nil {
		t.Fatalf("failed to serialize node: %v", err)
	}
	nodeDigest, err := o.Put(context.Background(), nodeBytes)
	if err != nil {
		t.Fatalf("failed to put string: %v", err)
	}

	v := &Field{}

	err = GetStruct(o, nodeDigest, v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// if v.FieldID != 1988 {
	// 	t.Fatalf("unexpected value: %v", v.FieldID)
	// }
	// if v.Name != "hello" {
	// 	t.Fatalf("unexpected value: %v", v.Name)
	// }
}

type T struct {
	A uint32   `ent:"0"`
	B string   `ent:"1"`
	C []uint32 `ent:"2"`
	D []string `ent:"3"`
	E U        `ent:"4"`
	F []U      `ent:"5"`
	G uint64   `ent:"6"`
}

type U struct {
	A uint32 `ent:"0"`
	B string `ent:"1"`
}

func TestRoundTrip(t *testing.T) {
	o := objectstore.Store{
		Inner: datastore.InMemory{
			Inner: make(map[string][]byte),
		},
	}

	vv := []*T{
		{},
		{
			A: 123,
			B: "hello",
			C: []uint32{0, 1, 2},
			D: []string{"a", "b", "c"},
			E: U{
				A: 456,
				B: "world",
			},
			F: []U{
				{
					A: 78,
					B: "mondo",
				},
				{
					A: 90,
					B: "monde",
				},
			},
			G: 123456789,
		},
	}

	for _, v := range vv {
		h, err := PutStruct(o, v)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		v1 := &T{}

		err = GetStruct(o, h, v1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !reflect.DeepEqual(v, v1) {
			t.Fatalf("unexpected value: %v, expected: %v", v1, v)
		}
	}
}
