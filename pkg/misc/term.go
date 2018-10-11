package misc

import (
	"os"
	"os/exec"
	"runtime"

	"golang.org/x/crypto/ssh/terminal"
)

// IsTerminal -
func IsTerminal() bool {
	if !terminal.IsTerminal(int(os.Stdout.Fd())) {
		return false
	}
	return true
}

// PauseTerminal -
func PauseTerminal() {
	cmdStr := []string{"read", "-rsp", "Press any key to continue...\n", "-n1", "key"}
	if runtime.GOOS == "windows" {
		cmdStr = []string{"cmd", "/C", "pause"}
	}
	cmd := exec.Command(cmdStr[0], cmdStr[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Run()
}
