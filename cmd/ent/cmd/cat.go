package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/google/ent/index"
	"github.com/google/ent/utils"
	"github.com/ipfs/go-cid"
	format "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	"github.com/spf13/cobra"
)

func catByHash(hash string) {
	log.Printf("%s\n", hash)
	u := indexBasePath + "/" + path.Join(index.HashToPath(hash), index.EntryFilename)
	log.Printf("fetching entry from %s\n", u)
	entryRes, err := http.Get(u)
	if err != nil {
		log.Fatalf("could not fetch entry: %s", err)
	}
	entry := index.IndexEntry{}
	err = json.NewDecoder(entryRes.Body).Decode(&entry)
	if err != nil {
		log.Fatalf("could not parse entry as JSON: %s", err)
	}
	log.Printf("parsed entry: %+v", entry)
	targetRes, err := http.Get(entry.URLS[0])
	if err != nil {
		log.Fatalf("could not fetch target: %s", err)
	}
	target, err := ioutil.ReadAll(targetRes.Body)
	if err != nil {
		log.Fatalf("could not download target: %s", err)
	}
	targetHash := utils.ComputeHash(target)
	if targetHash != hash {
		log.Fatalf("hash mismatch, wanted: %q, got %q", hash, targetHash)
	}
	os.Stdout.Write(target)
}

var catCmd = &cobra.Command{
	Use:  "cat [hash] [path]",
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		hash := args[0]
		filePath := ""
		if len(args) >= 2 {
			filePath = args[1]
		}

		if strings.HasPrefix(hash, "sha256:") {
			catByHash(hash)
			return
		}

		base, err := cid.Decode(hash)
		if err != nil {
			log.Fatalf("could not decode cid: %v", err)
		}

		pathSegments := filepath.SplitList(filePath)

		var node format.Node

		for {
			obj, err := nodeService.GetObject(context.Background(), base.Hash())
			if err != nil {
				log.Fatalf("could not fetch object: %v", err)
			}
			node, err = utils.ParseNodeFromBytes(base, obj)
			if len(pathSegments) > 0 {
				s := pathSegments[0]
				pathSegments = pathSegments[1:]

				switch node := node.(type) {
				case *merkledag.ProtoNode:
					link, err := node.GetNodeLink(s)
					if err != nil {
						log.Fatalf("could not get node link: %v", err)
					}
					base = link.Cid
				case *merkledag.RawNode:
					log.Fatalf("invalid state")
				}
			} else {
				break
			}
		}

		os.Stdout.Write(printNode(node))
	},
}

func printNode(node format.Node) []byte {
	switch node := node.(type) {
	case *merkledag.ProtoNode:
		listing := ""
		for _, l := range node.Links() {
			listing += fmt.Sprintf("%s %s\n", l.Cid, l.Name)
		}
		return []byte(listing)
	case *merkledag.RawNode:
		return node.RawData()
	default:
		log.Fatalf("invalid format")
		return nil
	}
}
