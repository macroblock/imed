package tagname

import (
	"regexp"
	"strings"

	"github.com/macroblock/imed/pkg/hash"
)

var rtForm = `
entry       =  (@_hackHD3D @sdhd, @year, @_hack3D,| !(,) @sdhd, @year,) @name [,snen] [,@comment] [DIV taglist] @type @ext$;

sdhd        = ['sd'|'hd'];
_hackHD3D   = 'hd' !('hd',|'3d',);
_hack3D     = '3d';
type        = 'trailer'|'film'|'teaser'| 'logo' | poster;

poster      = ('poster' sizetag) | 'logo';

taglist     = {!(type ('.'|$)) tags,};
DIV         = '__';
EONAME      = DIV|'.'|$;

INVALID_TAG = 'sd'|'hd'|'3d'|'logo'|'poster';
` + body

var rtNormalSchema = &TSchema{
	parser:                  &rtParser,
	// MustHaveByType:          []string{"name", "year", "sdhd", "type", "ext"},
	// NonUniqueByType:         nil,
	// Invalid:                 nil,
	ToStringHeadOrderByType: []string{"sdhd", "year", "_hack3D", "name", "sxx", "sname", "exx", "ename", "comment", "_", "alreadyagedtag", "agetag", "qtag", "atag", "stag"},
	ToStringTailOrderByType: []string{"datetag", "hashtag", "type", "ext"},
	UnmarshallFilter:        fnFromRTFilter,
	MarshallFilter:          fnToRTFilter,
}

var localBuffer = ""

var reRes = regexp.MustCompile(`\d+x\d+`)

// func getOneTag(tags *TTags, typ string) (string, error) {
// list := tags.GetTags(typ)
// switch {
// case len(list) == 1:
// return list[0], nil
// case len(list) > 1:
// return "", fmt.Errorf("too many '%v' tags", typ)
// default:
// return "", fmt.Errorf("tag '%v' not found", typ)
// }
// // unreachable
// }

func genHashTag(tags *TTags) string {
	name, _ := tags.GetTag("name")
	sxx, _ := tags.GetTag("sxx")
	year, _ := tags.GetTag("year")
	sdhd, _ := tags.GetTag("sdhd")
	comment, _ := tags.GetTag("comment")
	key := name + "_" + sxx + "_" + year + "_" + sdhd + "_" + comment
	return "x" + hash.Get(key)
}

func fnFromRTFilter(in, out *TTags, typ, val string, firstRun bool) error {
	if typ == "" && val == "" {
		// for last run only
		if !firstRun {
			err := fixATag(out)
			return err
		}
		return nil
	}

	typ, val = filterFixCommonTags(typ, val)
	if typ == "" {
		return nil
	}

	switch typ {
	case "hashtag", "_hackHD3D", "_hack3D":
		return nil
	case "sdhd":
		if val == "" {
			val = "3d"
		}
	case "name":
		val = strings.ToLower(val)
	case "sizetag":
		val = strings.ReplaceAll(val, "-", "x")
	case "type":
		switch {
		case strings.HasPrefix(val, "poster"):
			size := strings.TrimPrefix(val, "poster")
			size = strings.ReplaceAll(size, "-", "x")
			out.AddTag("sizetag", size)
			val = "poster"
		case val == "logo":
			val = "poster"
			out.AddTag("sizetag", "logo")
		}
	}

	out.AddTag(typ, val)
	return nil
}

func fnToRTFilter(in, out *TTags, typ, val string, firstRun bool) error {
	// if typ == "" && val == "" {
		// // for first run only
		// if firstRun {
			// err := unfixATag(in)
			// return err
		// }
		// return nil
	// }

	switch typ {
	case "sdhd":
		if val == "3d" {
			val = "hd"
			out.AddTag("_hack3D", "3d")
		}
	// case "name":
		// val = strings.Title(val)
	case "sizetag":
		t, _ := in.GetTag("type")
		if t == "poster" {
			return nil
		}
	case "type":
		switch val {
		case "poster":
			size, err := in.GetTag("sizetag")
			if err != nil {
				return err
			}
			if size == "logo" {
				val = "logo"
			} else {
				val = "poster" + size
			}
		// case "poster.logo":
			// val = "logo.poster"
		case "poster.gp":
		case "film":
			hash := genHashTag(in)
			out.AddTag("hashtag", hash)
		}
	}

	out.AddTag(typ, val)
	return nil
}
