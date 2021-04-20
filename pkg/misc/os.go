package misc

import (
	"fmt"
	"os/exec"
	"strings"
)

func CommandExists(names ...string) error {
	list := []string{}
	for _, name := range names {
		_, err := exec.LookPath(name)
		if err != nil {
			list = append(list, name)
		}
	}
	return fmt.Errorf("command(s) not found: %v", strings.Join(list, ", "))
}
