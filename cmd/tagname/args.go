package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/macroblock/ptool/pkg/ptool"
)

func splitFn(r rune) bool {
	return r == '|' || r == ':' || r == ';' || r == ','
}

func splitGroups(r rune) bool {
	return r == ';' || r == ' '
}

func splitItems(r rune) bool {
	return r == ',' || r == '|'
}

func splitAssign(r rune) bool {
	return r == '='
}

func splitEquation(r rune) bool {
	return r == ':'
}

func argSplit(s string) []string {
	list := strings.FieldsFunc(s, splitFn)
	ret := []string{}
	for _, s := range list {
		s = strings.TrimSpace(s)
		if s != "" {
			ret = append(ret, s)
		}
	}
	return ret
}

func argFilename(name string) ([]string, error) {
	name = strings.TrimSpace(name)

	// fmt.Printf("name %q\n", name)
	path := getListPath(name)
	// fmt.Println("1 ", path)
	if path != "" {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return []string{name}, nil
		}

		list := strings.Split(string(data), "\n")
		ret := []string{}
		for _, s := range list {
			s = strings.Split(s, "//")[0] // get rid of comments
			s = strings.TrimSpace(s)
			if s != "" {
				ret = append(ret, s)
			}
		}
		return ret, nil
	}
	return []string{name}, nil
}

type (
	tFilterItem struct {
		name    string
		filter  []tVal
		replace *string
	}
	tVal struct {
		neg bool
		val string
	}
)

const filterParserSource = `
entry    = [@item]{';'[@item]}$;
item     = @name (':' val{','val} ['=' @replace] | '=' @replace);
val      = negative|@pos;
negative = '!' @neg;
neg      = [ident];
pos      = [ident];
replace  = [ident];
name     = ident;
digit    = '0'..'9';
letter   = 'a'..'z'|'A'..'Z';
symbol   = letter|digit|'_'|'#'|'.';
ident	 = symbol{symbol};
`

func argSplitFilter(arg string) ([]tFilterItem, error) {
	tree, err := filterParser.Parse(arg)
	if err != nil {
		return nil, err
	}
	fmt.Println("-------------")
	s := ptool.TreeToString(tree, filterParser.ByID)
	fmt.Println(s)
	fmt.Println("-------------")
	filter := []tFilterItem(nil)
	for _, group := range tree.Links {
		item := tFilterItem{}
		for _, v := range group.Links {
			switch filterParser.ByID(v.Type) {
			case "name":
				item.name = v.Value
			case "replace":
				x := v.Value
				item.replace = &x
			case "pos":
				item.filter = append(item.filter, tVal{val: v.Value, neg: false})
			case "neg":
				item.filter = append(item.filter, tVal{val: v.Value, neg: true})
			}
		}
		filter = append(filter, item)
	}
	return filter, nil
}
