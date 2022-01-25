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
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/ent/datastore"
	"github.com/google/ent/nodeservice"
	"github.com/google/ent/objectstore"
	"github.com/google/ent/utils"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-merkledag"
	"google.golang.org/appengine"
)

var (
	blobStore nodeservice.NodeService
	tagStore  datastore.DataStore

	handlerBrowse http.Handler
	handlerWWW    http.Handler
)

const objectsBucketName = "ent-objects"
const tagsBucketName = "multiverse-312721-key"

const wwwSegment = "www"
const tagsSegment = "tags"

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

func redirectToCid(c *gin.Context, target cid.Cid, path string) {
	c.Redirect(http.StatusFound, fmt.Sprintf("//%s.%s.%s%s", target.String(), wwwSegment, domainName, path))
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
		blobStore = nodeservice.DataStore{
			Inner: objectstore.Store{
				Inner: datastore.File{
					DirName: "data/objects",
				},
			},
		}
		tagStore = datastore.File{
			DirName: "data/tags",
		}
	} else {
		blobStore = nodeservice.DataStore{
			Inner: objectstore.Store{
				Inner: datastore.Cloud{
					Client:     storageClient,
					BucketName: objectsBucketName,
				},
			},
		}
		tagStore = datastore.Cloud{
			Client:     storageClient,
			BucketName: tagsBucketName,
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

		router.POST("/api/v1/blobs/get", apiGetHandler)
		router.POST("/api/v1/blobs/put", apiPutHandler)

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

	s := &http.Server{
		Addr:           ":8088",
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

func postTagHandler(c *gin.Context) {
	segments := parsePath(c.Param("path"))
	tagName := segments[1]
	tagValueString, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Print(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	tagValue, err := cid.Decode(string(tagValueString))
	if err != nil {
		log.Print(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	err = tagStore.Set(c, tagName, []byte(tagValue.String()))
	if err != nil {
		log.Print(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
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

type PutRequest struct {
	Blobs [][]byte
}

type UploadResponse struct {
	Hash []string `json:"hash"`
}

type MutateRequest struct{}

type GetRequest struct {
	Items []GetItem `json:"items"`
}

type GetItem struct {
	Root string
	Path []utils.Selector
}

type GetResponse struct {
	Items map[string][]byte `json:"items"`
}

func apiPutHandler(c *gin.Context) {
	var req PutRequest
	err := json.NewDecoder(c.Request.Body).Decode(&req)
	if err != nil {
		log.Printf("could not parse request: %v", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var res UploadResponse
	res.Hash = make([]string, 0, len(req.Blobs))
	for _, blob := range req.Blobs {
		h, err := blobStore.AddObject(c, blob)
		if err != nil {
			log.Printf("error adding blob: %s", err)
			continue
		}
		log.Printf("added blob: %s", h)
		res.Hash = append(res.Hash, string(h))
	}

	log.Printf("res: %#v", res)
	c.JSON(http.StatusOK, res)
}

func fetchNodes(c *gin.Context, root utils.Hash, depth uint) ([][]byte, error) {
	log.Printf("fetching nodes for %s depth %d", root, depth)
	var nodes [][]byte

	blob, err := blobStore.GetObject(c, root)
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
	var req GetRequest
	json.NewDecoder(c.Request.Body).Decode(&req)
	log.Printf("req: %#v", req)

	var depth uint = 10

	var res GetResponse
	res.Items = make(map[string][]byte, len(req.Items))
	for _, item := range req.Items {
		h, err := utils.ParseHash(item.Root)
		if err != nil {
			log.Printf("error parsing hash %q: %s", item.Root, err)
			continue
		}
		blobs, err := fetchNodes(c, h, depth)
		if err != nil {
			log.Printf("error getting blob %q: %s", h, err)
			continue
		}
		for _, blob := range blobs {
			res.Items[string(utils.ComputeHash(blob))] = blob
		}
	}

	log.Printf("res: %#v", res)
	c.JSON(http.StatusOK, res)
}

func traverse(c context.Context, root utils.Hash, segments []utils.Selector) (utils.Hash, error) {
	if len(segments) == 0 {
		return root, nil
	} else {
		nodeRaw, err := blobStore.GetObject(c, root)
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

func traverseAdd(c context.Context, root cid.Cid, segments []string, nodeToAdd cid.Cid) (cid.Cid, error) {
	log.Printf("traverseAdd %v/%#v", root, segments)
	if len(segments) == 0 {
		return nodeToAdd, nil
	} else {
		node, err := blobStore.Get(c, root)
		if err != nil {
			return cid.Undef, fmt.Errorf("could not get blob %s", root)
		}
		switch node := node.(type) {
		case *merkledag.ProtoNode:
			head := segments[0]
			var next cid.Cid
			next, err = utils.GetLink(node, head)
			if err == merkledag.ErrLinkNotFound {
				// Ok
				newNode := utils.NewProtoNode()
				err = blobStore.Add(c, newNode)
				// TODO
				next = node.Cid()
			} else if err != nil {
				return cid.Undef, fmt.Errorf("could not get link: %v", err)
			}
			log.Printf("next: %v", next)

			newHash, err := traverseAdd(c, next, segments[1:], nodeToAdd)
			if err != nil {
				return cid.Undef, fmt.Errorf("could not call recursively: %v", err)
			}

			err = utils.SetLink(node, head, newHash)
			if err != nil {
				return cid.Undef, fmt.Errorf("could not add link: %v", err)
			}
			return node.Cid(), blobStore.Add(c, node)
		default:
			return cid.Undef, fmt.Errorf("incorrect node type")
		}
	}
}

func traverseRemove(c context.Context, root cid.Cid, segments []string) (cid.Cid, error) {
	log.Printf("traverseRemove %v/%#v", root, segments)
	node, err := blobStore.Get(c, root)
	if err != nil {
		return cid.Undef, fmt.Errorf("could not get node %s", root)
	}
	switch node := node.(type) {
	case *merkledag.ProtoNode:
		if len(segments) == 1 {
			utils.RemoveLink(node, segments[0])
		} else {
			head := segments[0]
			var next cid.Cid
			next, err = utils.GetLink(node, head)
			if err == merkledag.ErrLinkNotFound {
				// Ok
				newNode := utils.NewProtoNode()
				err = blobStore.Add(c, newNode)
				// TODO
				next = newNode.Cid()

			} else if err != nil {
				return cid.Undef, fmt.Errorf("could not get link: %v", err)
			}
			log.Printf("next: %v", next)

			newHash, err := traverseRemove(c, next, segments[1:])
			if err != nil {
				return cid.Undef, fmt.Errorf("could not call recursively: %v", err)
			}

			err = utils.SetLink(node, head, newHash)
			if err != nil {
				return cid.Undef, fmt.Errorf("could not add link: %v", err)
			}
		}
		return node.Cid(), blobStore.Add(c, node)
	default:
		return cid.Undef, nil
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
	nodeRaw, err := blobStore.GetObject(c, target)
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

	root := cid.Undef
	var err error

	switch hostSegments[1] {
	case wwwSegment:
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
	case tagsSegment:
		// tagValueBytes, err := tagStore.Get(c, hostSegments[0])
		// if err != nil {
		// 	log.Print(err)
		// 	c.AbortWithStatus(http.StatusInternalServerError)
		// 	return
		// }
		// tagValue, err := cid.Decode(string(tagValueBytes))
		// if err != nil {
		// 	log.Print(err)
		// 	c.AbortWithStatus(http.StatusInternalServerError)
		// 	return
		// }
		// serveWWW(c, tagValue, segments)
		return
	default:
		log.Printf("invalid segment")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// serveWWW(c, root, segments)
}
