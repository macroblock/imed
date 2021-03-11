package tagname

import (
	"fmt"
	"sort"
	"strings"

	"github.com/macroblock/imed/pkg/ptool"
)

// TSchema -
type TSchema struct {
	name                    string
	parser                  **ptool.TParser
	// MustHaveByType          []string
	// NonUniqueByType         []string // can be placed multiple times
	// Valid                   []string
	// ValidByType             []string // if empty then any tag is valid
	// Invalid                 []string
	// InvalidByType           []string
	ToStringHeadOrderByType []string
	ToStringTailOrderByType []string
	// ReadFilter              func(typ, val string) (string, string, error)
	// WriteFilter             func(typ, val string) (string, string, error)
	UnmarshallFilter func(in, out *TTags, typ, val string, firstRun bool) error
	MarshallFilter   func(in, out *TTags, typ, val string, firstRun bool) error

	// HackFilter func(tags *TTags)
	// checker *tChecker
}

// tags:
// name, snen, comment, year, tail, ext
// sdhd, qtag, atag, agetag, metatag, unktag;
// "_" - divider
//# snen     = 's' digit digit ['_' digit digit [digit] {, !(      COMPREFIX,|EONAME) ident}];
var body = `
,        = '_';
ZZZ      = 'zzz';

snen	 = @sxx [,@sname] [,@exx [,@ename]];
sxx      = 's' (digit digit | 'xx' | 'XX' | 'xX' | 'Xx');
exx      = !(EONAME) digit digit [digit] ['a'|'b'];
name     =                     ident {, !(sxx,|ZZZ,|EONAME) ident};
sname    = !(exx,|ZZZ,|EONAME) ident {, !(exx,|ZZZ,|EONAME) ident};
ename    = !(     ZZZ,|EONAME) ident {, !(     ZZZ,|EONAME) ident};
` +
// "comment  = ZZZ,      !(EONAME) ident {, !(          EONAME) ident};" +
"comment  = ZZZ      {, !(          EONAME) ident};" +
`

year     = digit digit digit digit !symbol;
hex      = '#' symbol symbol symbol symbol symbol symbol symbol symbol;

tags     = @INVALID_TAG | @EXCLUSIVE_TAGS
         |@qtag|@atag|@smktag|@alreadyagedtag|@agetag|@stag|@vtag
	 |@hardsubtag|@sbstag|@mtag|@sizetag|@datetag|@aligntag|@hashtag
         |@ERR_agetag|@ERR_atag|@UNKNOWN_TAG;

qtag      = 'q'digit('w'|'s')digit !symbol;
atag      = 'a' ( letter letter letter | 'r' | 'e' ) digit {( letter letter letter | 'r' | 'e' ) digit} !symbol;
stag      = 's' staglang {staglang} !symbol;
agetag    = ('00'|'06'|'12'|'16'|'18'|'99') !symbol;
alreadyagedtag = digit digit 'aged' !symbol;
vtag      = 'v' ('goblin'|'kurazhbambey'|'lostfilm'|'newstudio'|'pozitiv'|ERR_invalid_vtag) !symbol;
hardsubtag= ('mhardsub'|'hardsub'|'xhardsub') !symbol;
smktag    = ('xsmoking'|'xsmk'|'msmoking'|'msmk'|'smoking'|'smk') !symbol;
sbstag    = ('msbs'|'sbs'|'xsbs') !symbol;
mtag      = 'm' symbol {symbol} !symbol;
sizetag   = ('logo' | digit digit {digit} ('x'|'-') digit digit {digit}) !symbol;
aligntag  = ('center'|'left') !symbol;
datetag   = 'd' digit digit digit digit digit digit digit digit digit digit !symbol;
hashtag   = 'x' symbol symbol symbol symbol symbol symbol symbol symbol symbol symbol !symbol;

EXCLUSIVE_TAGS = ('amed'|'abc'|'pb'|'vp'|'disney'|('dop'{symbol})|'oscar'|'dk'|'ru'|'pryamoiz'
		 |'newstudio'|'pozitiv'|'lostfilm'
                 ) !symbol;

UNKNOWN_TAG = !'poster' symbol{symbol};

staglang = 'r'|'s'|ERR_unsupported_subtitle_language;

ERR_atag                          = 'a' {symbol};
ERR_agetag                        = digit digit !symbol;
ERR_invalid_vtag                  = {symbol};
ERR_unsupported_subtitle_language = letter;

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

// func dummyFilter(typ, val string) (string, string, error) { return typ, val, nil }
// func dummyHackFilter(_ *TTags)                            {}

// RegisterSchema - parameter name is caseinsensitive
func RegisterSchema(name string, schema *TSchema) {
	schema.name = name
	name = strings.ToLower(name)
	globSchemas[name] = schema
	// initChecker(schema)
}

// Schema -
func Schema(name string) (*TSchema, error) {
	name = strings.ToLower(name)
	ret, ok := globSchemas[name]
	if !ok {
		return nil, fmt.Errorf("%q is not a registered settings name", name)
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

func filterFixCommonTags(typ, val string) (string, string) {
	switch typ {
	case "EXCLUSIVE_TAGS":
		switch val {
		default:
			typ = "mtag"
			val = "m"+val
		case "newstudio", "pozitiv", "lostfilm":
			typ = "vtag"
			val = "v"+val
		}
	case "UNKNOWN_TAG":
		switch val {
		case "fix":
			typ = ""
			val = ""
		case "hSub":
			typ = "hardsubtag"
			val = "xhardsub"
		case "SD":
			typ = "sdhd"
			val = "sd"
		case "HD":
			typ = "sdhd"
			val = "hd"
		case "goblin":
			typ = "vtag"
			val = "vgoblin"
		}
	case "mtag":
		switch val {
		case "m4o", "mnew":
			typ = ""
			val = ""
		case "mgoblin", "mgolbin":
			typ = "vtag"
			val = "vgoblin"
		case "mkurazhbambey":
			typ = "vtag"
			val = "vkurazhbambey"
		case "mKurazhbombey":
			typ = "vtag"
			val = "vkurazhbambey"
		case "mhurdsub", "mahardsub":
			typ = "hardsubtag"
			val = "xhardsub"
		}
	case "smktag":
		val = "xsmoking"
	case "sbstag", "hardsubtag":
		val = strings.TrimPrefix(val, "m")
		val = strings.TrimPrefix(val, "x")
		val = "x" + val

	}
	return typ, val
}

// fixATag - add ar6 for hd movies or ar2 for sd movies or trailers. Do nothing for other types
func fixATag(tags *TTags) error {
	typ, err := tags.GetTag("type")
	if err != nil {
		return err
	}
	if typ != "trailer" && typ != "film" {
		return nil
	}

	atag, _ := tags.GetTag("atag")
	if atag != "" {
		return nil
	}
	sdhd, err := tags.GetTag("sdhd")
	if err != nil {
		return err
	}

	if sdhd == "sd" || typ == "trailer" {
		tags.AddTag("atag", "ar2")
		return nil
	}
	tags.AddTag("atag", "ar6")
	return nil
}

// unfixATag - remove ar6 for hd movies or ar2 for sd movies or trailers. Do nothing for other types
func unfixATag(tags *TTags) error {
	typ, err := tags.GetTag("type")
	if err != nil {
		return err
	}
	if typ != "trailer" && typ != "film" {
		return nil
	}

	atag, err := tags.GetTag("atag")
	if err != nil {
		return err
	}

	sdhd, err := tags.GetTag("sdhd")
	if err != nil {
		return err
	}

	switch {
	case atag == "ar2" && (sdhd == "sd" || typ == "trailer"):
		tags.RemoveTags("atag")
	case atag == "ar6" && (sdhd == "hd" || sdhd == "3d") && typ == "film":
		tags.RemoveTags("atag")
	}
	return nil
}
