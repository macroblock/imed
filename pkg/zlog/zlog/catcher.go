package zlog

import "github.com/macroblock/imed/pkg/zlog/loglevel"

var throw bool

const catcherMessage = "TCatcher: you forgot 'defer xxx.Catch()'"

// TCatcher -
type TCatcher struct {
	log    *TLog
	thrown bool
}

// Catch -
func (o *TCatcher) Catch() {
	if r := recover(); r != nil {
		//o.log.Debug("TCatcher: catched")
	}
}

//Return -
func (o *TCatcher) Return(condition interface{}) {
	ok, err := getErrorCondition(condition)
	if ok || err != nil {
		// o.log.Log(loglevel.Panic, 0, err, text...)
		panic(catcherMessage)
	}
}

// Panic -
func (o *TCatcher) Panic(condition interface{}, text ...interface{}) {
	ok, err := getErrorCondition(condition)
	if ok || err != nil {
		o.log.Log(loglevel.Panic, 0, err, text...)
		panic(catcherMessage)
	}
}

// Error -
func (o *TCatcher) Error(condition interface{}, text ...interface{}) {
	ok, err := getErrorCondition(condition)
	if ok || err != nil {
		o.log.Log(loglevel.Error, 0, err, text...)
		panic(catcherMessage)
	}
}

// Warning -
func (o *TCatcher) Warning(condition interface{}, text ...interface{}) {
	ok, err := getErrorCondition(condition)
	if ok || err != nil {
		o.log.Log(loglevel.Warning, 0, err, text...)
		panic(catcherMessage)
	}
}

// // Reset -
// func (o *TCatcher) Reset(resetFilter loglevel.TFilter, text ...interface{}) {
// 	o.Log(loglevel.Reset, resetFilter, nil, text...)
// }

// // Notice -
// func (o *TCatcher) Notice(text ...interface{}) {
// 	o.Log(loglevel.Notice, 0, nil, text...)
// }

// // Info -
// func (o *TCatcher) Info(text ...interface{}) {
// 	o.Log(loglevel.Info, 0, nil, text...)
// }

// // Debug -
// func (o *TCatcher) Debug(text ...interface{}) {
// 	o.Log(loglevel.Debug, 0, nil, text...)
// }
