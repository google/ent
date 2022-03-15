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
	"fmt"
	"io/ioutil"

	"cloud.google.com/go/storage"
	"github.com/google/ent/log"
)

// Cloud is an implementation of DataStore using a Google Cloud Storage bucket.
type Cloud struct {
	Client     *storage.Client
	BucketName string
}

func (s Cloud) Get(ctx context.Context, name string) ([]byte, error) {
	rc, err := s.Client.Bucket(s.BucketName).Object(name).NewReader(ctx)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("error reading from cloud storage: %v", err)
	}
	err = rc.Close()
	if err != nil {
		return nil, fmt.Errorf("error closing reader from cloud storage: %v", err)
	}
	return body, nil
}

func (s Cloud) Put(ctx context.Context, name string, value []byte) error {
	o := s.Client.Bucket(s.BucketName).Object(name)
	attr, err := o.Attrs(ctx)
	if err == storage.ErrObjectNotExist {
		wc := o.NewWriter(ctx)
		_, err := wc.Write(value)
		if err != nil {
			return fmt.Errorf("error writing to cloud storage: %v", err)
		}
		err = wc.Close()
		if err != nil {
			return fmt.Errorf("error closing writer to cloud storage: %v", err)
		}
		return nil
	} else if err != nil {
		return fmt.Errorf("error getting attrs from cloud storage: %v", err)
	}
	log.Infof(ctx, "object %q already exists in cloud storage: %+v", name, attr)
	return nil
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
