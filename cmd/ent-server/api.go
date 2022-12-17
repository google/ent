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

package main

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/ent/api"
	"github.com/google/ent/log"
	"github.com/google/ent/utils"
	"google.golang.org/appengine/v2"
)

func apiGetHandler(c *gin.Context) {
	ctx := appengine.NewContext(c.Request)

	accessItem := &LogItemGet{
		LogItem: BaseLogItem(c),
		Source:  SourceAPI,
	}
	defer LogGet(ctx, accessItem)

	apiKey := getAPIKey(c)
	accessItem.APIKey = apiKey
	if apiKey != readAPIKey && apiKey != readWriteAPIKey {
		log.Warningf(ctx, "invalid API key: %q", apiKey)
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	var req api.GetRequest
	json.NewDecoder(c.Request.Body).Decode(&req)
	log.Debugf(ctx, "req: %#v", req)

	var res api.GetResponse
	res.Items = make(map[string][]byte, len(req.Items))
	for _, item := range req.Items {
		nodeID := item.NodeID
		accessItem.Digest = append(accessItem.Digest, nodeID.Root.Hash().String())
		blobs, err := fetchNodes(ctx, nodeID.Root, item.Depth)
		if err != nil {
			log.Warningf(ctx, "error getting blob %q: %s", nodeID.Root, err)
			accessItem.NotFound = append(accessItem.NotFound, nodeID.Root.Hash().String())
			continue
		}
		for _, blob := range blobs {
			digest := utils.ComputeDigest(blob)
			digestString := digest.String()
			accessItem.Found = append(accessItem.Found, string(digestString))
			res.Items[digestString] = blob
		}
	}

	c.JSON(http.StatusOK, res)
}

func apiPutHandler(c *gin.Context) {
	ctx := appengine.NewContext(c.Request)

	accessItem := &LogItemPut{
		LogItem: BaseLogItem(c),
		Source:  SourceAPI,
	}
	defer LogPut(ctx, accessItem)

	apiKey := getAPIKey(c)
	accessItem.APIKey = apiKey
	if apiKey != readWriteAPIKey {
		log.Warningf(ctx, "invalid API key: %q", apiKey)
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	var req api.PutRequest
	err := json.NewDecoder(c.Request.Body).Decode(&req)
	if err != nil {
		log.Warningf(ctx, "could not parse request: %v", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var res api.PutResponse
	res.Digest = make([]utils.Digest, 0, len(req.Blobs))
	for _, blob := range req.Blobs {
		digest := utils.ComputeDigest(blob)
		exists, err := blobStore.Has(ctx, digest)
		if err != nil {
			log.Errorf(ctx, "error checking blob existence: %s", err)
			accessItem.NotCreated = append(accessItem.NotCreated, digest.String())
			continue
		}
		if exists {
			log.Infof(ctx, "blob %q already exists", digest)
			accessItem.NotCreated = append(accessItem.NotCreated, digest.String())
			continue
		}
		digest1, err := blobStore.Put(ctx, blob)
		if !bytes.Equal(digest1, digest) {
			log.Errorf(ctx, "mismatching digest, expected %q, got %q", digest.String(), digest1.String())
		}
		accessItem.Digest = append(accessItem.Digest, digest1.String())
		if err != nil {
			log.Errorf(ctx, "error adding blob: %s", err)
			accessItem.NotCreated = append(accessItem.NotCreated, digest1.String())
			continue
		}
		log.Infof(ctx, "added blob: %q", digest1.String())
		accessItem.Created = append(accessItem.Created, digest1.String())
		res.Digest = append(res.Digest, digest1)
	}

	log.Debugf(ctx, "res: %#v", res)
	c.JSON(http.StatusOK, res)
}