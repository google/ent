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

package schema

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"strconv"

	"github.com/google/ent/log"
	"github.com/google/ent/nodeservice"
	"github.com/google/ent/utils"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
)

type Schema struct {
	// #0 is the root node kind.
	Kinds []Kind `ent:"0"`
}

type Kind struct {
	KindID uint32  `ent:"0"`
	Name   string  `ent:"1"`
	Fields []Field `ent:"2"`
}

type Field struct {
	FieldID uint32 `ent:"0"`
	Name    string `ent:"1"`
	KindID  uint32 `ent:"2"`
	Raw     uint32 `ent:"3"`
}

func ResolveLink(o nodeservice.ObjectGetter, base utils.Digest, path []utils.Selector) (utils.Digest, error) {
	if len(path) == 0 {
		return base, nil
	} else {
		object, err := o.Get(context.Background(), base)
		if err != nil {
			return utils.Digest{}, fmt.Errorf("failed to get object: %v", err)
		}
		node, err := utils.ParseDAGNode(object)
		if err != nil {
			return utils.Digest{}, fmt.Errorf("failed to parse object: %v", err)
		}
		selector := path[0]
		newBase := node.Links[selector]
		return ResolveLink(o, utils.Digest(newBase.Hash()), path[1:])
	}
}

func GetFieldWithId(v interface{}, fieldID uint64) (reflect.Value, error) {
	rv := reflect.ValueOf(v).Elem()
	for i := 0; i < rv.NumField(); i++ {
		typeField := rv.Type().Field(i)
		tag := typeField.Tag.Get("ent")
		id, err := strconv.Atoi(tag)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("failed to parse field id: %v", err)
		}
		if uint64(id) == fieldID {
			return rv.Field(i), nil
		}
	}
	return reflect.Value{}, fmt.Errorf("failed to find field with id: %v", fieldID)
}

func GetStruct(o nodeservice.ObjectGetter, digest utils.Digest, v interface{}) error {
	log.Debugf(context.Background(), "getting struct %s", digest.String())
	object, err := o.Get(context.Background(), digest)
	if err != nil {
		return fmt.Errorf("failed to get struct object: %v", err)
	}
	node, err := utils.ParseDAGNode(object)
	if err != nil {
		return fmt.Errorf("failed to parse struct object: %v", err)
	}
	log.Debugf(context.Background(), "parsed node: %+v", node)
	r := bytes.NewReader(node.Bytes)
	linkIndex := 0
	for {
		field, err := utils.DecodeField(r)
		if err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("failed to decode field: %v", err)
		}
		fieldValue, err := GetFieldWithId(v, field.ID)
		if err != nil {
			return fmt.Errorf("failed to get field with id: %v", err)
		}
		switch field.Type {
		case utils.FieldTypeInt:
			switch fieldValue.Kind() {
			case reflect.Slice:
				// Append to existing slice.
				fieldValue.Set(reflect.Append(fieldValue, reflect.ValueOf(field.UintValue).Convert(fieldValue.Type().Elem())))
			case reflect.Int32, reflect.Int64, reflect.Int:
				fieldValue.SetInt(int64(field.UintValue))
			case reflect.Uint32, reflect.Uint64, reflect.Uint:
				fieldValue.SetUint(uint64(field.UintValue))
			default:
				return fmt.Errorf("unexpected field type for int: %v", fieldValue.Kind())
			}
		case utils.FieldTypeBytes:
			switch fieldValue.Kind() {
			case reflect.Slice:
				switch fieldValue.Type().Elem().Kind() {
				case reflect.Uint8: // []byte
					fieldValue.SetBytes(field.BytesValue)
				case reflect.String: // []string
					fieldValue.Set(reflect.Append(fieldValue, reflect.ValueOf(string(field.BytesValue))))
				default:
					return fmt.Errorf("unexpected field type for bytes: %v", fieldValue.Type().Elem().Kind())
				}
			case reflect.String:
				fieldValue.SetString(string(field.BytesValue))
			default:
				return fmt.Errorf("unsupported field type: %v", fieldValue.Kind())
			}
		case utils.FieldTypeMsg:
			switch fieldValue.Kind() {
			case reflect.Slice:
				if field.UintValue != 1 {
					return fmt.Errorf("no presence bit for repeated field: %v", field.UintValue)
				}
				link := node.Links[linkIndex]
				linkIndex++
				v := reflect.New(fieldValue.Type().Elem())
				if err := GetStruct(o, utils.Digest(link.Hash()), v.Interface()); err != nil {
					return fmt.Errorf("failed to get struct: %v", err)
				}
				fieldValue.Set(reflect.Append(fieldValue, v.Elem()))
			case reflect.Ptr:
			case reflect.Struct:
				if field.UintValue != 1 {
					return fmt.Errorf("no presence bit for repeated field: %v", field.UintValue)
				}
				link := node.Links[linkIndex]
				linkIndex++
				if err := GetStruct(o, utils.Digest(link.Hash()), fieldValue.Addr().Interface()); err != nil {
					return fmt.Errorf("failed to get struct: %v", err)
				}
			}
		default:
			return fmt.Errorf("unsupported field type: %v", field.Type)
		}
	}
	return nil
}

func PutStruct(o nodeservice.ObjectStore, v interface{}) (utils.Digest, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	bytes := &bytes.Buffer{}
	links := []cid.Cid{}
	// TODO: Traverse in order of field id.
	for i := 0; i < rv.NumField(); i++ {
		typeField := rv.Type().Field(i)
		tag := typeField.Tag.Get("ent")
		fieldID, err := strconv.Atoi(tag)
		if err != nil {
			return utils.Digest{}, fmt.Errorf("failed to parse field id: %v", err)
		}
		fieldValue := rv.Field(i)
		switch typeField.Type.Kind() {
		case reflect.Uint32, reflect.Uint64, reflect.Uint:
			log.Infof(nil, "putting int: %+v", fieldValue)
			utils.EncodeField(bytes, &utils.Field{
				ID:         uint64(fieldID),
				Type:       utils.FieldTypeInt,
				UintValue:  fieldValue.Uint(),
				BytesValue: nil,
			})
		case reflect.String:
			log.Infof(nil, "putting string: %+v", fieldValue)
			utils.EncodeField(bytes, &utils.Field{
				ID:         uint64(fieldID),
				Type:       utils.FieldTypeBytes,
				UintValue:  0,
				BytesValue: []byte(fieldValue.String()),
			})
		case reflect.Struct:
			digest, err := PutStruct(o, fieldValue.Interface())
			if err != nil {
				return utils.Digest{}, fmt.Errorf("failed to put struct field: %v", err)
			}
			utils.EncodeField(bytes, &utils.Field{
				ID:         uint64(fieldID),
				Type:       utils.FieldTypeMsg,
				UintValue:  1, // Present
				BytesValue: nil,
			})
			links = append(links, cid.NewCidV1(utils.TypeDAG, multihash.Multihash(digest)))
		case reflect.Slice:
			switch typeField.Type.Elem().Kind() {
			case reflect.Uint32, reflect.Uint64, reflect.Uint:
				for i := 0; i < fieldValue.Len(); i++ {
					iv := fieldValue.Index(i).Uint()
					utils.EncodeField(bytes, &utils.Field{
						ID:         uint64(fieldID),
						Type:       utils.FieldTypeInt,
						UintValue:  iv,
						BytesValue: nil,
					})
				}
			case reflect.String:
				for i := 0; i < fieldValue.Len(); i++ {
					iv := fieldValue.Index(i).String()
					utils.EncodeField(bytes, &utils.Field{
						ID:         uint64(fieldID),
						Type:       utils.FieldTypeBytes,
						UintValue:  0,
						BytesValue: []byte(iv),
					})
				}
			case reflect.Struct:
				for i := 0; i < fieldValue.Len(); i++ {
					iv := fieldValue.Index(i).Interface()
					digest, err := PutStruct(o, iv)
					if err != nil {
						return utils.Digest{}, fmt.Errorf("failed to put string field: %v", err)
					}
					utils.EncodeField(bytes, &utils.Field{
						ID:         uint64(fieldID),
						Type:       utils.FieldTypeMsg,
						UintValue:  1, // Present
						BytesValue: nil,
					})
					links = append(links, cid.NewCidV1(utils.TypeDAG, multihash.Multihash(digest)))
				}
			default:
				return utils.Digest{}, fmt.Errorf("unsupported field type: %v", typeField.Type.Elem().Kind())
			}
		default:
			return utils.Digest{}, fmt.Errorf("unsupported field type: %v", typeField.Type.Kind())
		}
	}
	node := utils.DAGNode{
		Bytes: bytes.Bytes(),
		Links: links,
	}
	b, err := utils.SerializeDAGNode(&node)
	if err != nil {
		return utils.Digest{}, fmt.Errorf("failed to serialize node: %v", err)
	}
	digest, err := o.Put(context.Background(), b)
	log.Infof(nil, "putting node: %+v -> %s", node, digest)
	return digest, err
}

func GetUint32(o nodeservice.ObjectGetter, digest utils.Digest) (uint32, error) {
	b, err := o.Get(context.Background(), digest)
	if err != nil {
		return 0, fmt.Errorf("failed to get struct object: %v", err)
	}
	v := binary.BigEndian.Uint32(b)
	return v, nil
}

func PutUint32(o nodeservice.ObjectStore, v uint32) (utils.Digest, error) {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, v)
	return o.Put(context.Background(), b)
}

func GetString(o nodeservice.ObjectGetter, digest utils.Digest) (string, error) {
	b, err := o.Get(context.Background(), digest)
	if err != nil {
		return "", fmt.Errorf("failed to get struct object: %v", err)
	}
	v := string(b)
	return v, nil
}

func PutString(o nodeservice.ObjectStore, v string) (utils.Digest, error) {
	b := []byte(v)
	return o.Put(context.Background(), b)
}
