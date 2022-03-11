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
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/ent/log"
	"github.com/google/ent/utils"
	"google.golang.org/appengine/v2"
)

func webGetHandler(c *gin.Context) {
	ctx := appengine.NewContext(c.Request)

	accessItem := &LogItemGet{
		Timestamp:     time.Now(),
		IP:            c.ClientIP(),
		UserAgent:     c.Request.UserAgent(),
		RequestMethod: c.Request.Method,
		RequestURI:    c.Request.RequestURI,
		Source:        SourceWeb,
		APIKey:        "www",
	}
	defer LogGet(ctx, accessItem)

	if strings.HasSuffix(c.Request.URL.Path, "/") {
		to := strings.TrimSuffix(c.Request.URL.Path, "/")
		log.Infof(ctx, "redirecting to: %q", to)
		c.Redirect(http.StatusMovedPermanently, to)
		return
	}

	digest, err := utils.ParseHash(c.Param("digest"))
	if err != nil {
		log.Warningf(ctx, "could not parse digest: %s", err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	accessItem.Digest = append(accessItem.Digest, string(digest))
	path, err := utils.ParsePath(c.Param("path"))
	if err != nil {
		log.Warningf(ctx, "invalid path: %v", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	target, err := traverse(ctx, digest, path)
	if err != nil {
		log.Warningf(ctx, "could not traverse: %s", err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	log.Infof(ctx, "target: %v", target)
	nodeRaw, err := blobStore.Get(ctx, target)
	if err != nil {
		log.Warningf(ctx, "could not get blob %s: %s", target, err)
		accessItem.NotFound = append(accessItem.NotFound, string(digest))
		c.Abort()
		return
	}
	accessItem.Found = append(accessItem.Found, string(digest))
	node := &utils.Node{}
	err = json.Unmarshal(nodeRaw, node)
	if err != nil {
		log.Warningf(ctx, "could not parse blob %s: %s", target, err)
		node = nil
	}

	serveUI1(c, target, path, nodeRaw, node)
}
