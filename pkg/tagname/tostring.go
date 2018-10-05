package tagname

import (
	"strings"
)

func toString(tagname *TTags, fromSchema, toSchema *TSchema) (string, error) {
	used := map[string]bool{}
	toSchema.HackFilter(tagname)

	appendNonEmptyStrings := func(src, list []string, aType string) ([]string, error) {
		err := error(nil)
		ret := src
		for _, v := range list {
			typ := aType
			typ, v, err = fromSchema.ReadFilter(typ, v)
			if err != nil {
				return nil, err
			}
			typ, v, err = toSchema.WriteFilter(typ, v)
			if err != nil {
				return nil, err
			}
			if v != "" {
				ret = append(ret, v)
			}
		}
		return ret, nil
	}

	getTagsByList := func(list []string) ([]string, error) {
		ret := []string{}
		err := error(nil)
		for _, typ := range list {
			used[typ] = true
			if typ == "_" {
				ret = append(ret, "")
			} else {
				ret, err = appendNonEmptyStrings(ret, tagname.byType[typ], typ)
				if err != nil {
					return nil, err
				}
			}
		}
		return ret, nil
	}

	head, err := getTagsByList(toSchema.ToStringHeadOrderByType)
	if err != nil {
		return "", err
	}
	tail, err := getTagsByList(toSchema.ToStringTailOrderByType)
	if err != nil {
		return "", err
	}

	freeTags := []string{}
	for typ, list := range tagname.byType {
		// skip types that has "_" prefix
		if !used[typ] && !strings.HasPrefix(typ, "_") {
			freeTags, err = appendNonEmptyStrings(freeTags, list, typ)
			if err != nil {
				return "", err
			}
		}
	}

	head = multiJoin(head, freeTags, tail)
	hstr := strings.Join(head, "_")

	hstr = strings.TrimRight(hstr, "_")
	hstr = strings.Replace(hstr, "_.", ".", -1)
	hstr = strings.Replace(hstr, "_.", ".", -1)

	return hstr, nil
}
