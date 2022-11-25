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
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/google/ent/schema"
	"github.com/google/ent/utils"
	"github.com/spf13/cobra"
)

var digestCmd = &cobra.Command{
	Use:  "digest [filename]",
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := ""
		if len(args) > 0 {
			filename = args[0]
		}
		if filename == "" {
			_, err := digestStdin()
			if err != nil {
				log.Fatal(err)
			}
		} else {
			_, err := digestFileOrDir(filename)
			if err != nil {
				log.Fatal(err)
			}
		}

	},
}

func digestStdin() (utils.Digest, error) {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("could not read stdin: %v", err)
	}
	return digestData(data)
}

func digestFileOrDir(filename string) (utils.Digest, error) {
	info, err := os.Stat(filename)
	if err != nil {
		return "", fmt.Errorf("could not stat %s: %v", filename, err)
	}
	if info.IsDir() {
		return digestDir(filename)
	} else {
		return digestFile(filename)
	}
}

// TODO: Also return schema in parallel.
func digestDir(dirname string) (utils.Digest, error) {
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		return "", fmt.Errorf("could not read directory %s: %v", dirname, err)
	}
	links := make(map[uint][]utils.Link)
	kinds := make([]schema.Kind, 0, len(files))
	for i, file := range files {
		digest, err := digestFileOrDir(dirname + "/" + file.Name())
		if err != nil {
			return "", err
		}
		links[uint(i)] = []utils.Link{{Digest: digest}}
		kinds = append(kinds, schema.Kind{
			KindID: uint32(i),
			Name:   file.Name(),
		})
	}
	// schema := schema.Schema{
	// 	Kinds: kinds,
	// }
	// fmt.Printf("schema: %v\n", schema)
	dagNode := utils.DAGNode{
		Links: links,
	}
	fmt.Printf("DAG node: %v\n", dagNode)
	serialized, err := utils.SerializeDAGNode(&dagNode)
	if err != nil {
		return "", err
	}
	digest := utils.ComputeDigest(serialized)
	fmt.Printf("%s", formatDigest(digest, dirname+"/"))
	return digest, nil
}

func digestFile(filename string) (utils.Digest, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("could not read file %q: %v", filename, err)
	}
	fmt.Printf("%s", formatDigest(utils.ComputeDigest(data), filename))
	return digestData(data)
}

func digestData(data []byte) (utils.Digest, error) {
	return utils.ComputeDigest(data), nil
}

func formatDigest(digest utils.Digest, name string) string {
	return fmt.Sprintf("%s %s\n", color.YellowString(string(digest)), name)
}
