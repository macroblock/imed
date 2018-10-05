package utils

type idestroyer interface {
	Destroy()
}

// func destroy(o interface{}) {
// 	// if o == nil {
// 	// 	log.Warning(true, "destroy: can't get address of argument <nil>")
// 	// 	return
// 	// }
// 	switch v := o.(type) {
// 	case nil:
// 	case idestroyer:
// 		v.Destroy()
// 	default:
// 		log.Warning(true, "destroy: variable hasn't got Destroy() method")
// 	}
// 	o = nil
// }

// Abs -
func Abs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}

// Min -
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Max -
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// type TLine {x1,y1,x2,y2 int}  //линии??
