package tagname

import "strings"

var oldForm = `
entry    = @name [,snen] [,@comment] ,@year [DIV taglist] @type @ext$;

sdhd     = 'sd'|'hd'|'3d';
type     = ['.trailer'| '_'poster];

poster   = digit{digit} '-' digit{digit} '.poster' [hex];

taglist  = [(@sdhd|tags){,(@sdhd|tags)}];
EONAME   = year (DIV|'.'|$);

INVALID_TAG = 'trailer'|'film';
` + body

var oldNormalSchema = &TSchema{
	parser:                  &oldParser,
	MustHaveByType:          []string{"name", "year", "sdhd", "type", "ext"},
	NonUniqueByType:         nil,
	Invalid:                 nil, // []string{"trailer", "film"},
	ToStringHeadOrderByType: []string{"name", "sxx", "sname", "exx", "ename", "comment", "year", "_", "sdhd", "alreadyagedtag", "agetag", "qtag", "atag", "stag"},
	ToStringTailOrderByType: []string{"m4otag", "datetag", "hashtag", "type", "ext"},
	ReadFilter:              fnOldSchemaReadFilter,
	WriteFilter:             fnOldSchemaWriteFilter,
}

func fnOldSchemaReadFilter(typ, val string) (string, string, error) {
	err := error(nil)
	switch typ {
	case "type":
		switch val {
		case "":
			val = "film"
		case ".trailer":
			val = "trailer"
		default:
			if strings.Contains(val, "poster") {
				x := strings.Split(val+"#", "#")
				val = x[0]
				val = strings.TrimPrefix(val, "_")
				val = strings.TrimSuffix(val, ".poster")
				val = strings.Replace(val, "-", "x", -1) + "#" + x[1]
			}
		}
	case "snen":
		val, err = fixSnen(val)
	case "name":
		val = strings.ToLower(val)
	}
	return typ, val, err
}

func fnOldSchemaWriteFilter(typ, val string) (string, string, error) {
	err := error(nil)
	switch typ {
	case "type":
		switch val {
		case "film":
			val = ""
		case "trailer":
			val = ".trailer"
		default:
			if strings.Contains(val, "#") {
				x := strings.Split(val, "#")
				val = x[0]
				val = strings.Replace(val, "x", "-", -1) + ".poster"
				if len(x[1]) != 0 {
					val += "#" + x[1]
				}
			}
		}
	case "snen":
		val, err = unfixSnen(val)
	case "name":
		val = strings.Title(val)
	}
	return typ, val, err
}
