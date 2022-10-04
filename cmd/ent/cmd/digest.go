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
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/google/ent/utils"
	"github.com/spf13/cobra"
)

var digestCmd = &cobra.Command{
	Use:  "digest [filename]",
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := ""
		if len(args) > 0 {
			filename = args[0]
		}
		if filename == "" {
			err := digestStdin()
			if err != nil {
				log.Fatal(err)
			}
		} else {
			err := digestFile(filename)
			if err != nil {
				log.Fatal(err)
			}
		}

	},
}

func digestStdin() error {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("could not read stdin: %v", err)
	}
	return digestData(data)
}

func digestFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("could not read file %q: %v", filename, err)
	}
	return digestData(data)
}

func digestData(data []byte) error {
	localDigest := utils.ComputeDigest(data)
	fmt.Printf("%s\n", color.YellowString(string(localDigest)))
	return nil
}
