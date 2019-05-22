package tagname

import (
	"fmt"
	"strings"
)

type tChecker struct {
	tabMustHaveType  map[string]bool
	tabValid         map[string]bool
	tabValidType     map[string]bool
	tabInvalid       map[string]bool
	tabInvalidType   map[string]bool
	tabNonUniqueType map[string]bool
}

func initChecker(schema *TSchema) {
	if schema == nil {
		panic("settings is nil")
	}
	o := &tChecker{}
	o.tabMustHaveType = initTable(schema.MustHaveByType)
	o.tabNonUniqueType = initTable(schema.NonUniqueByType)
	o.tabValid = initTable(schema.Valid)
	o.tabValidType = initTable(schema.ValidByType)
	o.tabInvalid = initTable(schema.Invalid)
	o.tabInvalidType = initTable(schema.InvalidByType)
	schema.checker = o
}

func initTable(list []string) map[string]bool {
	o := map[string]bool{}
	for _, v := range list {
		o[v] = false
	}
	return o
}

func isExist(table map[string]bool, typ string) bool {
	if len(table) == 0 {
		return true
	}
	_, ok := table[typ]
	return ok
}

func isNotExist(table map[string]bool, typ string) bool {
	_, ok := table[typ]
	return ok
}

func checkTags(tags *TTags, isStrictCheck bool) error {
	err := tags.State()
	if err != nil {
		return err
	}
	schema := tags.schema
	o := schema.checker
	for _, typ := range schema.MustHaveByType {
		// fmt.Println("###")
		_, ok := tags.byType[typ]
		if !ok {
			return fmt.Errorf("%q type does not exist", typ)
		}
	}
	errors := []string{}
	for typ, list := range tags.byType {
		if strings.HasPrefix(typ, "ERR_") {
			errors = append(errors, fmt.Sprintf("%v: %v", strings.TrimPrefix(typ, "ERR_"), strings.Join(list, ", ")))
			continue
		}
		if typ == "UNKNOWN_TAG" {
			if isStrictCheck {
				return fmt.Errorf("UNKNOWN type tag(s) are present: %v", list)
			}
			continue
		}
		if typ == "INVALID_TAG" {
			return fmt.Errorf("invalid tag(s) are present: %v", list)
		}
		if _, ok := o.tabNonUniqueType[typ]; !ok {
			if len(list) > 1 {
				return fmt.Errorf("%q type must be unique", typ)
			}
		}
		if !isExist(o.tabValidType, typ) {
			return fmt.Errorf("%q is not a valid type", typ)
		}
		if isNotExist(o.tabInvalidType, typ) {
			return fmt.Errorf("%q is an invalid type", typ)
		}
		for _, val := range list {
			if !isExist(o.tabValid, val) {
				return fmt.Errorf("tag (%q,%q) has not a valid value", typ, val)
			}
			if isNotExist(o.tabInvalid, val) {
				return fmt.Errorf("tag (%q,%q) has an invalid value", typ, val)
			}
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("some error(s):\n        %v", strings.Join(errors, "\n        "))
	}

	tags.RemoveTags("m4otag")

	t, err := tags.FilterTag("type")
	if err != nil {
		return err
	}
	if t != "film" && t != "trailer" {
		return nil
	}
	sdhd, err := tags.FilterTag("sdhd")
	if err != nil {
		return err
	}
	atag, err := tags.FilterTag("atag")
	if err != nil {
		return nil
	}

	switch {
	case atag == "ar2" && (sdhd == "sd" || t == "trailer"):
		tags.RemoveTags("atag")
	case atag == "ar6" && (sdhd == "hd" || sdhd == "3d") && t == "film":
		tags.RemoveTags("atag")
	}

	return nil
}
