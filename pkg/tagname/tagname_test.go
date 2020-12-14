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
			input: "//test\\path/Sobibor_2018__sd_12_q0w2.trailer.mpg",
			check: "//test\\path/sd_2018_sobibor__12_q0w2_ar2_trailer.mpg"},
		{settings: "old",
			input: "test/sd_2018_Sobibor__12_q0w2_trailer.mpg",
			check: "test/sobibor_2018__sd_12_q0w2_ar2.trailer.mpg"},
		{settings: "",
			input: "sd_2018_sobibor__12_q0w2_ar2_trailer.mpg"},
		{settings: "",
			input: "sobibor_2018__sd_12_q0w2_ar2.trailer.mpg"},
		{settings: "rt",
			input: "Sobibor_2018__3d_12_q0w2.trailer.mpg",
			check: "hd_2018_3d_sobibor__12_q0w2_ar2_trailer.mpg"},
		{settings: "old",
			input: "hd_2018_3d_Sobibor__12_q0w2_trailer.mpg",
			check: "sobibor_2018__3d_12_q0w2_ar2.trailer.mpg"},
		{settings: "",
			input: "hd_2018_3d_sobibor__12_q0w2_ar2_trailer.mpg"},
		{settings: "",
			input: "sobibor_2018__3d_12_q0w2_ar2.trailer.mpg"},
		{settings: "",
			input: "rrrrr_2018__3d_12_q0w2_mkurazhbambey_ar2.trailer.mpg",
			check: "rrrrr_2018__3d_12_q0w2_ar2_vkurazhbambey.trailer.mpg"},
		{settings: "",
			input: "rrrrr_2018__3d_12_q0w2_pozitiv_ar2.trailer.mpg",
			check: "rrrrr_2018__3d_12_q0w2_ar2_vpozitiv.trailer.mpg"},
		{settings: "",
			input: "qqqq_2018__12_q0w2_HD_ar2.trailer.mpg",
			check: "qqqq_2018__hd_12_q0w2_ar2.trailer.mpg"},
		{settings: "",
			input: "zvezda_rodilas_2018__hd_190x230.poster.jpg"},
		{settings: "",
			input: "hd_2018_bezumno_bogatye_aziaty__mpr_poster525x300.jpg"},
		{settings: "rt",
			input: "//test/path/Babnik_2008__hd_1620-996.poster.jpg",
			check: "//test/path/hd_2008_babnik__poster1620x996.jpg"},
		{settings: "old",
			input: "//test/path/sd_2018_Proigrannoe_mesto__mpryamoiz_mtest_poster525x300.jpg",
			check: "//test/path/proigrannoe_mesto_2018__sd_mpryamoiz_mtest_525x300.poster.jpg"},
		{settings: "old",
			input: "//test/path/sd_2018_Proigrannoe_mesto__logo.jpg",
			check: "//test/path/proigrannoe_mesto_2018__sd_logo.poster.jpg"},
		{settings: "rt",
			input: "//test/path/proigrannoe_mesto_2018__sd_logo.poster.jpg",
			check: "//test/path/sd_2018_proigrannoe_mesto__logo.jpg"},
		// {settings: "old",
			// input: "Proigrannoe_mesto_2018__sd_mpryamoiz_mtest_525x300_poster.jpg",
			// check: "proigrannoe_mesto_2018__sd_mpryamoiz_mtest_525x300.poster.jpg"},
		// {settings: "rt",
		// input: "Vse_elki_2018__sd_1140-726.poster#203b17b5.jpg",
		// check: "sd_2018_Vse_elki__poster1140x726#203b17b5.jpg"},
		// {settings: "old",
		// input: "sd_2018_Vse_elki__poster1140x726#203b17b5.jpg",
		// check: "Vse_elki_2018__sd_1140-726.poster#203b17b5.jpg"},
		{settings: "old",
			input: "//test/path/Proigrannoe_mesto_2018__525x300.jpg",
			check: "//test/path/proigrannoe_mesto_2018__525x300.jpg"},
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
		// "sd_2018_Sobibor__12_q0w2_trailer.mpg",
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
		{input: "hd_2018_3d_The_name_sXX_a_sname_01b_a_ename_zzz_comment__q0w0_amed_film",
			tags: []ttag{
				{typ: "name", val: "the_name"},
				// {typ: "snen", val: "s01e01_a_subname"},
				{typ: "sxx", val: "sXX"},
				{typ: "exx", val: "01b"},
				{typ: "sname", val: "a_sname"},
				{typ: "ename", val: "a_ename"},
				{typ: "comment", val: "zzz_comment"},
				{typ: "sdhd", val: "3d"},
				{typ: "mtag", val: "mamed"},
				{typ: "qtag", val: "q0w0"},
				{typ: "type", val: "film"},
				{typ: "ext", val: ""},
			},
		},
		{input: "The_xxx_name_s01_2015__sd_16_q3w0_xhardsub.trailer.mpg",
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
				{typ: "hardsubtag", val: "xhardsub"},
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
				{typ: "sizetag", val: "1620x996"},
				{typ: "type", val: "poster"},
				{typ: "ext", val: ".jpg"},
			},
		},
		{input: "babnik_gp_2008__1620x996.jpg",
			tags: []ttag{
				{typ: "name", val: "babnik_gp"},
				{typ: "year", val: "2008"},
				{typ: "sizetag", val: "1620x996"},
				{typ: "type", val: "poster.gp"},
				{typ: "ext", val: ".jpg"},
			},
		},
		{input: "babnik_3000_gp_2008_sd_logo.poster.jpg",
			tags: []ttag{
				{typ: "name", val: "babnik_3000_gp"},
				{typ: "year", val: "2008"},
				{typ: "sizetag", val: "logo"},
				{typ: "type", val: "poster"},
				{typ: "ext", val: ".jpg"},
			},
		},
		{input: "babnik_gp_4000_sxx_x_5000_01_y_6000_zzz_7000_2008_1620x1996.jpg",
			tags: []ttag{
				{typ: "name", val: "babnik_gp_4000"},
				{typ: "sxx", val: "sxx"},
				{typ: "sname", val: "x_5000"},
				{typ: "exx", val: "01"},
				{typ: "ename", val: "y_6000"},
				{typ: "comment", val: "zzz_7000"},
				{typ: "year", val: "2008"},
				{typ: "sizetag", val: "1620x1996"},
				{typ: "type", val: "poster.gp"},
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
			t.Errorf("\n%q\nConvertTo() error: %v", v, err)
			continue
		}

		check := v.check
		if check == "" {
			check = v.input
		}
		if res != check {
			t.Errorf("\nnot equivalent \nin : %q\nres: %q\nchk: %q", v.input, res, check)
			t.Errorf("srcTags: %v", tagname.srcTags)
			t.Errorf("tags: %v", tagname.tags)
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
