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
	"log"
	"os"

	"github.com/google/ent/cmd/ent/config"
	"github.com/google/ent/nodeservice"
	"github.com/google/ent/utils"
	"github.com/spf13/cobra"
)

var urlFlag string

func get(digest utils.Digest) ([]byte, error) {
	config := config.ReadConfig()
	objectGetter := getMultiplexObjectGetter(config)
	return objectGetter.Get(context.Background(), digest)
}

var getCmd = &cobra.Command{
	Use:  "get [digest]",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		digest, err := utils.ParseDigest(args[0])
		if err != nil {
			log.Fatalf("could not parse digest: %v", err)
			return
		}
		var object []byte
		if urlFlag != "" {
			// If a URL is specified, just fetch the object directly from there and ensure that it
			// has the correct digest.
			object, err = nodeservice.DownloadFromURL(digest, urlFlag)
		} else {
			object, err = get(digest)
		}
		if err != nil {
			log.Fatalf("could not get object: %v", err)
			return
		}
		os.Stdout.Write(object)
	},
}

func init() {
	getCmd.PersistentFlags().StringVar(&urlFlag, "url", "", "optional URL of the node to fetch")
}
