package utils

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Node struct {
	Kind  string
	Links map[uint][]Link
}

type Link struct {
	Type uint
	Hash string
}

type Selector struct {
	FieldID uint
	Index   uint
}

func ParseNode(b []byte) (*Node, error) {
	node := Node{}
	err := json.Unmarshal(b, &node)
	if err != nil {
		return nil, fmt.Errorf("invalid node: %w", err)
	}
	return &node, nil
}

// Parse a selector of the form "0[1]"
func ParseSelector(s string) (*Selector, error) {
	var fieldID, index uint
	_, err := fmt.Sscanf(s, "%d[%d]", &fieldID, &index)
	if err != nil {
		return nil, fmt.Errorf("invalid selector: %w", err)
	}
	return &Selector{
		FieldID: fieldID,
		Index:   index,
	}, nil
}

func PrintSelector(s Selector) string {
	return fmt.Sprintf("%d[%d]", s.FieldID, s.Index)
}

func ParsePath(s string) ([]Selector, error) {
	selectors := []Selector{}
	for _, s := range strings.Split(s, "/") {
		if s == "" {
			continue
		}
		selector, err := ParseSelector(s)
		if err != nil {
			return nil, fmt.Errorf("invalid selector: %w", err)
		}
		selectors = append(selectors, *selector)
	}
	return selectors, nil
}

func PrintPath(selectors []Selector) string {
	path := ""
	for _, s := range selectors {
		path += "/" + PrintSelector(s)
	}
	return path
}
