package cli

import (
	"strings"
)

// Text -
func Text(text string) string {
	text = strings.Replace(text, "!PROG!", programName, -1)
	return text
}

func initElements(o Interface, elements ...IElement) {
	flag := &TFlag{}
	cmd := &TCommand{}
	switch t := o.(type) {
	default:
		internalPanic("something went wrong")
	case *TFlag:
		flag = t
	case *TCommand:
		cmd = t
		flag = &t.TFlag
	}
	for _, elem := range elements {
		text := ""
		strPtr := (*string)(nil)
		switch t := elem.(type) {
		default:
			internalPanicf("intiElements(any) got unsupported type of element %T", t)
		case *tErrorHandler:
			flag.onError = t
		case *tUsage:
			strPtr = &flag.usage
			text = t.text
		case *tHint:
			strPtr = &flag.hint
			text = t.text
		case *tDoc:
			strPtr = &flag.doc
			text = t.text
		case Interface:
			if cmd == nil {
				internalPanicf("initElements(flag) got unsupported type of element %T", t)
			}
			cmd.elements = append(cmd.elements, t)
		}
		if strPtr != nil {
			*strPtr = compLine("", *strPtr, "\n") + text
		}
	}
}
