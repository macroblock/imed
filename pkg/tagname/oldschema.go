package tagname

import "strings"

var oldForm = `
entry    = @name [,snen] [,@comment] ,@year [DIV taglist] ['.' @type] @ext$;

sdhd     = ('sd'|'hd'|'3d'|'4k') !symbol;
type     = 'trailer'| 'poster' | 'teaser';
` +
	// 999x999.poster
	// logo.poster

	// poster   = (((digit{digit} ('-'|'x') digit{digit}) | 'logo') '.poster') | 'logo';
	`
taglist  = [(@sdhd|tags){,(@sdhd|tags)}];
` +
	// "EONAME   = year (DIV|'.'|$);" +
	"EONAME   = year !({ '_' !(year) ident } '_' year) (DIV|'.'|$);" +
	`
DIV = '__'|'_';

INVALID_TAG = 'asdfafdadf!!';
` + body

var oldNormalSchema = &TSchema{
	parser: &oldParser,
	// MustHaveByType:          []string{"name", "year", "type"},
	// NonUniqueByType:         nil,
	// Invalid:                 nil, //[]string{"trailer", "film", "logo", "poster"},
	ToStringHeadOrderByType: []string{"name", "sxx", "sname", "exx", "ename", "comment", "year", "_", "sdhd", "alreadyagedtag", "agetag", "qtag", "atag", "stag"},
	ToStringTailOrderByType: []string{"datetag", "prttag", "hashtag", "type", "aligntag", "ext"},
	UnmarshallFilter:        fnFromOldFilter,
	MarshallFilter:          fnToOldFilter,
}

func fixTypeTag(tags *TTags) error {
	typ, _ := tags.GetTag("type")
	switch typ {
	default:
		return nil
	case "":
		sdhd, _ := tags.GetTag("sdhd")
		typ = "film"
		if sdhd == "" {
			typ = "poster.gp"
		}
		tags.AddTag("type", typ)
	}
	return nil
}

func fnFromOldFilter(in, out *TTags, typ, val string, firstRun bool) error {
	if typ == "" && val == "" {
		// for last run only
		if !firstRun {
			err := fixTypeTag(out)
			if err != nil {
				return err
			}
			err = fixATag(out)
			return err
		}
		return nil
	}

	typ, val = filterFixCommonTags(typ, val)
	if typ == "" {
		return nil
	}

	switch typ {
	case "hashtag":
		return nil
	case "name":
		val = strings.ToLower(val)
	case "sizetag":
		val = strings.ReplaceAll(val, "-", "x")

		// case "type":
		// val = strings.TrimPrefix(val, "poster")
		// switch {
		// case val == "":
		// val = "film"
		// _, err := in.GetTag("sdhd")
		// if err != nil {
		// val = "poster.gp"
		// }
		// case val == ".trailer":
		// val = "trailer"
		// case strings.HasPrefix(val, "logo"):
		// val = "poster.logo"
		// case strings.HasSuffix(val, ".poster"):
		// val = strings.TrimSuffix(val, ".poster")
		// val = strings.TrimPrefix(val, "_")
		// size := strings.ReplaceAll(val, "-", "x")
		// out.AddTag("sizetag", size)
		// val = "poster"
		// }
	}

	out.AddTag(typ, val)
	return nil
}

func fnToOldFilter(in, out *TTags, typ, val string, firstRun bool) error {
	// if typ == "" && val == "" {
	// // for first run only
	// if firstRun {
	// err := unfixATag(in)
	// return err
	// }
	// return nil
	// }

	switch typ {
	// case "name":
	// val = strings.Title(val)
	case "sizetag":
		t, _ := in.GetTag("type")
		if t == "poster" || t == "poster.gp" {
			return nil
		}
	case "type":
		switch val {
		case "poster":
			size, err := in.GetTag("sizetag")
			if err != nil && size != "logo" {
				return err
			}
			// size = strings.ReplaceAll(size, "x", "-")
			val = size + ".poster"
		case "poster.logo":
			val = "logo.poster"
		case "poster.gp":
			size, err := in.GetTag("sizetag")
			if err != nil {
				return err
			}
			val = size
		case "film":
			val = ""
		case "trailer":
			val = ".trailer"
		}
	}

	out.AddTag(typ, val)
	return nil
}
