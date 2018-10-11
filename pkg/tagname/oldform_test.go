package tagname

import (
	"testing"
)

var (
	tableOldFormParseCorrect = []struct {
		inputVal, check string
	}{
		//23456789012345678901234567890
		{inputVal: "Sobibor_2018__sd_12_q0w2.trailer.mpg"},
		{inputVal: "451_gradus_po_farengeytu_2018__hd_q0w0_16.trailer",
			check: "451_gradus_po_farengeytu_2018__hd_16_q0w0.trailer"},
		{inputVal: "Test_name_2018__hd_16_q0w0"},
		{inputVal: "Test_name_2018_sdok_2000__hd_16_q0w0.trailer"},
		{inputVal: "A_2000"},
		{inputVal: "A_1999_2000"},
		{inputVal: "b_2000__",
			check: "B_2000"},
		{inputVal: "A_2000.trailer.ext"},
		{inputVal: "a_2000__.trailer.ext",
			check: "A_2000.trailer.ext"},
		{inputVal: "b_s01_01_2000__",
			check: "B_s01_01_2000"},
		{inputVal: "The_name_s01_002_zzz_a_comment_2018__hd_q0w0"},
		{inputVal: "The_name_s01_002_a_subname_zzz_a_comment_2018__hd_q0w0"},
		{inputVal: "The_name_s01_002_a_subname_2018__hd_q0w0"},
		{inputVal: "The_name_zzz_a_comment_2018__hd_q0w0"},
		{inputVal: "Vostochnye_zheny_s01_2015__sd_16_q3w0.trailer.mpg"},
	}
	tableOldFormParseIncorrect = []string{
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
		"The_name_2018__sd_q0w0_mdisney_mhardsub",
		"The_name_2018__sd_q0w0_mhardsub_q1s3_trailer",
	}
)

// TestOldFormParseCorrect -
func TestOldFormParseCorrect(t *testing.T) {
	for _, v := range tableOldFormParseCorrect {
		tagname, err := Parse(v.inputVal, "old.normal")
		if err != nil {
			t.Errorf("\n%q\nParse() error: %v", v.inputVal, err)
			continue
		}
		res, err := ToString(tagname, "old.normal", "old.normal")
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

// TestOldFormParseIncorrect -
func TestOldFormParseIncorrect(t *testing.T) {
	for _, v := range tableOldFormParseIncorrect {
		_, err := Parse(v, "old.normal")
		if err == nil {
			t.Errorf("\n%q\nhas no error", v)
			// fmt.Println("#### unk:", x.GetTags("unktag"))
			// fmt.Println("#### sdhd:", x.GetTags("sdhd"))
			continue
		}
	}
}
