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
	res.Items = make(map[utils.Digest][]byte, len(req.Items))
	for _, item := range req.Items {
		nodeID := item.NodeID
		accessItem.Digest = append(accessItem.Digest, string(nodeID.Root.Digest))
		blobs, err := fetchNodes(ctx, nodeID.Root, item.Depth)
		if err != nil {
			log.Warningf(ctx, "error getting blob %q: %s", nodeID.Root, err)
			accessItem.NotFound = append(accessItem.NotFound, string(nodeID.Root.Digest))
			continue
		}
		for _, blob := range blobs {
			digest := utils.ComputeDigest(blob)
			accessItem.Found = append(accessItem.Found, string(digest))
			res.Items[digest] = blob
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
		h := utils.ComputeDigest(blob)
		exists, err := blobStore.Has(ctx, h)
		if err != nil {
			log.Errorf(ctx, "error checking blob existence: %s", err)
			accessItem.NotCreated = append(accessItem.NotCreated, string(h))
			continue
		}
		if exists {
			log.Infof(ctx, "blob %q already exists", h)
			accessItem.NotCreated = append(accessItem.NotCreated, string(h))
			continue
		}
		h1, err := blobStore.Put(ctx, blob)
		if h1 != h {
			log.Errorf(ctx, "mismatching hash, expected %s, got %s", h, h1)
		}
		accessItem.Digest = append(accessItem.Digest, string(h1))
		if err != nil {
			log.Errorf(ctx, "error adding blob: %s", err)
			accessItem.NotCreated = append(accessItem.NotCreated, string(h1))
			continue
		}
		log.Infof(ctx, "added blob: %s", h1)
		accessItem.Created = append(accessItem.Created, string(h1))
		res.Digest = append(res.Digest, h1)
	}

	log.Debugf(ctx, "res: %#v", res)
	c.JSON(http.StatusOK, res)
}
