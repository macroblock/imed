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
	flagSet := cli.New("!PROG! is a programm that shows how to use a structured command line arguments.", mainFunc)
	flagSet.Elements(
		cli.Usage("!PROG! {<flag>} {command {flag}}"),
		cli.Hint("Use '!PROG! help <command>' for more information about that topic."),
		cli.Flag("-v --version   : get version", versionFunc), //.Terminator(),
		cli.Flag("-i -x --import : import file for further work", &flagV),

		cli.Command("aaaa  : tiny aaaa section", func() error { fmt.Println("=== aaaa called"); return nil },
			cli.Flag("-x --xxx  : get version", &flagX),
			cli.Command("subsection : brief", func() error { fmt.Println("=== aaaa subsection"); return nil },
				cli.Flag("-s --sss : string parameter", &flagS),
			),
		),
		cli.Command("bbbb : super-duper b section", nil,
			cli.Flag("-y --yyy : get version", &flagY),
			cli.Command("final : for test purposes", nil,
				cli.Flag("-y --yyy : get version", &flagY),
			),
		),
		cli.OnError("run !PROG! -h for more information"),
	)

	err := flagSet.PrintHelp()
	if err != nil {
		fmt.Println("print help error: ", err)
	}

	fmt.Println("\n-----------")

	args := []string{"progName", "--version", "aaaa", "-x", "subsection", "-s", "data", "xxxx"}
	fmt.Println(strings.Join(args, " "))

	err = flagSet.Parse(args)
	if err != nil {
		fmt.Println("parse error: ", err)
		fmt.Println("hint: ", flagSet.GetHint())
	}
	fmt.Println("-v:", flagV, "\n-x:", flagX, "\n-y:", flagY, "\n-s:", flagS)
}
