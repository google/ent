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

	"github.com/fatih/color"
	"github.com/google/ent/cmd/ent/config"
	"github.com/google/ent/log"
	"github.com/google/ent/utils"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:  "status [digest]",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		digest, err := utils.ParseDigest(args[0])
		if err != nil {
			log.Criticalf(ctx, "parse digest: %v", err)
			os.Exit(1)
		}
		status(digest)
	},
}

func status(digest utils.Digest) {
	c := config.ReadConfig()
	s := []<-chan string{}
	for _, remote := range c.Remotes {
		c := make(chan string)
		s = append(s, c)
		go func(remote config.Remote, c chan<- string) {
			objectGetter := getObjectGetter(remote)
			marker := color.GreenString("✓")
			_, err := objectGetter.Get(context.Background(), digest)
			if err != nil {
				marker = color.RedString("✗")
			}
			c <- fmt.Sprintf("%s %s [%s]\n", color.YellowString(digest.String()), marker, remote.Name)
		}(remote, c)
	}
	for _, c := range s {
		fmt.Printf(<-c)
	}
}
