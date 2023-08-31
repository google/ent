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
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/google/ent/cmd/ent/config"
	"github.com/google/ent/cmd/ent/remote"
	"github.com/google/ent/log"
	"github.com/google/ent/nodeservice"
	"github.com/google/ent/schema"
	"github.com/google/ent/utils"
	"github.com/spf13/cobra"
)

var (
	schemaFlag string
	s          schema.Schema
)

func tree(o nodeservice.ObjectGetter, digest utils.Digest, indent int, kindID uint32) {
	ctx := context.Background()
	object, err := o.Get(ctx, digest)
	if err != nil {
		log.Criticalf(ctx, "download target: %s", err)
		os.Exit(1)
	}
	node, err := utils.ParseDAGNode(object)
	if err != nil {
		fmt.Printf("%s %s\n", strings.Repeat("  ", indent), object)
		return
	}
	k := kind(s, kindID)
	kindName := k.Name
	fmt.Printf("%s %s\n", strings.Repeat("  ", indent), color.GreenString(kindName))
	for i, link := range node.Links {
		selector := fmt.Sprintf("%d", i)
		fmt.Printf("%s %s %s\n", strings.Repeat("  ", indent), color.BlueString(selector), color.YellowString(link.String()))
		tree(o, utils.Digest(link.Hash()), indent+1, 0 /* TODO */)
	}
}

func kind(s schema.Schema, kindID uint32) schema.Kind {
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
	Use:  "tree [digest]",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		digest, err := utils.ParseDigest(args[0])
		if err != nil {
			log.Criticalf(ctx, "parse digest: %v", err)
			os.Exit(1)
		}

		config := config.ReadConfig()
		r := config.Remotes[0]
		if remoteFlag != "" {
			var err error
			r, err = remote.GetRemote(config, remoteFlag)
			if err != nil {
				log.Criticalf(ctx, "use remote: %v", err)
				os.Exit(1)
			}
		}
		o := remote.GetObjectStore(r)
		// o1 := nodeservice.Cached{
		// 	Cache: make(map[utils.DigestArray][]byte),
		// 	Inner: o,
		// }

		if schemaFlag != "" {
			schemaDigest, err := utils.ParseDigest(schemaFlag)
			if err != nil {
				log.Criticalf(ctx, "parse schema digest: %v", err)
				os.Exit(1)
			}
			err = schema.GetStruct(o, schemaDigest, &s)
			if err != nil {
				log.Criticalf(ctx, "load schema: %v", err)
				os.Exit(1)
			}
			log.Infof(ctx, "loaded schema: %+v", s)
		}
		tree(o, digest, 0, 0)
	},
}

func init() {
	treeCmd.PersistentFlags().StringVar(&schemaFlag, "schema", "", "digest of schema")
	treeCmd.PersistentFlags().StringVar(&remoteFlag, "remote", "", "remote")
}
