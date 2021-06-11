package ptool

import (
	"fmt"
	"reflect"
	"strings"
)

func inspectPreOrder(root *TNode, fn func(*TNode) (bool, error)) error {
	ok, err := fn(root)
	if err != nil || !ok {
		return err
	}
	for i := range root.Links {
		err = inspectPreOrder(root.Links[i], fn)
		if err != nil {
			return err
		}
	}
	return nil
}

// func rebuildBranch(root *TNode, lut map[string]*tLutItem) error {
// 	switch root.Type {
// 	case cIdent:
// 		item, ok := lut[root.Value]
// 		if !ok {
// 			return fmt.Errorf("undefined statement %q", root.Value)
// 		}
// 		*root = *item.node.Links[1]
// 	}
// 	for i := range root.Links {
// 		err := rebuildBranch(root.Links[i], lut)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// func optimizeZBNFTree(root *TNode, entries ...string) error {
// 	lut := buildLUT(root)
// 	for _, entry := range entries {
// 		s := fmt.Sprintf("entry %v", entry)
// 		item, ok := lut[entry]
// 		if !ok {
// 			return fmt.Errorf("optimize: undefined statement %q", entry)
// 		}
// 		fmt.Println(s)
// 		//fmt.Println(item.node)

// 		err := rebuildBranch(item.node.Links[1], lut)
// 		if err != nil {
// 			return err
// 		}
// 		//fmt.Println(item.node)
// 	}
// 	return nil
// }

func print(code []TInstruction) {
	line := ""
	for _, instr := range code {
		switch instr.opcode {
		default:
			line = fmt.Sprintf("\t%v", instr.opcode)
			if instr.data != nil {
				line = fmt.Sprintf("%v %v", line, instr.data)
			}
		case opCHECKRUNE, opCHECKSTR:
			line = fmt.Sprintf("\t%v %q", instr.opcode, instr.data)
		case opCHECKRANGE:
			data := instr.data.([2]rune)
			line = fmt.Sprintf("\t%v %q..%q", instr.opcode, data[0], data[1])
		case opLABEL:
			line = fmt.Sprintf("%v", instr.data)
		case opNOP:
			line = fmt.Sprintf("    %v", instr.opcode)
		}
		fmt.Println(line)
	}
}

// TreeToString -
func TreeToString(root *TNode, fn func(int) string) string {
	if root == nil {
		return "<nil>"
	}
	s := root.CustomString(func(node *TNode) string {
		return fmt.Sprintf("%3v %-8v %q", node.Type, fn(node.Type), node.Value)
	})
	return s
}

// Unmarshall -
func Unmarshall(root *TNode, byID func(int) string, obj interface{}) error {
	if root == nil {
		return nil
	}
	errors := []string{}
	for idx := range root.Links {
		name := byID(root.Links[idx].Type)
		val := root.Links[idx].Value
		// setStruct(obj, strings.Title(name), val)
		ok, err := SetStructField(obj, strings.Title(name), val)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%q: %v", strings.Title(name), err))
			continue
		}
		if !ok {
			errors = append(errors, fmt.Sprintf("can't set field %q", strings.Title(name)))
			continue
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("unmarshall() error(s):\n  %v\n", strings.Join(errors, "\n  "))
	}

	return nil
}

// SetStructField -
func SetStructField(obj interface{}, name, value string) (bool, error) {
	val := reflect.ValueOf(obj)
	if val.Kind() != reflect.Ptr {
		return false, fmt.Errorf("obj is unaddressable value")
	}
	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return false, fmt.Errorf("obj is not a struct")
	}
	v := val.FieldByName(name)
	if !v.IsValid() || !v.CanSet() { // not found
		return false, nil
	}
	switch v.Kind() {
	default:
		return false, fmt.Errorf("unsupported type %v", v.Kind())
	case reflect.String:
		v.SetString(value)
	// case reflect.Slice:
	}

	return true, nil
}

func setStruct(ob interface{}, name, value string) error {
	val := reflect.ValueOf(ob).Elem()
	typ := val.Type()
	for idx := 0; idx < val.NumField(); idx++ {
		v := val.Field(idx)
		t := typ.Field(idx)

		name1 := name
		name2 := t.Name

		// tag := t.Tag.Get("ptool")
		// if tag != "lowercase" {
		// 	name1 = strings.ToLower(name1)
		// 	name2 = strings.ToLower(name2)
		// }
		// fmt.Printf("%q == %q\n", name1, name2)
		if name1 != name2 {
			continue
		}
		switch v.Kind() {
		default:
			fmt.Printf("unsupported type %s\n", v.Kind())
		case reflect.String:
			// value := v.String()
			// fmt.Printf("string %q\n", t.Name)
			v.SetString(value)
			// if tag == "lowercase" {
			// 	value = strings.ToUpper(value)
			// 	v.SetString(value)
			// }
		}
	}
	return nil
}

