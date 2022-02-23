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
	"io/ioutil"
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/google/ent/nodeservice"
	"github.com/google/ent/utils"
	"github.com/spf13/cobra"
)

var remoteFlag string

var putCmd = &cobra.Command{
	Use:  "put [filename]",
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
			err := putFile(filename)
			if err != nil {
				log.Fatal(err)
			}
		}

	},
}

func putStdin() error {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("could not read stdin: %v", err)
	}
	return putData(data)
}

func putFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("could not read file %q: %v", filename, err)
	}
	return putData(data)
}

func putData(data []byte) error {
	config := readConfig()
	remote := config.Remotes[0]
	if remoteFlag != "" {
		var err error
		remote, err = getRemote(config, remoteFlag)
		if err != nil {
			return fmt.Errorf("could not use remote: %v", err)
		}
	}
	nodeService := getObjectStore(remote)

	localHash := utils.ComputeHash(data)
	if exists(nodeService, localHash) {
		marker := color.GreenString("✓")
		fmt.Printf("%s %s [%s]\n", color.YellowString(string(localHash)), marker, remote.Name)
		return nil
	} else {
		_, err := nodeService.Put(context.Background(), data)
		if err != nil {
			return fmt.Errorf("could not add object: %v", err)
		}
		marker := color.BlueString("↑")
		fmt.Printf("%s %s [%s]\n", color.YellowString(string(localHash)), marker, remote.Name)
		return nil
	}
}

func exists(nodeService nodeservice.ObjectGetter, hash utils.Hash) bool {
	_, err := nodeService.Get(context.Background(), hash)
	if err == nodeservice.ErrNotFound {
		return false
	} else if err != nil {
		log.Fatalf("could not check existence of %q: %v", hash, err)
	}
	return true
}

func init() {
	putCmd.PersistentFlags().StringVar(&remoteFlag, "remote", "", "remote")
}
