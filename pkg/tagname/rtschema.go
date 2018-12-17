package tagname

import (
	"fmt"
	"regexp"
	"strings"
)

var rtForm = `
entry       =  (@_hackHD3D @sdhd, @year, @_hack3D,| !(,) @sdhd, @year,) @name [,snen] [,@comment] [DIV taglist] @type @ext$;

sdhd        = ['sd'|'hd'];
_hackHD3D   = 'hd' !('hd',|'3d',);
_hack3D     = '3d';
type        = 'trailer'|'film'| poster;

poster      = 'poster' digit{digit} 'x' digit{digit};

taglist     = {!(type ('.'|$)) tags,};
EONAME      = DIV|'.'|$;

INVALID_TAG = 'sd'|'hd'|'3d';
` + body

var rtNormalSchema = &TSchema{
	parser:                  &rtParser,
	MustHaveByType:          []string{"name", "year", "sdhd", "type", "ext"},
	NonUniqueByType:         nil,
	Invalid:                 nil,
	ToStringHeadOrderByType: []string{"_hackHD3D", "sdhd", "year", "_hack3D", "name", "sxx", "sname", "exx", "ename", "comment", "_", "alreadyagedtag", "agetag", "qtag", "atag", "stag"},
	ToStringTailOrderByType: []string{"m4o", "type", "ext"},
	ReadFilter:              fnRtSchemaReadFilter,
	WriteFilter:             fnRtSchemaWriteFilter,
	HackFilter:              fnHackRtFilter,
}

var localBuffer = ""

var reRes = regexp.MustCompile(`\d+x\d+`)

func fnRtSchemaReadFilter(typ, val string) (string, string, error) {
	err := error(nil)
	switch typ {
	case "sdhd":
		if val == "" {
			val = "3d"
		}
	case "_hackSDHD":
		val = ""
	case "snen":
		val, err = fixSnen(val)
	case "name":
		val = strings.ToLower(val)
	case "type":
		val = strings.TrimPrefix(val, "poster")
	}
	return typ, val, err
}

func fnRtSchemaWriteFilter(typ, val string) (string, string, error) {
	err := error(nil)
	switch typ {
	case "sdhd":
		if val == "3d" {
			val = ""
		}
	case "snen":
		val, err = unfixSnen(val)
	case "name":
		val = strings.Title(val)
	case "type":
		if reRes.MatchString(val) {
			val = "poster" + val
		}
	}
	return typ, val, err
}

func fnHackRtFilter(tags *TTags) {
	list := tags.GetTags("sdhd")
	tags.RemoveTags("_hackHD3D")
	tags.RemoveTags("_hack3D")
	if len(list) == 0 {
		return
	}
	_, val, err := fnRtSchemaReadFilter("sdhd", list[0])
	if err != nil {
		panic(fmt.Sprintf("fnHackRTFilter() error: %v", err))
	}
	if val == "3d" {
		tags.AddTag("_hackHD3D", "hd")
		tags.AddTag("_hack3D", "3d")
	}
}
