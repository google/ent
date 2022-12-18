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

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/ent/cmd/ent/cmd"
	"github.com/google/ent/log"
	"github.com/google/ent/nodeservice"
	"github.com/google/ent/utils"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

const (
	wwwSegment  = "www"
	defaultPort = 27334
)

var domainName = "localhost:27334"

func main() {
	ctx := context.Background()

	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	router.GET("/*path", webGetHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = fmt.Sprintf("%d", defaultPort)
		log.Infof(ctx, "Defaulting to port %s", port)
	}

	s := &http.Server{
		Addr:           ":" + port,
		Handler:        h2c.NewHandler(router, &http2.Server{}),
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Infof(ctx, "Running locally")
	log.Criticalf(ctx, "%v", s.ListenAndServe())
}

func webGetHandler(c *gin.Context) {
	ctx := c
	hostSegments := hostSegments(c.Request.Host)
	if len(hostSegments) == 0 {
		log.Warningf(ctx, "no host segments")
		pathSegments := strings.Split(strings.TrimPrefix(c.Param("path"), "/"), "/")
		log.Warningf(ctx, "path segments: %#v", pathSegments)
		if len(pathSegments) == 1 {
			base := pathSegments[0]
			digest, err := utils.ParseDigest(base)
			if err != nil {
				log.Warningf(ctx, "could not parse digest: %s", err)
				c.AbortWithStatus(http.StatusNotFound)
				return
			}
			link := cid.NewCidV1(utils.TypeDAG, multihash.Multihash(digest))
			c.Redirect(http.StatusMovedPermanently, fmt.Sprintf("http://%s.www.%s", link.String(), domainName))
			return
		}
	}
	if len(hostSegments) != 2 {
		log.Warningf(ctx, "invalid host segments: %#v", hostSegments)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	if hostSegments[len(hostSegments)-1] != wwwSegment {
		log.Warningf(ctx, "invalid host segments: %#v", hostSegments)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	rootLink, err := cid.Parse(hostSegments[0])
	if err != nil {
		log.Warningf(ctx, "could not parse root link from %q: %s", hostSegments[0], err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	path := strings.Split(strings.TrimPrefix(c.Param("path"), "/"), "/")
	og := cmd.GetObjectGetter()
	log.Debugf(ctx, "root link: %s", rootLink.String())
	log.Debugf(ctx, "path: %#v", path)
	target, err := traverseString(ctx, og, rootLink, path)
	if err != nil {
		log.Warningf(ctx, "could not traverse: %s", err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	log.Debugf(ctx, "target: %s", target.String())
	c.Header("ent-digest", target.String())
	nodeRaw, err := og.Get(ctx, utils.Digest(target.Hash()))
	if err != nil {
		log.Warningf(ctx, "could not get blob %s: %s", target, err)
		c.Abort()
		return
	}
	switch target.Type() {
	case utils.TypeRaw:
		contentType := http.DetectContentType(nodeRaw)
		log.Debugf(ctx, "content type: %s", contentType)
		c.Data(http.StatusOK, contentType, nodeRaw)
	case utils.TypeDAG:
		renderDag(c, rootLink, target, nodeRaw, path)
	default:
		log.Warningf(ctx, "invalid target type: %s", target.Type())
		c.AbortWithStatus(http.StatusNotFound)
	}
}

func renderDag(c *gin.Context, root cid.Cid, target cid.Cid, nodeRaw []byte, path []string) {
	ctx := c
	node, err := utils.ParseDAGNode(nodeRaw)
	if err != nil {
		log.Warningf(ctx, "could not parse dag node: %s", err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	parents := []UILink{}
	parents = append(parents, UILink{
		Name: root.String(),
		URL:  "/",
	})
	for i, s := range path {
		parents = append(parents, UILink{
			Name: s,
			URL:  "/" + strings.Join(path[:i+1], "/"),
		})
	}

	links := []UILink{}
	names := strings.Split(string(node.Bytes), "\n")
	prefix := "/" + strings.Join(path, "/")
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	for i, link := range node.Links {
		links = append(links, UILink{
			Name: names[i],
			Raw:  link.Type() == utils.TypeRaw,
			URL:  prefix + names[i],
		})
	}

	c.HTML(http.StatusOK, "basic.html", gin.H{
		"links":   links,
		"parents": parents,
		"root":    target.String(),
	})
}

type UILink struct {
	Name string
	Raw  bool
	URL  string
}

func hostSegments(host string) []string {
	host = strings.TrimSuffix(host, domainName)
	host = strings.TrimSuffix(host, ".")
	hostSegments := strings.Split(host, ".")
	if len(hostSegments) > 0 && hostSegments[0] == "" {
		return hostSegments[1:]
	} else {
		return hostSegments
	}
}

func traverseString(ctx context.Context, og nodeservice.ObjectGetter, link cid.Cid, segments []string) (cid.Cid, error) {
	if len(segments) == 0 || link.Type() == utils.TypeRaw {
		return link, nil
	} else {
		digest := utils.Digest(link.Hash())
		nodeRaw, err := og.Get(ctx, digest)
		if err != nil {
			return cid.Cid{}, fmt.Errorf("could not get blob %s: %w", digest, err)
		}
		node, err := utils.ParseDAGNode(nodeRaw)
		if err != nil {
			return cid.Cid{}, fmt.Errorf("could not parse node %s: %w", digest, err)
		}
		selector := segments[0]
		if selector != "" {
			names := strings.Split(string(node.Bytes), "\n")
			// Find the name corresponding to the selector
			log.Debugf(ctx, "selector: %v", selector)
			log.Debugf(ctx, "names: %#v", names)
			linkIndex := -1
			for i, name := range names {
				if name == selector {
					linkIndex = i
					break
				}
			}
			if linkIndex == -1 {
				return cid.Cid{}, fmt.Errorf("could not find link %s/%v", digest, selector)
			}
			next := node.Links[linkIndex]
			if err != nil {
				return cid.Cid{}, fmt.Errorf("could not traverse %s/%v: %w", digest, selector, err)
			}
			log.Debugf(ctx, "next: %v", next)
			return traverseString(ctx, og, next, segments[1:])
		} else {
			return link, nil
		}
	}
}
