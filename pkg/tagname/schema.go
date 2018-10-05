package tagname

import (
	"fmt"
	"sort"
	"strings"

	"github.com/macroblock/imed/pkg/ptool"
)

// TSchema -
type TSchema struct {
	parser                  **ptool.TParser
	MustHaveByType          []string
	NonUniqueByType         []string // can be placed multiple times
	Valid                   []string
	ValidByType             []string // if empty then any tag is valid
	Invalid                 []string
	InvalidByType           []string
	ToStringHeadOrderByType []string
	ToStringTailOrderByType []string
	ReadFilter              func(typ, val string) (string, string, error)
	WriteFilter             func(typ, val string) (string, string, error)
	HackFilter              func(tags *TTags)
	checker                 *tChecker
}

// tags:
// name, snen, comment, year, tail, ext
// sdhd, qtag, atag, agetag, metatag, unktag;
// "_" - divider
//# snen     = 's' digit digit ['_' digit digit [digit] {, !(      COMPREFIX,|EONAME) ident}];
var body = `
,        = '_';
DIV      = '__';
ZZZ      = 'zzz';

snen	 = @sxx [,@sname] [,@exx [,@ename]];
sxx      = 's' (digit digit | 'xx' | 'XX' | 'xX' | 'Xx');
exx      = !(EONAME) digit digit [digit];
name     =                     ident {, !(sxx,|ZZZ,|EONAME) ident};
sname    = !(exx,|ZZZ,|EONAME) ident {, !(exx,|ZZZ,|EONAME) ident};
ename    = !(     ZZZ,|EONAME) ident {, !(     ZZZ,|EONAME) ident};
comment  = ZZZ,      !(EONAME) ident {, !(          EONAME) ident};

postertype = 'poster' digit{digit} 'x' digit{digit};

year     = digit digit digit digit;

tags     = @qtag|@atag|@stag|@agetag|@m4otag|@mtag|@unktag;

qtag     = 'q'digit('w'|'s')digit;
atag     = 'a'letter digit{letter digit};
stag     = 's'letter {letter};
agetag   = '00'|'06'|'12'|'16'|'18'|'99';
m4otag   = 'm4o';
mtag     = 'm'letter{letter};
unktag   = ident;

ext      = ['.'ident];

digit    = '0'..'9';
letter   = 'a'..'z'|'A'..'Z';
symbol   = letter|digit;
ident	 = symbol{symbol};
`

func fixSnen(val string) (string, error) {
	if val[3] != '_' {
		return val, fmt.Errorf("wrong snen-tag format %q", val)
	}
	return val[:3] + "e" + val[4:], nil
}

func unfixSnen(val string) (string, error) {
	if val[3] != 'e' {
		return val, fmt.Errorf("wrong snen-tag format %q", val)
	}
	return val[:3] + "_" + val[4:], nil
}

var globSchemas map[string]*TSchema

func dummyFilter(typ, val string) (string, string, error) { return typ, val, nil }
func dummyHackFilter(_ *TTags)                            {}

// RegisterSchema - parameter name is caseinsensitive
func RegisterSchema(name string, schema *TSchema) {
	name = strings.ToLower(name)
	globSchemas[name] = schema

	if schema.ReadFilter == nil {
		schema.ReadFilter = dummyFilter
	}
	if schema.WriteFilter == nil {
		schema.WriteFilter = dummyFilter
	}
	if schema.HackFilter == nil {
		schema.HackFilter = dummyHackFilter
	}
	initChecker(schema)
}

// Schema -
func Schema(name string) (*TSchema, error) {
	name = strings.ToLower(name)
	ret, ok := globSchemas[name]
	if !ok {
		return nil, fmt.Errorf("%q is not registered settings name", name)
	}
	return ret, nil
}

// Schemas -
func Schemas() []string {
	keys := make([]string, 0, len(globSchemas))
	for key := range globSchemas {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
