package cmd

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/google/ent/utils"
	"github.com/ipfs/go-cid"
	format "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:  "pull [hash] [target directory]",
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		hash := args[0]
		targetDir := "."
		if len(args) >= 2 {
			targetDir = args[1]
		}
		targetDir, err := filepath.Abs(targetDir)
		if err != nil {
			log.Fatalf("could not normalize target directory %q, %v", targetDir, err)
		}

		base, err := cid.Decode(hash)
		if err != nil {
			log.Fatalf("could not decode cid: %v", err)
		}

		pull(base, targetDir, false)

		log.Printf("pull %s %s", hash, targetDir)
	},
}

func pull(base cid.Cid, targetPath string, executable bool) {
	_, err := os.Stat(targetPath)
	if os.IsNotExist(err) {
		// Continue.
	} else if err != nil {
		log.Fatalf("could not stat target path: %v", err)
	}
	if !os.IsNotExist(err) {
		log.Printf("target path %s already exists; skipping", targetPath)
		// Skip?
		return
	}
	files, err := ioutil.ReadDir(targetPath)
	if os.IsNotExist(err) {
		// Continue.
	} else if err != nil {
		log.Fatalf("could not read target directory %q: %v", targetPath, err)
	}
	if len(files) > 0 {
		log.Fatalf("cannot pull to non-empty directory %q", targetPath)
	}
	traverseRemote(base, "", func(p string, node format.Node) error {
		fullPath := filepath.Join(targetPath, p)
		log.Printf("%s\n", fullPath)
		switch node := node.(type) {
		case *merkledag.ProtoNode:
			err := os.MkdirAll(fullPath, 0755)
			if err != nil {
				log.Fatalf("could not create directory %q: %v", fullPath, err)
			}
		case *merkledag.RawNode:
			err := os.MkdirAll(path.Dir(fullPath), 0755)
			if err != nil {
				log.Fatalf("could not create directory %q: %v", fullPath, err)
			}
			mode := 0644
			if executable {
				mode = 0755
			}
			err = ioutil.WriteFile(fullPath, node.RawData(), os.FileMode(mode))
			if err != nil {
				log.Fatalf("could not create file %q: %v", fullPath, err)
			}
		}
		return nil
	})
}

func traverseRemote(base cid.Cid, relativeFilename string, f func(string, format.Node) error) {
	obj, err := nodeService.GetObject(context.Background(), base.Hash())
	if err != nil {
		log.Fatal(err)
	}

	switch base.Prefix().Codec {
	case cid.DagProtobuf:
		node, err := utils.ParseProtoNode(obj)
		if err != nil {
			log.Fatal(err)
		}
		err = f(relativeFilename, node)
		if err != nil {
			log.Fatal(err)
		}

		for _, l := range node.Links() {
			if l.Name == "" {
				continue
			}
			newRelativeFilename := path.Join(relativeFilename, l.Name)
			traverseRemote(l.Cid, newRelativeFilename, f)
		}
	case cid.Raw:
		node, err := utils.ParseRawNode(obj)
		if err != nil {
			log.Fatal(err)
		}
		err = f(relativeFilename, node)
		if err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("invalid codec: %v", base.Prefix().Codec)
	}
}
