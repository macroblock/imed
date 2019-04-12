package ptool

import (
	"fmt"
	"testing"
)

// TestPTool -
func TestTagnameCorrect(t *testing.T) {
	rools := `
entry = '' bom {@rus | @eng} @any $;
bom = \ufeff;
rusrune = 'а'..'я'|'А'..'Я';
engrune = 'a'..'z'|'A'..'Z';
rus = rusrune#{#rusrune};
eng = engrune#{#engrune};
any = anyrune#{#anyrune};
anyrune = \x00..$;
= {' '}
	`
	builder := NewBuilder().FromString(rools).Entries("entry")
	fmt.Println("build -------")
	p, err := builder.Build()
	if err != nil {
		t.Errorf("builder error: %v\n", err)
		return
	}
	// fmt.Println(builder.TreeToString())
	fmt.Println("parse -------")
	tree, err := p.Parse(string([]byte{0xef, 0xbb, 0xbf}) + "xxzабвгддд qwerty   asdfфываdsa 1111")
	if err != nil {
		t.Errorf("parser error: %v\n", err)
		return
	}
	fmt.Println(TreeToString(tree, p.ByID))

}
