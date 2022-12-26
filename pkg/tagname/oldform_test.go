package tagname

import (
	"testing"
)

type tValueCheckSlice struct {
	inputVal string
	check    string
}

var (
	tableOldSchemaParseCorrect = []tValueCheckSlice{
		//23456789012345678901234567890
		{inputVal: "sobibor_2018__sd_12_q0w2_ar2.trailer.mpg"},
		{inputVal: "451_gradus_po_farengeytu_2018__hd_q0w0_16.trailer",
			check: "451_gradus_po_farengeytu_2018__hd_16_q0w0_ar2.trailer"},
		{inputVal: "test_name_2018__hd_16_q0w0_ar6"},
		{inputVal: "test_name_2018_hdok_2000__hd_16_q0w0_ar2.trailer"},
		{inputVal: "a_2000__sd_ar2"},
		{inputVal: "a_1999_2000__hd_ar6"},
		{inputVal: "b_2000__3d",
			check: "b_2000__3d_ar6"},
		{inputVal: "a_2000__sd_ar2.trailer.ext"},
		{inputVal: "a_2000__q0w0_sd.trailer.ext",
			check: "a_2000__sd_q0w0_ar2.trailer.ext"},
		{inputVal: "b_s01_01_2000__hd_x1234567890",
			check: "b_s01_01_2000__hd_ar6"},
		{inputVal: "the_name_s01_002_zzz_a_comment_2018__hd_q0w0_ar6"},
		{inputVal: "the_name_s01_002_a_subname_zzz_a_comment_2018__hd_q0w0_achn6"},
		{inputVal: "the_name_s01_002_a_subname_2018__hd_q0w0_aesp6"},
		{inputVal: "the_name_zzz_a_comment_2018__hd_q0w0_aqqq6"},
		{inputVal: "vostochnye_zheny_s01_2015__sd_16_q3w0_ar2.trailer.mpg"},
		{inputVal: "xxx_s01_01_2000__300x400.jpg",
			check: "xxx_s01_01_2000__300x400.jpg"},
		{inputVal: "xxx_s01_01_2000__center_300x400_sbs_mxxx.jpg",
			check: "xxx_s01_01_2000__sbs_mxxx_300x400_center.jpg"},
		{inputVal: "b_s01_01_2000__hd_x1234567890_PRT123456789012_mxxx",
			check: "b_s01_01_2000__hd_ar6_mxxx_prt123456789012"},
	}
	tableOldSchemaParseIncorrect = []string{
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
		// "The_name_s01_zzz_2018__hd_q0w0",
		// double tags
		"The_name_2018__hd_q0w0_hd",
		"The_name_2018__sd_q0w0_hd",
		"The_name_2018__sd_q0w0_mdisney_film.trailer",
		"The_name_2018__sd_q0w0_mhardsub_q1s3.trailer",
		"The_name_2018__sd_q0w0_prt12345",
		"The_name_2018__sd_q0w0_Prt123456789012",
		// "Sobibor_2018__sd_12_q0w2.trailer.mpg",
	}
)

func parseCorrect(t *testing.T, schemaName string, isStrictLevel bool, table []tValueCheckSlice) {
	for _, v := range table {
		tagname, err := Parse(v.inputVal, schemaName)
		if err != nil {
			t.Errorf("\n%q\nParse() error: %v", v.inputVal, err)
			continue
		}
		// err = tagname.Check(isStrictLevel)
		schema, err := Schema(schemaName)
		if err != nil {
			t.Errorf("\n%q\nSchema() error: %v", v.inputVal, err)
			continue
		}

		tn, err := TranslateTags(tagname, schema.UnmarshallFilter)
		if err != nil {
			t.Errorf("\n%q\nTranslateTags() error: %v", v.inputVal, err)
			continue
		}
		tagname = tn

		err = CheckTags(tagname) //, schema)
		if err != nil {
			t.Errorf("\n%q\nCheck() error: %v", v.inputVal, err)
			continue
		}
		res, err := ToString(tagname, schemaName)
		if err != nil {
			t.Errorf("\n%q\nToString() error: %v", v, err)
			continue
		}
		check := v.check
		if check == "" {
			check = v.inputVal
		}
		if res != check {
			t.Errorf("\nnot equivalent \nin : %q\nres: %q\nchk: %q", v.inputVal, res, check)
			t.Errorf("tags: %v", tagname)
			continue
		}
	}
}

func parseIncorrect(t *testing.T, schemaName string, isStrictLevel bool, table []string) {
	for _, v := range table {
		tn, err := Parse(v, schemaName)
		if err != nil {
			continue
		}
		// err = tn.Check(isStrictLevel)

		schema, err := Schema(schemaName)
		if err != nil {
			t.Errorf("\n%q\nSchema() error: %v", v, err)
			continue
		}

		tags, err := TranslateTags(tn, schema.UnmarshallFilter)
		if err != nil {
			// t.Errorf("\n%q\nTranslateTags() error: %v", v, err)
			continue
		}
		tn = tags

		err = CheckTags(tn) //, schema)
		if err != nil {
			continue
		}
		t.Errorf("\n%q\nhas no error", v)
		// fmt.Println("#### unk:", x.GetTags("unktag"))
		// fmt.Println("#### sdhd:", x.GetTags("sdhd"))
	}
}

// TestOldSchemaParseCorrect -
func TestOldSchemaParseCorrect(t *testing.T) {
	parseCorrect(t, "old", false, tableOldSchemaParseCorrect)
}

// TestOldSchemaParseIncorrect -
func TestOldSchemaParseIncorrect(t *testing.T) {
	parseIncorrect(t, "old", false, tableOldSchemaParseIncorrect)
}
