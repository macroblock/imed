package tagname

import "testing"

var (
	tableRtFormParseCorrect = []struct {
		inputVal, check string
	}{
		//          123456789012345678901234567890
		{inputVal: "sd_2018_Sobibor__12_q0w2_trailer.mpg"},
		{inputVal: "hd_2018_451_gradus_po_farengeytu__q0w0_16_trailer",
			check: "hd_2018_451_gradus_po_farengeytu__16_q0w0_trailer"},
		{inputVal: "hd_2018_Test_name__16_q0w0_trailer"},
		{inputVal: "hd_2000_Test_name__16_q0w0_trailerx_trailer"},
		{inputVal: "sd_2000_A__film"},
		{inputVal: "hd_2000_3d_A__film"},
		{inputVal: "sd_2000_b__film",
			check: "sd_2000_B__film"},
		{inputVal: "sd_2000_A__trailer.ext"},
		{inputVal: "sd_2000_a__trailer.ext",
			check: "sd_2000_A__trailer.ext"},
		{inputVal: "sd_2000_b_s01_01__film",
			check: "sd_2000_B_s01_01__film"},
		{inputVal: "sd_2018_The_name_s01_002_zzz_a_comment__q0w0_film"},
		{inputVal: "sd_2018_The_name_s01_002_a_subname_zzz_a_comment__q0w0_film"},
		{inputVal: "sd_2018_The_name_s01_002_a_subname__q0w0_film"},
		{inputVal: "sd_2018_The_name_zzz_a_comment__q0w0_film"},
	}
	tableRtFormParseIncorrect = []string{
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
		"hd_2018_The_name_s01_a_subname__q0w0",
		"hd_2018_The_name_s01_zzz__q0w0_film",
		// double tags
		"hd_2018_The_name_2018__q0w0_hd_film",
		"hd_2018_The_name_2018__q0w0_sd_film",
		"sd_2018_The_name_2018__hd_q0w0_mdisney_mhardsub_film",
		"hd_2018_The_name_2018__q0w0_mhardsub_q1s3_trailer",
	}
)

// TestOldFormParseCorrect -
func TestRtFormParseCorrect(t *testing.T) {
	for _, v := range tableRtFormParseCorrect {
		tagname, err := Parse(v.inputVal, "rt.normal")
		if err != nil {
			t.Errorf("\n%q\nParse() error: %v", v.inputVal, err)
			continue
		}
		res, err := ToString(tagname, "rt.normal", "rt.normal")
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
func TestRtFormParseIncorrect(t *testing.T) {
	for _, v := range tableRtFormParseIncorrect {
		_, err := Parse(v, "rt.normal")
		if err == nil {
			t.Errorf("\n%q\nhas no error", v)
			continue
		}
	}
}
