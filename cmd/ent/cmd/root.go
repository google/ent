//
// Copyright 2021 The Ent Authors.
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
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/google/ent/nodeservice"
	"github.com/spf13/cobra"
)

const (
	indexBasePath = "https://raw.githubusercontent.com/tiziano88/ent-index/main"
)

type Config struct {
	Remotes []Remote
}

// TODO: auth

type Remote struct {
	Name      string
	URL       string
	Index     bool
	APIKey    string `toml:"api_key"`
	Write     bool
	ReadGroup uint
}

func readConfig() Config {
	s, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("could not load config dir: %v", err)
	}
	s = filepath.Join(s, "ent.toml")
	f, err := ioutil.ReadFile(s)
	if err != nil {
		log.Printf("could not read config: %v", err)
		return defaultConfig()
	}
	config := Config{}
	err = toml.Unmarshal(f, &config)
	if err != nil {
		log.Fatalf("could not parse config: %v", err)
	}
	return config
}

func defaultConfig() Config {
	return Config{
		Remotes: []Remote{
			{
				Name:  "default",
				URL:   indexBasePath,
				Index: true,
			},
		},
	}
}

func getObjectStore(config Config) nodeservice.ObjectStore {
	for _, remote := range config.Remotes {
		if remote.Write {
			return nodeservice.Remote{
				APIURL: remote.URL,
			}
		}
	}
	return nil
}

func getObjectGetter(config Config) nodeservice.ObjectGetter {
	inner := make([]nodeservice.ObjectGetter, 0)
	for _, remote := range config.Remotes {
		if remote.Index {
			inner = append(inner, nodeservice.IndexClient{
				BaseURL: remote.URL,
			})
		} else {
			inner = append(inner, nodeservice.Remote{
				APIURL: remote.URL,
				APIKey: remote.APIKey,
			})
		}
	}
	return nodeservice.Multiplex{
		Inner: inner,
	}
}

var rootCmd = &cobra.Command{
	Use: "ent",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

var (
	remoteName string
)

func init() {
	rootCmd.PersistentFlags().StringVar(&remoteName, "remote", "", "")

	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(putCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(treeCmd)
	rootCmd.AddCommand(createSchemaCmd)
}
