//
// Copyright 2021 The Ent Authors.
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

package nodeservice

import (
	"context"
	"fmt"

	"github.com/google/ent/objectstore"
	"github.com/google/ent/utils"
	"github.com/ipfs/go-cid"
	format "github.com/ipfs/go-ipld-format"
	"github.com/multiformats/go-multihash"
)

const hashType = multihash.SHA2_256

type DataStore struct {
	Inner objectstore.Store
}

func (s DataStore) GetObject(ctx context.Context, h multihash.Multihash) ([]byte, error) {
	return s.Inner.Get(ctx, h)
}

func (s DataStore) AddObject(ctx context.Context, b []byte) (multihash.Multihash, error) {
	return s.Inner.Add(ctx, b)
}

func (s DataStore) Has(ctx context.Context, c cid.Cid) (bool, error) {
	_, err := s.Inner.Get(ctx, c.Hash())
	return err == nil, nil
}

func (s DataStore) Get(ctx context.Context, c cid.Cid) (format.Node, error) {
	bytes, err := s.Inner.Get(ctx, c.Hash())
	if err != nil {
		return nil, err
	}
	switch c.Prefix().Codec {
	case cid.DagProtobuf:
		return utils.ParseProtoNode(bytes)
	case cid.Raw:
		return utils.ParseRawNode(bytes)
	default:
		return nil, fmt.Errorf("invalid codec")
	}
}

func (s DataStore) GetMany(ctx context.Context, cc []cid.Cid) <-chan *format.NodeOption {
	return nil
}

func (s DataStore) Add(ctx context.Context, node format.Node) error {
	_, err := s.Inner.Add(ctx, node.RawData())
	return err
}

func (s DataStore) AddMany(ctx context.Context, nodes []format.Node) error {
	return fmt.Errorf("not implemented")
}

func (s DataStore) Remove(ctx context.Context, c cid.Cid) error {
	return fmt.Errorf("not implemented")
}

func (s DataStore) RemoveMany(ctx context.Context, cc []cid.Cid) error {
	return fmt.Errorf("not implemented")
}
