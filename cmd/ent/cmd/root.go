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
	"fmt"
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

func getRemote(config Config, remoteName string) (Remote, error) {
	for _, remote := range config.Remotes {
		if remote.Name == remoteName {
			return remote, nil
		}
	}
	return Remote{}, fmt.Errorf("remote %q not found", remoteName)
}

func getObjectStore(remote Remote) nodeservice.NodeService {
	if remote.Write {
		return nodeservice.Remote{
			APIURL: remote.URL,
			APIKey: remote.APIKey,
		}
	} else {
		return nil
	}
}

func getMultiplexObjectGetter(config Config) nodeservice.ObjectGetter {
	inner := make([]nodeservice.Inner, 0)
	for _, remote := range config.Remotes {
		inner = append(inner, nodeservice.Inner{
			Name:         remote.Name,
			ObjectGetter: getObjectGetter(remote)})
	}
	return nodeservice.Multiplex{
		Inner: inner,
	}
}

func getObjectGetter(remote Remote) nodeservice.ObjectGetter {
	if remote.Index {
		return nodeservice.IndexClient{
			BaseURL: remote.URL,
		}
	} else {
		return nodeservice.Remote{
			APIURL: remote.URL,
			APIKey: remote.APIKey,
		}
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

func init() {
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(putCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(treeCmd)
	rootCmd.AddCommand(createSchemaCmd)
	rootCmd.AddCommand(printSchemaCmd)
	rootCmd.AddCommand(uploadCmd)
}
