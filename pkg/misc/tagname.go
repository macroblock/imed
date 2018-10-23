package misc

import (
	"fmt"
	"strings"

	"github.com/macroblock/imed/pkg/tagname"
)

// StringToCheckLevel -
func StringToCheckLevel(s string) (int, error) {
	ret := 0
	switch strings.ToLower(s) {
	default:
		return ret, fmt.Errorf("unknown check level string: %q", s)
	case "none":
		ret = tagname.CheckNone
	case "normal":
		ret = tagname.CheckNormal
	case "strict":
		ret = tagname.CheckStrict
	case "deep", "deepnormal", "normaldeep":
		ret = tagname.CheckDeepNormal
	case "deepstrict", "strictdeep":
		ret = tagname.CheckDeepNormal
	}
	return ret, nil
}
