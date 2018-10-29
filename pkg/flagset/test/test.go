package main

import (
	"fmt"
	"strings"

	"github.com/macroblock/imed/pkg/flagset"
)

var (
	flagV bool
	flagX bool
	flagY bool
	flagS string
)

func main() {
	flag := flagset.Flag
	cmd := flagset.Command
	Usage := flagset.Usage
	Hint := flagset.Hint

	flagSet := flagset.New("!PROG! is a programm that shows how to use a structured command line arguments.", nil,
		Usage("!PROG! {<flag>} {command {flag}} "),
		Hint("Use \"!PROG! help <command>\" for more information about that topic."),
		flag("-v --version   \000 get version", &flagV),
		flag("-i -x --import \000 import file for further work", &flagV),

		cmd("aaaa \000 tiny aaaa section", nil,
			flag("-x --xxx  \000 get version", &flagX),
			cmd("subsection \000 brief", nil,
				flag("-s --sss \000 string parameter", &flagS),
			),
		),
		cmd("bbbb \000 super-duper b section", nil,
			flag("-y --yyy \000 get version", &flagY),
			cmd("final \000 for test purposes", nil,
				flag("-y --yyy \000 get version", &flagY),
			),
		),
	)

	err := flagSet.PrintHelp()
	if err != nil {
		fmt.Println("error: ", err)
	}

	fmt.Println("\n-----------")

	args := []string{"progName", "--version", "aaaa", "-x", "subsection", "-ss", "data"}
	fmt.Println(strings.Join(args, " "))

	err = flagSet.Parse(args)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("-v:", flagV, "\n-x:", flagX, "\n-y:", flagY, "\n-s:", flagS)
}
