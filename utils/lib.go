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

package utils

import (
	"encoding/hex"
	"fmt"

	"github.com/ipfs/go-cid"
	format "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	"github.com/multiformats/go-multihash"
)

func NewProtoNode() *merkledag.ProtoNode {
	node := merkledag.ProtoNode{}
	node.SetCidBuilder(merkledag.V1CidPrefix())
	return &node
}

func ParseProtoNode(b []byte) (*merkledag.ProtoNode, error) {
	node, err := merkledag.DecodeProtobuf(b)
	if err != nil {
		return nil, err
	}
	node.SetCidBuilder(merkledag.V1CidPrefix())
	return node, nil
}

func ParseRawNode(b []byte) (*merkledag.RawNode, error) {
	node, err := merkledag.NewRawNodeWPrefix(b, merkledag.V1CidPrefix())
	if err != nil {
		return nil, err
	}
	return node, nil
}

func ParseNodeFromBytes(c cid.Cid, b []byte) (format.Node, error) {
	codec := c.Prefix().Codec
	switch codec {
	case cid.DagProtobuf:
		return ParseProtoNode(b)
	case cid.Raw:
		return ParseRawNode(b)
	default:
		return nil, fmt.Errorf("invalid codec: %d (%s)", codec, cid.CodecToStr[codec])
	}
}

func GetLink(node *merkledag.ProtoNode, name string) (cid.Cid, error) {
	link, err := node.GetNodeLink(name)
	if err != nil {
		return cid.Undef, err
	}
	return link.Cid, nil
}

func SetLink(node *merkledag.ProtoNode, name string, hash cid.Cid) error {
	node.RemoveNodeLink(name) // Ignore errors
	return node.AddRawLink(name, &format.Link{
		Cid: hash,
	})
}

func RemoveLink(node *merkledag.ProtoNode, name string) error {
	return node.RemoveNodeLink(name)
}

func Hash(c cid.Cid) string {
	return hex.EncodeToString([]byte(c.Hash()))
	// return c.Hash().B58String()
}

func ParseHash(s string) (multihash.Multihash, error) {
	return multihash.FromHexString(s)
}
