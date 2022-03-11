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
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/ent/log"
	"github.com/google/ent/utils"
	"google.golang.org/appengine/v2"
)

func rawGetHandler(c *gin.Context) {
	ctx := appengine.NewContext(c.Request)

	accessItem := &LogItemGet{
		Timestamp:     time.Now(),
		IP:            c.ClientIP(),
		UserAgent:     c.Request.UserAgent(),
		RequestMethod: c.Request.Method,
		RequestURI:    c.Request.RequestURI,
		Source:        SourceRaw,
	}
	defer LogGet(ctx, accessItem)

	apiKey := getAPIKey(c)
	accessItem.APIKey = apiKey
	if apiKey != readAPIKey && apiKey != readWriteAPIKey {
		log.Errorf(ctx, "invalid API key: %q", apiKey)
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	digest, err := utils.ParseHash(c.Param("digest"))
	if err != nil {
		log.Errorf(ctx, "could not parse digest: %s", err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	accessItem.Digest = append(accessItem.Digest, string(digest))

	target := digest

	nodeRaw, err := blobStore.Get(ctx, target)
	if err != nil {
		log.Errorf(ctx, "could not get blob %s: %s", target, err)
		accessItem.NotFound = append(accessItem.NotFound, string(digest))
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	accessItem.Found = append(accessItem.Found, string(digest))

	c.Data(http.StatusOK, "text/plain; charset=utf-8", nodeRaw)
}

func rawPutHandler(c *gin.Context) {
	ctx := appengine.NewContext(c.Request)

	accessItem := &LogItemPut{
		Timestamp:     time.Now(),
		IP:            c.ClientIP(),
		UserAgent:     c.Request.UserAgent(),
		RequestMethod: c.Request.Method,
		RequestURI:    c.Request.RequestURI,
		Source:        SourceRaw,
	}
	defer LogPut(ctx, accessItem)

	apiKey := getAPIKey(c)
	accessItem.APIKey = apiKey
	if apiKey != readWriteAPIKey {
		log.Errorf(ctx, "invalid API key: %q", apiKey)
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	nodeRaw, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Errorf(ctx, "could not read node: %s", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	h, err := blobStore.Put(ctx, nodeRaw)
	accessItem.Digest = append(accessItem.Digest, string(h))
	if err != nil {
		log.Errorf(ctx, "could not put blob: %s", err)
		accessItem.NotCreated = append(accessItem.NotCreated, string(h))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	accessItem.Created = append(accessItem.Created, string(h))

	location := fmt.Sprintf("/raw/%s", h)
	log.Infof(ctx, "new object location: %q", location)

	c.Header("Location", location)
	// https://stackoverflow.com/questions/797834/should-a-restful-put-operation-return-something
	c.Status(http.StatusCreated)
}
