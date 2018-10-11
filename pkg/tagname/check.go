package tagname

import (
	"fmt"
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

func check(tagname *TTags, schema *TSchema) error {
	o := schema.checker
	for _, typ := range schema.MustHaveByType {
		_, ok := tagname.byType[typ]
		if !ok {
			return fmt.Errorf("%q type does not exist", typ)
		}
	}
	// for _, typ := range schema.NonUniqueByType {
	// 	_, ok := o.tabNonUniqueType[typ]
	// 	fmt.Println("#####", typ)
	// 	if !ok {
	// 		if len(tagname.byType[typ]) > 1 {
	// 			fmt.Println("#####")
	// 			return fmt.Errorf("%q type must be unique", typ)
	// 		}
	// 	}
	// }
	for typ, list := range tagname.byType {
		if typ == "INVALID_TAGS" {
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
	return nil
}
