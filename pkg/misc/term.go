package misc

import (
	"os"
	"os/exec"
	"runtime"
	"strings"

	ansi "github.com/k0kubun/go-ansi"
	"github.com/macroblock/imed/pkg/zlog/zlog"
	"golang.org/x/crypto/ssh/terminal"
)

// TTerminalColor -
type TTerminalColor string

// -
const (
	ColorReset       = TTerminalColor("0")
	ColorBlack       = TTerminalColor("30")
	ColorRed         = TTerminalColor("31")
	ColorGreen       = TTerminalColor("32")
	ColorYellow      = TTerminalColor("33")
	ColorBlue        = TTerminalColor("34")
	ColorPurple      = TTerminalColor("35")
	ColorCyan        = TTerminalColor("36")
	ColorWhite       = TTerminalColor("37")
	ColorLightBlack  = TTerminalColor("30;1")
	ColorLightRed    = TTerminalColor("31;1")
	ColorLightGreen  = TTerminalColor("32;1")
	ColorLightYellow = TTerminalColor("33;1")
	ColorLightBlue   = TTerminalColor("34;1")
	ColorLightPurple = TTerminalColor("35;1")
	ColorLightCyan   = TTerminalColor("36;1")
	ColorLightWhite  = TTerminalColor("37;1")
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

var lastStrLen int

// CPrint -
func CPrint(color TTerminalColor, s string) {
	lastStrLen = len(s)
	ansi.Printf("\x1b[%vm%v\x1b[0m", color, s)
}

// CPrintUndo -
func CPrintUndo() {
	goback := strings.Repeat("\b", lastStrLen)
	clean := strings.Repeat(" ", lastStrLen)
	ansi.Print(goback)
	ansi.Print(clean)
	ansi.Print(goback)
	lastStrLen = 0
}
