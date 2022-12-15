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
	"strconv"
	"strings"

	"github.com/google/ent/utils"
	"github.com/spf13/cobra"
)

var uploadCmd = &cobra.Command{
	Use:  "upload [plan filename]",
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := ""
		if len(args) > 0 {
			filename = args[0]
		} else {
			log.Fatal("no filename specified")
		}
		err := upload(filename)
		if err != nil {
			log.Fatal(err)
		}
	},
}

type Plan struct {
	Entries []Entry
}

type Entry struct {
	FieldID  uint
	Filename string
}

func parsePlan(planFilename string) (Plan, error) {
	data, err := ioutil.ReadFile(planFilename)
	if err != nil {
		return Plan{}, fmt.Errorf("could not read plan file: %v", err)
	}
	lines := strings.Split(string(data), "\n")
	plan := Plan{
		Entries: []Entry{},
	}
	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Split(line, " ")
		fieldID, err := strconv.Atoi(fields[0])
		if err != nil {
			return Plan{}, fmt.Errorf("could not parse field id: %v", err)
		}

		plan.Entries = append(plan.Entries, Entry{
			FieldID:  uint(fieldID),
			Filename: fields[1],
		})
	}
	return plan, nil
}

func upload(planFilename string) error {
	plan, err := parsePlan(planFilename)
	if err != nil {
		return fmt.Errorf("could not parse plan file: %v", err)
	}
	log.Printf("parsed plan: %#v", plan)

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

	dagNode := utils.DAGNode{
		Links: []utils.Link{},
	}

	for _, entry := range plan.Entries {
		data, err := ioutil.ReadFile(entry.Filename)
		if err != nil {
			return fmt.Errorf("could not read file: %v", err)
		}
		log.Printf("read file: %s", entry.Filename)
		digest, err := nodeService.Put(context.Background(), data)
		if err != nil {
			return fmt.Errorf("could not put file: %v", err)
		}
		dagNode.Links = append(dagNode.Links, utils.Link{
			Type:   utils.TypeRaw,
			Digest: digest,
		})
	}

	log.Printf("dag node: %#v", dagNode)

	dagNodeBytes, err := utils.SerializeDAGNode(&dagNode)
	if err != nil {
		return fmt.Errorf("could not serialize dag node: %v", err)
	}

	root, err := nodeService.Put(context.Background(), dagNodeBytes)
	if err != nil {
		return fmt.Errorf("could not put dag node: %v", err)
	}
	log.Printf("root: %s", root)

	return nil
}

func init() {
	uploadCmd.PersistentFlags().StringVar(&remoteFlag, "remote", "", "remote")
}
