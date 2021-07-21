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

	"cloud.google.com/go/storage"
)

// Cloud is an implementation of DataStore using a Google Cloud Storage bucket.
type Cloud struct {
	Client     *storage.Client
	BucketName string
}

func (s Cloud) Set(ctx context.Context, name string, value []byte) error {
	wc := s.Client.Bucket(s.BucketName).Object(name).NewWriter(ctx)
	defer wc.Close()
	_, err := wc.Write(value)
	return err
}

func (s Cloud) Get(ctx context.Context, name string) ([]byte, error) {
	rc, err := s.Client.Bucket(s.BucketName).Object(name).NewReader(ctx)
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	body, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (s Cloud) Has(ctx context.Context, name string) (bool, error) {
	_, err := s.Client.Bucket(s.BucketName).Object(name).Attrs(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}
