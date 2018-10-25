package zflag

// TFlagBuilder -
type TFlagBuilder struct {
	flag TFlag
}

// // New -
// func New(keys ...string) *TFlagBuilder {
// 	ret := &TFlagBuilder{}
// 	for _, key := range keys {
// 		if key != "" {
// 			ret.flag.keys = append(ret.flag.keys, key)
// 		}
// 	}
// 	log.Warning(len(ret.flag.keys) == 0, "flag without keys")
// 	return ret
// }

// Done -
// func (o *TFlagBuilder) Done() *TFlag {
// 	ret := o.flag
// 	return &ret
// }

// // Command -
// func (o *TFlagBuilder) Command(fn func()) *TFlagBuilder {
// 	o.flag.cmd = fn
// 	return o
// }

// // Brief -
// func (o *TFlagBuilder) Brief(text string) *TFlagBuilder {
// 	o.flag.brief = text
// 	return o
// }

// // Usage -
// func (o *TFlagBuilder) Usage(text string) *TFlagBuilder {
// 	o.flag.usage = text
// 	return o
// }

// // Hint -
// func (o *TFlagBuilder) Hint(text string) *TFlagBuilder {
// 	o.flag.hint = text
// 	return o
// }
