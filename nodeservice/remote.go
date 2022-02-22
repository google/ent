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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/ent/api"
	"github.com/google/ent/log"
	"github.com/google/ent/utils"
)

type Remote struct {
	APIURL string
	APIKey string
}

var (
	ErrNotFound = fmt.Errorf("not found")
)

func (s Remote) Get(ctx context.Context, h utils.Hash) ([]byte, error) {
	req := api.GetRequest{
		Items: []utils.NodeID{{
			Root: h,
		}},
	}
	reqBytes := bytes.Buffer{}
	err := json.NewEncoder(&reqBytes).Encode(req)
	if err != nil {
		return nil, fmt.Errorf("error encoding JSON request: %w", err)
	}
	r, err := http.Post(s.APIURL+api.APIV1BLOBSGET+"?key="+s.APIKey, "application/json", &reqBytes)
	if err != nil {
		return nil, err
	}
	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: %v", r.Status)
	}

	res := api.GetResponse{}
	err = json.NewDecoder(r.Body).Decode(&res)
	if err != nil {
		return nil, fmt.Errorf("error decoding JSON response: %w", err)
	}

	item, ok := res.Items[h]
	if !ok {
		return nil, ErrNotFound
	}

	return item, nil
}

func (s Remote) Put(ctx context.Context, b []byte) (utils.Hash, error) {
	req := api.PutRequest{
		Blobs: [][]byte{b},
	}
	reqBytes := bytes.Buffer{}
	err := json.NewEncoder(&reqBytes).Encode(req)
	if err != nil {
		return "", fmt.Errorf("error encoding JSON request: %w", err)
	}
	r, err := http.Post(s.APIURL+api.APIV1BLOBSPUT+"?key="+s.APIKey, "application/json", &reqBytes)
	if err != nil {
		return "", err
	}
	if r.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error: %v", r.Status)
	}

	res := api.PutResponse{}
	err = json.NewDecoder(r.Body).Decode(&res)
	if err != nil {
		return "", fmt.Errorf("error decoding JSON response: %w", err)
	}

	return res.Hash[0], nil
}

func (s Remote) Has(ctx context.Context, h utils.Hash) (bool, error) {
	req := api.GetRequest{
		Items: []utils.NodeID{{
			Root: h,
		}},
	}
	reqBytes := bytes.Buffer{}
	json.NewEncoder(&reqBytes).Encode(req)
	r, err := http.Post(s.APIURL+api.APIV1BLOBSGET+"?key="+s.APIKey, "application/json", &reqBytes)
	if err != nil {
		log.Errorf(ctx, "error sending request: %v", err)
	}
	if r.StatusCode == http.StatusOK {
		return true, nil
	}
	if r.StatusCode == http.StatusNotFound {
		return false, nil
	}
	return false, fmt.Errorf("invalid status code: %d", r.StatusCode)
}
