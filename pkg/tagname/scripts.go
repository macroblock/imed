package tagname


import (
	// "context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	// "github.com/d5/tengo/v2"
	"github.com/d5/tengo"
	"github.com/d5/tengo/stdlib"
)

var ErrTagnameIsNil = fmt.Errorf("Tagname object is <nil>")

type TScript struct {
	compiled *tengo.Compiled
}

func NewScript(src string) (*TScript, error) {
	script := tengo.NewScript([]byte(src))

	moduleMap := tengo.NewModuleMap()
	moduleMap.AddMap(stdlib.GetModuleMap(stdlib.AllModuleNames()...))

	tnType := newTnModType()
	moduleMap.AddBuiltinModule("imed", tnType.ModuleMap())

	script.SetImports(moduleMap)

	err := script.Add("filename", "###error###")
	if err != nil {
		return nil, err
	}

	// compiled, err := script.RunContext(context.Background())
	compiled, err := script.Compile()
	if err != nil {
		return nil, err
	}
	return &TScript{compiled}, err
}

func (o *TScript) Run(arg string) ([]*TTagname, error) {
	err := o.compiled.Set("filename", arg)
	if err != nil {
		return nil, err
	}
	err = o.compiled.Run()
	if err != nil {
		return nil, err
	}
	v := o.compiled.Get("filename")
	if v == nil {
		panic("!!!")
	}

	if err, ok := v.Value().(error); ok {
		return nil, err
	}

	arr, ok := v.Value().([]interface{})
	if !ok {
		panic("???")
	}
	ret := []*TTagname(nil)

	for _, elem := range arr {
		if elem == nil {
			// panic("here")
			continue
		}
		tn, ok := elem.(*tnType)
		if !ok {
			return nil, fmt.Errorf("one of the returned values is not Tagname")
			// panic(fmt.Sprintf("--- %T", elem))
		}
		if tn == nil || tn.tn == nil {
			continue
		}
		ret = append(ret, tn.tn)
	}
	// panic(fmt.Sprintf("--- %v", len(arr)))
	return ret, err
}

type tnModType struct {
	tengo.ObjectImpl
	moduleMap map[string]tengo.Object
	mtx sync.Mutex
}

func newTnModType() *tnModType {
	ret := &tnModType{}
	ret.moduleMap = map[string]tengo.Object{
		// "xstr": &tengo.UserFunction{Value: ret.customString},
		"tagname": &tengo.UserFunction{Value: ret.fnParse},
		"filebase": &tengo.UserFunction{Value: funcASRS(filepath.Base)},
		"filedir": &tengo.UserFunction{Value: funcASRS(filepath.Dir)},
		"fileext": &tengo.UserFunction{Value: funcASRS(filepath.Ext)},
		// "err": &tengo.UserFunction{Value: ret.fnError},
	}
	return ret
}

func (o *tnModType) TypeName() string {
	return "TagnameModule"
}

func (o *tnModType) String() string {
	v := tengo.ImmutableMap{Value: o.moduleMap}
	return v.String()
}

func (o *tnModType) ModuleMap() map[string]tengo.Object {
	return o.moduleMap
}

func (o *tnModType) fnParse(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 2 {
		return nil, tengo.ErrWrongNumArguments
	}
	val := args[0]
	s1, ok := tengo.ToString(val)
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name: "first",
			Expected: "string(compatible)",
			Found: val.TypeName(),
		}
	}
	val = args[1]
	b2, ok := tengo.ToBool(val)
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name: "second",
			Expected: "bool",
			Found: val.TypeName(),
		}
	}

	tn, err := NewFromFilename(s1, b2)
	tnType := newTnType(tn, err)

	return tnType, nil
}


type tnType struct {
	tengo.ObjectImpl
	tn *TTagname
	err error
}

func newTnType(tn *TTagname, err error) *tnType {
	ret := &tnType{}
	ret.tn = tn
	ret.err = err
	return ret
}

func (o *tnType) TypeName() string {
	return "TagnameType"
}

func (o *tnType) String() string {
	if o == nil {
		panic("tnType.String()")
	}
	return fmt.Sprintf("%v", o.tn)
}

func (o *tnType) IndexGet(index tengo.Object) (tengo.Object, error) {
	val := index
	s, ok := tengo.ToString(val)
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name: "",
			Expected: "string(compatible)",
			Found: val.TypeName(),
		}
	}
	switch strings.ToLower(s) {
	default:
		return nil, fmt.Errorf("cannot call %v for object %v",s, o.TypeName())
	case "err", "error":
		return &tengo.UserFunction{Value: o.fnError}, nil
	case "len":
		return &tengo.UserFunction{Value: o.fnLen}, nil
	case "schema":
		return &tengo.UserFunction{Value: funcRS(o.tn.Schema)}, nil
	case "convertto", "convert_to":
		return &tengo.UserFunction{Value: funcASRSE(o, o.tn.ConvertTo)}, nil
	case "listtags", "list_tags":
		return &tengo.UserFunction{Value: funcRSs(o.tn.ListTags)}, nil
	case "gettag", "get_tag":
		return &tengo.UserFunction{Value: funcASRSE(o, o.tn.GetTag)}, nil
	case "gettags", "get_tags":
		return &tengo.UserFunction{Value: funcASRSs(o.tn.GetTags)}, nil
	case "removetags", "remove_tags":
		return &tengo.UserFunction{Value: funcAS(o.tn.RemoveTags)}, nil
	case "settag", "set_tag":
		return &tengo.UserFunction{Value: funcASS(o.tn.SetTag)}, nil
	case "addtag", "add_tag":
		return &tengo.UserFunction{Value: funcASS(o.tn.AddTag)}, nil
	case "gatherextension", "gatherext":
		return &tengo.UserFunction{Value: funcRSE(o, o.tn.GatherExtension)}, nil
	case "gathersizetag", "gather_sizetag":
		return &tengo.UserFunction{Value: funcRSE(o, o.tn.GatherSizeTag)}, nil
	case "gatheratag", "gather_atag":
		return &tengo.UserFunction{Value: funcRSE(o, o.tn.GatherATag)}, nil
	case "gatherstag", "gather_stag":
		return &tengo.UserFunction{Value: funcRSE(o, o.tn.GatherSTag)}, nil
	case "gathersdhd", "gather_sdhd":
		return &tengo.UserFunction{Value: funcRSE(o, o.tn.GatherSDHD)}, nil
	}
}

func (o *tnType) setError(err error) {
	if err == nil || o == nil {
		return
	}
	o.err = err
}

func (o *tnType) fnError(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 0 {
		return nil, tengo.ErrWrongNumArguments
	}
	if o == nil {
		return wrapError(ErrTagnameIsNil), nil
	}
	return wrapError(o.err), nil
}

func (o *tnType) fnLen(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 0 {
		return nil, tengo.ErrWrongNumArguments
	}
	if o == nil || o.tn == nil {
		return &tengo.Int{Value: -1}, nil
	}
	return &tengo.Int{Value: int64(o.tn.Len())}, nil
}

func (o *tnType) fnCheck(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}
	val := args[0]
	v, ok := tengo.ToBool(val)
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name: "first",
			Expected: "bool",
			Found: val.TypeName(),
		}
	}
	if o == nil {
		return wrapError(ErrTagnameIsNil), nil
	}
	err := o.tn.Check(v)
	o.setError(err)
	return tengo.UndefinedValue, nil
}

func funcRS(fn func() string) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		if len(args) != 0 {
			return nil, tengo.ErrWrongNumArguments
		}
		ret := fn()
		return &tengo.String{Value: ret}, nil
	}
}

func funcAS(fn func(string)) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		if len(args) != 1 {
			return nil, tengo.ErrWrongNumArguments
		}
		val := args[0]
		s1, ok := tengo.ToString(val)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name: "first",
				Expected: "string(compatible)",
				Found: val.TypeName(),
			}
		}
		fn(s1)
		return tengo.UndefinedValue, nil
	}
}

func funcASS(fn func(string, string)) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		if len(args) != 2 {
			return nil, tengo.ErrWrongNumArguments
		}
		val := args[0]
		s1, ok := tengo.ToString(val)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name: "first",
				Expected: "string(compatible)",
				Found: val.TypeName(),
			}
		}
		val = args[1]
		s2, ok := tengo.ToString(val)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name: "second",
				Expected: "string(compatible",
				Found: val.TypeName(),
			}
		}
		fn(s1, s2)
		return tengo.UndefinedValue, nil
	}
}

func funcRSE(o *tnType, fn func() (string, error)) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		if o == nil || o.tn == nil {
			return nil, ErrTagnameIsNil
		}

		if len(args) != 0 {
			return nil, tengo.ErrWrongNumArguments
		}
		ret, err := fn()
		o.setError(err)
		return &tengo.String{Value: ret}, nil
	}
}

func funcASRS(fn func(string) string) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		if len(args) != 1 {
			return nil, tengo.ErrWrongNumArguments
		}
		val := args[0]
		s1, ok := tengo.ToString(val)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name: "first",
				Expected: "string(compatible)",
				Found: val.TypeName(),
			}
		}
		ret := fn(s1)
		return &tengo.String{Value: ret}, nil
	}
}

func funcASRSE(o *tnType, fn func(string) (string, error)) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		if o == nil || o.tn == nil {
			return nil, ErrTagnameIsNil
		}

		if len(args) != 1 {
			return nil, tengo.ErrWrongNumArguments
		}
		val := args[0]
		s1, ok := tengo.ToString(val)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name: "first",
				Expected: "string(compatible)",
				Found: val.TypeName(),
			}
		}
		ret, err := fn(s1)
		o.setError(err)
		return &tengo.String{Value: ret}, nil
	}
}

func funcRSs(fn func() []string) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		if len(args) != 0 {
			return nil, tengo.ErrWrongNumArguments
		}
		arr := &tengo.Array{}
		elems := fn()
		for _, elem := range elems {
			if len(elem) > tengo.MaxStringLen {
				return nil, tengo.ErrStringLimit
			}
			arr.Value = append(arr.Value, &tengo.String{Value: elem})
		}
		return arr, nil
	}
}

func funcASRSs(fn func(string) []string) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		if len(args) != 1 {
			return nil, tengo.ErrWrongNumArguments
		}
		val := args[0]
		s1, ok := tengo.ToString(val)
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name: "first",
				Expected: "string(compatible)",
				Found: val.TypeName(),
			}
		}
		arr := &tengo.Array{}
		elems := fn(s1)
		for _, elem := range elems {
			if len(elem) > tengo.MaxStringLen {
				return nil, tengo.ErrStringLimit
			}
			arr.Value = append(arr.Value, &tengo.String{Value: elem})
		}
		return arr, nil
	}
}

func wrapError(err error) tengo.Object {
	if err == nil {
		// return tengo.TrueValue
		return tengo.UndefinedValue
	}
	return &tengo.Error{Value: &tengo.String{Value: err.Error()}}
}

