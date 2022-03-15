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
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/ent/api"
	"github.com/google/ent/datastore"
	"github.com/google/ent/log"
	"github.com/google/ent/nodeservice"
	"github.com/google/ent/objectstore"
	"github.com/google/ent/utils"
	"google.golang.org/appengine/v2"
)

var (
	blobStore nodeservice.ObjectStore

	handlerBrowse http.Handler
	handlerWWW    http.Handler

	enableMemcache = false
	enableBigquery = false

	readAPIKey      = ""
	readWriteAPIKey = ""
)

const (
	wwwSegment  = "www"
	defaultPort = 27333
)

var domainName = "localhost:8088"

type UINode struct {
	Kind         string
	Value        string
	Digest       string
	Links        []UILink
	URL          string
	ParentURL    string
	PathSegments []UIPathSegment
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

func main() {
	ctx := context.Background()
	if appengine.IsAppEngine() {
		ctx = appengine.BackgroundContext()
	}

	domainNameEnv := os.Getenv("DOMAIN_NAME")
	if domainNameEnv != "" {
		domainName = domainNameEnv
	}
	log.Infof(ctx, "domain name: %s", domainName)

	if os.Getenv("ENABLE_MEMCACHE") != "" {
		enableMemcache = true
		log.Infof(ctx, "memcache enabled")
	} else {
		log.Infof(ctx, "memcache disabled")
	}

	readAPIKey = os.Getenv("READ_API_KEY")
	if readAPIKey != "" {
		log.Infof(ctx, "read API key: %q", readAPIKey)
	}

	readWriteAPIKey = os.Getenv("READ_WRITE_API_KEY")
	if readWriteAPIKey != "" {
		log.Infof(ctx, "read write API key: %q", readWriteAPIKey)
	}

	if os.Getenv("ENABLE_BIGQUERY") != "" {
		enableBigquery = true
		log.Infof(ctx, "bigquery enabled")
	} else {
		log.Infof(ctx, "bigquery disabled")
	}

	var ds datastore.DataStore

	objectsBucketName := os.Getenv("CLOUD_STORAGE_BUCKET")
	if objectsBucketName != "" {
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

	if enableMemcache {
		ds = datastore.Memcache{
			Inner: ds,
		}
	}

	if enableBigquery {
		bigqueryDataset := os.Getenv("BIGQUERY_DATASET")
		log.Infof(ctx, "bigquery dataset: %q", bigqueryDataset)
		InitBigquery(ctx, bigqueryDataset)
	}

	blobStore = objectstore.Store{
		Inner: ds,
	}

	{
		router := gin.Default()

		config := cors.DefaultConfig()
		config.AllowAllOrigins = true
		router.Use(cors.New(config))
		/*
			router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
				ctx := appengine.NewContext(param.Request)
				log.Infof(ctx, "%s", param.ErrorMessage)
				return "\n"
			}))
		*/

		router.RedirectTrailingSlash = false
		router.RedirectFixedPath = false
		router.LoadHTMLGlob("templates/*")

		// Uninterpreted bytes by digest, no DAG traversal.
		// router.GET("/api/objects/:digest", apiObjectsGetHandler)
		// router.POST("/api/objects", apiObjectsUpdateHandler)

		// router.POST("/api/objects/get", apiObjectsGetHandler)
		// router.POST("/api/objects/update", apiObjectsUpdateHandler)

		router.POST(api.APIV1BLOBSGET, apiGetHandler)
		router.POST(api.APIV1BLOBSPUT, apiPutHandler)

		router.GET("/raw/:digest", rawGetHandler)
		router.PUT("/raw", rawPutHandler)

		router.GET("/web/:digest", webGetHandler)
		router.GET("/web/:digest/*path", webGetHandler)

		router.StaticFile("/static/tailwind.min.css", "./templates/tailwind.min.css")

		handlerBrowse = router
	}
	{
		router := gin.Default()
		router.GET("/*path", renderHandler)
		handlerWWW = router
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = fmt.Sprintf("%d", defaultPort)
		log.Infof(ctx, "Defaulting to port %s", port)
	}

	s := &http.Server{
		Addr:           ":" + port,
		Handler:        http.HandlerFunc(handlerRoot),
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if appengine.IsAppEngine() {
		log.Infof(ctx, "Running on App Engine")
		http.HandleFunc("/", handlerRoot)
		appengine.Main()
	} else {
		log.Infof(ctx, "Running locally")
		log.Criticalf(ctx, "%v", s.ListenAndServe())
	}
}

func handlerRoot(w http.ResponseWriter, r *http.Request) {
	// hostSegments := hostSegments(r.Host)
	// log.Printf("host segments: %#v", hostSegments)
	// if len(hostSegments) == 0 {
	handlerBrowse.ServeHTTP(w, r)
	// } else {
	// handlerWWW.ServeHTTP(w, r)
	// }
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
			URL:      path.Join("/", "web", string(root), utils.PrintPath(segments[0:i+1])),
		})
	}
	links := []UILink{}
	if node != nil {
		for fieldID, ll := range node.Links {
			for index, l := range ll {
				selector := utils.Selector{
					FieldID: fieldID,
					Index:   uint(index),
				}
				linkPath := segments
				linkPath = append(linkPath, selector)
				links = append(links, UILink{
					Selector: selector,
					Digest:   string(l.Digest),
					URL:      path.Join("/", "web", string(root), utils.PrintPath(linkPath)),
				})
			}
		}
	}
	currentURL := path.Join("/", "web", string(root))
	parentURL := ""
	if len(segments) > 0 {
		parentURL = path.Join("/", "web", string(root), utils.PrintPath(segments[0:len(segments)-1]))
	}
	uiNode := UINode{
		Value:        string(rawData),
		Digest:       string(root),
		PathSegments: templateSegments,
		Links:        links,
		URL:          currentURL,
		ParentURL:    parentURL,
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

func fetchNodes(ctx context.Context, base utils.Link, depth uint) ([][]byte, error) {
	log.Debugf(ctx, "fetching nodes for %v, depth %d", base, depth)
	var nodes [][]byte

	blob, err := blobStore.Get(ctx, base.Digest)
	if err != nil {
		return nil, fmt.Errorf("error getting blob %q: %w", base.Digest, err)
	}

	nodes = append(nodes, blob)

	// Nothing to recurse here.
	if base.Type == utils.TypeRaw || depth == 0 {
		return nodes, nil
	}

	dagNode, err := utils.ParseDAGNode(blob)
	if err != nil {
		log.Warningf(ctx, "error parsing blob %q: %s", base, err)
		return nodes, nil
	}

	for _, links := range dagNode.Links {
		for _, link := range links {
			nn, err := fetchNodes(ctx, link, depth-1)
			if err != nil {
				log.Warningf(ctx, "error fetching nodes: %s", err)
				continue
			}
			nodes = append(nodes, nn...)
		}
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
			return "", fmt.Errorf("could not get blob %s: %w", digest, err)
		}
		node, err := utils.ParseDAGNode(nodeRaw)
		if err != nil {
			return "", fmt.Errorf("could not parse node %s: %w", digest, err)
		}
		selector := segments[0]
		next := node.Links[selector.FieldID][selector.Index]
		if err != nil {
			return "", fmt.Errorf("could not traverse %s/%v: %w", digest, selector, err)
		}
		log.Debugf(ctx, "next: %v", next)
		return traverse(ctx, next.Digest, segments[1:])
	}
}

func renderHandler(c *gin.Context) {
	ctx := appengine.NewContext(c.Request)

	root, err := utils.ParseDigest(c.Param("digest"))
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
		APIKey:  "www",
		Digest:  []string{string(target)},
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
