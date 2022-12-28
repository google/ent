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

type Sequence struct {
	Inner []Inner
}

type Inner struct {
	Name         string
	ObjectGetter ObjectGetter
}

func (s Sequence) Get(ctx context.Context, digest utils.Digest) ([]byte, error) {
	for _, ss := range s.Inner {
		b, err := ss.ObjectGetter.Get(ctx, digest)
		if err == ErrNotFound {
			log.Infof(ctx, "object %s not found in %s", digest, ss.Name)
			continue
		} else if err != nil {
			log.Errorf(ctx, "error fetching (get %q) from remote %q: %v", digest, ss.Name, err)
			continue
		}
		log.Infof(ctx, "fetched from remote %q", ss.Name)
		return b, nil
	}
	return nil, ErrNotFound
}

func (s Sequence) Has(ctx context.Context, digest utils.Digest) (bool, error) {
	for _, ss := range s.Inner {
		b, err := ss.ObjectGetter.Has(ctx, digest)
		if err != nil {
			log.Errorf(ctx, "error fetching (has %q) from remote %q: %v", digest, ss.Name, err)
			continue
		}
		if b {
			return b, nil
		}
	}
	return false, nil
}

func (s Sequence) Put(ctx context.Context, b []byte) (utils.Digest, error) {
	// return s.Inner[0].Put(ctx, b)
	return utils.Digest{}, fmt.Errorf("not implemented")
}
