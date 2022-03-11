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
					KindID: 0,
					Name:   "Root",
					Fields: []schema.Field{
						{
							FieldID: 0,
							Name:    "book",
							KindID:  1,
						},
						{
							FieldID: 1,
							Name:    "film",
							KindID:  2,
						},
					},
				},
				{
					KindID: 1,
					Name:   "Book",
					Fields: []schema.Field{
						{
							FieldID: 0,
							Name:    "title",
							Raw:     1,
						},
						{
							FieldID: 1,
							Name:    "author",
							Raw:     1,
						},
					},
				},
				{
					KindID: 2,
					Name:   "Film",
					Fields: []schema.Field{
						{
							FieldID: 0,
							Name:    "title",
							Raw:     1,
						},
						{
							FieldID: 1,
							Name:    "director",
							Raw:     1,
						},
					},
				},
				{
					KindID: 3,
					Name:   "Docker",
					Fields: []schema.Field{
						{
							FieldID: 0,
							Name:    "command_build",
							KindID:  4,
						},
						{
							FieldID: 1,
							Name:    "command_run",
							KindID:  5,
						},
					},
				},
				{
					KindID: 4,
					Name:   "DockerBuild",
					Fields: []schema.Field{
						{
							FieldID: 0,
							Name:    "add-host",
							Raw:     1,
						},
						{
							FieldID: 1,
							Name:    "build-arg",
							Raw:     1,
						},
						{
							FieldID: 2,
							Name:    "cache-from",
							Raw:     1,
						},
						{
							FieldID: 3,
							Name:    "compress",
							Raw:     1,
						},
					},
				},
				{
					KindID: 5,
					Name:   "DockerRun",
					Fields: []schema.Field{
						{
							FieldID: 0,
							Name:    "attach",
							Raw:     1,
						},
						{
							FieldID: 1,
							Name:    "cap-add",
							Raw:     1,
						},
						{
							FieldID: 2,
							Name:    "cap-drop",
							Raw:     1,
						},
						{
							FieldID: 3,
							Name:    "detach",
							Raw:     1,
						},
					},
				},
			},
		}
		h, err := schema.PutStruct(nodeService, &s)
		if err != nil {
			log.Fatalf("could not create schema: %v", err)
		}
		log.Printf("created schema with digest %s", h)
	},
}
