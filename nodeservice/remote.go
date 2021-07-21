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
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"

	"github.com/google/ent/utils"
	"github.com/ipfs/go-cid"
	format "github.com/ipfs/go-ipld-format"
	"github.com/multiformats/go-multihash"
)

type Remote struct {
	APIURL string
}

type UploadRequest struct {
	Root  string
	Blobs []UploadBlob
}

type UploadBlob struct {
	Type    string // file | directory
	Path    string
	Content []byte
}

type UploadResponse struct {
	Root string
}

type GetRequest struct {
	Root string
	Path string
}

type GetResponse struct {
	Content []byte
}

var (
	ErrNotFound = fmt.Errorf("not found")
)

func (s Remote) GetObject(ctx context.Context, h multihash.Multihash) ([]byte, error) {
	// res, err := http.Get(s.APIURL + "/api/objects/" + h.HexString())
	u, _ := url.Parse(s.APIURL)
	u.Path = path.Join(u.Path, h.HexString())
	res, err := http.Get(u.String())
	if err != nil {
		log.Fatal(err)
	}
	if res.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: %v", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (s Remote) AddObject(ctx context.Context, b []byte) (multihash.Multihash, error) {
	// res, err := http.Post(s.APIURL+"/api/objects", "", bytes.NewReader(b))
	res, err := http.Post(s.APIURL, "", bytes.NewReader(b))
	if err != nil {
		log.Fatal(err)
	}
	if res.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("not found")
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: %v", res.Status)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	h, err := multihash.FromHexString(string(body))
	if err != nil {
		return nil, err
	}
	return h, nil
}

func (s Remote) Has(ctx context.Context, c cid.Cid) (bool, error) {
	r := GetRequest{
		Root: c.String(),
		Path: "",
	}
	buf := bytes.Buffer{}
	json.NewEncoder(&buf).Encode(r)
	res, err := http.Post(s.APIURL+"/api/get", "", &buf)
	if err != nil {
		log.Fatal(err)
	}
	if res.StatusCode == http.StatusOK {
		return true, nil
	}
	if res.StatusCode == http.StatusNotFound {
		return false, nil
	}
	return false, fmt.Errorf("invalid status code: %d", res.StatusCode)
}

func (s Remote) Get(ctx context.Context, c cid.Cid) (format.Node, error) {
	r := GetRequest{
		Root: c.String(),
		Path: "",
	}
	buf := bytes.Buffer{}
	json.NewEncoder(&buf).Encode(r)
	res, err := http.Post(s.APIURL+"/api/get", "", &buf)
	if err != nil {
		log.Fatal(err)
	}
	if res.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("not found")
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: %v", res.Status)
	}

	response := GetResponse{}
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	switch c.Prefix().Codec {
	case cid.DagProtobuf:
		return utils.ParseProtoNode(response.Content)
	case cid.Raw:
		return utils.ParseRawNode(response.Content)
	default:
		return nil, fmt.Errorf("invalid codec")
	}
}

func (s Remote) GetMany(ctx context.Context, cc []cid.Cid) <-chan *format.NodeOption {
	return nil
}

func (s Remote) Add(ctx context.Context, node format.Node) error {
	blobType := ""
	switch node.Cid().Prefix().Codec {
	case cid.Raw:
		blobType = "file"
	case cid.DagProtobuf:
		blobType = "directory"
	}
	r := UploadRequest{
		Root: "",
		Blobs: []UploadBlob{
			{
				Type:    blobType,
				Path:    "",
				Content: node.RawData(),
			},
		},
	}
	buf := bytes.Buffer{}
	json.NewEncoder(&buf).Encode(r)
	res, err := http.Post(s.APIURL+"/api/update", "", &buf)
	if err != nil {
		return fmt.Errorf("could not POST request: %v", err)
	}
	resJson := UploadResponse{}
	err = json.NewDecoder(res.Body).Decode(&resJson)
	if err != nil {
		return fmt.Errorf("could not read response body: %v", err)
	}
	log.Printf("uploaded: %#v", resJson)
	remoteHash := resJson.Root
	if node.Cid().String() != remoteHash {
		return fmt.Errorf("hash mismatch; local: %s, remote: %s", node.Cid().String(), remoteHash)
	}
	return nil
}

func (s Remote) AddMany(ctx context.Context, nodes []format.Node) error {
	return fmt.Errorf("not implemented")
}

func (s Remote) Remove(ctx context.Context, c cid.Cid) error {
	return fmt.Errorf("not implemented")
}

func (s Remote) RemoveMany(ctx context.Context, cc []cid.Cid) error {
	return fmt.Errorf("not implemented")
}
