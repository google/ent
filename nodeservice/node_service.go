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

	"github.com/ipfs/go-cid"
	format "github.com/ipfs/go-ipld-format"
	"github.com/multiformats/go-multihash"
)

type ObjectStore interface {
	GetObject(ctx context.Context, h multihash.Multihash) ([]byte, error)
	AddObject(ctx context.Context, b []byte) (multihash.Multihash, error)
}

// https://github.com/ipfs/go-ipld-format/blob/579737706ba5da3e550111621e2ab1bf122ed53f/merkledag.go
type NodeService interface {
	Has(context.Context, cid.Cid) (bool, error)
	// Get(context.Context, cid.Cid) (format.Node, error)
	// Add(context.Context, format.Node) error
	format.DAGService

	ObjectStore
}
