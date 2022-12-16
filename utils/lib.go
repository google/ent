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
	"encoding/base32"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
)

type Digest multihash.Multihash
type DigestArray [32 + 4]byte
type DigestString string

const hash = multihash.SHA2_256

func ParseDigest(s string) (Digest, error) {
	digest, err := multihash.FromHexString(s)
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

func ToBase32(d Digest) (string, error) {
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(d), nil
}

func FromBase32(s string) (Digest, error) {
	d, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(s)
	if err != nil {
		return nil, err
	}
	return Digest(d), nil
}

func (d Digest) String() string {
	return multihash.Multihash(d).HexString()
}

func (d Digest) Array() DigestArray {
	var a DigestArray
	if len(d) != len(a) {
		panic("invalid digest")
	}
	copy(a[:], d[:])
	return a
}
