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
	"github.com/google/ent/utils"
	"github.com/spf13/cobra"
)

var printSchemaCmd = &cobra.Command{
	Use:  "print-schema [digest]",
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		schemaDigest, err := utils.ParseDigest(args[0])
		if err != nil {
			log.Fatalf("could not parse digest: %v", err)
			return
		}

		config := readConfig()
		o := getMultiplexObjectGetter(config)
		if err != nil {
			log.Fatalf("could not parse schema digest: %v", err)
			return
		}
		err = schema.GetStruct(o, schemaDigest, &s)
		if err != nil {
			log.Fatalf("could not load schema: %v", err)
			return
		}
		log.Printf("loaded schema: %+v", s)
	},
}
