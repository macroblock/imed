package ffmpeg

import (
	"fmt"
	"strconv"
	"strings"
)

// TArgMap -
type TArgMap struct {
	table map[string]string
	err   error
}

// NewArgMap -
func NewArgMap() *TArgMap {
	return &TArgMap{table: map[string]string{}}
}

// Add -
func (o *TArgMap) Add(name, val string) {
	o.table[name] = val
}

// Error -
func (o *TArgMap) Error() error {
	return o.err
}

// GetS -
func (o *TArgMap) GetS(name, suffix string, val *string) {
	if o.err != nil {
		return
	}
	if s, ok := o.table[name]; ok {
		v, err := valToS(s, suffix)
		if err != nil {
			o.err = err
			return
		}
		if val != nil {
			*val = v
		}
		return
	}
	o.err = fmt.Errorf("%q not found", name)
}

// GetI -
func (o *TArgMap) GetI(name, suffix string, val *int) {
	if o.err != nil {
		return
	}
	if s, ok := o.table[name]; ok {
		v, err := valToI(s, suffix)
		if err != nil {
			o.err = err
			return
		}
		if val != nil {
			*val = v
		}
		return
	}
	o.err = fmt.Errorf("%q not found", name)
}

// GetF -
func (o *TArgMap) GetF(name, suffix string, val *float64) {
	if o.err != nil {
		return
	}
	if s, ok := o.table[name]; ok {
		v, err := valToF(s, suffix)
		if err != nil {
			o.err = err
			return
		}
		if val != nil {
			*val = v
		}
		return
	}
	o.err = fmt.Errorf("%q not found", name)
}

func valToS(val, trimSuffix string) (string, error) {
	s := strings.TrimSpace(val)
	if !strings.HasSuffix(val, trimSuffix) {
		return "", fmt.Errorf("%q does not have suffix %q", val, trimSuffix)
	}
	s = strings.TrimSuffix(s, trimSuffix)
	s = strings.TrimSpace(s)
	return s, nil
}

func valToI(val, trimSuffix string) (int, error) {
	s, err := valToS(val, trimSuffix)
	if err != nil {
		return 0, err
	}
	ret, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return int(ret), nil
}

func valToF(val, trimSuffix string) (float64, error) {
	s, err := valToS(val, trimSuffix)
	if err != nil {
		return 0.0, err
	}
	ret, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0, err
	}
	return ret, nil
}

func getValS(st map[string]string, name, suffix string) (string, error) {
	if s, ok := st[name]; ok {
		return valToS(s, suffix)
	}
	return "", fmt.Errorf("%v not found", name)
}

func getValI(st map[string]string, name, suffix string) (int, error) {
	if s, ok := st[name]; ok {
		return valToI(s, suffix)
	}
	return 0, fmt.Errorf("%v not found", name)
}

func getValF(st map[string]string, name, suffix string) (float64, error) {
	if s, ok := st[name]; ok {
		return valToF(s, suffix)
	}
	return 0.0, fmt.Errorf("%v not found", name)
}

func parseNameVal(line, delim string) (string, string, error) {
	s := strings.TrimSpace(line)
	if s == "" {
		return "", "", nil
	}
	x := strings.Split(s, delim)
	if len(x) > 2 {
		return "", "", fmt.Errorf("more than one delimeter in line %q", line)
	}
	if len(x) < 2 {
		return "", "", fmt.Errorf("no delimeter in line %q", line)
	}
	return strings.TrimSpace(x[0]), strings.TrimSpace(x[1]), nil
}
