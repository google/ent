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
	"encoding/hex"
	"fmt"
	"strings"
)

type Hash string

func ParseHash(s string) (Hash, error) {
	if strings.HasPrefix(s, "sha256:") {
		rest := s[7:]
		if len(rest) != 64 {
			return "", fmt.Errorf("invalid hash: %q", s)
		}
		for _, c := range rest {
			if !strings.Contains("0123456789abcdef", string(c)) {
				return "", fmt.Errorf("invalid hash: %q", s)
			}
		}
		_, err := hex.DecodeString(s[7:])
		if err != nil {
			return "", err
		}
		return Hash(s), nil
	} else {
		return "", fmt.Errorf("invalid hash: %q", s)
	}
}

func ComputeHash(b []byte) Hash {
	h := sha256.Sum256(b)
	return Hash("sha256:" + hex.EncodeToString(h[:]))
}

type NodeID struct {
	Root Hash
	Path []Selector
}
