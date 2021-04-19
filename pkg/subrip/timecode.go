package subrip

import (
	"strings"
	"fmt"

	"github.com/macroblock/imed/pkg/types"
)

type Timecode = types.Timecode

func NewTimecode(h, m, s float64) Timecode {
	return Timecode(types.NewTimecode(h, m, s))
}

func ParseTimecode(in string) (Timecode, error) {
	s := strings.ReplaceAll(in, ",", ".")
	ret, err := types.ParseTimecode(s)
	if err != nil {
		err = fmt.Errorf(
			strings.Replace(err.Error(), s, in, 1),
		)
		return ret, err
	}
	return ret, nil
}
