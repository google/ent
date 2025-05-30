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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/google/ent/api"
	"github.com/google/ent/cmd/ent/config"
	"github.com/google/ent/log"
	"github.com/google/ent/utils"
	"github.com/spf13/cobra"
)

var (
	urlFlag    string
	outFlag    string
	digestFlag string
)

var getCmd = &cobra.Command{
	Use:  "get",
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		digest, err := utils.ParseDigest(digestFlag)
		if err != nil {
			log.Criticalf(ctx, "parse digest: %v", err)
			os.Exit(1)
		}
		// Make API request to get entry metadata and mirrors
		resp, err := getEntry(ctx, digest)
		if err != nil {
			log.Criticalf(ctx, "get entry request failed: %v", err)
			os.Exit(1)
		}
		log.Debugf(ctx, "got entry response: metadata=%+v mirrors=%+v", resp.Metadata, resp.Mirrors)
		fmt.Printf("size: %v\n", resp.Metadata.LengthBytes)
		for _, mirror := range resp.Mirrors {
			log.Debugf(ctx, "mirror: %v", mirror)
			object, err := http.Get(mirror.URL)
			if err != nil {
				log.Criticalf(ctx, "get mirror: %v", err)
				os.Exit(1)
			}
			defer object.Body.Close()

			// Read the full response body
			body, err := io.ReadAll(object.Body)
			if err != nil {
				log.Criticalf(ctx, "read mirror body: %v", err)
				os.Exit(1)
			}

			// Verify the digest matches
			actualDigest := utils.ComputeDigest(body)
			if !bytes.Equal(actualDigest, digest) {
				log.Criticalf(ctx, "digest mismatch: got %v, want %v", actualDigest, digest)
				continue
			}
			log.Debugf(ctx, "digest matches")

			if outFlag != "" {
				os.WriteFile(outFlag, body, 0644)
			}

			os.Exit(0)
		}
		os.Exit(0)
	},
}

func getEntry(ctx context.Context, digest utils.Digest) (*api.GetEntryResponse, error) {
	req := api.GetEntryRequest{
		Digests: utils.DigestToApi(digest),
	}
	log.Debugf(ctx, "sending request: %v", req)
	var resp api.GetEntryResponse
	config := config.ReadConfig()
	client := &http.Client{}
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		config.Remotes[0].URL+"/"+api.GET_ENTRY_METHOD_ID,
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %v", httpResp.Status)
	}

	err = json.NewDecoder(httpResp.Body).Decode(&resp)
	if err != nil {
		return nil, err
	}

	return &api.GetEntryResponse{
		Metadata: resp.Metadata,
		Mirrors:  resp.Mirrors,
	}, nil
}

func init() {
	getCmd.PersistentFlags().StringVar(&digestFlag, "digest", "", "digest of the object to fetch")
	getCmd.PersistentFlags().StringVar(&urlFlag, "url", "", "optional URL of the object to fetch")
	getCmd.PersistentFlags().StringVar(&outFlag, "out", "", "optional output file")
}
