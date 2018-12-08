package tagname

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/macroblock/imed/pkg/zlog/zlog"
)

var (
	log = zlog.Instance("tagname")
)

// TTagname -
type TTagname struct {
	dir    string
	src    string
	schema string
	tags   *TTags
}

// NewFromString -
func NewFromString(str string, checkLevel int, schemaNames ...string) (*TTagname, error) {
	var err error
	var tags *TTags
	var schema string
	var errors []string

	schemas := schemaNames
	if len(schemas) == 0 {
		schemas = []string{"rt", "old"}
	}

	for _, schema = range schemas {
		tags, err = Parse(str, schema)
		if err == nil {
			break
		}
		errors = append(errors, fmt.Sprintf("%-16q: %v", schema, err))
	}

	if err != nil {
		s := strings.Join(errors, "\n  ")
		return nil, fmt.Errorf("multiple ones:\n  %v", s)
	}

	tn := &TTagname{schema: schema, src: str, tags: tags}

	err = tn.Check(checkLevel)
	if err != nil {
		return nil, err
	}
	return tn, nil
}

// NewFromFilename -
func NewFromFilename(path string, checkLevel int, schemaNames ...string) (*TTagname, error) {
	src := filepath.Base(path)
	ret, err := NewFromString(src, checkLevel, schemaNames...)
	if err != nil {
		return nil, err
	}
	ret.dir = filepath.Dir(path)
	return ret, nil
}

// State -
func (o *TTagname) State() error {
	if o == nil {
		return fmt.Errorf("tagname object is nil")
	}
	return o.tags.State()
}

// ConvertTo -
func (o *TTagname) ConvertTo(schemaName string) (string, error) {
	if schemaName == "" {
		schemaName = o.schema
	}
	_, err := Schema(schemaName)
	if err != nil {
		return "", err
	}
	ret, err := ToString(o.tags, o.schema, schemaName)
	if err != nil {
		return "", err
	}
	ret = filepath.Join(o.dir, ret)
	return ret, nil
}

// Check -
func (o *TTagname) Check(checkLevel int) error {
	if err := o.State(); err != nil {
		return err
	}

	if checkLevel < 0 {
		return nil
	}

	isStrictCheck := checkLevel&CheckStrict != 0
	isDeepCheck := checkLevel&CheckDeep != 0
	err := o.tags.Check(isStrictCheck)
	if err != nil || !isDeepCheck {
		return err
	}
	err = checkDeep(o)
	return err
}

// GetTag -
func (o *TTagname) GetTag(typ string) (string, error) {
	list := o.tags.GetTags(typ)
	if len(list) == 0 {
		return "", fmt.Errorf("%q has no tags of %q type", o.src, typ)
	}
	if len(list) > 1 {
		return "", fmt.Errorf("GetTag() cannot return multiple tags of %q type in %q", typ, o.src)
	}
	schema, err := Schema(o.schema)
	if err != nil {
		return "", err
	}
	val := list[0]
	typ, val, err = schema.ReadFilter(typ, val)
	if err != nil {
		return "", err
	}
	return val, nil
}

// GetTags -
func (o *TTagname) GetTags(typ string) []string {
	list := o.tags.GetTags(typ)
	// if len(list) == 0 {
	// 	return nil, fmt.Errorf("%q has no tags of %q type", o.src, typ)
	// }
	schema, err := Schema(o.schema)
	if err != nil {
		fmt.Println("Schema() error at tagname.GetTag")
		panic(err)
		// return "", err
	}
	ret := []string{}
	for _, s := range list {
		_, val, err := schema.ReadFilter(typ, s)
		if err != nil {
			fmt.Println("ReadFilter() error at tagname.GetTag")
			panic(err)
			// return "", err
		}
		ret = append(ret, val)
	}
	return ret
}

// RemoveTags -
func (o *TTagname) RemoveTags(typ string) {
	o.tags.RemoveTags(typ)
}

// SetTag -
func (o *TTagname) SetTag(typ string, val string) {
	o.tags.RemoveTags(typ)
	o.tags.AddTag(typ, val)
}

// Schema -
func (o *TTagname) Schema() string {
	return o.schema
}

// GetFormat -
func (o *TTagname) GetFormat() (string, error) {
	return o.GetTag("sdhd")
}

// GetType -
func (o *TTagname) GetType() (string, error) {
	return o.GetTag("type")
}

// TQuality -
type TQuality struct {
	Quality    int
	Widescreen bool
	CacheType  int
}

// GetQuality -
func (o *TTagname) GetQuality() (*TQuality, error) {
	q, err := o.GetTag("qtag")
	if err != nil {
		return nil, fmt.Errorf("qtag is absent")
	}
	quality, err := strconv.Atoi(string(q[1]))
	if err != nil {
		panic(err)
	}
	cachetype, err := strconv.Atoi(string(q[3]))
	if err != nil {
		panic(err)
	}
	wide := false
	switch q[2] {
	default:
		panic("unreachable")
	case 'w':
		wide = true
	case 's':
		wide = false
	}
	return &TQuality{Quality: quality, Widescreen: wide, CacheType: cachetype}, nil
}

// TAudio -
type TAudio struct {
	language string
	channels int
}

// GetAudio -
func (o *TTagname) GetAudio() ([]TAudio, error) {
	val, err := o.GetTag("atag")
	if err != nil {
		// trying to describe by format and type
		typ, err1 := o.GetType()
		frm, err2 := o.GetFormat()
		if err1 != nil || err2 != nil {
			return nil, fmt.Errorf("%v", "cannot get audio tag (fomat or/and type tags are missing)")
		}

		if typ != "film" && typ != "trailer" {
			return nil, fmt.Errorf("%v", "cannot get audio tag (fomat or/and type tags are missing)")
		}
		val = "ar2"
		if frm != "sd" && typ == "film" {
			val = "ar6"
		}
	}
	// fill a ret struct
	ret := []TAudio{}
	lang := ""
	for _, r := range val[1:] {
		if r < '0' || r > '9' {
			lang += string(r)
			continue
		}
		ch, err := strconv.Atoi(string(r))
		if err != nil {
			panic("strconv")
		}
		switch lang {
		case "r":
			lang = "rus"
		case "e":
			lang = "eng"
		}
		ret = append(ret, TAudio{lang, ch})
		lang = ""
	}
	return ret, nil
}

// GetSubtitle -
func (o *TTagname) GetSubtitle() ([]string, error) {
	val, err := o.GetTag("stag")
	if err != nil {
		return nil, nil
	}
	// fill a ret struct
	ret := []string{}
	lang := ""
	for _, r := range val[1:] {
		lang = string(r)
		switch lang {
		default:
			lang = "???"
		case "r":
			lang = "rus"
		case "e":
			lang = "eng"
		}
		ret = append(ret, lang)
	}
	return ret, nil
}

// TResolution -
type TResolution struct {
	W, H int
}

func (o TResolution) String() string {
	return fmt.Sprintf("%vx%v", o.W, o.H)
}

// TFormat -
type TFormat struct {
	resolution TResolution
	Sar        string
	Audio      []TAudio
	Subtitle   []string
	Quality    int
	CacheType  int
	Sbs        bool
}

func newFormat() *TFormat {
	return &TFormat{resolution: TResolution{-1, -1}, Quality: -1, CacheType: -1, Sbs: false}
}

// Describe -
func (o *TTagname) Describe() (*TFormat, error) {
	format := newFormat()
	quality, err := o.GetQuality()
	// if err != nil {
	// 	return nil, err
	// }

	frm, err := o.GetFormat()
	switch frm {
	default:
		if err == nil {
			err = fmt.Errorf("unsupported format %q of the tagname %v", frm, o.src)
		}
		return nil, err
	case "hd", "3d":
		format.resolution = TResolution{1920, 1080}
		format.Sar = "1:1"
	case "sd":
		format.resolution = TResolution{720, 576}
		if quality == nil {
			break
		}
		format.Sar = "16:15"
		if quality.Widescreen {
			format.Sar = "64:45"
		}
	}

	format.Audio, err = o.GetAudio()
	if err != nil {
		return nil, err
	}
	format.Subtitle, err = o.GetSubtitle()
	if err != nil {
		return nil, err
	}

	if quality != nil {
		format.Quality = quality.Quality
		format.CacheType = quality.CacheType
	}
	return format, nil
}
