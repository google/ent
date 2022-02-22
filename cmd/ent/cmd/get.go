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

	"github.com/google/ent/utils"
	"github.com/spf13/cobra"
)

func get(hash utils.Hash) {
	config := readConfig()
	objectGetter := getMultiplexObjectGetter(config)
	object, err := objectGetter.Get(context.Background(), hash)
	if err != nil {
		log.Fatalf("could not download target: %s", err)
	}
	os.Stdout.Write(object)
}

var getCmd = &cobra.Command{
	Use:  "get [hash]",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		hash, err := utils.ParseHash(args[0])
		if err != nil {
			log.Fatalf("could not parse hash: %v", err)
			return
		}
		get(hash)
	},
}
