package main

import "sync"

// TSyncMap -
type TSyncMap struct {
	sync.Mutex
	data map[string]bool
}

// NewSyncMap -
func NewSyncMap() *TSyncMap {
	o := &TSyncMap{}
	o.data = map[string]bool{}
	return o
}

// Got -
func (o *TSyncMap) Got(key string) bool {
	o.Lock()
	defer o.Unlock()
	_, ok := o.data[key]
	return ok
}

// Add -
func (o *TSyncMap) Add(key string) {
	o.Lock()
	defer o.Unlock()
	o.data[key] = true
}

// Keys -
func (o *TSyncMap) Keys() []string {
	o.Lock()
	defer o.Unlock()
	ret := []string{}
	for key := range o.data {
		ret = append(ret, key)
	}
	return ret
}

// Clear -
func (o *TSyncMap) Clear() {
	o.Lock()
	o.data = map[string]bool{}
	o.Unlock()
}
