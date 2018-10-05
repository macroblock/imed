package main

import (
	"runtime"
	"sync"
)

type (
	// TJobList -
	TJobList struct {
		sync.Mutex
		ch      chan bool
		numJobs int
	}

	// TWorkers -
	TWorkers struct {
		wg sync.WaitGroup
	}
)

// NewWorkers -
func NewWorkers() *TWorkers {
	return &TWorkers{}
}

// Add -
func (o *TWorkers) Add(num int, fn func()) {
	if num < 0 {
		num = runtime.NumCPU()*2/3 + 1
	}
	for i := 0; i < num; i++ {
		go func() {
			o.wg.Add(1)
			fn()
			o.wg.Done()
		}()
	}
}

// Wait -
func (o *TWorkers) Wait() {
	o.wg.Wait()
}

// NewJobList -
func NewJobList(numJobs int) *TJobList {
	o := &TJobList{numJobs: numJobs}
	if o.numJobs <= 0 {
		o.numJobs = runtime.NumCPU()*2/3 + 1
	}
	o.ch = make(chan bool, o.numJobs)
	return o
}

// AddFn -
func (o *TJobList) AddFn(fn func()) {
	o.ch <- true
	if fn == nil {
		<-o.ch
		return
	}
	go func() {
		fn()
		<-o.ch
	}()
}

// Wait -
func (o *TJobList) Wait() {
	for i := 0; i < cap(o.ch); i++ {
		o.ch <- true
	}
}
