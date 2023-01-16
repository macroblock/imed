package tagname

import "testing"

var (
	tableRtSchemaParseCorrect = []tValueCheckSlice{
		//          123456789012345678901234567890
		{inputVal: "sd_2018_sobibor__12_q0w2_ar2_trailer.mpg"},
		{inputVal: "hd_2018_451_gradus_po_farengeytu__q0w0_16_trailer",
			check: "hd_2018_451_gradus_po_farengeytu__16_q0w0_ar2_trailer"},
		{inputVal: "hd_2018_test_name__16_q0w0_ar2_trailer"},
		// {inputVal: "hd_2000_Test_name__16_q0w0_trailerx_trailer"},
		{inputVal: "sd_2000_A__film",
			check: "sd_2000_a__ar2_xy4cS2uPAq0_film"},
		{inputVal: "hd_2000_3d_A__xEORZqiYOl7_film",
			check: "hd_2000_3d_a__ar6_xEORZqiYOl7_film"},
		{inputVal: "sd_2000_b__trailer",
			check: "sd_2000_b__ar2_trailer"},
		{inputVal: "sd_2000_a__ar2_trailer.ext"},
		{inputVal: "sd_2000_a__trailer.ext",
			check: "sd_2000_a__ar2_trailer.ext"},
		{inputVal: "sd_2000_b_s01_01__trailer",
			check: "sd_2000_b_s01_01__ar2_trailer"},
		{inputVal: "sd_2018_the_name_s01_002_zzz_a_comment__q0w0_ar2_trailer"},
		{inputVal: "sd_2018_the_name_s01_002_a_subname_zzz_a_comment__q0w0_ar2_trailer"},
		{inputVal: "sd_2018_the_name_s01_002_a_subname__q0w0_ar2_trailer"},
		{inputVal: "sd_2018_the_name_zzz_a_comment__q0w0_ar2_trailer"},
		{inputVal: "sd_2018_the_name_zzz_a_comment__q0w0_ar2_poster300x600"},
		{inputVal: "sd_2018_the_name_zzz_a_comment__q0w0_ar2_logo"},
		{inputVal: "sd_2018_the_name_zzz_a_comment__q0w0_ar2_poster300-600",
			check: "sd_2018_the_name_zzz_a_comment__q0w0_ar2_poster300x600"},
		// sorting issues
		{inputVal: "hd_1996_arliss_s01_07__ae2_xsmoking_xhardsub_xo6hDsJGYXd_film.mp4",
			check: "hd_1996_arliss_s01_07__ae2_xhardsub_xsmoking_xo6hDsJGYXd_film.mp4"},
		{inputVal: "hd_1996_arliss_s01_07__ae2_PRT123456789012_xsmoking_xhardsub_xo6hDsJGYXd_film.mp4",
			check: "hd_1996_arliss_s01_07__ae2_xhardsub_xsmoking_prt123456789012_xo6hDsJGYXd_film.mp4"},
	}
	tableRtSchemaParseIncorrect = []string{
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
		// "hd_2018_The_name_s01_zzz__q0w0_film",
		// double tags
		"hd_2018_The_name_2018__q0w0_hd_film",
		"hd_2018_The_name_2018__q0w0_sd_film",
		"sd_2018_The_name_2018__hd_q0w0_mdisney_mhardsub_film",
		"hd_2018_The_name_2018__q0w0_mhardsub_q1s3_trailer",
		"sd_2018_The_name_2018__q0w0_prt12345_film",
		"sd_2018_The_name_2018__q0w0_Prt123456789012_film",
		"sd_2018_The_name_prt123456789012_2018__q0w0_film",
		"sd_2018_name_prt123456789012_xxx_2018__q0w0_film",
		// "sd_2018_Sobibor__12_q0w2_trailer.mpg",
	}
)

// TestOldSchemaParseCorrect -
func TestRtSchemaParseCorrect(t *testing.T) {
	parseCorrect(t, "rt", false, tableRtSchemaParseCorrect)
}

// TestOldSchemaParseIncorrect -
func TestRtSchemaParseIncorrect(t *testing.T) {
	parseIncorrect(t, "rt", false, tableRtSchemaParseIncorrect)
}
