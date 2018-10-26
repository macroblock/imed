package main

import (
	"fmt"
	"strings"

	"github.com/macroblock/imed/pkg/zflag"
)

var (
	flagV bool
	flagX bool
	flagY bool
	flagS string
)

func main() {
	flagSet := zflag.NewSection("", nil,
		"!progname! is a programm that shows how to use a structured command line arguments.",
		"!progname! {<flag>} {command {flag}} ",
		"Use \"!progname! help <command>\" for more information about that topic.",
		"",
		zflag.New("-v --version", &flagV, nil, "get version", "", "hint", ""),
		zflag.New("-i -x --import", &flagV, nil, "import file for further work", "????", "hint", ""),
		zflag.NewSection("aaaa", nil,
			"tiny aaaa section",
			"usage",
			"hint",
			"",
			zflag.New("-x --xxx", &flagX, nil, "get version", "", "hint", ""),
			zflag.NewSection("subsection", nil,
				"brief",
				"usage",
				"hint",
				"",
				zflag.New("-s --sss", &flagS, nil, "string parameter", "", "hint", ""),
			),
		),
		zflag.NewSection("bbbb", nil,
			"super-duper b section",
			"usage",
			"hint",
			"",
			zflag.New("-y --yyy", &flagY, nil, "get version", "", "hint", ""),
		),
		zflag.NewSection("final", nil,
			"for test purposes",
			"usage",
			"hint",
			"",
			zflag.New("-y --yyy", &flagY, nil, "get version", "", "hint", ""),
		),
	)

	err := flagSet.PrintHelp()
	if err != nil {
		fmt.Println("error: ", err)
	}

	fmt.Println("\n-----------")

	args := []string{"progName", "--version", "aaaa", "-x", "subsection", "-s", "data"}
	fmt.Println(strings.Join(args, " "))

	err = flagSet.Parse(&args)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("-v:", flagV, "\n-x:", flagX, "\n-y:", flagY, "\n-s:", flagS)
}
