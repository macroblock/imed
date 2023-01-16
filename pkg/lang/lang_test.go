package lang

import (
	"testing"
)

// TestPTool -
func TestTagnameCorrect(t *testing.T) {
	check(t, "Ассирийский", "", "ass")
	check(t, "Гэльский", "", "gae")
	check(t, "Гаэльский", "Гэльский", "gae")
	check(t, "Шотландский", "Гэльский", "gae")
}

func check(t *testing.T, name string, realName string, lat string) {
	if realName == "" {
		realName = name
	}
	if l := ByName(name); l != nil {
		if l.Lat != lat {
			t.Errorf("%q must be %q, not %q\n", name, lat, l.Lat)
		}
	} else {
		t.Errorf("%q not found\n", name)
	}

	if l := ByLat(lat); l != nil {
		if l.Name != realName {
			t.Errorf("%q must be %q, not %q\n", lat, realName, l.Name)
		}
	} else {
		t.Errorf("%q not found\n", lat)
	}
}
