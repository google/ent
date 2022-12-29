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

package main

type Config struct {
	ListenAddress string

	DomainName string

	RedisEnabled  bool
	RedisEndpoint string

	BigqueryEnabled bool
	BigqueryDataset string

	CloudStorageEnabled bool
	CloudStorageBucket  string

	GinMode  string
	LogLevel string

	Remotes []Remote

	Users []User
}

type Remote struct {
	Name string
}

type User struct {
	ID       uint64
	Name     string
	APIKey   string
	CanRead  bool
	CanWrite bool
}
