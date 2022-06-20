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
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"fmt"
	"strings"
)

type Digest string

func ParseDigest(s string) (Digest, error) {
	if strings.HasPrefix(s, "sha256:") {
		rest := s[7:]
		if len(rest) != 64 {
			return "", fmt.Errorf("invalid digest: %q", s)
		}
		for _, c := range rest {
			if !strings.Contains("0123456789abcdef", string(c)) {
				return "", fmt.Errorf("invalid digest: %q", s)
			}
		}
		_, err := hex.DecodeString(s[7:])
		if err != nil {
			return "", err
		}
		return Digest(s), nil
	} else {
		return "", fmt.Errorf("invalid digest: %q", s)
	}
}

func ComputeDigest(b []byte) Digest {
	h := sha256.Sum256(b)
	return Digest("sha256:" + hex.EncodeToString(h[:]))
}

type NodeID struct {
	Root Link
	Path Path
}

func ToBase32(d Digest) (string, error) {
	h, err := hex.DecodeString(string(d[7:]))
	if err != nil {
		return "", fmt.Errorf("invalid digest: %q", d)
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(h), nil
}
