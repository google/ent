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
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"cloud.google.com/go/firestore"
	"github.com/google/ent/index"
	"github.com/google/ent/utils"
	"github.com/spf13/cobra"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

var (
	indexFlag string

	firebaseCredentials string
	firebaseProject     string
	concurrency         int

	urlFlag string
)

type URL struct {
	URL string
}

func server(cmd *cobra.Command, args []string) {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, firebaseProject, option.WithCredentialsFile(firebaseCredentials))
	if err != nil {
		log.Fatalf("could not create client: %v", err)
	}
	if indexFlag == "" {
		log.Fatal("index flag is required")
	}
	iter := client.Collection("urls").Documents(ctx)
	defer iter.Stop()
	wg := sync.WaitGroup{}
	tokens := make(chan struct{}, concurrency)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("error iterating over URLs: %v", err)
		}
		url := doc.Data()["url"].(string)
		if url == "" {
			continue
		}
		wg.Add(1)
		tokens <- struct{}{}
		go func() {
			defer func() {
				wg.Done()
				<-tokens
			}()
			fetch(url)
		}()
	}
	wg.Wait()
}

func get(cmd *cobra.Command, args []string) {
	fetch(urlFlag)
}

func fetch(url string) {
	log.Print(url)
	res, err := http.Get(url)
	if err != nil {
		log.Printf("could not get URL %q: %v", url, err)
		return
	}
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("could not read HTTP body: %v", err)
		return
	}
	h := utils.ComputeHash(data)
	l := filepath.Join(indexFlag, index.HashToPath(h), index.EntryFilename)
	e := index.IndexEntry{
		Hash: h,
		Size: len(data),
		URLS: []string{url},
	}
	es, err := json.Marshal(e)
	if err != nil {
		log.Printf("could not marshal JSON: %v", err)
		return
	}
	log.Printf("%s", es)
	err = os.MkdirAll(filepath.Dir(l), 0755)
	if err != nil {
		log.Printf("could not create file: %v", err)
		return
	}
	ioutil.WriteFile(l, es, 0644)
}

func main() {
	var rootCmd = &cobra.Command{Use: "indexer"}

	serverCmd :=
		&cobra.Command{
			Use:   "server",
			Short: "Run the indexer server",
			Run:   server,
		}
	serverCmd.PersistentFlags().StringVar(&indexFlag, "index", "", "path to index repository")
	serverCmd.PersistentFlags().StringVar(&firebaseProject, "firebase-project", "", "Firebase project name")
	serverCmd.PersistentFlags().StringVar(&firebaseCredentials, "firebase-credentials", "", "file with Firebase credentials")
	serverCmd.PersistentFlags().IntVar(&concurrency, "concurrency", 10, "HTTP fetch concurrency")
	rootCmd.AddCommand(serverCmd)

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Index a single URL",
		Run:   get,
	}
	getCmd.PersistentFlags().StringVar(&indexFlag, "index", "", "path to index repository")
	getCmd.PersistentFlags().StringVar(&urlFlag, "url", "", "url of the entry to index")
	rootCmd.AddCommand(getCmd)

	rootCmd.Execute()
}
