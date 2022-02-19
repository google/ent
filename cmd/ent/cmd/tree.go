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
	"fmt"
	"log"
	"strings"

	"github.com/fatih/color"
	"github.com/google/ent/nodeservice"
	"github.com/google/ent/schema"
	"github.com/google/ent/utils"
	"github.com/spf13/cobra"
)

var (
	schemaFlag string
	s          schema.Schema
)

func tree(o nodeservice.ObjectGetter, hash utils.Hash, indent int) {
	object, err := o.Get(context.Background(), hash)
	if err != nil {
		log.Fatalf("could not download target: %s", err)
	}
	node, err := utils.ParseNode(object)
	if err != nil {
		fmt.Printf("%s %s\n", strings.Repeat("  ", indent), object)
		return
	}
	k := kind(node.Kind)
	kindName := k.Name
	if kindName == "" {
		kindName = node.Kind
	}
	fmt.Printf("%s %s\n", strings.Repeat("  ", indent), color.GreenString(kindName))
	for fieldID, links := range node.Links {
		f := field(k, uint32(fieldID))
		fieldName := f.Name
		if fieldName == "" {
			fieldName = fmt.Sprintf("%d", f.FieldID)
		}
		for index, link := range links {
			selector := fmt.Sprintf("%s[%d]", fieldName, index)
			fmt.Printf("%s %s %s\n", strings.Repeat("  ", indent), color.BlueString(selector), color.YellowString(string(link.Hash)))
			tree(o, link.Hash, indent+1)
		}
	}
}

func kind(kindID string) schema.Kind {
	for _, k := range s.Kinds {
		if k.KindID == kindID {
			return k
		}
	}
	return schema.Kind{}
}

func field(k schema.Kind, fieldID uint32) schema.Field {
	for _, f := range k.Fields {
		if f.FieldID == fieldID {
			return f
		}
	}
	return schema.Field{}
}

var treeCmd = &cobra.Command{
	Use:  "tree [hash]",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		hash, err := utils.ParseHash(args[0])
		if err != nil {
			log.Fatalf("could not parse hash: %v", err)
			return
		}

		config := readConfig()
		o := getObjectGetter(config)
		if schemaFlag != "" {
			schemaHash, err := utils.ParseHash(schemaFlag)
			if err != nil {
				log.Fatalf("could not parse schema hash: %v", err)
				return
			}
			err = schema.GetStruct(o, schemaHash, &s)
			if err != nil {
				log.Fatalf("could not load schema: %v", err)
				return
			}
			log.Printf("loaded schema: %+v", s)
		}
		tree(o, hash, 0)
	},
}

func init() {
	treeCmd.PersistentFlags().StringVar(&schemaFlag, "schema", "", "digest of schema")
}
