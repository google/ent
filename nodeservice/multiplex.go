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

	"github.com/ipfs/go-cid"
	format "github.com/ipfs/go-ipld-format"
	"github.com/multiformats/go-multihash"
)

type Multiplex struct {
	Inner []NodeService
}

func (s Multiplex) GetObject(ctx context.Context, h multihash.Multihash) ([]byte, error) {
	for _, i := range s.Inner {
		b, err := i.GetObject(ctx, h)
		if err != nil {
			continue
		}
		return b, nil
	}
	return nil, fmt.Errorf("not found")
}

func (s Multiplex) AddObject(ctx context.Context, b []byte) (multihash.Multihash, error) {
	return s.Inner[0].AddObject(ctx, b)
}

func (s Multiplex) Has(ctx context.Context, c cid.Cid) (bool, error) {
	for _, i := range s.Inner {
		ok, _ := i.Has(ctx, c)
		if ok {
			return ok, nil
		}
	}
	return false, nil
}

func (s Multiplex) Get(ctx context.Context, c cid.Cid) (format.Node, error) {
	for _, i := range s.Inner {
		node, err := i.Get(ctx, c)
		if err != nil {
			continue
		}
		return node, nil
	}
	return nil, fmt.Errorf("not found")
}

func (s Multiplex) GetMany(ctx context.Context, cc []cid.Cid) <-chan *format.NodeOption {
	return nil
}

func (s Multiplex) Add(ctx context.Context, node format.Node) error {
	return s.Inner[0].Add(ctx, node)
}

func (s Multiplex) AddMany(ctx context.Context, nodes []format.Node) error {
	return fmt.Errorf("not implemented")
}

func (s Multiplex) Remove(ctx context.Context, c cid.Cid) error {
	return fmt.Errorf("not implemented")
}

func (s Multiplex) RemoveMany(ctx context.Context, cc []cid.Cid) error {
	return fmt.Errorf("not implemented")
}
