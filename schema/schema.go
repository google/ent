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
	"context"
	"encoding/binary"
	"fmt"
	"reflect"
	"strconv"

	"github.com/google/ent/log"
	"github.com/google/ent/nodeservice"
	"github.com/google/ent/utils"
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

func ResolveLink(o nodeservice.ObjectGetter, base utils.Hash, path []utils.Selector) (utils.Hash, error) {
	if len(path) == 0 {
		return base, nil
	} else {
		object, err := o.Get(context.Background(), base)
		if err != nil {
			return "", fmt.Errorf("failed to get object: %v", err)
		}
		node, err := utils.ParseNode(object)
		if err != nil {
			return "", fmt.Errorf("failed to parse object: %v", err)
		}
		selector := path[0]
		newBase := node.Links[selector.FieldID][selector.Index]
		return ResolveLink(o, newBase.Hash, path[1:])
	}
}

func GetStruct(o nodeservice.ObjectGetter, digest utils.Hash, v interface{}) error {
	log.Debugf(context.Background(), "getting struct %s", digest)
	object, err := o.Get(context.Background(), digest)
	if err != nil {
		return fmt.Errorf("failed to get struct object: %v", err)
	}
	node, err := utils.ParseNode(object)
	if err != nil {
		return fmt.Errorf("failed to parse struct object: %v", err)
	}
	log.Debugf(context.Background(), "parsed node: %+v", node)
	rv := reflect.ValueOf(v).Elem()
	for i := 0; i < rv.NumField(); i++ {
		typeField := rv.Type().Field(i)
		tag := typeField.Tag.Get("ent")
		fieldID, err := strconv.Atoi(tag)
		if err != nil {
			return fmt.Errorf("failed to parse field id: %v", err)
		}
		links := node.Links[uint(fieldID)]
		fieldValue := rv.Field(i)
		switch typeField.Type.Kind() {
		case reflect.Uint32:
			if len(links) != 1 {
				return fmt.Errorf("expected 1 link for field %q with field id %d, got %d", typeField.Name, fieldID, len(links))
			}
			f, err := GetUint32(o, links[0].Hash)
			if err != nil {
				return fmt.Errorf("failed to get uint32 field: %v", err)
			}
			fieldValue.SetUint(uint64(f))
		case reflect.String:
			if len(links) != 1 {
				return fmt.Errorf("expected 1 link for field %q with field id %d, got %d", typeField.Name, fieldID, len(links))
			}
			f, err := GetString(o, links[0].Hash)
			if err != nil {
				return fmt.Errorf("failed to get uint32 field: %v", err)
			}
			fieldValue.SetString(f)
		case reflect.Struct:
			if len(links) != 1 {
				return fmt.Errorf("expected 1 link for field %q with field id %d, got %d", typeField.Name, fieldID, len(links))
			}
			err = GetStruct(o, links[0].Hash, fieldValue.Addr().Interface())
			if err != nil {
				return fmt.Errorf("failed to get struct field: %v", err)
			}
		case reflect.Slice:
			switch typeField.Type.Elem().Kind() {
			case reflect.Uint32:
				for _, link := range links {
					v, err := GetUint32(o, link.Hash)
					if err != nil {
						return fmt.Errorf("failed to get uint32 field: %v", err)
					}
					fieldValue.Set(reflect.Append(fieldValue, reflect.ValueOf(v)))
				}
			case reflect.String:
				for _, link := range links {
					v, err := GetString(o, link.Hash)
					if err != nil {
						return fmt.Errorf("failed to get string field: %v", err)
					}
					fieldValue.Set(reflect.Append(fieldValue, reflect.ValueOf(v)))
				}
			case reflect.Struct:
				for _, link := range links {
					v := reflect.New(typeField.Type.Elem())
					err := GetStruct(o, link.Hash, v.Interface())
					if err != nil {
						return fmt.Errorf("failed to get struct field: %v", err)
					}
					fieldValue.Set(reflect.Append(fieldValue, v.Elem()))
				}
			default:
				return fmt.Errorf("unsupported field type: %v", typeField.Type.Elem().Kind())
			}
		default:
			return fmt.Errorf("unsupported field type: %v", typeField.Type.Kind())
		}
	}
	return nil
}

func PutStruct(o nodeservice.ObjectStore, v interface{}) (utils.Hash, error) {
	node := utils.Node{
		Links: make(map[uint][]utils.Link),
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	for i := 0; i < rv.NumField(); i++ {
		typeField := rv.Type().Field(i)
		tag := typeField.Tag.Get("ent")
		fieldID, err := strconv.Atoi(tag)
		if err != nil {
			return "", fmt.Errorf("failed to parse field id: %v", err)
		}
		fieldValue := rv.Field(i)
		switch typeField.Type.Kind() {
		case reflect.Uint32:
			h, err := PutUint32(o, uint32(fieldValue.Uint()))
			if err != nil {
				return "", fmt.Errorf("failed to put uint32 field: %v", err)
			}
			node.Links[uint(fieldID)] = append(node.Links[uint(fieldID)], utils.Link{
				Hash: h,
			})
		case reflect.String:
			h, err := PutString(o, fieldValue.String())
			if err != nil {
				return "", fmt.Errorf("failed to put string field: %v", err)
			}
			node.Links[uint(fieldID)] = append(node.Links[uint(fieldID)], utils.Link{
				Hash: h,
			})
		case reflect.Struct:
			h, err := PutStruct(o, fieldValue.Interface())
			if err != nil {
				return "", fmt.Errorf("failed to put struct field: %v", err)
			}
			node.Links[uint(fieldID)] = append(node.Links[uint(fieldID)], utils.Link{
				Hash: h,
			})
		case reflect.Slice:
			switch typeField.Type.Elem().Kind() {
			case reflect.Uint32:
				for i := 0; i < fieldValue.Len(); i++ {
					iv := fieldValue.Index(i).Uint()
					h, err := PutUint32(o, uint32(iv))
					if err != nil {
						return "", fmt.Errorf("failed to put uint32 field: %v", err)
					}
					node.Links[uint(fieldID)] = append(node.Links[uint(fieldID)], utils.Link{
						Hash: h,
					})
				}
			case reflect.String:
				for i := 0; i < fieldValue.Len(); i++ {
					iv := fieldValue.Index(i).String()
					h, err := PutString(o, iv)
					if err != nil {
						return "", fmt.Errorf("failed to put string field: %v", err)
					}
					node.Links[uint(fieldID)] = append(node.Links[uint(fieldID)], utils.Link{
						Hash: h,
					})
				}
			case reflect.Struct:
				for i := 0; i < fieldValue.Len(); i++ {
					iv := fieldValue.Index(i).Interface()
					h, err := PutStruct(o, iv)
					if err != nil {
						return "", fmt.Errorf("failed to put string field: %v", err)
					}
					node.Links[uint(fieldID)] = append(node.Links[uint(fieldID)], utils.Link{
						Hash: h,
					})
				}
			default:
				return "", fmt.Errorf("unsupported field type: %v", typeField.Type.Elem().Kind())
			}
		default:
			return "", fmt.Errorf("unsupported field type: %v", typeField.Type.Kind())
		}
	}
	b, err := utils.SerializeNode(&node)
	if err != nil {
		return "", fmt.Errorf("failed to serialize node: %v", err)
	}
	return o.Put(context.Background(), b)
}

func GetUint32(o nodeservice.ObjectGetter, digest utils.Hash) (uint32, error) {
	b, err := o.Get(context.Background(), digest)
	if err != nil {
		return 0, fmt.Errorf("failed to get struct object: %v", err)
	}
	v := binary.BigEndian.Uint32(b)
	return v, nil
}

func PutUint32(o nodeservice.ObjectStore, v uint32) (utils.Hash, error) {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, v)
	return o.Put(context.Background(), b)
}

func GetString(o nodeservice.ObjectGetter, digest utils.Hash) (string, error) {
	b, err := o.Get(context.Background(), digest)
	if err != nil {
		return "", fmt.Errorf("failed to get struct object: %v", err)
	}
	v := string(b)
	return v, nil
}

func PutString(o nodeservice.ObjectStore, v string) (utils.Hash, error) {
	b := []byte(v)
	return o.Put(context.Background(), b)
}
