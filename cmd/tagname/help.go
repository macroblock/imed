package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
)

func haveArg(entry tEntry) bool {
	return entry.arg != nil
}

func callArg(entry tEntry) {
	if entry.arg != nil {
		entry.arg()
	}
}

func callCmd(command string, entry tEntry) {
	if entry.cmd != nil {
		entry.cmd(command, entry)
	}
}

func printEntryHelp(entry tEntry) {
	fmt.Println()
	if entry.brief != "" {
		fmt.Printf("%v\n", entry.brief)
	}
	if entry.usage != "" {
		fmt.Printf("\nUsage: %v\n", entry.usage)
	}
	if entry.argList != "" {
		s := "    " + strings.Replace(entry.argList, "\n", "\n    ", -1)
		fmt.Printf("\n%v\n", s)
	}
	if haveArg(entry) {
		fmt.Println()
		flag.PrintDefaults()
	}
	if entry.hint != "" {
		fmt.Printf("\n%v\n", entry.hint)
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func initBasicHelpArgs() {
	cmdList := []string{}
	maxW := 0
	for key := range entries {
		if key != "" {
			maxW = maxInt(maxW, len(key))
			cmdList = append(cmdList, key)
		}
	}
	sort.Strings(cmdList)
	argList := []string{}
	frmt := fmt.Sprintf("%%-%vv   %%v", maxW)
	for _, s := range cmdList {
		e := entries[s]
		argList = append(argList, fmt.Sprintf(frmt, s, e.brief))
	}
	e, ok := entries[""]
	if !ok {
		panic("unreachable")
	}

	e.argList = strings.Join(argList, "\n")
	entries[""] = e
}

func cmdBasicHelp(command string, entry tEntry) {
	callArg(entry)
	printEntryHelp(entry)
}

func cmdHelp(command string, entry tEntry) {
	args := os.Args[1:]
	argc := len(args)
	if argc > 1 {
		fmt.Printf("Too many arguments given.\n")
		return
	}
	command = ""
	if argc == 1 {
		command = args[0]
	}
	ok := false
	entry, ok = entries[command]
	if !ok {
		fmt.Printf("Unknown help topic %q. Run '%v help'.\n", command, progName)
		return
	}
	callArg(entry)
	printEntryHelp(entry)
}
