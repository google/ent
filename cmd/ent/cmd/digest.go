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
			_, err := traverseFileOrDir(filename, print)
			if err != nil {
				log.Fatal(err)
			}
		}

	},
}

type traverseF func([]byte, utils.Digest, string)

func print(bytes []byte, digest utils.Digest, name string) {
	fmt.Printf("%s", formatDigest(digest, name))
}

func digestStdin() (utils.Digest, error) {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("could not read stdin: %v", err)
	}
	return digestData(data)
}

func traverseFileOrDir(filename string, f traverseF) (utils.Digest, error) {
	info, err := os.Stat(filename)
	if err != nil {
		return "", fmt.Errorf("could not stat %s: %v", filename, err)
	}
	if info.IsDir() {
		return traverseDir(filename, f)
	} else {
		return traverseDir(filename, f)
	}
}

func traverseDir(dirname string, f traverseF) (utils.Digest, error) {
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		return "", fmt.Errorf("could not read directory %s: %v", dirname, err)
	}
	links := make(map[uint][]utils.Link)
	for _, file := range files {
		filename := dirname + "/" + file.Name()
		info, err := os.Stat(filename)
		if err != nil {
			return "", fmt.Errorf("could not stat %q: %v", filename, err)
		}
		if info.IsDir() {
			digest, err := traverseDir(filename, f)
			if err != nil {
				return "", err
			}
			links[0] = append(links[0], utils.Link{
				Type:   utils.TypeDAG,
				Digest: digest,
			})
		} else {
			digest, err := traverseFile(filename, f)
			if err != nil {
				return "", err
			}
			links[2] = append(links[2], utils.Link{
				Type:   utils.TypeRaw,
				Digest: digest,
			})
		}
	}
	dagNode := utils.DAGNode{
		Links: links,
	}
	fmt.Printf("DAG node: %v\n", dagNode)
	serialized, err := utils.SerializeDAGNode(&dagNode)
	if err != nil {
		return "", err
	}
	digest := utils.ComputeDigest(serialized)
	f(serialized, digest, dirname+"/")
	return digest, nil
}

func traverseFile(filename string, f traverseF) (utils.Digest, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("could not read file %q: %v", filename, err)
	}
	f(data, utils.ComputeDigest(data), filename)
	return digestData(data)
}

func digestData(data []byte) (utils.Digest, error) {
	return utils.ComputeDigest(data), nil
}

func formatDigest(digest utils.Digest, name string) string {
	return fmt.Sprintf("%s %s\n", color.YellowString(string(digest)), name)
}
