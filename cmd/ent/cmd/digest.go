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

package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/fatih/color"
	"github.com/google/ent/log"
	"github.com/google/ent/utils"
	"github.com/ipfs/go-cid"
	mbase "github.com/multiformats/go-multibase"
	"github.com/multiformats/go-multihash"
	"github.com/spf13/cobra"
)

var digestCmd = &cobra.Command{
	Use:  "digest [filename]",
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		filename := ""
		if len(args) > 0 {
			filename = args[0]
		}
		if filename == "" {
			_, err := digestStdin()
			if err != nil {
				log.Criticalf(ctx, "computing digest of stdin: %v", err)
				os.Exit(1)
			}
		} else {
			_, err := traverseFileOrDir(filename, print)
			if err != nil {
				log.Criticalf(ctx, "traversing file or dir: %v", err)
				os.Exit(1)
			}
		}

	},
}

type traverseF func([]byte, cid.Cid, string) error

func print(bytes []byte, link cid.Cid, name string) error {
	fmt.Printf("%s", formatLink(link, name))
	return nil
}

func digestStdin() (utils.Digest, error) {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return utils.Digest{}, fmt.Errorf("could not read stdin: %v", err)
	}
	return digestData(data)
}

func traverseFileOrDir(filename string, f traverseF) (utils.Digest, error) {
	info, err := os.Stat(filename)
	if err != nil {
		return utils.Digest{}, fmt.Errorf("could not stat %s: %v", filename, err)
	}
	if info.IsDir() {
		return traverseDir(filename, f)
	} else {
		return traverseFile(filename, f)
	}
}

func traverseDir(dirname string, f traverseF) (utils.Digest, error) {
	ctx := context.Background()
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		return utils.Digest{}, fmt.Errorf("could not read directory %s: %v", dirname, err)
	}
	links := make([]cid.Cid, 0, len(files))
	data := ""
	for _, file := range files {
		filename := dirname + "/" + file.Name()
		info, err := os.Stat(filename)
		if err != nil {
			return utils.Digest{}, fmt.Errorf("could not stat %q: %v", filename, err)
		}
		if info.IsDir() {
			digest, err := traverseDir(filename, f)
			if err != nil {
				return utils.Digest{}, err
			}
			links = append(links, cid.NewCidV1(utils.TypeDAG, multihash.Multihash(digest)))
			data += file.Name() + "\n"
		} else {
			digest, err := traverseFile(filename, f)
			if err != nil {
				return utils.Digest{}, err
			}
			links = append(links, cid.NewCidV1(utils.TypeRaw, multihash.Multihash(digest)))
			data += file.Name() + "\n"
		}
	}
	dagNode := utils.DAGNode{
		Links: links,
		Bytes: []byte(data),
	}
	log.Infof(ctx, "DAG node: %v\n", dagNode)
	serialized, err := utils.SerializeDAGNode(&dagNode)
	if err != nil {
		return utils.Digest{}, err
	}
	digest := utils.ComputeDigest(serialized)
	link := cid.NewCidV1(utils.TypeDAG, multihash.Multihash(digest))
	err = f(serialized, link, dirname+"/")
	if err != nil {
		log.Infof(ctx, "could not traverse directory %q: %v", dirname, err)
	}
	return digest, nil
}

func traverseFile(filename string, f traverseF) (utils.Digest, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return utils.Digest{}, fmt.Errorf("could not read file %q: %v", filename, err)
	}
	link := cid.NewCidV1(utils.TypeRaw, multihash.Multihash(utils.ComputeDigest(data)))
	f(data, link, filename)
	return digestData(data)
}

func digestData(data []byte) (utils.Digest, error) {
	return utils.ComputeDigest(data), nil
}

func formatDigest(digest utils.Digest, name string) string {
	digestString := utils.FormatDigest(digest, digestFormatFlag)
	return fmt.Sprintf("%s %s\n", color.YellowString(digestString), name)
}

func formatLink(link cid.Cid, name string) string {
	linkString, err := link.StringOfBase(mbase.Base32)
	if err != nil {
		// Should never happen.
		panic(err)
	}
	return fmt.Sprintf("%s %s\n", color.YellowString(linkString), name)
}
