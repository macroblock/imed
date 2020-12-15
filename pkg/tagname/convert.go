package tagname

import (
	"strings"
	// "fmt"
)

// TranslateTags -
func TranslateTags(srcTags *TTags, fnFilter func(in, out *TTags, typ, val string, done bool) error) (*TTags, error) {
	dstTags := &TTags{}

	err := fnFilter(srcTags, dstTags, "", "", true)
	if err != nil {
		return nil, err
	}

	for typ, list := range srcTags.byType {
		for _, val := range list {
			err := fnFilter(srcTags, dstTags, typ, val, false)
			// fmt.Printf("--- tag %v, %v ", typ, val)
			if err != nil {
				return nil, err
			}
		}
	}
	// fmt.Printf("end of tags %v\n", dstTags)
	err = fnFilter(srcTags, dstTags, "", "", false)
	if err != nil {
		return nil, err
	}

	return dstTags, nil
}

func toString(tagname *TTags, toSchema *TSchema) (string, error) {
	err := error(nil)
	used := map[string]bool{}

	tagname, err = TranslateTags(tagname, toSchema.MarshallFilter)
	if err != nil {
		return "", err
	}

	appendNonEmptyStrings := func(src, list []string, aType string) ([]string, error) {
		ret := src
		for _, v := range list {
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
				tags, _ := tagname.byType[typ]
				ret, err = appendNonEmptyStrings(ret, tags, typ)
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
