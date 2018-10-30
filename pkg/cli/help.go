package cli

import (
	"fmt"
	"strings"

	"github.com/macroblock/imed/pkg/misc"
)

func compLine(prefix, base, postfix string) string {
	if base == "" {
		return ""
	}
	return prefix + base + postfix
}

func formatFlags(elements []Interface, option tOptions, optionText string) string {
	if len(elements) == 0 {
		return ""
	}
	maxKeyStr := 0
	lines := []string{}
	for _, elem := range elements {
		keys := strings.Join(elem.GetKeys(), ", ")
		maxKeyStr = misc.MaxInt(maxKeyStr, len(keys))
		lines = append(lines, keys)
	}
	printOptionText := false
	for i, elem := range elements {
		opt := ' '
		if elem.GetOption(option) {
			printOptionText = true
			opt = '*'
		}
		lines[i] = fmt.Sprintf("  %c %v %v%c %v", opt, lines[i], strings.Repeat(" ", maxKeyStr-len(lines[i])), opt, elem.GetBrief())
	}
	ret := strings.Join(lines, "\n")
	if printOptionText {
		ret += compLine("\n", optionText, "")
	}
	return ret
}

func defaultHelp(o *TCommand, args ...string) error {
	flags := []Interface{}
	sections := []Interface{}
	for _, elem := range o.elements {
		switch t := elem.(type) {
		default:
			return fmt.Errorf("something went wrong")
		case *TFlag:
			flags = append(flags, t)
		case *TCommand:
			sections = append(sections, t)
		}
	}
	text := fmt.Sprintf("%v%v%v%v%v",
		compLine("", o.GetBrief(), "\n"),
		compLine("\nUsage:\n    ", o.GetUsage(), "\n"),
		compLine("\nThe flags are:\n", formatFlags(flags, optTerminator, "* this flag terminates the working flow when it had been processed."), "\n"),
		compLine("\nThe commands are:\n", formatFlags(sections, optInvalid, ""), "\n"),
		compLine("\n", o.GetHint(), "\n"),
	)
	text = strings.Replace(text, "!PROG!", programName, -1)

	fmt.Print(text)

	return nil
}

// PrintHelp -
func (o *TCommand) PrintHelp(args ...string) error {
	return defaultHelp(o, args...)
}
