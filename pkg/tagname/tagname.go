package tagname

import (
	"fmt"
	"path/filepath"
	"strings"
)

// TTagname -
type TTagname struct {
	dir    string
	src    string
	schema string
	tags   *TTags
}

// NewFromString -
func NewFromString(str string, checkLevel int, schemaNames ...string) (*TTagname, error) {
	var err error
	var tags *TTags
	var schema string
	var errors []string

	schemas := schemaNames
	if len(schemas) == 0 {
		schemas = []string{"rt", "old"}
	}

	for _, schema = range schemas {
		tags, err = Parse(str, schema)
		if err == nil {
			break
		}
		errors = append(errors, fmt.Sprintf("%-16q: %v", schema, err))
	}

	if err != nil {
		s := strings.Join(errors, "\n  ")
		return nil, fmt.Errorf("multiple ones:\n  %v", s)
	}

	tn := &TTagname{schema: schema, src: str, tags: tags}
	err = tn.Check(checkLevel)
	if err != nil {
		return nil, err
	}
	return tn, nil
}

// NewFromFilename -
func NewFromFilename(path string, checkLevel int, schemaNames ...string) (*TTagname, error) {
	src := filepath.Base(path)
	ret, err := NewFromString(src, checkLevel, schemaNames...)
	if err != nil {
		return nil, err
	}
	ret.dir = filepath.Dir(path)
	return ret, nil
}

// State -
func (o *TTagname) State() error {
	if o == nil {
		return fmt.Errorf("tagname object is nil")
	}
	return o.tags.State()
}

// ConvertTo -
func (o *TTagname) ConvertTo(schemaName string) (string, error) {
	if schemaName == "" {
		schemaName = o.schema
	}
	_, err := Schema(schemaName)
	if err != nil {
		return "", err
	}
	ret, err := ToString(o.tags, o.schema, schemaName)
	if err != nil {
		return "", err
	}
	ret = filepath.Join(o.dir, ret)
	return ret, nil
}

// Check -
func (o *TTagname) Check(checkLevel int) error {
	if err := o.State(); err != nil {
		return err
	}

	if checkLevel < 0 {
		return nil
	}

	isStrictCheck := checkLevel&CheckStrict != 0
	isDeepCheck := checkLevel&CheckDeep != 0
	err := o.tags.Check(isStrictCheck)
	if err != nil || !isDeepCheck {
		return err
	}
	err = checkDeep(o)
	return err
}

// GetTag -
func (o *TTagname) GetTag(typ string) (string, error) {
	list := o.tags.GetTags(typ)
	if len(list) == 0 {
		return "", fmt.Errorf("%q has no tags of %q type", o.src, typ)
	}
	if len(list) > 1 {
		return "", fmt.Errorf("GetTag() cannot return multiple tags of %q type in %q", typ, o.src)
	}
	schema, err := Schema(o.schema)
	if err != nil {
		return "", err
	}
	val := list[0]
	typ, val, err = schema.ReadFilter(typ, val)
	if err != nil {
		return "", err
	}
	return val, nil
}

// GetTags -
func (o *TTagname) GetTags(typ string) []string {
	list := o.tags.GetTags(typ)
	// if len(list) == 0 {
	// 	return nil, fmt.Errorf("%q has no tags of %q type", o.src, typ)
	// }
	schema, err := Schema(o.schema)
	if err != nil {
		fmt.Println("Schema() error at tagname.GetTag")
		panic(err)
		// return "", err
	}
	ret := []string{}
	for _, s := range list {
		_, val, err := schema.ReadFilter(typ, s)
		if err != nil {
			fmt.Println("ReadFilter() error at tagname.GetTag")
			panic(err)
			// return "", err
		}
		ret = append(ret, val)
	}
	return ret
}

// RemoveTags -
func (o *TTagname) RemoveTags(typ string) {
	o.tags.RemoveTags(typ)
}

// SetTag -
func (o *TTagname) SetTag(typ string, val string) {
	o.tags.RemoveTags(typ)
	o.tags.AddTag(typ, val)
}

// Schema -
func (o *TTagname) Schema() string {
	return o.schema
}
