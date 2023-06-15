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

	"github.com/google/ent/api"
	pb "github.com/google/ent/proto"
	"github.com/google/ent/utils"
)

type ObjectGetter interface {
	Get(ctx context.Context, h utils.Digest) ([]byte, error)
	Has(ctx context.Context, h utils.Digest) (bool, error)
}

type ObjectStore interface {
	ObjectGetter
	Put(ctx context.Context, b []byte) (utils.Digest, error)
}

type NodeGetter interface {
	ObjectGetter
	GetNodes(ctx context.Context, req api.GetRequest) (api.GetResponse, error)
}

type NodeService interface {
	NodeGetter
	ObjectStore
	PutNodes(ctx context.Context, req api.PutRequest) (api.PutResponse, error)
	MapSet(ctx context.Context, m *pb.MapSetRequest) error
}
