package main

import (
	"fmt"
	"strings"

	"github.com/macroblock/imed/pkg/cli"
)

var (
	flagV bool
	flagX bool
	flagY bool
	flagS string
)

func mainFunc() error {
	fmt.Println("=== mainFunc called")
	return nil
}

func versionFunc() error {
	fmt.Println("=== versionFunc called")
	return nil
}

func main() {
	flag := cli.Flag
	cmd := cli.Command
	Usage := cli.Usage
	Hint := cli.Hint

	_ = flag

	flagSet := cli.New("!PROG! is a programm that shows how to use a structured command line arguments.", mainFunc)
	flagSet.Elements(
		Usage("!PROG! {<flag>} {command {flag}}"),
		Hint("Use '!PROG! help <command>' for more information about that topic."),
		flag("-v --version   : get version", versionFunc).Terminator(),
		flag("-i -x --import : import file for further work", &flagV),

		cmd("aaaa  : tiny aaaa section", nil,
			flag("-x --xxx  : get version", &flagX),
			cmd("subsection : brief", nil,
				flag("-s --sss : string parameter", &flagS),
			),
		),
		cmd("bbbb : super-duper b section", nil,
			flag("-y --yyy : get version", &flagY),
			cmd("final : for test purposes", nil,
				flag("-y --yyy : get version", &flagY),
			),
		),
	)

	err := flagSet.PrintHelp()
	if err != nil {
		fmt.Println("error: ", err)
	}

	fmt.Println("\n-----------")

	args := []string{"progName", "--version", "aaaa", "-x", "subsection", "-s", "data"}
	fmt.Println(strings.Join(args, " "))

	err = flagSet.Parse(args)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("-v:", flagV, "\n-x:", flagX, "\n-y:", flagY, "\n-s:", flagS)
}
