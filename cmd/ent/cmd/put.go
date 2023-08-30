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
	"fmt"
	"io/ioutil"
	"os"

	"github.com/fatih/color"
	"github.com/google/ent/cmd/ent/config"
	"github.com/google/ent/cmd/ent/remote"
	"github.com/google/ent/log"
	"github.com/google/ent/nodeservice"
	"github.com/google/ent/utils"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
	"github.com/spf13/cobra"
	"github.com/tonistiigi/units"
)

var (
	remoteFlag       string
	digestFormatFlag string
	porcelainFlag    bool
)

var putCmd = &cobra.Command{
	Use:  "put [filename]",
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := ""
		if len(args) > 0 {
			filename = args[0]
		}
		ctx := context.Background()
		if filename == "" {
			err := putStdin()
			if err != nil {
				log.Criticalf(ctx, "could not read from stdin: %v", err)
				return
			}
		} else {
			_, err := traverseFileOrDir(filename, put)
			if err != nil {
				log.Criticalf(ctx, "could not traverse file: %v", err)
				return
			}
		}

	},
}

func putStdin() error {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("could not read stdin: %v", err)
	}
	digest := utils.ComputeDigest(data)
	link := cid.NewCidV1(utils.TypeRaw, multihash.Multihash(digest))
	return put(data, link, "-")
}

// func putFile(filename string) error {
// 	data, err := os.ReadFile(filename)
// 	if err != nil {
// 		return fmt.Errorf("could not read file %q: %v", filename, err)
// 	}
// 	return putData(data)
// }

func put(b []byte, link cid.Cid, name string) error {
	ctx := context.Background()
	config := config.ReadConfig()
	r := config.Remotes[0]
	if remoteFlag != "" {
		var err error
		r, err = remote.GetRemote(config, remoteFlag)
		if err != nil {
			return fmt.Errorf("could not use remote: %v", err)
		}
	}
	nodeService := remote.GetObjectStore(r)
	size := len(b)

	switch link.Type() {
	case utils.TypeRaw:
		digest := utils.Digest(link.Hash())
		digestString := utils.FormatDigest(digest, digestFormatFlag)
		marker := color.GreenString("-")
		if exists(nodeService, digest) {
			marker = color.GreenString("✓")
		} else {
			log.Infof(ctx, "putting object %q", digestString)
			_, err := nodeService.Put(ctx, uint64(size), bytes.NewReader(b))
			if err != nil {
				log.Errorf(ctx, "could not put object: %v", err)
				return fmt.Errorf("could not put object: %v", err)
			}
			marker = color.BlueString("↑")
		}
		if porcelainFlag {
			fmt.Printf("%s\n", digestString)
		} else {
			fmt.Printf("%s [%s %s] %s %.0f\n", color.YellowString(digestString), marker, r.Name, name, units.Bytes(size))
		}
		return nil
	case utils.TypeDAG:
		return nil
	default:
		return fmt.Errorf("unknown type: %v", link.Type())
	}
}

func exists(nodeService nodeservice.ObjectGetter, digest utils.Digest) bool {
	ctx := context.Background()
	ok, err := nodeService.Has(ctx, digest)
	if err != nil {
		log.Errorf(ctx, "could not check existence of %q: %v", digest, err)
		return false
	}
	return ok
}

func init() {
	putCmd.PersistentFlags().StringVar(&remoteFlag, "remote", "", "remote")
	putCmd.PersistentFlags().StringVar(&digestFormatFlag, "digest-format", "b58", "format [human, hex, b58]")
	putCmd.PersistentFlags().BoolVar(&porcelainFlag, "porcelain", false, "porcelain output (parseable by machines)")
}
