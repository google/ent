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

package index

import (
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestDigestToPath(t *testing.T) {
	assert.Equal(t, "sha256/4f/20/63/ae/a9/fc/8a/fa/12/36/17/d1/bf/9e/51/8b/58/46/95/62", DigestToPath("sha256:4f2063aea9fc8afa123617d1bf9e518b58469562"))
}
