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
	"log"

	"github.com/google/ent/schema"
	"github.com/spf13/cobra"
)

var createSchemaCmd = &cobra.Command{
	Use:  "create-schema [filename]",
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		config := readConfig()
		remote := config.Remotes[0]
		nodeService := getObjectStore(remote)
		s := schema.Schema{
			Kinds: []schema.Kind{
				{
					KindID: "d76d88c5-2094-48b4-b4ed-dbf8df15fa59",
					Name:   "User",
					Fields: []schema.Field{
						{
							FieldID: 0,
							Name:    "whatever",
						},
					},
				},
			},
		}
		h, err := schema.PutStruct(nodeService, &s)
		if err != nil {
			log.Fatalf("could not create schema: %v", err)
		}
		log.Printf("created schema with hash %s", h)
	},
}
