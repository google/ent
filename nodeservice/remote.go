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
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
)

type Remote struct {
	APIURL string
	APIKey string
}

const (
	APIKeyHeader = "x-api-key"
)

var (
	ErrNotFound = fmt.Errorf("not found")
)

func DoRequest(req *http.Request) (*http.Response, error) {
	httpClient := http.Client{
		// Transport: &http2.Transport{
		// 	AllowHTTP: true,
		// 	DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
		// 		return net.Dial(network, addr)
		// 	},
		// },
	}
	return httpClient.Do(req)
}

func (s Remote) Get(ctx context.Context, digest utils.Digest) ([]byte, error) {
	req := api.GetRequest{
		Items: []api.GetRequestItem{{
			NodeID: utils.NodeID{
				Root: cid.NewCidV1(utils.TypeRaw, multihash.Multihash(digest)),
			},
		}},
	}
	reqBytes := bytes.Buffer{}
	err := json.NewEncoder(&reqBytes).Encode(req)
	if err != nil {
		return nil, fmt.Errorf("error encoding JSON request: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, s.APIURL+api.APIV1BLOBSGET, &reqBytes)
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %w", err)
	}
	httpReq.Header.Set(APIKeyHeader, s.APIKey)
	httpRes, err := DoRequest(httpReq)
	if err != nil {
		return nil, err
	}
	if httpRes.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: %v", httpRes.Status)
	}

	res := api.GetResponse{}
	err = json.NewDecoder(httpRes.Body).Decode(&res)
	if err != nil {
		return nil, fmt.Errorf("error decoding JSON response: %w", err)
	}

	item, ok := res.Items[digest.String()]
	if !ok {
		return nil, ErrNotFound
	}

	return item, nil
}

func (s Remote) Put(ctx context.Context, b []byte) (utils.Digest, error) {
	req := api.PutRequest{
		Blobs: [][]byte{b},
	}

	res, err := s.PutNodes(ctx, req)
	if err != nil {
		return utils.Digest{}, fmt.Errorf("error getting response: %w", err)
	}
	if len(res.Digest) != 1 {
		return utils.Digest{}, fmt.Errorf("expected 1 digest, got %d", len(res.Digest))
	}

	return res.Digest[0], nil
}

func (s Remote) GetNodes(ctx context.Context, req api.GetRequest) (api.GetResponse, error) {
	reqBytes := bytes.Buffer{}
	err := json.NewEncoder(&reqBytes).Encode(req)
	if err != nil {
		return api.GetResponse{}, fmt.Errorf("error encoding JSON request: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, s.APIURL+api.APIV1BLOBSGET, &reqBytes)
	if err != nil {
		return api.GetResponse{}, fmt.Errorf("error creating HTTP request: %w", err)
	}
	httpReq.Header.Set(APIKeyHeader, s.APIKey)
	httpRes, err := DoRequest(httpReq)
	if err != nil {
		return api.GetResponse{}, fmt.Errorf("error sending request: %w", err)
	}
	if httpRes.StatusCode != http.StatusOK {
		return api.GetResponse{}, fmt.Errorf("error getting get response: %s", httpRes.Status)
	}

	res := api.GetResponse{}
	err = json.NewDecoder(httpRes.Body).Decode(&res)
	if err != nil {
		return api.GetResponse{}, fmt.Errorf("error decoding JSON response: %w", err)
	}

	return res, nil
}

func (s Remote) PutNodes(ctx context.Context, req api.PutRequest) (api.PutResponse, error) {
	reqBytes := bytes.Buffer{}
	err := json.NewEncoder(&reqBytes).Encode(req)
	if err != nil {
		return api.PutResponse{}, fmt.Errorf("error encoding JSON request: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, s.APIURL+api.APIV1BLOBSPUT, &reqBytes)
	if err != nil {
		return api.PutResponse{}, fmt.Errorf("error creating HTTP request: %w", err)
	}
	httpReq.Header.Set(APIKeyHeader, s.APIKey)
	httpRes, err := DoRequest(httpReq)
	if err != nil {
		return api.PutResponse{}, fmt.Errorf("error sending request: %w", err)
	}
	if httpRes.StatusCode != http.StatusOK {
		return api.PutResponse{}, fmt.Errorf("error getting put response: %s", httpRes.Status)
	}

	res := api.PutResponse{}
	err = json.NewDecoder(httpRes.Body).Decode(&res)
	if err != nil {
		return api.PutResponse{}, fmt.Errorf("error decoding JSON response: %w", err)
	}

	return res, nil
}

func (s Remote) Has(ctx context.Context, digest utils.Digest) (bool, error) {
	req := api.GetRequest{
		Items: []api.GetRequestItem{{
			NodeID: utils.NodeID{
				Root: cid.NewCidV1(utils.TypeRaw, multihash.Multihash(digest)),
			},
		}},
	}
	reqBytes := bytes.Buffer{}
	json.NewEncoder(&reqBytes).Encode(req)

	httpReq, err := http.NewRequest(http.MethodPost, s.APIURL+api.APIV1BLOBSGET, &reqBytes)
	if err != nil {
		return false, fmt.Errorf("error creating HTTP request: %w", err)
	}
	httpReq.Header.Set(APIKeyHeader, s.APIKey)
	httpRes, err := DoRequest(httpReq)
	if err != nil {
		log.Errorf(ctx, "error sending request: %v", err)
	}
	if httpRes.StatusCode == http.StatusOK {
		return true, nil
	}
	if httpRes.StatusCode == http.StatusNotFound {
		return false, nil
	}
	return false, fmt.Errorf("invalid status code: %d", httpRes.StatusCode)
}
