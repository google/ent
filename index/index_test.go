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
	"github.com/google/ent/utils"
)

func TestDigestToPath(t *testing.T) {
	digest, err := utils.ParseDigest("sha256:366ac3bdad37d1bdc0ca87e2ea60111872e2c8d7aac8a18f2588d791056e658f")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assert.Equal(t, "sha256/36/6a/c3/bd/ad/37/d1/bd/c0/ca/87/e2/ea/60/11/18/72/e2/c8/d7/aa/c8/a1/8f/25/88/d7/91/05/6e/65/8f", DigestToPath(digest))
}
