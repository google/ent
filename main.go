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
	"encoding/json"
	"fmt"
	"log"
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
	"github.com/google/ent/nodeservice"
	"github.com/google/ent/objectstore"
	"github.com/google/ent/utils"
	"google.golang.org/appengine"
)

var (
	blobStore nodeservice.ObjectStore

	handlerBrowse http.Handler
	handlerWWW    http.Handler
)

const objectsBucketName = "ent-objects"

const wwwSegment = "www"

var domainName = "localhost:8088"

type UINode struct {
	Kind         string
	Value        string
	Hash         string
	Links        []UILink
	URL          string
	ParentURL    string
	PathSegments []UIPathSegment
}

type UILink struct {
	Selector utils.Selector
	Hash     string
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
	domainNameEnv := os.Getenv("DOMAIN_NAME")
	if domainNameEnv != "" {
		domainName = domainNameEnv
	}
	log.Printf("domain name: %s", domainName)

	ctx := context.Background()
	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		log.Print(err)
		blobStore =
			objectstore.Store{
				Inner: datastore.File{
					DirName: "data/objects",
				},
			}
	} else {
		blobStore =
			objectstore.Store{
				Inner: datastore.Cloud{
					Client:     storageClient,
					BucketName: objectsBucketName,
				},
			}
	}

	{
		router := gin.Default()

		config := cors.DefaultConfig()
		config.AllowAllOrigins = true
		router.Use(cors.New(config))

		router.RedirectTrailingSlash = false
		router.RedirectFixedPath = false
		router.LoadHTMLGlob("templates/*")

		// Uninterpreted bytes by hash, no DAG traversal.
		// router.GET("/api/objects/:objecthash", apiObjectsGetHandler)
		// router.POST("/api/objects", apiObjectsUpdateHandler)

		// router.POST("/api/objects/get", apiObjectsGetHandler)
		// router.POST("/api/objects/update", apiObjectsUpdateHandler)

		router.POST(api.APIV1BLOBSGET, apiGetHandler)
		router.POST(api.APIV1BLOBSPUT, apiPutHandler)

		router.POST("/api/v1/links/get", apiGetHandler)
		router.POST("/api/v1/links/update", apiPutHandler)

		router.GET("/web/:root", browseBlobHandler)
		router.GET("/web/:root/*path", browseBlobHandler)

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
		port = "8088"
		log.Printf("Defaulting to port %s", port)
	}

	s := &http.Server{
		Addr:           ":" + port,
		Handler:        http.HandlerFunc(handlerRoot),
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())

	appengine.Main()
}

func handlerRoot(w http.ResponseWriter, r *http.Request) {
	hostSegments := hostSegments(r.Host)
	log.Printf("host segments: %#v", hostSegments)
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

func serveUI1(c *gin.Context, root utils.Hash, segments []utils.Selector, rawData []byte, node *utils.Node) {
	templateSegments := []UIPathSegment{}
	for i, s := range segments {
		templateSegments = append(templateSegments, UIPathSegment{
			Name:     utils.PrintSelector(s),
			Selector: s,
			URL:      path.Join("/", "web", string(root), utils.PrintPath(segments[0:i+1])),
		})
	}
	kind := "unknown"
	links := []UILink{}
	if node != nil {
		kind = node.Kind
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
					Hash:     l.Hash,
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
		Kind:         kind,
		Value:        string(rawData),
		Hash:         string(root),
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
		})
	} else {
		c.HTML(http.StatusOK, "browse.tmpl", gin.H{
			"type":     "file",
			"wwwHost":  wwwSegment + "." + domainName,
			"blob":     rawData,
			"blob_str": string(rawData),
		})
	}
}

func serveWWW(c *gin.Context, root utils.Hash, segments []utils.Selector) {
	/*
		target, err := traverse(c, root, segments)
		if err != nil {
			log.Print(err)
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		log.Printf("target: %s", target)

		node, err := blobStore.GetObject(c, target)
		if err != nil {
			log.Print(err)
			c.Abort()
			return
		}
			switch node := node.(type) {
			case *merkledag.RawNode:
				c.Header("ent-hash", target.String())
				ext := filepath.Ext(segments[len(segments)-1])
				contentType := mime.TypeByExtension(ext)
				if contentType == "" {
					contentType = http.DetectContentType(node.RawData())
				}
				c.Header("Content-Type", contentType)
				c.Data(http.StatusOK, "", node.RawData())
				return
			case *merkledag.ProtoNode:
				serveUI(c, root, segments, target, node)
			default:
				log.Printf("unknown codec: %v", target.Prefix().Codec)
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
	*/
}

type RenameRequest struct {
	Root     string
	FromPath string
	ToPath   string
}

type RemoveRequest struct {
	Root string
	Path string
}

type MutateRequest struct{}

func apiPutHandler(c *gin.Context) {
	var req api.PutRequest
	err := json.NewDecoder(c.Request.Body).Decode(&req)
	if err != nil {
		log.Printf("could not parse request: %v", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var res api.PutResponse
	res.Hash = make([]utils.Hash, 0, len(req.Blobs))
	for _, blob := range req.Blobs {
		h, err := blobStore.Put(c, blob)
		if err != nil {
			log.Printf("error adding blob: %s", err)
			continue
		}
		log.Printf("added blob: %s", h)
		res.Hash = append(res.Hash, h)
	}

	log.Printf("res: %#v", res)
	c.JSON(http.StatusOK, res)
}

func fetchNodes(c *gin.Context, root utils.Hash, depth uint) ([][]byte, error) {
	log.Printf("fetching nodes for %s depth %d", root, depth)
	var nodes [][]byte

	blob, err := blobStore.Get(c, root)
	if err != nil {
		log.Printf("error getting blob %q: %s", root, err)
		return nil, err
	}

	nodes = append(nodes, blob)

	node, err := utils.ParseNode(blob)
	if err != nil {
		log.Printf("error parsing blob %q: %s", root, err)
		return nodes, nil
	}

	if depth == 0 {
		return nodes, nil
	} else {
		for _, links := range node.Links {
			for _, link := range links {
				hash, err := utils.ParseHash(link.Hash)
				if err != nil {
					log.Printf("error parsing link: %s", err)
					continue
				}
				nn, err := fetchNodes(c, hash, depth-1)
				if err != nil {
					log.Printf("error fetching nodes: %s", err)
					continue
				}
				nodes = append(nodes, nn...)
			}
		}
		return nodes, nil
	}
}

func apiGetHandler(c *gin.Context) {
	var req api.GetRequest
	json.NewDecoder(c.Request.Body).Decode(&req)
	log.Printf("req: %#v", req)

	var depth uint = 10

	var res api.GetResponse
	res.Items = make(map[utils.Hash][]byte, len(req.Items))
	for _, item := range req.Items {
		blobs, err := fetchNodes(c, item.Root, depth)
		if err != nil {
			log.Printf("error getting blob %q: %s", item.Root, err)
			continue
		}
		for _, blob := range blobs {
			res.Items[utils.ComputeHash(blob)] = blob
		}
	}

	log.Printf("res: %#v", res)
	c.JSON(http.StatusOK, res)
}

func traverse(c context.Context, root utils.Hash, segments []utils.Selector) (utils.Hash, error) {
	if len(segments) == 0 {
		return root, nil
	} else {
		nodeRaw, err := blobStore.Get(c, root)
		if err != nil {
			return "", fmt.Errorf("could not get blob %s: %w", root, err)
		}
		node, err := utils.ParseNode(nodeRaw)
		if err != nil {
			return "", fmt.Errorf("could not parse node %s: %w", root, err)
		}
		selector := segments[0]
		next := node.Links[selector.FieldID][selector.Index]
		if err != nil {
			return "", fmt.Errorf("could not traverse %s/%v: %w", root, selector, err)
		}
		nextHash, err := utils.ParseHash(next.Hash)
		if err != nil {
			return "", fmt.Errorf("invalid hash: %w", err)
		}
		log.Printf("next: %v", next)
		return traverse(c, nextHash, segments[1:])
	}
}

func browseBlobHandler(c *gin.Context) {
	pathString := c.Param("path")
	log.Printf("path: %q", pathString)

	if strings.HasSuffix(c.Request.URL.Path, "/") {
		to := strings.TrimSuffix(c.Request.URL.Path, "/")
		log.Printf("redirecting to: %q", to)
		c.Redirect(http.StatusMovedPermanently, to)
		return
	}

	root, err := utils.ParseHash(c.Param("root"))
	if err != nil {
		log.Print(err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	path, err := utils.ParsePath(c.Param("path"))
	if err != nil {
		log.Printf("invalid path: %v", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	target, err := traverse(c, root, path)
	if err != nil {
		log.Print(err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	log.Printf("root: %v", root)
	nodeRaw, err := blobStore.Get(c, target)
	if err != nil {
		log.Print(err)
		c.Abort()
		return
	}
	node := &utils.Node{}
	err = json.Unmarshal(nodeRaw, node)
	if err != nil {
		log.Printf("invalid node: %v", err)
		node = nil
	}
	serveUI1(c, root, path, nodeRaw, node)
}

func renderHandler(c *gin.Context) {
	hostSegments := hostSegments(c.Request.Host)
	pathString := c.Param("path")
	log.Printf("path: %v", pathString)
	segments := parsePath(pathString)
	log.Printf("segments: %#v", segments)
	if pathString != "/" && strings.HasSuffix(pathString, "/") {
		c.Redirect(http.StatusMovedPermanently, strings.TrimSuffix(pathString, "/"))
		return
	}

	switch hostSegments[1] {
	case wwwSegment:
		/*
			baseDomain := hostSegments[0]
			log.Printf("base domain: %s", baseDomain)
			if baseDomain == "empty" {
					newNode := utils.NewProtoNode()
					err := blobStore.Add(c, newNode)
					if err != nil {
						log.Print(err)
						c.AbortWithStatus(http.StatusNotFound)
						return
					}
					target := newNode.Cid()
					log.Printf("target: %s", target.String())
					redirectToCid(c, target, "")
				return
			}

			root, err = cid.Decode(baseDomain)
			if err != nil {
				log.Print(err)
				c.AbortWithStatus(http.StatusNotFound)
				return
			}
			log.Printf("root: %v", root)
		*/
	default:
		log.Printf("invalid segment")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// serveWWW(c, root, segments)
}
