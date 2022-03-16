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

package nodeservice

import (
	"context"
	"fmt"

	"github.com/google/ent/api"
	"github.com/google/ent/log"
	"github.com/google/ent/utils"
)

type Cached struct {
	Cache map[utils.Digest][]byte
	Inner NodeService
}

func (s Cached) Get(ctx context.Context, h utils.Digest) ([]byte, error) {
	if b, ok := s.Cache[h]; ok {
		log.Debugf(ctx, "cache hit for %s", h)
		return b, nil
	}
	s.fillCache(ctx, h)
	if b, ok := s.Cache[h]; ok {
		return b, nil
	}
	return nil, ErrNotFound
}

func (s Cached) Has(ctx context.Context, h utils.Digest) (bool, error) {
	if _, ok := s.Cache[h]; ok {
		log.Debugf(ctx, "cache hit for %s", h)
		return true, nil
	}
	s.fillCache(ctx, h)
	if _, ok := s.Cache[h]; ok {
		return true, nil
	}
	return false, nil
}

func (s Cached) fillCache(ctx context.Context, h utils.Digest) error {
	req := api.GetRequest{
		Items: []api.GetRequestItem{{
			NodeID: utils.NodeID{
				Root: utils.Link{
					Type:   utils.TypeDAG,
					Digest: h,
				},
			},
			Depth: 10,
		}},
	}
	res, err := s.Inner.GetNodes(ctx, req)
	if err != nil {
		return fmt.Errorf("error fetching from remote: %v", err)
	}
	log.Debugf(ctx, "retrieved %d nodes", len(res.Items))
	for _, r := range res.Items {
		s.Cache[h] = r
	}
	return nil
}
