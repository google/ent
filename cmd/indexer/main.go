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
	"net/url"
	"os"
	"path/filepath"
	"sort"
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
			_, err := fetch(url)
			if err != nil {
				log.Printf("could not fetch URL: %v", err)
			}
		}()
	}
	wg.Wait()
}

func fetchCmd(cmd *cobra.Command, args []string) {
	e, err := fetch(urlFlag)
	if err != nil {
		log.Fatalf("could not fetch URL: %v", err)
	}
	// Print hash to stdout.
	fmt.Printf("%s\n", e.Digest)
}

func fetch(urlString string) (*index.IndexEntry, error) {
	log.Print(urlString)
	parsedURL, err := url.Parse(urlString)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.User != nil {
		return nil, fmt.Errorf("non-empty user in URL")
	}
	if parsedURL.Fragment != "" {
		return nil, fmt.Errorf("non-empty fragment in URL")
	}

	urlString = parsedURL.String()
	log.Printf("fetching %q", urlString)

	res, err := http.Get(urlString)
	if err != nil {
		return nil, fmt.Errorf("could not fetch URL %q: %w", urlString, err)
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid error code %d (%s)", res.StatusCode, res.Status)
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read HTTP body: %w", err)
	}
	h := utils.ComputeHash(data)

	l := filepath.Join(indexFlag, index.HashToPath(h), index.EntryFilename)

	var e index.IndexEntry
	if _, err := os.Stat(l); err == nil {
		log.Printf("index entry existing: %q", l)
		bytes, err := ioutil.ReadFile(l)
		if err != nil {
			return nil, fmt.Errorf("could not read index entry: %w", err)
		}
		err = json.Unmarshal(bytes, &e)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal JSON for index entry: %w", err)
		}

		if sort.SearchStrings(e.URLS, urlString) != -1 {
			// URL already in the indexed entry.
		} else {
			// Add the new URL.
			e.URLS = append(e.URLS, urlString)
			sort.Strings(e.URLS)
		}

		// Fix all fields just in case.
		e.MediaType = http.DetectContentType(data)
		e.Digest = h
		e.Size = len(data)
	} else {
		e = index.IndexEntry{
			MediaType: http.DetectContentType(data),
			Digest:    h,
			Size:      len(data),
			URLS:      []string{urlString},
		}
	}
	log.Printf("index entry to create: %+v", e)

	es, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("could not marshal JSON: %w", err)
	}
	err = os.MkdirAll(filepath.Dir(l), 0755)
	if err != nil {
		return nil, fmt.Errorf("could not create file: %w", err)
	}
	err = ioutil.WriteFile(l, es, 0644)
	if err != nil {
		return nil, fmt.Errorf("could not write to file: %w", err)
	}

	log.Printf("index entry created: %q", l)

	return &e, nil
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
		Use:   "fetch",
		Short: "Fetch and index a single URL",
		Run:   fetchCmd,
	}
	getCmd.PersistentFlags().StringVar(&indexFlag, "index", "", "path to index repository")
	getCmd.PersistentFlags().StringVar(&urlFlag, "url", "", "url of the entry to index")
	rootCmd.AddCommand(getCmd)

	rootCmd.Execute()
}
