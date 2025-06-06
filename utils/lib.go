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

	"github.com/google/ent/api"
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
			parts := strings.Split(s, ":")
			if len(parts) == 2 {
				code, ok := multihash.Names[strings.ToLower(parts[0])]
				if !ok {
					return nil, fmt.Errorf("invalid digest code: %q", parts[0])
				}
				ss, err := hex.DecodeString(parts[1])
				if err != nil {
					return nil, err
				}
				digest, err = multihash.Encode(ss, code)
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
	m, err := multihash.Decode(d)
	if err != nil {
		panic(err)
	}
	codeString := multihash.Codes[m.Code]
	return fmt.Sprintf(codeString + ":" + hex.EncodeToString(m.Digest))
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

func FormatDigest(digest Digest, format string) string {
	switch format {
	case "human":
		return DigestToHumanString(digest)
	case "hex":
		return digest.HexString()
	case "b58":
		return digest.B58String()
	default:
		return "-"
	}
}

func DigestForLog(digest Digest) string {
	return digest.B58String()
}

func DigestToApi(digest Digest) api.HexDigests {
	mh, err := multihash.Decode(digest)
	if err != nil {
		panic(err)
	}
	switch mh.Code {
	case multihash.SHA2_256:
		return api.HexDigests{Sha2_256: hex.EncodeToString(mh.Digest)}
	case multihash.SHA2_512:
		return api.HexDigests{Sha2_512: hex.EncodeToString(mh.Digest)}
	case multihash.SHA3_256:
		return api.HexDigests{Sha3_256: hex.EncodeToString(mh.Digest)}
	case multihash.SHA3_384:
		return api.HexDigests{Sha3_384: hex.EncodeToString(mh.Digest)}
	case multihash.SHA3_512:
		return api.HexDigests{Sha3_512: hex.EncodeToString(mh.Digest)}
	default:
		panic(fmt.Sprintf("unsupported hash code: %v", mh.Code))
	}
}
