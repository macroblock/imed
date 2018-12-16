package tagname

import (
	"testing"
)

var (
	tableTagnameCorrect = []struct {
		settings, input, check string
	}{
		//23456789012345678901234567890
		{settings: "rt",
			input: "//test/path/Sobibor_2018__sd_12_q0w2.trailer.mpg",
			check: "\\\\test\\path\\sd_2018_Sobibor__12_q0w2_trailer.mpg"},
		{settings: "old",
			input: "test\\sd_2018_Sobibor__12_q0w2_trailer.mpg",
			check: "test\\Sobibor_2018__sd_12_q0w2.trailer.mpg"},
		{settings: "",
			input: "sd_2018_Sobibor__12_q0w2_trailer.mpg"},
		{settings: "",
			input: "Sobibor_2018__sd_12_q0w2.trailer.mpg"},
		{settings: "rt",
			input: "Sobibor_2018__3d_12_q0w2.trailer.mpg",
			check: "hd_2018_3d_Sobibor__12_q0w2_trailer.mpg"},
		{settings: "old",
			input: "hd_2018_3d_Sobibor__12_q0w2_trailer.mpg",
			check: "Sobibor_2018__3d_12_q0w2.trailer.mpg"},
		{settings: "",
			input: "hd_2018_3d_Sobibor__12_q0w2_trailer.mpg"},
		{settings: "",
			input: "Sobibor_2018__3d_12_q0w2.trailer.mpg"},
		{settings: "",
			input: "Zvezda_rodilas_2018__hd_190-230.poster.jpg"},
		{settings: "",
			input: "hd_2018_Bezumno_bogatye_aziaty__pr_poster525x300.jpg"},
		{settings: "rt",
			input: "//test/path/Babnik_2008__hd_1620-996.poster.jpg",
			check: "\\\\test\\path\\hd_2008_Babnik__poster1620x996.jpg"},
		{settings: "old",
			input: "//test/path/sd_2018_Proigrannoe_mesto__pryamoiz_poster525x300.jpg",
			check: "\\\\test\\path\\Proigrannoe_mesto_2018__sd_pryamoiz_525-300.poster.jpg"},
	}
	tableTagnameIncorrect = []string{
		//23456789012345678901234567890
		"a",
		"a__",
		"2000",
		"2000__",
		"a_200",
		"a_20000",
		"_a_2000",
		"a-#_2000",
		"a_2000.trailer.ext.zzz",
		"a_2000.ext.zzz",
		"a_2000__.ext.zzz",
		"a_2000__tag__tag2",
		"a__2000",
		"The_name_s01_a_subname_2018__q0w0",
		"The_name_s01_a_subname_2018__hd_q0w0_",
		"The_name_s01_zzz_2018__hd_q0w0",
	}

	tableTagnameGetCorrect = []struct {
		input string
		tags  []ttag
	}{
		{input: "The_name_sXX_a_sname_01_a_ename_zzz_comment_2018__hd_q0w0",
			tags: []ttag{
				{typ: "name", val: "the_name"},
				// {typ: "snen", val: "s01e01_a_subname"},
				{typ: "sxx", val: "sXX"},
				{typ: "exx", val: "01"},
				{typ: "sname", val: "a_sname"},
				{typ: "ename", val: "a_ename"},
				{typ: "comment", val: "zzz_comment"},
				{typ: "sdhd", val: "hd"},
				{typ: "qtag", val: "q0w0"},
				{typ: "type", val: "film"},
				{typ: "ext", val: ""},
			},
		},
		{input: "hd_2018_3d_The_name_sXX_a_sname_01_a_ename_zzz_comment__q0w0_film",
			tags: []ttag{
				{typ: "name", val: "the_name"},
				// {typ: "snen", val: "s01e01_a_subname"},
				{typ: "sxx", val: "sXX"},
				{typ: "exx", val: "01"},
				{typ: "sname", val: "a_sname"},
				{typ: "ename", val: "a_ename"},
				{typ: "comment", val: "zzz_comment"},
				{typ: "sdhd", val: "3d"},
				{typ: "qtag", val: "q0w0"},
				{typ: "type", val: "film"},
				{typ: "ext", val: ""},
			},
		},
		{input: "The_xxx_name_s01_2015__sd_16_q3w0.trailer.mpg",
			tags: []ttag{
				{typ: "name", val: "the_xxx_name"},
				// {typ: "snen", val: "s01e01_a_subname"},
				{typ: "sxx", val: "s01"},
				// {typ: "exx", val: ""},
				// {typ: "sname", val: ""},
				// {typ: "ename", val: ""},
				// {typ: "comment", val: ""},
				{typ: "sdhd", val: "sd"},
				{typ: "qtag", val: "q3w0"},
				{typ: "agetag", val: "16"},
				{typ: "type", val: "trailer"},
				{typ: "ext", val: ".mpg"},
			},
		},
		{input: "Babnik_2008__hd_1620-996.poster.jpg",
			tags: []ttag{
				{typ: "name", val: "babnik"},
				{typ: "year", val: "2008"},
				{typ: "sdhd", val: "hd"},
				{typ: "type", val: "1620x996"},
				{typ: "ext", val: ".jpg"},
			},
		},
	}
)

type ttag struct {
	typ, val string
}

// TestTagnameParseCorrect -
func TestTagnameCorrect(t *testing.T) {
	for _, v := range tableTagnameCorrect {
		tagname, err := NewFromFilename(v.input, CheckNormal)
		if err != nil {
			t.Errorf("\n%q\nNewFormFromFile() error:\n%v", v.input, err)
			continue
		}
		res, err := tagname.ConvertTo(v.settings)
		if err != nil {
			t.Errorf("\n%q\nConverTo() error: %v", v, err)
			continue
		}

		check := v.check
		if check == "" {
			check = v.input
		}
		if res != check {
			t.Errorf("\nnot equivalent \nin : %q\nres: %q\nchk: %q", v.input, res, check)
			continue
		}
	}
}

// TestTagnameIncorrect -
func TestTagnameIncorrect(t *testing.T) {
	for _, v := range tableTagnameIncorrect {
		_, err := NewFromFilename(v, CheckNormal)
		if err == nil {
			t.Errorf("\n%q\nhas no error", v)
			// fmt.Println("#### unk:", x.GetTags("unktag"))
			// fmt.Println("#### sdhd:", x.GetTags("sdhd"))
			continue
		}
	}
}

// TestTagnameGetCorrect -
func TestTagnameGetCorrect(t *testing.T) {
	for _, v := range tableTagnameGetCorrect {
		tagname, err := NewFromFilename(v.input, CheckNormal)
		if err != nil {
			t.Errorf("\n%q\nNewFromFilename() error:\n%v", v.input, err)
			continue
		}
		for _, v := range v.tags {
			val, err := tagname.GetTag(v.typ)
			if err != nil {
				t.Errorf("\n%q\nGetTag() error: %v", v, err)
				continue
			}
			if val != v.val {
				t.Errorf("\n%q\ngot tag (%v,%v) expected (%v,%v)", tagname.src, v.typ, val, v.typ, v.val)
				continue
			}
		}
	}
}
