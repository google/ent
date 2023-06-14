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
	"context"
	"encoding/base32"
	"flag"
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/BurntSushi/toml"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/ent/api"
	"github.com/google/ent/datastore"
	"github.com/google/ent/log"
	"github.com/google/ent/nodeservice"
	"github.com/google/ent/objectstore"
	"github.com/google/ent/utils"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

var (
	blobStore nodeservice.ObjectStore

	enableMemcache = false
	enableBigquery = false

	apiKeyToUser = map[string]*User{}
)

const (
	wwwSegment  = "www"
	defaultPort = 27333
)

var domainName = "localhost:8088"

var configPath = flag.String("config", "", "path to config file")

type UINode struct {
	Kind         string
	Value        string
	Digest       string
	Links        []UILink
	URL          string
	ParentURL    string
	PathSegments []UIPathSegment
	WWWURL       string
}

type UILink struct {
	Selector utils.Selector
	Digest   string
	Raw      bool
	URL      string
}

type UIPathSegment struct {
	Selector utils.Selector
	Name     string
	URL      string
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

func readConfig() Config {
	config := Config{}
	_, err := toml.DecodeFile(*configPath, &config)
	if err != nil {
		panic(err)
	}
	return config
}

func main() {
	flag.Parse()

	ctx := context.Background()
	if *configPath == "" {
		log.Errorf(ctx, "must specify config")
		return
	}
	log.Infof(ctx, "loading config from %q", *configPath)
	config := readConfig()
	log.Infof(ctx, "loaded config: %#v", config)

	log.InitLog(config.ProjectID)

	domainName = config.DomainName
	log.Infof(ctx, "domain name: %s", domainName)

	if config.RedisEnabled {
		enableMemcache = true
		log.Infof(ctx, "memcache enabled")
	} else {
		log.Infof(ctx, "memcache disabled")
	}

	for _, user := range config.Users {
		// Must make a copy first, or else the map will point to the same.
		u := user
		apiKeyToUser[user.APIKey] = &u
	}
	for apiKey, user := range apiKeyToUser {
		log.Infof(ctx, "user %q: %q %d", redact(apiKey), user.Name, user.ID)
	}

	if config.BigqueryEnabled {
		enableBigquery = true
		log.Infof(ctx, "bigquery enabled")
	} else {
		log.Infof(ctx, "bigquery disabled")
	}

	var ds datastore.DataStore

	if config.CloudStorageEnabled {
		objectsBucketName := config.CloudStorageBucket
		if objectsBucketName == "" {
			log.Errorf(ctx, "must specify Cloud Storage bucket name")
			return
		}
		log.Infof(ctx, "using Cloud Storage bucket: %q", objectsBucketName)
		storageClient, err := storage.NewClient(ctx)
		if err != nil {
			log.Errorf(ctx, "could not create Cloud Storage client: %v", err)
			return
		}
		ds = datastore.Cloud{
			Client:     storageClient,
			BucketName: objectsBucketName,
		}
	} else {
		log.Infof(ctx, "using local file system")
		ds = datastore.File{
			DirName: "data/objects",
		}
	}

	if config.RedisEnabled {
		log.Infof(ctx, "using Redis: %q", config.RedisEndpoint)
		rdb := redis.NewClient(&redis.Options{
			Addr: config.RedisEndpoint,
		})
		err := rdb.Set(ctx, "key", "value", 0).Err()
		if err != nil {
			log.Errorf(ctx, "could not connect to Redis: %v", err)
		} else {
			ds = datastore.Memcache{
				Inner: ds,
				RDB:   rdb,
			}
		}
	}

	if config.BigqueryEnabled {
		bigqueryDataset := config.BigqueryDataset
		log.Infof(ctx, "bigquery dataset: %q", bigqueryDataset)
		InitBigquery(ctx, config.ProjectID, config.BigqueryDataset)
	}

	blobStore = objectstore.Store{
		Inner: ds,
	}

	gin.SetMode(config.GinMode)
	router := gin.Default()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	router.Use(cors.New(corsConfig))

	router.RedirectTrailingSlash = false
	router.RedirectFixedPath = false
	router.LoadHTMLGlob("templates/*")

	router.POST(api.APIV1BLOBSGET, apiGetHandler)
	router.POST(api.APIV1BLOBSPUT, apiPutHandler)

	router.GET("/raw/:digest", rawGetHandler)
	router.PUT("/raw", rawPutHandler)

	router.GET("/browse/:digest", webGetHandler)
	router.GET("/browse/:digest/*path", webGetHandler)

	router.StaticFile("/static/tailwind.min.css", "./templates/tailwind.min.css")

	s := &http.Server{
		Addr:           config.ListenAddress,
		Handler:        h2c.NewHandler(router, &http2.Server{}),
		ReadTimeout:    600 * time.Second,
		WriteTimeout:   600 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Infof(ctx, "server running")
	log.Criticalf(ctx, "%v", s.ListenAndServe())
}

func indexHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl", gin.H{})
}

func parsePath(p string) []string {
	if p == "/" || p == "" {
		return []string{}
	} else {
		return strings.Split(strings.TrimPrefix(p, "/"), "/")
	}
}

func parseHost(p string) []string {
	if p == "/" || p == "" {
		return []string{}
	} else {
		return strings.Split(strings.TrimPrefix(p, "/"), "/")
	}
}

func serveUI1(c *gin.Context, root utils.Digest, segments []utils.Selector, rawData []byte, node *utils.DAGNode) {
	templateSegments := []UIPathSegment{}
	for i, s := range segments {
		templateSegments = append(templateSegments, UIPathSegment{
			Name:     utils.PrintSelector(s),
			Selector: s,
			URL:      path.Join("/", "browse", string(root), utils.PrintPath(segments[0:i+1])),
		})
	}
	links := []UILink{}
	if node != nil {
		for linkIndex, link := range node.Links {
			linkPath := segments
			linkPath = append(linkPath, utils.Selector(linkIndex))
			links = append(links, UILink{
				Selector: utils.Selector(linkIndex),
				Digest:   link.Hash().String(),
				URL:      path.Join("/", "browse", string(root), utils.PrintPath(linkPath)),
			})
		}
	}
	currentURL := path.Join("/", "browse", string(root))
	parentURL := ""
	if len(segments) > 0 {
		parentURL = path.Join("/", "browse", string(root), utils.PrintPath(segments[0:len(segments)-1]))
	}

	link := cid.NewCidV1(utils.TypeDAG, multihash.Multihash(root))

	uiNode := UINode{
		Value:        string(rawData),
		Digest:       string(root),
		PathSegments: templateSegments,
		Links:        links,
		URL:          currentURL,
		ParentURL:    parentURL,
		WWWURL:       fmt.Sprintf("http://%s.%s.%s/", link.String(), wwwSegment, domainName),
	}
	if node != nil {
		c.HTML(http.StatusOK, "browse.tmpl", gin.H{
			"type":    "directory",
			"wwwHost": wwwSegment + "." + domainName,
			"node":    uiNode,
			"digest":  string(root),
		})
	} else {
		c.HTML(http.StatusOK, "browse.tmpl", gin.H{
			"type":     "file",
			"wwwHost":  wwwSegment + "." + domainName,
			"blob":     rawData,
			"blob_str": string(rawData),
			"digest":   string(root),
		})
	}
}

func fetchNodes(ctx context.Context, base cid.Cid, depth uint) ([][]byte, error) {
	log.Debugf(ctx, "fetching nodes for %v, depth %d", base, depth)
	var nodes [][]byte

	blob, err := blobStore.Get(ctx, utils.Digest(base.Hash()))
	if err != nil {
		return nil, fmt.Errorf("error getting blob %q: %w", base.Hash(), err)
	}

	nodes = append(nodes, blob)

	// Nothing to recurse here.
	if base.Type() == utils.TypeRaw || depth == 0 {
		return nodes, nil
	}

	dagNode, err := utils.ParseDAGNode(blob)
	if err != nil {
		log.Warningf(ctx, "error parsing blob %q: %s", base, err)
		return nodes, nil
	}

	for _, link := range dagNode.Links {
		nn, err := fetchNodes(ctx, link, depth-1)
		if err != nil {
			log.Warningf(ctx, "error fetching nodes: %s", err)
			continue
		}
		nodes = append(nodes, nn...)
	}
	return nodes, nil
}

func getAPIKey(c *gin.Context) string {
	// See https://cloud.google.com/endpoints/docs/openapi/openapi-limitations#api_key_definition_limitations
	const header = "x-api-key"
	return c.Request.Header.Get(header)
}

func traverse(ctx context.Context, digest utils.Digest, segments []utils.Selector) (utils.Digest, error) {
	if len(segments) == 0 {
		return digest, nil
	} else {
		nodeRaw, err := blobStore.Get(ctx, digest)
		if err != nil {
			return utils.Digest{}, fmt.Errorf("could not get blob %s: %w", digest, err)
		}
		node, err := utils.ParseDAGNode(nodeRaw)
		if err != nil {
			return utils.Digest{}, fmt.Errorf("could not parse node %s: %w", digest, err)
		}
		selector := segments[0]
		next := node.Links[selector]
		if err != nil {
			return utils.Digest{}, fmt.Errorf("could not traverse %s/%v: %w", digest, selector, err)
		}
		log.Debugf(ctx, "next: %v", next)
		return traverse(ctx, utils.Digest(next.Hash()), segments[1:])
	}
}

func renderHandler(c *gin.Context) {
	ctx := c.Request.Context()

	hostSegments := hostSegments(c.Request.Host)
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

	digestBytes, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(hostSegments[0]))
	if err != nil {
		log.Warningf(ctx, "could not base32 decode host segment: %v: %v", hostSegments[0], err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	digest := fmt.Sprintf("sha256:%x", digestBytes)
	log.Infof(ctx, "digest: %s", digest)

	root, err := utils.ParseDigest(string(digest))
	if err != nil {
		log.Warningf(ctx, "could not parse digest: %s", err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	path, err := utils.ParsePath(c.Param("path"))
	if err != nil {
		log.Warningf(ctx, "invalid path: %v", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	target, err := traverse(ctx, root, path)
	if err != nil {
		log.Warningf(ctx, "could not traverse: %s", err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	log.Infof(ctx, "root: %v", root)
	nodeRaw, err := blobStore.Get(ctx, target)
	if err != nil {
		log.Warningf(ctx, "could not get blob %s: %s", target, err)
		c.Abort()
		return
	}

	LogGet(ctx, &LogItemGet{
		LogItem: BaseLogItem(c),
		Source:  SourceWeb,
		// TODO: UserID
		Digest: []string{string(target)},
	})

	c.Header("ent-digest", string(target))
	contentType := http.DetectContentType(nodeRaw)
	c.Data(http.StatusOK, contentType, nodeRaw)
}

func BaseLogItem(c *gin.Context) LogItem {
	return LogItem{
		Timestamp:     time.Now(),
		IP:            c.ClientIP(),
		UserAgent:     c.Request.UserAgent(),
		RequestMethod: c.Request.Method,
		RequestURI:    c.Request.RequestURI,
	}
}
