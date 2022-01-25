package utils

import (
	"reflect"
	"testing"
)

func TestParseSelector(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want *Selector
	}{
		{
			name: "empty",
			s:    "",
			want: nil,
		},
		{
			name: "invalid",
			s:    "0[1",
			want: nil,
		},
		{
			name: "valid",
			s:    "0[1]",
			want: &Selector{
				FieldID: 0,
				Index:   1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := ParseSelector(tt.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseSelector() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParsePath(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want []Selector
	}{
		{
			name: "empty",
			s:    "",
			want: nil,
		},
		{
			name: "invalid",
			s:    "0[1",
			want: nil,
		},
		{
			name: "valid",
			s:    "0[1]/1[2]",
			want: []Selector{
				{
					FieldID: 0,
					Index:   1,
				},
				{
					FieldID: 1,
					Index:   2,
				},
			},
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
