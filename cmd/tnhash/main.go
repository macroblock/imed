package main

import (
	"os"

	"github.com/macroblock/imed/cmd/tnhash/hash"
	"github.com/macroblock/imed/pkg/misc"
	"github.com/macroblock/imed/pkg/tagname"
	"github.com/macroblock/imed/pkg/zlog/loglevel"
	"github.com/macroblock/imed/pkg/zlog/zlog"
)

var (
	log       = zlog.Instance("main")
	retif     = log.Catcher()
	logFilter = loglevel.Warning.OrLower()
)

func doProcess(path string, schema string, checkLevel int) {
	defer retif.Catch()
	log.Info("")
	log.Info("rename: " + path)
	tn, err := tagname.NewFromFilename(path, checkLevel)
	retif.Error(err, "cannot parse filename")

	tag, err := tn.GetTag("type")
	if tag != "film" {
		log.Info("skip")
		return
	}

	name, _ := tn.GetTag("name")
	sxx, _ := tn.GetTag("sxx")
	year, _ := tn.GetTag("year")
	sdhd, _ := tn.GetTag("sdhd")
	comment, _ := tn.GetTag("comment")

	key := hash.Key(name, sxx, year, sdhd, comment)
	tag = hash.Get(key)

	tn.SetTag("hashtag", "h"+tag)

	if schema == "" {
		schema = tn.Schema()
	}
	newPath, err := tn.ConvertTo(schema)
	retif.Error(err, "cannot convert to '"+schema+"'")

	err = os.Rename(path, newPath)
	retif.Error(err, "cannot rename file")

	log.Notice(schema, " > ", newPath)
}

func main() {

	// fmt.Printf("Size: %v\n", sha256.Size)
	// s := "abcdefg"
	// s2 := hash.Get(s, 32)
	// fmt.Printf("%q -> %q\n", s, s2)

	// s = "abcdefh"
	// s2 = hash.Get(s, 32)
	// fmt.Printf("%q -> %q\n", s, s2)

	// setup log
	newLogger := misc.NewSimpleLogger
	if misc.IsTerminal() {
		newLogger = misc.NewAnsiLogger
	}
	log.Add(
		newLogger(loglevel.Warning.OrLower(), ""),
		newLogger(loglevel.Info.Only().Include(loglevel.Notice.Only()), "~x\n"),
	)

	defer func() {
		if log.State().Intersect(loglevel.Warning.OrLower()) != 0 {
			misc.PauseTerminal()
		}
	}()

	// process command line arguments
	if len(os.Args) <= 1 {
		log.Warning(true, "not enough parameters")
		log.Info("Usage:\n    tnhash [-rt|-old] {filename}\n")
		return
	}

	// main job
	args := os.Args[1:]
	schema := ""
	switch args[0] {
	case "-rt":
		schema = "rt"
		args = args[1:]
	case "-old":
		schema = "old"
		args = args[1:]
	}

	// wasError := false
	for _, path := range args {
		doProcess(path, schema, tagname.CheckNormal) //tagname.CheckDeepStrict)
	}
}
