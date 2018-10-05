package misc

import (
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
)

// SliceToFile -
func SliceToFile(filePath string, values interface{}) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	err = SliceToWriter(f, values)
	if err != nil {
		return err
	}
	return nil
}

// SliceToWriter -
func SliceToWriter(f io.Writer, values interface{}) error {
	rv := reflect.ValueOf(values)
	if rv.Kind() != reflect.Slice {
		return errors.New("value is not a slice")
	}
	for i := 0; i < rv.Len(); i++ {
		fmt.Fprintln(f, rv.Index(i).Interface())
	}
	return nil
}
