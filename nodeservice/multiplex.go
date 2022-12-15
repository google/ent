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

	"github.com/google/ent/log"
	"github.com/google/ent/utils"
)

type Multiplex struct {
	Inner []Inner
}

type Inner struct {
	Name         string
	ObjectGetter ObjectGetter
}

func (s Multiplex) Get(ctx context.Context, digest utils.Digest) ([]byte, error) {
	for _, ss := range s.Inner {
		b, err := ss.ObjectGetter.Get(ctx, digest)
		if err == ErrNotFound {
			log.Infof(ctx, "object %s not found in %s", digest, ss.Name)
			continue
		} else if err != nil {
			log.Errorf(ctx, "error fetching from remote %q: %v", ss.Name, err)
			continue
		}
		log.Infof(ctx, "fetched from remote %q", ss.Name)
		return b, nil
	}
	return nil, fmt.Errorf("not found")
}

func (s Multiplex) Has(ctx context.Context, digest utils.Digest) (bool, error) {
	for _, ss := range s.Inner {
		b, err := ss.ObjectGetter.Has(ctx, digest)
		if err != nil {
			continue
		}
		return b, nil
	}
	return false, nil
}

func (s Multiplex) Put(ctx context.Context, b []byte) (utils.Digest, error) {
	// return s.Inner[0].Put(ctx, b)
	return utils.Digest{}, fmt.Errorf("not implemented")
}
