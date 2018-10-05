package tagname

import (
	"fmt"
	"sort"

	"github.com/macroblock/imed/pkg/ptool"
)

// TTags -
type TTags struct {
	byType map[string][]string
	// settings *TSettings
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
	RegisterSchema("old.normal", oldNormalSchema)
	RegisterSchema("rt.normal", rtNormalSchema)
}

// NewTagname -
func NewTagname(tree *ptool.TNode, parser *ptool.TParser) (*TTags, error) {
	if parser == nil {
		return nil, fmt.Errorf("parser is null")
	}
	tagname := &TTags{}
	for _, node := range tree.Links {
		val := node.Value
		typ := parser.ByID(node.Type)

		// err := settings.ReadFilter(&typ, &val)
		// if err != nil {
		// 	return nil, err
		// }

		tagname.AddTag(typ, val)
	}
	return tagname, nil
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

	tagname, err := NewTagname(tree, parser)
	if err != nil {
		return nil, err
	}

	err = check(tagname, schema)
	if err != nil {
		return nil, err
	}

	// tagname.settings = checker.settings
	for _, list := range tagname.byType {
		sort.Strings(list)
	}
	return tagname, nil
}

// Check -
func Check(tagname *TTags, schemaName string) error {
	if tagname == nil {
		return fmt.Errorf("tagname is nil")
	}
	schema, err := Schema(schemaName)
	if err != nil {
		return err
	}
	err = check(tagname, schema)
	if err != nil {
		return err
	}
	return nil
}

// ToString -
func ToString(tagname *TTags, fromSchemaName, toSchemaName string) (string, error) {
	if tagname == nil {
		return "", fmt.Errorf("tagname is nil")
	}
	fromSchema, err := Schema(fromSchemaName)
	if err != nil {
		return "", err
	}
	toSchema, err := Schema(toSchemaName)
	if err != nil {
		return "", err
	}
	s, err := toString(tagname, fromSchema, toSchema)
	return s, err
}
