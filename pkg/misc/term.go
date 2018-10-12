package misc

import (
	"os"
	"os/exec"
	"runtime"

	"github.com/macroblock/imed/pkg/zlog/zlog"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	log   = zlog.Instance("main")
	retif = log.Catcher()
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

// RunCommand -
func RunCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	// cmd.Stdout =
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), err
	}
	return string(out), nil
	// cmd.Run()
}
