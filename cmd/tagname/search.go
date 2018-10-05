package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/macroblock/imed/pkg/tagname"
)

func getPresetPath(name string) string {
	if cfgPath == "" {
		return ""
	}
	return filepath.Join(cfgPath, constConfigDir, progName+"."+name+constListExt)
}

func getListPath(name string) string {
	if cwdPath == "" {
		return ""
	}
	return filepath.Join(cwdPath, name+constListExt)
}

func printCurrentElementToParse(path string) {
	fmt.Printf("\x0d%s", strings.Repeat(" ", 79))
	k := minInt(len(path), 79)
	fmt.Printf("\x0d%s", path[:k])
}

func cleanPath(s string) string {
	s = strings.TrimSpace(s)
	s = path.Clean(s)
	if osWindows {
		s = strings.ToLower(s)
	}
	return s
}

func readListOrPreset(path string, visitedMap *map[string]struct{}, unfold, firstRun bool) ([]string, []error) {
	path = cleanPath(path)
	if _, ok := (*visitedMap)[path]; ok {
		return nil, []error{fmt.Errorf("duplicated path %q", path)}
	}
	(*visitedMap)[path] = struct{}{}

	if !(unfold || firstRun) {
		return []string{path}, nil
	}
	firstRun = false

	listName := getListPath(path)
	if !IsFile(listName) {
		listName = getPresetPath(path)
		if !IsFile(listName) {
			dir, err := filepath.Abs(path)
			if err != nil {
				return nil, []error{err}
			}
			return []string{dir}, nil
		}
	}

	errList := []error(nil)
	dirList := []string(nil)

	data, err := ioutil.ReadFile(listName)
	if err != nil {
		errList = append(errList, err)
		return nil, errList
	}

	list := strings.Split(string(data), "\n")
	for _, s := range list {
		s = strings.Split(s, "//")[0] // get rid of comments
		recursive := false
		if strings.HasPrefix(s, "*") {
			s = s[1:]
			recursive = true
		}
		paths, errors := readListOrPreset(s, visitedMap, recursive, false) // recursive, there isn't first run
		errList = append(errList, errors...)
		dirList = append(dirList, paths...)
	}
	return dirList, errList
}

func osReadDir(root string) ([]string, []string, error) {
	f, err := os.Open(root)
	if err != nil {
		return nil, nil, err
	}
	fileInfo, err := f.Readdir(-1)
	f.Close()

	files := []string(nil)
	dirs := []string(nil)
	for _, info := range fileInfo {
		if info.IsDir() {
			dirs = append(dirs, filepath.Join(root, info.Name()))
			continue
		}
		files = append(files, filepath.Join(root, info.Name()))
	}
	return dirs, files, err
}

func readDir(path string) ([]string, []string, error) {
	if strings.IndexRune(path, '#') > -1 {
		return nil, nil, nil
	}
	dirList := []string(nil)
	fileList := []string(nil)

	// to do: is filepath.Clean needed ???
	info, err := os.Stat(path)
	if err != nil {
		return nil, nil, err
	}
	if !info.IsDir() {
		fileList = append(fileList, path)
		return nil, []string{path}, nil
	}

	dirs, files, err := osReadDir(path)

	nfiles := []string{}
	for _, s := range files {
		if strings.IndexRune(s, '#') < 0 {
			nfiles = append(nfiles, s)
		}
	}

	dirList = append(dirList, dirs...)
	fileList = append(fileList, nfiles...)

	return dirList, fileList, err
}

func traversDataSource(names []string) ([]string, []error) {
	errList := []error(nil)
	dirList := []string(nil)
	fileList := []string(nil)

	visitedMap := map[string]struct{}{}
	for _, name := range names {
		dirs, errors := readListOrPreset(name, &visitedMap, true, true) // recursive, firstRun
		dirList = append(dirList, dirs...)
		errList = append(errList, errors...)
	}

	visitedMap = map[string]struct{}{}
	for i := 0; i < len(dirList); i++ {
		path := cleanPath(dirList[i])
		if _, ok := visitedMap[path]; ok {
			errList = append(errList, fmt.Errorf("x duplicated path %q", path))
			continue
		}

		dirs, files, err := readDir(path)

		if err != nil {
			errList = append(errList, err)
		}
		fileList = append(fileList, files...)
		dirList = append(dirList, dirs...)
	}
	return fileList, errList
}

func getDuplicates(list []string) []string {
	filter := ""
	old := ""
	ret := []string{}
	for _, s := range list {
		if strings.ToLower(s) == filter && filter != old {
			ret = append(ret, s)
		}
		old = filter
		filter = strings.ToLower(s)
	}
	return ret
}

func filterHasTag(vals []tVal, tags []string) bool {
	// fmt.Printf("vals %v\n tags %v\n\n", vals, tags)
	for _, val := range vals {
		for _, tag := range tags {
			if (tag == val.val) != val.neg {
				return true
			}
		}
	}
	return false
}

func applyFilters(name string, filters [][]tFilterItem, schemaName string) (bool, error) {
	schemas := []string{strings.TrimSpace(schemaName)}
	if schemas[0] == "" {
		schemas = []string{"rt.normal", "old.normal"}
	}
	tn, err := tagname.NewFromFilename(name, schemas...)
	if err != nil {
		// fmt.Printf(" %q read error: %v\n", s, err)
		return false, fmt.Errorf("%v\n%v", name, err)
	}
	ret := true
	// OR loop
	for _, filter := range filters {
		ret = true
		// AND loop
		for _, val := range filter {
			tags := tn.GetTags(val.name)
			if len(tags) == 0 {
				tags = []string{""}
			}
			// OR check
			if !filterHasTag(val.filter, tags) {
				ret = false
				break
			}
		}
		if ret {
			break
		}
	}
	return ret, nil
}

func cmdSearch(command string, entry tEntry) {
	flag.Parse()
	// args := flag.Args()
	if len(flag.Args()) != 0 {
		fmt.Printf("Too many arguments given. (search)\n")
		return
	}
	fromSchema := (flagFromSchemas)
	dataSource := argSplit(flagDataSource)
	dataDest := (flagDataDest)
	dataSave := strings.TrimSpace(flagSave)

	// fmt.Println("#src is ", flagSrcSchemas, "\n", srcf)
	// fmt.Println("#dst is ", flagDstSchema)
	_ = fromSchema
	_ = dataDest
	_ = dataSave

	fmt.Println()
	filterList := [][]tFilterItem{}
	for _, val := range flagFilter {
		filters, err := argSplitFilter(val)
		if err != nil {
			fmt.Printf("Error while parsing a <filter> flag %q: %v\n", val, err)
			return
		}
		filterList = append(filterList, filters)
	}

	startTime := time.Now()
	dataSource, errors := traversDataSource(dataSource) // allow presets

	result := []string{}
	for _, path := range dataSource {
		// name := filepath.Base(path)
		// ext := filepath.Ext(path)
		// name = strings.TrimSuffix(name, ext)

		ok, err := applyFilters(path, filterList, fromSchema)
		if err != nil {
			errors = append(errors, err)
		}
		if ok {
			result = append(result, path)
		}
	}
	sort.Strings(result)

	dup := getDuplicates(result)
	for _, s := range dup {
		fmt.Printf("duplicate: %v\n", s)
	}

	// fmt.Println("=dataSource==============================================================")
	// for _, s := range dataSource {
	// 	fmt.Println(s)
	// }
	fmt.Printf("Data source: %v file(s)\n", len(dataSource))
	// fmt.Println("=errors==================================================================")
	if len(errors) > 0 {
		msg := "Error:"
		if len(errors) > 1 {
			msg = "Errors:"
		}
		fmt.Println(msg)
		for _, e := range errors {
			fmt.Println(e)
		}
	}
	fmt.Println("=========================================================================")

	elapsedTime := time.Since(startTime)
	log.Printf("Time elapsed %s", elapsedTime)

	fmt.Println("=result==================================================================")
	for _, s := range result {
		fmt.Println(s)
	}

	if dataSave != "" {
		err := saveResult(dataSave, result, errors)
		if err != nil {
			fmt.Printf("error while saving file: %v\n", err)
			return
		}
	}

	// 	tn, err := tagname.NewFromFilename(s, srcf...)
	// 	if err != nil {
	// 		fmt.Printf("### %q read error: %v\n", s, err)
	// 		continue
	// 	}
	// 	res, err := tn.ConvertTo(flagDstSchema)
	// 	if err != nil {
	// 		fmt.Printf("### %q write error: %v\n", s, err)
	// 		continue
	// 	}
	// 	fmt.Printf("===%q schema %v\n-> %q schema %v\n", filepath.Base(s), tn.Schema(), filepath.Base(res), flagDstSchema)
	// }
}

func saveSlice(filePath string, values interface{}) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	rv := reflect.ValueOf(values)
	if rv.Kind() != reflect.Slice {
		return errors.New("Not a slice")
	}
	for i := 0; i < rv.Len(); i++ {
		fmt.Fprintln(f, rv.Index(i).Interface())
	}
	return nil
}

func saveResult(path string, res []string, errors []error) error {
	err2 := error(nil)
	err := saveSlice(path+".list", res)
	if len(errors) > 0 {
		err2 = saveSlice(path+".err", errors)
	}
	if err == nil {
		err = err2
	}
	return err
}
