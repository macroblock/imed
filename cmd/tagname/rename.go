package main

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/macroblock/imed/pkg/tagname"
)

func cmdRename(command string, entry tEntry) {
	flag.Parse()
	srcf := argSplit(flagFromSchemas)
	// fmt.Println("#src is ", flagSrcSchemas, "\n", srcf)
	// fmt.Println("#dst is ", flagDstSchema)
	args, err := getDeGlobRestOfArgs()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	fmt.Println()
	for _, s := range args {
		tn, err := tagname.NewFromFilename(s, 0, srcf...)
		if err != nil {
			fmt.Printf("### %q read error: %v\n", s, err)
			continue
		}
		res, err := tn.ConvertTo(flagToSchema)
		if err != nil {
			fmt.Printf("### %q write error: %v\n", s, err)
			continue
		}
		fmt.Printf("===%q schema %v\n-> %q schema %v\n", filepath.Base(s), tn.Schema(), filepath.Base(res), flagToSchema)
	}
}
