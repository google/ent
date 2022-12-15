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
			_, err := traverseFileOrDir(filename, put)
			if err != nil {
				log.Fatal(err)
			}
		}

	},
}

func putStdin() error {
	// data, err := ioutil.ReadAll(os.Stdin)
	// if err != nil {
	// 	return fmt.Errorf("could not read stdin: %v", err)
	// }
	// return putData(data)
	return nil
}

// func putFile(filename string) error {
// 	data, err := os.ReadFile(filename)
// 	if err != nil {
// 		return fmt.Errorf("could not read file %q: %v", filename, err)
// 	}
// 	return putData(data)
// }

func put(bytes []byte, digest utils.Digest, name string) error {
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

	if exists(nodeService, digest) {
		marker := color.GreenString("✓")
		fmt.Printf("%s %s [%s] %s\n", color.YellowString(digest.String()), marker, remote.Name, name)
		return nil
	} else {
		_, err := nodeService.Put(context.Background(), bytes)
		if err != nil {
			log.Printf("could not put object: %v", err)
			return fmt.Errorf("could not put object: %v", err)
		}
		marker := color.BlueString("↑")
		fmt.Printf("%s %s [%s] %s\n", color.YellowString(digest.String()), marker, remote.Name, name)
		return nil
	}
}

func exists(nodeService nodeservice.ObjectGetter, digest utils.Digest) bool {
	_, err := nodeService.Get(context.Background(), digest)
	if err == nodeservice.ErrNotFound {
		return false
	} else if err != nil {
		log.Fatalf("could not check existence of %q: %v", digest, err)
	}
	return true
}

func init() {
	putCmd.PersistentFlags().StringVar(&remoteFlag, "remote", "", "remote")
}
