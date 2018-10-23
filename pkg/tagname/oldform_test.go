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
		{inputVal: "Sobibor_2018__sd_12_q0w2.trailer.mpg"},
		{inputVal: "451_gradus_po_farengeytu_2018__hd_q0w0_16.trailer",
			check: "451_gradus_po_farengeytu_2018__hd_16_q0w0.trailer"},
		{inputVal: "Test_name_2018__hd_16_q0w0"},
		{inputVal: "Test_name_2018_sdok_2000__hd_16_q0w0.trailer"},
		{inputVal: "A_2000__sd"},
		{inputVal: "A_1999_2000__hd"},
		{inputVal: "b_2000__3d",
			check: "B_2000__3d"},
		{inputVal: "A_2000__sd.trailer.ext"},
		{inputVal: "a_2000__q0w0_sd.trailer.ext",
			check: "A_2000__sd_q0w0.trailer.ext"},
		{inputVal: "b_s01_01_2000__hd",
			check: "B_s01_01_2000__hd"},
		{inputVal: "The_name_s01_002_zzz_a_comment_2018__hd_q0w0"},
		{inputVal: "The_name_s01_002_a_subname_zzz_a_comment_2018__hd_q0w0"},
		{inputVal: "The_name_s01_002_a_subname_2018__hd_q0w0"},
		{inputVal: "The_name_zzz_a_comment_2018__hd_q0w0"},
		{inputVal: "Vostochnye_zheny_s01_2015__sd_16_q3w0.trailer.mpg"},
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
		"The_name_s01_zzz_2018__hd_q0w0",
		// double tags
		"The_name_2018__hd_q0w0_hd",
		"The_name_2018__sd_q0w0_hd",
		"The_name_2018__sd_q0w0_mdisney_film.trailer",
		"The_name_2018__sd_q0w0_mhardsub_q1s3.trailer",
	}
)

func parseCorrect(t *testing.T, schemaName string, isStrictLevel bool, table []tValueCheckSlice) {
	for _, v := range table {
		tagname, err := Parse(v.inputVal, schemaName)
		if err != nil {
			t.Errorf("\n%q\nParse() error: %v", v.inputVal, err)
			continue
		}
		err = tagname.Check(isStrictLevel)
		if err != nil {
			t.Errorf("\n%q\nCheck() error: %v", v.inputVal, err)
			continue
		}
		res, err := ToString(tagname, schemaName, schemaName)
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
		err = tn.Check(isStrictLevel)
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
