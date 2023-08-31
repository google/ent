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

package cmd

import (
	"context"
	"os"

	"github.com/google/ent/cmd/ent/cmd/tag"
	"github.com/google/ent/cmd/ent/config"
	"github.com/google/ent/log"
	"github.com/google/ent/nodeservice"
	"github.com/spf13/cobra"
)

func getMultiplexObjectGetter(c config.Config) nodeservice.ObjectGetter {
	inner := make([]nodeservice.Inner, 0)
	for _, remote := range c.Remotes {
		inner = append(inner, nodeservice.Inner{
			Name:         remote.Name,
			ObjectGetter: getObjectGetter(remote)})
	}
	return nodeservice.Sequence{
		Inner: inner,
	}
}

func getObjectGetter(remote config.Remote) nodeservice.ObjectGetter {
	if remote.Index {
		return nodeservice.IndexClient{
			BaseURL: remote.URL,
		}
	} else {
		return nodeservice.Remote{
			APIURL: remote.URL,
			APIKey: remote.APIKey,
		}
	}
}

var rootCmd = &cobra.Command{
	Use: "ent",
}

func Execute() {
	ctx := context.Background()
	if err := rootCmd.Execute(); err != nil {
		log.Criticalf(ctx, "execute command: %v", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(digestCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(putFSCmd)
	rootCmd.AddCommand(putCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(treeCmd)
	rootCmd.AddCommand(createSchemaCmd)
	rootCmd.AddCommand(printSchemaCmd)
	rootCmd.AddCommand(keygenCmd)
	rootCmd.AddCommand(tag.TagCmd)
}

func GetObjectGetter() nodeservice.ObjectGetter {
	config := config.ReadConfig()
	return getMultiplexObjectGetter(config)
}
