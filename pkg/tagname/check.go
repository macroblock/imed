package tagname

import (
	"fmt"
	"strings"
)

type tCheckContext struct{
	ListMustHaveTypes []string
	TabNonUniqueTypes map[string]uint8
	TabInvalidTypes   map[string]uint8
	TabInvalidValues  map[string]uint8
	TabValidTypes     map[string]uint8
}

var (
	defaultCheckContext = &tCheckContext{
		ListMustHaveTypes: []string{"type", "name", "year"},
		TabNonUniqueTypes: map[string]uint8{"mtag": 0},
		TabInvalidTypes:   map[string]uint8{"UNKNOWN_TAG": 0},
		TabInvalidValues:  map[string]uint8{},
		TabValidTypes:     map[string]uint8{},
	}
	checkContext = defaultCheckContext
)

var (
	checkContextForFilms = &tCheckContext{
		ListMustHaveTypes: []string{"sdhd"},
		TabNonUniqueTypes: map[string]uint8{},
		TabInvalidTypes:   map[string]uint8{},
		TabInvalidValues:  map[string]uint8{},
		TabValidTypes:     map[string]uint8{},
	}
	checkContextForPosters = &tCheckContext{
		ListMustHaveTypes: []string{"sdhd", "sizetag"},
		TabNonUniqueTypes: map[string]uint8{},
		TabInvalidTypes:   map[string]uint8{},
		TabInvalidValues:  map[string]uint8{},
		TabValidTypes:     map[string]uint8{},
	}
	checkContextForGpPosters = &tCheckContext{
		ListMustHaveTypes: []string{"sizetag"},
		TabNonUniqueTypes: map[string]uint8{},
		TabInvalidTypes:   map[string]uint8{},
		TabInvalidValues:  map[string]uint8{},
		TabValidTypes:     map[string]uint8{},
	}
)

func updateCheckContext(out *tCheckContext, in1, in2 *tCheckContext) *tCheckContext {
	if out != nil {
		return out
	}
	return &tCheckContext{
		ListMustHaveTypes: updateList(nil, in1.ListMustHaveTypes, in2.ListMustHaveTypes),
		TabNonUniqueTypes: updateTable(nil, in1.TabNonUniqueTypes, in2.TabNonUniqueTypes),
		TabInvalidTypes:   updateTable(nil, in1.TabInvalidTypes, in2.TabInvalidTypes),
		TabInvalidValues:  updateTable(nil, in1.TabInvalidValues, in2.TabInvalidValues),
		TabValidTypes:     updateTable(nil, in1.TabValidTypes, in2.TabValidTypes),
	}
}

var filmsCheckContext *tCheckContext
func getFilmsCC() *tCheckContext {
	if filmsCheckContext != nil {
		return filmsCheckContext
	}
	filmsCheckContext = updateCheckContext(nil, defaultCheckContext, checkContextForFilms)
	return filmsCheckContext
}

var postersCheckContext *tCheckContext
func getPostersCC() *tCheckContext {
	if postersCheckContext != nil {
		return postersCheckContext
	}
	postersCheckContext = updateCheckContext(nil, defaultCheckContext, checkContextForPosters)
	return postersCheckContext
}

var gpPostersCheckContext *tCheckContext
func getGpPostersCC() *tCheckContext {
	if gpPostersCheckContext != nil {
		return gpPostersCheckContext
	}
	gpPostersCheckContext = updateCheckContext(nil, defaultCheckContext, checkContextForGpPosters)
	return gpPostersCheckContext
}

func updateList(out, in1, in2 []string) []string{
	if out != nil {
		return out
	}
	for _, s := range in1 {
		out = append(out, s)
	}
	for _, s := range in2 {
		out = append(out, s)
	}
	return out
}

func updateTable(out , in1, in2 map[string]uint8) map[string]uint8 {
	if out != nil {
		return out
	}
	out = map[string]uint8{}
	for key := range in1 {
		out[key] = 0
	}
	for key := range in2 {
		out[key] = 0
	}
	return out
}

func isExist(table map[string]uint8, typ string) bool {
	_, ok := table[typ]
	return ok
}

func isNotExist(table map[string]uint8, typ string) bool {
	if len(table) == 0 {
		return  false 
	}
	_, ok := table[typ]
	return !ok
}

func CheckTags(tags *TTags, isStrictCheck bool) error {

	err := tags.State()
	if err != nil {
		return err
	}

	typ, err := tags.GetTag("type")
	if err != nil {
		return err
	}

	cc := defaultCheckContext
	switch typ {
	case "film", "trailer":
		cc = getFilmsCC()
	case "poster", "poster.logo":
		cc = getPostersCC()
	case "poster.gp":
		cc = getGpPostersCC()
	default:
		return fmt.Errorf("check: unsupported tag 'type': %q", typ)
	}

	// o := schema.checker
	for _, typ := range cc.ListMustHaveTypes {
		// fmt.Println("###")
		_, ok := tags.byType[typ]
		if !ok {
			return fmt.Errorf("%q type does not exist", typ)
		}
	}
	errors := []string{}
	for typ, list := range tags.byType {
		if strings.HasPrefix(typ, "ERR_") {
			errors = append(errors, fmt.Sprintf("%v: %v", typ, strings.Join(list, ", ")))
			continue
		}
		// if typ == "UNKNOWN_TAG" {
		// // if isStrictCheck {
		// return fmt.Errorf("UNKNOWN type tag(s) are present: %v", list)
		// // }
		// continue
		// }
		if typ == "INVALID_TAG" {
			return fmt.Errorf("invalid tag(s) are present: %v", list)
		}
		if _, ok := cc.TabNonUniqueTypes[typ]; !ok {
			if len(list) > 1 {
				return fmt.Errorf("%q type must be unique", typ)
			}
		}
		if isNotExist(cc.TabValidTypes, typ) {
			return fmt.Errorf("%q is not a valid tag type", typ)
		}
		if isExist(cc.TabInvalidTypes, typ) {
			return fmt.Errorf("%q is an invalid tag type: %q", typ, list)
		}
		for _, val := range list {
		// if !isExist(o.tabValid, val) {
		// return fmt.Errorf("tag (%q,%q) has not a valid value", typ, val)
		// }
			if isExist(cc.TabInvalidValues, val) {
				return fmt.Errorf("tag (%q,%q) has an invalid value", typ, val)
			}
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("some error(s):\n        %v", strings.Join(errors, "\n        "))
	}

	switch typ {
	case "film", "trailer":
		return checkFilmsOrTrailers(tags, typ)
	case "poster", "poster.logo":
		return checkPostersOrLogo(tags, typ)
	case "poster.gp":
		return checkGpPosters(tags, typ)
	default:
		return fmt.Errorf("check: unsupported tag 'type': %q", typ)
	}
	// unreachable
}


func checkFilmsOrTrailers(tags *TTags, typ string) error {
	return nil
}

func checkPostersOrLogo(tags *TTags, t string) error {
	// switch t {
	// case "logo", "
	// }
	// return fmt.Errorf("unsupported tag type: %q", t)
	return nil
}


func checkGpPosters(tags *TTags, t string) error {
	// switch t {
	// case "logo", "
	// }
	// return fmt.Errorf("unsupported tag type: %q", t)
	return nil
}
