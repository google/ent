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

package utils

import (
	"bytes"
	"testing"
)

func TestParseDigest(t *testing.T) {
	digest0, err := ParseDigest("sha256:1f209f17903dc0310f9a0fe337d3a893193f20b4171895a74d0200d6019dedd6")
	if err != nil {
		t.Fatal(err)
	}
	digest1, err := ParseDigest("12201f209f17903dc0310f9a0fe337d3a893193f20b4171895a74d0200d6019dedd6")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(digest0, digest1) {
		t.Fatalf("digests should be equal")
	}
	digestString := digest0.String()
	if digestString != "12201f209f17903dc0310f9a0fe337d3a893193f20b4171895a74d0200d6019dedd6" {
		t.Fatalf("digest string should be equal")
	}

	base32Digest, err := ToBase32(digest0)
	if err != nil {
		t.Fatal(err)
	}
	expectedBase32Digest := "CIQB6IE7C6ID3QBRB6NA7YZX2OUJGGJ7EC2BOGEVU5GQEAGWAGO63VQ"
	if base32Digest != expectedBase32Digest {
		t.Fatalf("incorrect base32 digest")
	}
	back, err := FromBase32(base32Digest)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(digest0, back) {
		t.Fatalf("digests should be equal")
	}
}
