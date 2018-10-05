package ptool

import (
	"fmt"
)

type tLutItem struct {
	// visited  bool
	// terminal bool
	node *TNode
	id   int
	name string
	ip   TOffset
}

type tLut struct {
	items    []*tLutItem
	entries  []int
	defEntry int
}

func newLut() *tLut {
	ret := &tLut{}
	ret.addItem("", nil)
	return ret
}

func (o *tLut) find(name string) *tLutItem {
	for _, item := range o.items {
		if item.name == name && item.id > 0 {
			// fmt.Println(item.name)
			return item
		}
	}
	return nil
}

func (o *tLut) exists(name string) bool {
	if o.find(name) != nil {
		return true
	}
	return false
}

func (o *tLut) addItem(name string, node *TNode) error {
	item := o.find(name)
	if item != nil {
		return fmt.Errorf("duplicated element in items %q", name)
	}
	item = &tLutItem{name: name, node: node, id: len(o.items), ip: -1}
	o.items = append(o.items, item)
	return nil
}

func (o *tLut) addEntry(item *tLutItem) error {
	for _, v := range o.items {
		if v.id == item.id {
			return fmt.Errorf("duplicated element in entries %v %q", item.id, item.name)
		}
	}
	o.entries = append(o.entries, item.id)
	return nil
}

func (o *tLut) makeProgItems() []tProgItem {
	ret := make([]tProgItem, 0, len(o.items))
	for _, item := range o.items {
		ret = append(ret, tProgItem{name: item.name, ip: item.ip})
	}
	return ret
}

func (o *tLut) makeProgEntries() []int {
	if len(o.entries) != 0 {
		// might be should duplicate
		return o.entries
	}
	// should be check ret value
	return []int{1}
}

func (o *tLut) String() string {
	s := ""
	for _, item := range o.items {
		s += fmt.Sprintf("id: %v, ip: %v, name: %q\n", item.id, item.ip, item.name)
	}
	return s
}
