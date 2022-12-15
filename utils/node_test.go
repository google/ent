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
	"reflect"
	"testing"
)

func TestParseSelector(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want Selector
		err  bool
	}{
		{
			name: "empty",
			s:    "",
			want: Selector(0),
		},
		{
			name: "invalid",
			s:    "0[1",
			want: Selector(0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := ParseSelector(tt.s); !reflect.DeepEqual(got, tt.want) {
				if (err != nil) != tt.err {
					t.Errorf("ParseSelector() error = %v, wantErr %v", err, tt.err)
					return
				}
				t.Errorf("ParseSelector() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParsePath(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want Path
	}{
		{
			name: "empty",
			s:    "",
			want: []Selector{},
		},
		{
			name: "invalid",
			s:    "0[1",
			want: nil,
		},
		{
			name: "valid",
			s:    "0",
			want: []Selector{0},
		},
		{
			name: "valid",
			s:    "2",
			want: []Selector{2},
		},
		{
			name: "valid",
			s:    "0/1/2",
			want: []Selector{0, 1, 2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := ParsePath(tt.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParsePath() = %v, want %v", got, tt.want)
			}
		})
	}
}
