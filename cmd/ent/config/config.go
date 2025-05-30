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

package config

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const (
	indexBasePath = "https://raw.githubusercontent.com/tiziano88/ent-index/main"
	staticSpace   = "https://api.static.space/v1"
	localhost     = "http://localhost:8081/v1"
)

type Config struct {
	Remotes   []Remote
	SecretKey string `toml:"secret_key"`
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

func ReadConfig() Config {
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
				URL:   localhost,
				Index: true,
			},
		},
	}
}
