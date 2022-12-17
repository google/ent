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
}

func TestDigestToArray(t *testing.T) {
	digest, err := ParseDigest("12201f209f17903dc0310f9a0fe337d3a893193f20b4171895a74d0200d6019dedd6")
	if err != nil {
		t.Fatal(err)
	}
	digestArray := [64]byte(DigestToArray(digest))
	expectedDigestArray := [64]byte{
		0x12, 0x20,
		0x1f, 0x20, 0x9f, 0x17, 0x90, 0x3d, 0xc0, 0x31, 0x0f, 0x9a, 0x0f, 0xe3, 0x37, 0xd3, 0xa8, 0x93,
		0x19, 0x3f, 0x20, 0xb4, 0x17, 0x18, 0x95, 0xa7, 0x4d, 0x02, 0x00, 0xd6, 0x01, 0x9d, 0xed, 0xd6,
	}
	if digestArray != expectedDigestArray {
		t.Fatalf("digest array should be equal:\n%x\n%x", digestArray, expectedDigestArray)
	}
}
