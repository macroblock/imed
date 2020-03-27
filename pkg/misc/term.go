package misc

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	ansi "github.com/k0kubun/go-ansi"
	"github.com/macroblock/imed/pkg/zlog/zlog"
	"golang.org/x/crypto/ssh/terminal"
)

// TTerminalColor -
type TTerminalColor int

// -
const (
	ColorReset TTerminalColor = iota
	ColorBold
	ColorFaint
)

// -
const (
	ColorBlack TTerminalColor = 30 + iota
	ColorRed
	ColorGreen
	ColorYellow
	ColorBlue
	ColorMagenta
	ColorCyan
	ColorWhite
)

// -
const (
	ColorBgBlack TTerminalColor = 40 + iota
	ColorBgRed
	ColorBgGreen
	ColorBgYellow
	ColorBgBlue
	ColorBgMagenta
	ColorBgCyan
	ColorBgWhite
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

// BatchFileExt -
func BatchFileExt() string {
	ret := ".sh"
	if runtime.GOOS == "windows" {
		ret = ".bat"
	}
	return ret
}

// PauseTerminalCmd -
func PauseTerminalCmd() []string {
	ret := []string{"/bin/bash", "-c", "echo Press any key to continue...; read -rs -n 1 key"}

	if runtime.GOOS == "windows" {
		ret = []string{"cmd", "/C", "pause"}
	}
	return ret
}

// PauseTerminalStr -
func PauseTerminalStr() string {
	args := PauseTerminalCmd()
	ret := args[0]
	for _, s := range args[1:] {
		if strings.Contains(s, " ") {
			s = "\"" + s + "\""
		}
		ret += " " + s
	}
	return ret
}

// PauseTerminal -
func PauseTerminal() {
	cmdStr := PauseTerminalCmd()
	cmd := exec.Command(cmdStr[0], cmdStr[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
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
	ansi.Printf("%v%v%v", Color(color), s, Color())
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

// Color -
func Color(colors ...TTerminalColor) string {
	ret := "\033[0"
	for _, c := range colors {
		ret += ";" + strconv.Itoa(int(c))
	}
	ret += "m"
	return ret
}
