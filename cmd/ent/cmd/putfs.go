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
	"log"

	"github.com/fatih/color"
	"github.com/google/ent/cmd/ent/config"
	"github.com/google/ent/cmd/ent/remote"
	"github.com/google/ent/utils"
	"github.com/ipfs/go-cid"
	"github.com/spf13/cobra"
)

var putFSCmd = &cobra.Command{
	Use:  "putfs [filename]",
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := ""
		if len(args) > 0 {
			filename = args[0]
		}
		if filename == "" {
			err := putStdin()
			if err != nil {
				log.Fatal(err)
			}
		} else {
			_, err := traverseFileOrDir(filename, putFS)
			if err != nil {
				log.Fatal(err)
			}
		}

	},
}

func putFS(b []byte, link cid.Cid, name string) error {
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
	digest := utils.Digest(link.Hash())
	if exists(nodeService, digest) {
		marker := color.GreenString("✓")
		fmt.Printf("%s %s [%s] %s\n", color.YellowString(link.String()), marker, r.Name, name)
		return nil
	} else {
		_, err := nodeService.Put(context.Background(), uint64(size), bytes.NewReader(b))
		if err != nil {
			log.Printf("could not put object: %v", err)
			return fmt.Errorf("could not put object: %v", err)
		}
		marker := color.BlueString("↑")
		fmt.Printf("%s %s [%s] %s\n", color.YellowString(link.String()), marker, r.Name, name)
		return nil
	}
}

func init() {
	putFSCmd.PersistentFlags().StringVar(&remoteFlag, "remote", "", "remote")
}
