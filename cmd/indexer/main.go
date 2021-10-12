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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/ent/index"
	"github.com/google/ent/utils"
	"github.com/spf13/cobra"
)

var (
	indexFlag string
	in        string
)

func server(cmd *cobra.Command, args []string) {
	if indexFlag == "" {
		fmt.Println("index flag is required")
		os.Exit(1)
		return
	}
	if in == "" {
		fmt.Println("in flag is required")
		os.Exit(1)
		return
	}
	lines, err := ioutil.ReadFile(in)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
		return
	}
	for _, line := range strings.Split(string(lines), "\n") {
		if line == "" {
			continue
		}
		fmt.Println(line)
		res, err := http.Get(line)
		if err != nil {
			fmt.Println(err)
			continue
		}
		data, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Printf("size: %d\n", len(data))
		h := utils.ComputeHash(data)
		fmt.Printf("hash: %s\n", h)
		l := filepath.Join(indexFlag, index.HashToPath(h), index.EntryFilename)
		e := index.IndexEntry{
			Hash: h,
			Size: len(data),
			URLS: []string{line},
		}
		es, err := json.Marshal(e)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Printf("%s\n", es)
		err = os.MkdirAll(filepath.Dir(l), 0755)
		if err != nil {
			fmt.Println(err)
			continue
		}
		ioutil.WriteFile(l, es, 0644)
	}
}

func main() {
	var rootCmd = &cobra.Command{Use: "indexer"}
	rootCmd.AddCommand(
		&cobra.Command{
			Use:   "server",
			Short: "Run the indexer server",
			Run:   server,
		})
	rootCmd.PersistentFlags().StringVar(&indexFlag, "index", "", "path to index repository")
	rootCmd.PersistentFlags().StringVar(&in, "in", "", "file with input URLs")
	rootCmd.Execute()
}
