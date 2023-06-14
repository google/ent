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

package utils

import (
	"encoding/hex"
	"fmt"
	"strings"

	pb "github.com/google/ent/proto"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
)

type Digest = multihash.Multihash
type DigestArray [64]byte

const hash = multihash.SHA2_256

func ParseDigest(s string) (Digest, error) {
	digest, err := multihash.FromHexString(s)
	if err == nil {
		return Digest(digest), nil
	} else {
		digest, err := multihash.FromB58String(s)
		if err == nil {
			return Digest(digest), nil
		} else {
			if strings.HasPrefix(s, "sha256:") {
				s = strings.TrimPrefix(s, "sha256:")
				ss, err := hex.DecodeString(s)
				if err != nil {
					return nil, err
				}
				digest, err = multihash.Encode(ss, multihash.SHA2_256)
				if err != nil {
					return nil, err
				}
				return Digest(digest), err
			} else {
				return nil, fmt.Errorf("invalid digest: %v", err)
			}
		}
	}
}

func DigestToHumanString(d Digest) string {
	// TODO: Check digest type.
	m, err := multihash.Decode(d)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("sha256:%s", hex.EncodeToString(m.Digest))
}

func DigestToProto(d Digest) *pb.Digest {
	m, err := multihash.Decode(d)
	if err != nil {
		panic(err)
	}
	return &pb.Digest{
		Code:   m.Code,
		Digest: m.Digest,
	}
}

func DigestFromProto(d *pb.Digest) Digest {
	b, err := multihash.Encode(d.Digest, d.Code)
	if err != nil {
		panic(err)
	}
	return Digest(b)
}

func ComputeDigest(b []byte) Digest {
	d, err := multihash.Sum(b, hash, -1)
	if err != nil {
		panic(err)
	}
	return Digest(d)
}

type NodeID struct {
	Root cid.Cid
	Path Path
}

// Convert to a fixed size array. Only to be used for in-memory caching.
func DigestToArray(digest multihash.Multihash) DigestArray {
	var a DigestArray
	if len(digest) > len(a) {
		panic(fmt.Sprintf("invalid digest length; got %d, want > %d", len(digest), len(a)))
	}
	copy(a[:], digest[:])
	return a
}
