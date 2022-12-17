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
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-varint"
)

type DAGNode struct {
	Bytes []byte
	Links []cid.Cid
}

const (
	TypeRaw = cid.Raw
	TypeDAG = 0x70
)

type Path []Selector

// Index of the link to follow
type Selector uint

var order = binary.BigEndian

const (
	FieldTypeInt   = 0
	FieldTypeBytes = 1
	FieldTypeMsg   = 2
)

type Field struct {
	ID   uint64
	Type uint64

	UintValue  uint64
	BytesValue []byte
}

func ReadUint64(b *bytes.Reader) (uint64, error) {
	return varint.ReadUvarint(b)
}

func WriteUint64(b *bytes.Buffer, i uint64) error {
	_, err := b.Write(varint.ToUvarint(i))
	return err
}

func DecodeField(b *bytes.Reader) (*Field, error) {
	fieldID, err := ReadUint64(b)
	if err == io.EOF {
		return nil, err
	} else if err != nil {
		return nil, fmt.Errorf("could not read field ID: %w", err)
	}
	fieldType, err := ReadUint64(b)
	if err != nil {
		return nil, fmt.Errorf("could not read field type: %w", err)
	}
	switch fieldType {
	case FieldTypeInt:
		fieldValue, err := ReadUint64(b)
		if err != nil {
			return nil, fmt.Errorf("could not read field value: %w", err)
		}
		return &Field{
			ID:         fieldID,
			Type:       fieldType,
			UintValue:  fieldValue,
			BytesValue: nil,
		}, nil
	case FieldTypeBytes:
		fieldLength, err := ReadUint64(b)
		if err != nil {
			return nil, fmt.Errorf("could not read field length: %w", err)
		}
		fieldValue := make([]byte, fieldLength)
		if fieldLength > 0 {
			n, err := b.Read(fieldValue)
			if err != nil {
				return nil, fmt.Errorf("could not read field value: %w", err)
			}
			if n != int(fieldLength) {
				return nil, fmt.Errorf("could not read field value, read %d bytes, expected %d", n, fieldLength)
			}
		}
		return &Field{
			ID:         fieldID,
			Type:       fieldType,
			UintValue:  0,
			BytesValue: fieldValue,
		}, nil
	case FieldTypeMsg:
		fieldValue, err := ReadUint64(b)
		if err != nil {
			return nil, fmt.Errorf("could not read field value: %w", err)
		}
		return &Field{
			ID:         fieldID,
			Type:       fieldType,
			UintValue:  fieldValue,
			BytesValue: nil,
		}, nil
	default:
		return nil, fmt.Errorf("unknown field type: %d", fieldType)
	}
}

func EncodeField(b *bytes.Buffer, f *Field) error {
	if err := WriteUint64(b, f.ID); err != nil {
		return fmt.Errorf("could not write field ID: %w", err)
	}
	if err := WriteUint64(b, f.Type); err != nil {
		return fmt.Errorf("could not write field type: %w", err)
	}
	switch f.Type {
	case FieldTypeInt:
		if err := WriteUint64(b, f.UintValue); err != nil {
			return fmt.Errorf("could not write field value: %w", err)
		}
	case FieldTypeBytes:
		if err := WriteUint64(b, uint64(len(f.BytesValue))); err != nil {
			return fmt.Errorf("could not write field length: %w", err)
		}
		if _, err := b.Write(f.BytesValue); err != nil {
			return fmt.Errorf("could not write field bytes: %w", err)
		}
	case FieldTypeMsg:
		if err := WriteUint64(b, f.UintValue); err != nil {
			return fmt.Errorf("could not write field value: %w", err)
		}
	default:
		return fmt.Errorf("unknown field type: %d", f.Type)
	}
	return nil
}

func EncodeLink(b *bytes.Buffer, link cid.Cid) error {
	_, err := link.WriteBytes(b)
	return err
}

func DecodeLink(b *bytes.Reader) (cid.Cid, error) {
	_, link, err := cid.CidFromReader(b)
	return link, err
}

func ParseDAGNode(b []byte) (*DAGNode, error) {
	// https://fuchsia.dev/fuchsia-src/reference/fidl/language/wire-format#envelopes
	log.Printf("ParseDAGNode: %x", b)
	bytesNum := order.Uint64(b[0:8])
	log.Printf("bytesNum: %d", bytesNum)
	linksNum := order.Uint64(b[8:16])
	log.Printf("linksNum: %d", linksNum)
	linkReader := bytes.NewReader(b[16+bytesNum:])
	links := make([]cid.Cid, linksNum)
	for i := 0; i < int(linksNum); i++ {
		link, err := DecodeLink(linkReader)
		if err != nil {
			return nil, fmt.Errorf("could not decode link #%d: %w", i, err)
		}
		links[i] = link
	}
	return &DAGNode{
		Bytes: b[16 : 16+bytesNum],
		Links: links,
	}, nil
}

func SerializeDAGNode(node *DAGNode) ([]byte, error) {
	b := &bytes.Buffer{}
	{
		x := make([]byte, 8)
		order.PutUint64(x, uint64(len(node.Bytes)))
		b.Write(x)
	}
	{
		x := make([]byte, 8)
		order.PutUint64(x, uint64(len(node.Links)))
		b.Write(x)
	}
	b.Write(node.Bytes)
	for i, link := range node.Links {
		if err := EncodeLink(b, link); err != nil {
			return nil, fmt.Errorf("could not encode link #%d: %w", i, err)
		}
	}
	bb := b.Bytes()
	log.Printf("SerializeDAGNode: %x", bb)
	return bb, nil
}

// Parse a selector of the form "1"
func ParseSelector(s string) (Selector, error) {
	index, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid selector: %w", err)
	}
	return Selector(index), nil
}

func PrintSelector(s Selector) string {
	return fmt.Sprintf("%d", s)
}

func ParsePath(s string) (Path, error) {
	selectors := []Selector{}
	for _, s := range strings.Split(s, "/") {
		if s == "" {
			continue
		}
		selector, err := ParseSelector(s)
		if err != nil {
			return nil, fmt.Errorf("invalid selector: %w", err)
		}
		selectors = append(selectors, selector)
	}
	return selectors, nil
}

func PrintPath(path Path) string {
	out := ""
	for _, s := range path {
		out += "/" + PrintSelector(s)
	}
	return out
}
