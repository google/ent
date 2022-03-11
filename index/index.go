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

package index

import (
	"strings"

	"github.com/google/ent/utils"
)

const (
	EntryFilename = "entry.json"
)

// Vaguely similar to https://github.com/opencontainers/image-spec/blob/main/descriptor.md
type IndexEntry struct {
	MediaType string   `json:"mediaType"`
	Digest    string   `json:"digest"`
	Size      int      `json:"size"`
	URLS      []string `json:"urls"`
}

// Split the digest into its prefix, and then two character chunks, separated by slashes, so that
// each directory contains at most 255 entries.
func DigestToPath(digest utils.Digest) string {
	s := strings.Split(string(digest), ":")
	out := s[0]
	for i := 0; i < len(s[1])/2; i++ {
		out += "/" + s[1][i*2:(i+1)*2]
	}
	return out
}
