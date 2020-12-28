package tagname

import (
	"fmt"
	"sort"
	"strings"

	"github.com/macroblock/imed/pkg/ptool"
)

// -
const (
	CheckNone = iota - 1
	CheckNormal
	CheckStrict
	CheckDeep

	CheckDeepNormal = CheckDeep | CheckNormal
	CheckDeepStrict = CheckDeep | CheckStrict
)

// TTags -
type TTags struct {
	byType map[string][]string
}

var (
	oldParser *ptool.TParser
	rtParser  *ptool.TParser
)

func init() {
	p, err := ptool.NewBuilder().FromString(oldForm).Entries("entry").Build()
	if err != nil {
		fmt.Println("\n[old form] parser error: ", err)
		panic("")
	}
	oldParser = p
	p, err = ptool.NewBuilder().FromString(rtForm).Entries("entry").Build()
	if err != nil {
		fmt.Println("\n[RT form] parser error: ", err)
		panic("")
	}
	rtParser = p

	globSchemas = map[string]*TSchema{}
	RegisterSchema("old", oldNormalSchema)
	RegisterSchema("rt", rtNormalSchema)
}

// NewTags -
func NewTags(tree *ptool.TNode, parser *ptool.TParser, schema *TSchema) (*TTags, error) {
	if parser == nil {
		return nil, fmt.Errorf("NewTagname() parser is null")
	}
	tags := &TTags{}
	for _, node := range tree.Links {
		val := node.Value
		typ := parser.ByID(node.Type)

		tags.AddTag(typ, val)
	}
	return tags, nil
}

// AddTag -
func (o *TTags) AddTag(typ, val string) {
	if o.byType == nil {
		o.byType = map[string][]string{}
	}
	list, ok := o.byType[typ]
	if !ok {
		list = []string{val}
	} else {
		list = append(list, val)
	}
	o.byType[typ] = list
}

// GetTags -
func (o *TTags) GetTags(typ string) []string {
	list, ok := o.byType[typ]
	if !ok {
		return nil
	}
	return list
}

// RemoveTags -
func (o *TTags) RemoveTags(typ string) {
	delete(o.byType, typ)
}

// GetTag -
func (o *TTags) GetTag(typ string) (string, error) {
	list := o.GetTags(typ)
	if len(list) == 0 {
		return "", fmt.Errorf("must have tag of %q type", typ)
	}
	if len(list) > 1 {
		return "", fmt.Errorf("must have only one tag of %q type", typ)
	}
	return list[0], nil
}

// MustHave -
func (o *TTags) MustHave(args ...string) error {
	list := []string{}
	for _, arg := range args {
		if _, ok := o.byType[arg]; ok {
			continue
		}
		list = append(list, arg)
	}
	if len(list) == 0 {
		return nil
	}
	return fmt.Errorf("must have tags: %v", strings.Join(list, ", "))
}

// MustNotHave -
func (o *TTags) MustNotHave(args ...string) error {
	list := []string{}
	for _, arg := range args {
		if _, ok := o.byType[arg]; !ok {
			continue
		}
		list = append(list, arg)
	}
	if len(list) == 0 {
		return nil
	}
	return fmt.Errorf("must not have tags: %v", strings.Join(list, ", "))
}

// Parse -
func Parse(s string, schemaName string) (*TTags, error) {
	schema, err := Schema(schemaName)
	if err != nil {
		return nil, err
	}

	parser := *schema.parser
	tree, err := parser.Parse(s)
	if err != nil {
		return nil, err
	}

	tags, err := NewTags(tree, parser, schema)
	if err != nil {
		return nil, err
	}

	// tagname.settings = checker.settings
	for _, list := range tags.byType {
		sort.Strings(list)
	}
	return tags, nil
}

// State -
func (o *TTags) State() error {
	if o == nil {
		return fmt.Errorf("TTags object is nil")
	}
	return nil
}

// ToString -
func ToString(tags *TTags, schemaName string) (string, error) {
	if tags == nil {
		return "", fmt.Errorf("tagname is nil")
	}
	schema, err := Schema(schemaName)
	if err != nil {
		return "", err
	}
	s, err := toString(tags, schema)
	return s, err
}
