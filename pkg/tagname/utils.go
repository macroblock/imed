package tagname

func multiJoin(args ...[]string) []string {
	l := 0
	for i := range args {
		l += len(args[i])
	}
	ret := make([]string, 0, l)
	for i := range args {
		ret = append(ret, args[i]...)
	}
	return ret
}
