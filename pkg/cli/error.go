package cli

import (
	"fmt"
	"strings"
)

type (
	tErrorHandler func(error) (string, error)
	// IOnError -
	IOnError interface {
		HandleError(error) (string, error)
	}

	tInternalError struct {
		msg string
	}
	// TErrorNotEnoughArguments -
	TErrorNotEnoughArguments struct {
		msg string
	}
)

func (o *tErrorHandler) IElementUniquePattern() {}

func (o *tErrorHandler) Handle(err error) (string, error) {
	if o == nil || err == nil {
		return "", err
	}
	switch err.(type) {
	case *tInternalError:
		return "", err
	}
	return (*o)(err)
}

func internalPanic(msg string) {
	log.Panic(fmt.Sprintf("#intrnal panic# %v", msg))
}

func internalPanicf(format string, a ...interface{}) {
	log.Panic(fmt.Sprintf(format, a...))
}

func internalPanicErr(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func internalError(msg string) error {
	return &tInternalError{msg: msg}
}

func internalErrorf(format string, a ...interface{}) error {
	return internalError(fmt.Sprintf(format, a...))
}

func (o *tInternalError) Error() string {
	return fmt.Sprint("#Internal error#  ", o.msg)
}

func (o *TErrorNotEnoughArguments) Error() string {
	if o.msg == "" {
		return fmt.Sprint("not enough arguments")
	}
	return fmt.Sprint(o.msg)
}

// ErrorNotEnoughArguments -
func ErrorNotEnoughArguments(s ...string) error {
	return &TErrorNotEnoughArguments{msg: strings.Join(s, "")}
}

func newErrorHandler(i interface{}) (*tErrorHandler, error) {
	o := tErrorHandler(nil)
	switch t := i.(type) {
	default:
		return nil, internalErrorf("tOnError got unsupported type %T", i)
	case nil:
		o = func(err error) (string, error) { return "", err }
	case func(error) (string, error):
		o = t
	case func(error) error:
		o = func(err error) (string, error) { return "", t(err) }
	case func(error) string:
		o = func(err error) (string, error) { return t(err), err }
	case func(error):
		o = func(err error) (string, error) { t(err); return "", err }
	case func():
		o = func(err error) (string, error) { t(); return "", err }
	case func() string:
		o = func(err error) (string, error) { return t(), err }
	case func() error:
		o = func(err error) (string, error) { return "", t() }
	case func() (string, error):
		o = func(err error) (string, error) { return t() }
	case error:
		o = func(err error) (string, error) { return "", t }
	case string:
		o = func(err error) (string, error) { return Text(t), err }
	}
	return &o, nil
}
