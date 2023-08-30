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
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	"github.com/BurntSushi/toml"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/ent/datastore"
	"github.com/google/ent/log"
	"github.com/google/ent/nodeservice"
	"github.com/google/ent/objectstore"
	pb "github.com/google/ent/proto"
	"github.com/google/ent/utils"
	"github.com/ipfs/go-cid"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	// Server address with port, from config file. Example: "localhost:27333"
	domainName string

	blobStore nodeservice.ObjectStore
	store     Store

	apiKeyToUser = map[string]*User{}
)

const (
	// See https://cloud.google.com/endpoints/docs/openapi/openapi-limitations#api_key_definition_limitations
	apiKeyHeader = "x-api-key"

	wwwSegment = "www"
)

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

func readConfig() Config {
	config := Config{}
	_, err := toml.DecodeFile(*configPath, &config)
	if err != nil {
		panic(err)
	}
	return config
}

func initStore(ctx context.Context, projectID string) *firestore.Client {
	firestoreClient, err := firestore.NewClient(context.Background(), projectID)
	if err != nil {
		log.Criticalf(ctx, "Failed to create client: %v", err)
		os.Exit(1)
	}
	res, err := firestoreClient.Doc("test/test").Set(ctx, map[string]interface{}{})
	if err != nil {
		log.Criticalf(ctx, "Failed to write to firestore: %v", err)
		os.Exit(1)
	}
	log.Infof(ctx, "Wrote to firestore: %v", res)
	return firestoreClient
}

func main() {
	flag.Parse()

	ctx := context.Background()
	if *configPath == "" {
		log.Errorf(ctx, "must specify config")
		os.Exit(1)
	}
	log.Infof(ctx, "loading config from %q", *configPath)
	config := readConfig()
	log.Infof(ctx, "loaded config: %#v", config)

	log.InitLog(config.ProjectID)

	for _, user := range config.Users {
		// Must make a copy first, or else the map will point to the same.
		u := user
		apiKeyToUser[user.APIKey] = &u
	}
	for apiKey, user := range apiKeyToUser {
		log.Infof(ctx, "user %q: %q %d", redact(apiKey), user.Name, user.ID)
	}

	var ds datastore.DataStore

	if config.CloudStorageEnabled {
		objectsBucketName := config.CloudStorageBucket
		if objectsBucketName == "" {
			log.Errorf(ctx, "must specify Cloud Storage bucket name")
			os.Exit(1)
		}
		log.Infof(ctx, "using Cloud Storage bucket: %q", objectsBucketName)
		storageClient, err := storage.NewClient(ctx)
		if err != nil {
			log.Errorf(ctx, "could not create Cloud Storage client: %v", err)
			os.Exit(1)
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

	fs := initStore(ctx, config.ProjectID)
	store = Store{
		c: fs,
	}

	gin.SetMode(config.GinMode)
	router := gin.Default()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	router.Use(cors.New(corsConfig))

	router.RedirectTrailingSlash = false
	router.RedirectFixedPath = false

	router.GET("/raw/:digest", rawGetHandler)
	router.PUT("/raw", rawPutHandler)

	grpServer := grpc.NewServer()
	pb.RegisterEntServer(grpServer, grpcServer{})
	router.Any("/ent.server.api.Ent/*any", gin.WrapH(grpServer))

	s := &http.Server{
		Addr:           config.ListenAddress,
		Handler:        h2c.NewHandler(router, &http2.Server{}),
		ReadTimeout:    600 * time.Second,
		WriteTimeout:   600 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Infof(ctx, "server running")
	err := s.ListenAndServe()
	fmt.Printf("server exited: %v", err)
	log.Criticalf(ctx, "%v", err)
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
	return c.Request.Header.Get(apiKeyHeader)
}

func getAPIKeyGRPC(c context.Context) string {
	md, _ := metadata.FromIncomingContext(c)
	vv := md.Get(apiKeyHeader)
	if len(vv) == 0 {
		return ""
	}
	return vv[0]
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

func BaseLogItem(c *gin.Context) LogItem {
	return LogItem{
		Timestamp:     time.Now(),
		IP:            c.ClientIP(),
		UserAgent:     c.Request.UserAgent(),
		RequestMethod: c.Request.Method,
		RequestURI:    c.Request.RequestURI,
	}
}
