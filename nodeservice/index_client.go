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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"

	"github.com/google/ent/index"
	"github.com/google/ent/utils"
)

type IndexClient struct {
	BaseURL string
}

func (c IndexClient) Get(ctx context.Context, digest utils.Digest) ([]byte, error) {
	log.Printf("%s\n", string(digest))
	u := c.BaseURL + "/" + path.Join(index.DigestToPath(digest), index.EntryFilename)
	log.Printf("fetching entry from %s\n", u)
	entryRes, err := http.Get(u)
	if err != nil {
		return nil, fmt.Errorf("could not fetch index entry: %w", err)
	}
	if entryRes.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("could not fetch index entry: %s", entryRes.Status)
	}
	entry := index.IndexEntry{}
	err = json.NewDecoder(entryRes.Body).Decode(&entry)
	if err != nil {
		return nil, fmt.Errorf("could not parse index entry as JSON: %w", err)
	}
	log.Printf("parsed entry: %+v", entry)
	targetRes, err := http.Get(entry.URLS[0])
	if err != nil {
		return nil, fmt.Errorf("could not fetch target: %w", err)
	}
	target, err := ioutil.ReadAll(targetRes.Body)
	if err != nil {
		return nil, fmt.Errorf("could not download target: %w", err)
	}
	targetDigest := utils.ComputeDigest(target)
	if targetDigest != digest {
		return nil, fmt.Errorf("digest mismatch, wanted: %q, got %q", digest, targetDigest)
	}
	return target, nil
}

func (c IndexClient) Has(ctx context.Context, h utils.Digest) (bool, error) {
	return false, nil
}
