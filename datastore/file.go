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

package datastore

import (
	"context"
	"io/ioutil"
	"os"
	"path"
)

// File is an implementation of DataStore using the local file system, rooted at the
// specified directory.
type File struct {
	DirName string
}

func (s File) Set(ctx context.Context, name string, value []byte) error {
	return ioutil.WriteFile(path.Join(s.DirName, name), value, 0644)
}

func (s File) Get(ctx context.Context, name string) ([]byte, error) {
	return ioutil.ReadFile(path.Join(s.DirName, name))
}

func (s File) Has(ctx context.Context, name string) (bool, error) {
	_, err := os.Stat(path.Join(s.DirName, name))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}
