package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/macroblock/imed/pkg/mov"
)

type (
	MoovInfo struct {
		filename string
		posInfo  PositionInfo
		root     *mov.Atom
		tracks   []TrackInfo
	}

	TrackInfo struct {
		root *mov.Atom
		typ  string
	}

	PositionInfo struct {
		pos  int64
		end  int64
		size int64
		last bool
	}
)

func (o *MoovInfo) SetUdta(root *mov.Atom, path string, typ mov.Fcc, lc mov.LangCode, text string) error {

	p := "udta"
	if path != "" {
		p = path + "/udta"
	}

	atom, err := o.FindAtom(root, p)
	if err != nil {
		return err
	}

	if atom == nil {
		a, err := o.BuildPath(root, p)
		if err != nil {
			return err
		}
		atom = a
	}

	data := atom.Data()
	if data == nil {
		atom.SetData(mov.NewUdta())
		data = atom.Data()
	}

	udta, ok := data.(*mov.Udta)
	if !ok {
		return fmt.Errorf("data is not udta at %q", path)
	}

	udta.Set(typ, lc, text)

	return nil
}

func (o *MoovInfo) RemoveUdta(root *mov.Atom, path string, typ mov.Fcc, lc mov.LangCode) error {

	p := "udta"
	if path != "" {
		p = path + "/udta"
	}

	if typ == 0 {
		// remove whole atom
		err := o.RemoveAtom(root, p)
		return err
	}

	// otherwise remove only specified udta field

	atom, err := o.FindAtom(root, p)
	if err != nil {
		return err
	}

	data := atom.Data()
	if data == nil {
		atom.SetData(mov.NewUdta())
		data = atom.Data()
	}

	udta, ok := data.(*mov.Udta)
	if !ok {
		return fmt.Errorf("data is not udta at %q", path)
	}

	udta.Remove(typ, lc)

	return nil
}

func StrToPath(path string) ([]mov.Fcc, error) {
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	list := strings.Split(path, "/")
	p := make([]mov.Fcc, 0, len(list))
	for _, s := range list {
		fcc, err := mov.StrToFcc(s)
		if err != nil {
			return nil, err
		}
		p = append(p, fcc)
	}
	return p, nil
}

func atomIndex(atoms []*mov.Atom, typ mov.Fcc) int {
	for i, a := range atoms {
		if a.Type() == typ {
			return i
		}
	}
	return -1
}

func (o *MoovInfo) FindAtom(root *mov.Atom, path string) (*mov.Atom, error) {
	p, err := StrToPath(path)
	if err != nil {
		return nil, err
	}

	for _, typ := range p {
		atoms := root.Atoms()
		i := atomIndex(atoms, typ)
		if i < 0 {
			return nil, nil
		}
		root = atoms[i]
	}

	return root, nil
}

func (o *MoovInfo) BuildPath(root *mov.Atom, path string) (*mov.Atom, error) {
	p, err := StrToPath(path)
	if err != nil {
		return nil, err
	}

	ret := root
	for _, typ := range p {
		atoms := ret.Atoms()
		i := atomIndex(atoms, typ)
		if i < 0 {
			atoms = append(atoms, mov.NewAtom(typ))
			ret.SetAtoms(atoms)
			i = len(atoms) - 1
		}
		ret = atoms[i]
	}

	return ret, nil
}

func (o *MoovInfo) RemoveAtom(root *mov.Atom, path string) error {
	p, err := StrToPath(path)
	if err != nil {
		return err
	}

	prev := (*mov.Atom)(nil)
	i := 0
	for _, typ := range p {
		atoms := root.Atoms()
		i = atomIndex(atoms, typ)
		if i < 0 {
			return nil
		}
		prev = root
		root = atoms[i]
	}

	if prev == nil {
		return fmt.Errorf("cannot remove atom by path %q", path)
	}

	atoms := prev.Atoms()
	atoms = append(atoms[:i], atoms[i+1:]...)
	prev.SetAtoms(atoms)

	return nil
}

func (o *MoovInfo) MergeUdta(srcMoov *MoovInfo) error {
	mergeUdta := func(dstRoot, srcRoot *mov.Atom, path string) {
		src, err := o.FindAtom(srcRoot, path)
		if err != nil {
			panic("unreachable")
		}
		if src == nil || src.Data() == nil {
			return
		}
		srcUdta := src.Data().(*mov.Udta)

		a, err := o.BuildPath(dstRoot, "udta")
		if err != nil {
			panic("unreachable")
		}
		if a.Data() == nil {
			a.SetData(mov.NewUdta())
		}
		dstUdta := a.Data().(*mov.Udta)

		dstUdta.Merge(srcUdta)
	}

	if len(o.tracks) != len(srcMoov.tracks) {
		return fmt.Errorf("merge: unequal number of tracks")
	}

	mergeUdta(o.root, srcMoov.root, "udta")

	for i := range srcMoov.tracks {
		mergeUdta(o.tracks[i].root, srcMoov.tracks[i].root, "udta")
		mergeUdta(o.tracks[i].root, srcMoov.tracks[i].root, "mdia/udta")
	}
	return nil
}

func (o *MoovInfo) CleanAllUdta() error {
	err := o.RemoveAtom(o.root, "udta")
	if err != nil {
		return err
	}
	for _, track := range o.tracks {
		err = o.RemoveAtom(track.root, "udta")
		if err != nil {
			return err
		}
		err = o.RemoveAtom(track.root, "mdia/udta")
		if err != nil {
			return err
		}
	}
	return nil
}

func WriteToNewFile(filename string, root *mov.Atom) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	wr := mov.NewStreamWriterBE(f)
	defer wr.Flush()

	err = root.Write(wr)

	return err
}

func ReadMoovFromFile(filename string) ([]*mov.Atom, PositionInfo, error) {
	posInfo := PositionInfo{-1, -1, -1, false}

	f, err := os.Open(filename)
	if err != nil {
		return nil, posInfo, err
	}
	defer f.Close()

	rd := mov.NewStreamReaderBE(f)

	atoms, err := mov.WalkStream(rd, "",
		func(rd *mov.StreamReader, path string, header *mov.AtomHeader, atom *mov.Atom, fn mov.WalkStreamFn) error {
			//fmt.Printf("## %v\n", path)

			if !strings.HasPrefix(path, "/moov") {
				posInfo.last = false
				return mov.SkipAtom
			}

			err := error(nil)
			data := mov.IAtomData(nil)
			atoms := []*mov.Atom(nil)
			switch path {
			default:
				data, err = mov.ReadUnknown(rd)
			case "/moov":
				atoms, err = mov.WalkStream(rd, path, fn)
				posInfo.pos = header.Pos
				posInfo.size = header.TotalSize
				posInfo.last = true
			case "/moov/trak":
				atoms, err = mov.WalkStream(rd, path, fn)
			case "/moov/trak/mdia":
				atoms, err = mov.WalkStream(rd, path, fn)
			case "/moov/udta":
				var udta *mov.Udta
				udta, err = mov.ReadAnyUdta(rd)
				data = udta
				if atom.Data() != nil {
					data = atom.Data()
					data.(*mov.Udta).Merge(udta)
				}
			case "/moov/trak/udta":
				data, err = mov.ReadAnyUdta(rd)
			case "/moov/trak/mdia/udta":
				data, err = mov.ReadAnyUdta(rd)
			case "/moov/trak/mdia/hdlr":
				data, err = mov.ReadMoovTrakMdiaHdlr(rd)
			} // switch path
			atom.SetData(data)
			atom.SetAtoms(atoms)
			return err
		}) // Walk

	posInfo.end = rd.Pos()

	return atoms, posInfo, err
}

func CleanMetaData(atoms []*mov.Atom) []*mov.Atom {
	atoms = mov.ShallowCopy(atoms, "", func(path string, atom *mov.Atom) bool {
		ret := false
		switch path {
		case "/moov/udta",
			"/moov/trak/udta",
			"/moov/trak/mdia/udta":
			ret = atom.Data() != nil
		case "/moov", "/moov/trak/mdia/hdlr":
			ret = true
		}
		return ret
	})
	return atoms
}

func NewMoovInfo(filename string, atoms []*mov.Atom) *MoovInfo {
	moov := &MoovInfo{filename: filename}
	track := TrackInfo{}

	_ = mov.WalkAtomsBT(atoms, "", func(path string, atom *mov.Atom) {
		switch path {
		case "/moov":
			moov.root = atom
		case "/moov/trak":
			track.root = atom
			moov.tracks = append(moov.tracks, track)
			track = TrackInfo{}
			/*
				case "/moov/udta":
					moov.udta = atom
				case "/moov/trak/udta":
					track.udta = atom
				case "/moov/trak/mdia/udta":
					track.mdiaUdta = atom
			*/
		case "/moov/trak/mdia/hdlr":
			track.typ = atom.Data().(*mov.MoovTrakMdiaHdlr).ComponentSubtype.String()
		} // switch path
	})
	return moov
}

func indent(s string, indent string) string {
	return strings.Replace(
		strings.Replace(s, "\n", "\n"+indent, -1),
		"\\xa9", "(c)", -1)
}

func PrintMoov(moov *MoovInfo) {
	findAtom := func(root *mov.Atom, path string) *mov.Atom {
		atom, err := moov.FindAtom(root, path)
		if err != nil {
			panic("unreachable")
		}
		return atom
	}
	fmt.Printf("%v:\n", moov.filename)
	fmt.Printf("eof pos: 0x%08x\n", moov.posInfo.end)
	fmt.Printf("moov (%v):\n", moov.root.Size())
	s := "is not last atom in file"
	if moov.posInfo.last {
		s = "last atom in file"
	}
	fmt.Printf("    original size: %v, pos: 0x%08x - %v\n",
		moov.posInfo.size, moov.posInfo.pos, s,
	)
	fmt.Printf("tracks: %v\n", len(moov.tracks))
	atom := findAtom(moov.root, "udta")
	if atom != nil {
		fmt.Printf("  udta (%v):\n", atom.Size())
		fmt.Printf("    %v\n", indent(atom.Data().String(), "  "))
	}
	for i := range moov.tracks {
		track := moov.tracks[i]
		fmt.Printf("  %v: %v\n", i, track.typ)
		atom = findAtom(track.root, "udta")
		if atom != nil {
			fmt.Printf("    udta (%v):\n", atom.Size())
			fmt.Printf("      %v\n",
				indent(atom.Data().String(), "      "),
			)
		}
		atom = findAtom(track.root, "mdia/udta")
		if atom != nil {
			fmt.Printf("    mdia/udta (%v):\n", atom.Size())
			fmt.Printf("      %v\n",
				indent(atom.Data().String(), "      "),
			)
		}
	}
}

func PrintAtoms(atoms []*mov.Atom) {
	if atoms == nil {
		fmt.Println("<NIL>")
		return
	}

	printAtoms(atoms, "")
}

func printAtoms(atoms []*mov.Atom, path string) {
	for _, atom := range atoms {
		name := fmt.Sprintf("%v/%v", path, atom.Type())

		fmt.Printf("(% 4v) %q\n", atom.Size(), name)

		if atom.Data() != nil {
			text := atom.Data().String()
			if len(text) > 0 {
				text = strings.Replace(text, "\n", "\n            ", -1)
				fmt.Printf("            %v\n", text)
			}
		}
		printAtoms(atom.Atoms(), path+"/"+atom.Type().String())
	}
}

type State struct {
	moov *MoovInfo
}

func NewState() *State {
	return &State{}
}

func (o *State) Process(cmd Command) error {
	err := error(nil)
	switch cmd.cmd {
	default:
		return fmt.Errorf("command %v unimplemented yet", cmd.cmd)

	case "-i":
		fmt.Printf("Importing moov from: %q\n", cmd.path)
		atoms, posInfo, e := ReadMoovFromFile(cmd.path)
		err = e
		// PrintAtoms(atoms)
		// fmt.Printf("----------------------------------------\n")
		o.moov = NewMoovInfo(cmd.path, atoms)
		o.moov.posInfo = posInfo

	case "-export":
		fmt.Printf("Exporting metadata to: %q\n", cmd.path)
		atoms := []*mov.Atom{o.moov.root}
		atoms = CleanMetaData(atoms)
		root := atoms[0]
		err = WriteToNewFile(cmd.path, root)

	case "-merge":
		fmt.Printf("Merging with: %q\n", cmd.path)
		atoms, _, e := ReadMoovFromFile(cmd.path)
		err = e
		if err == nil {
			atoms = CleanMetaData(atoms)
			srcMoov := NewMoovInfo(cmd.path, atoms)
			err = o.moov.MergeUdta(srcMoov)
		}

	case "-write":
		fmt.Printf("Wrighting back to: %q\n", o.moov.filename)
		if o.moov.filename == "" {
			return fmt.Errorf("no filename to write")
		}
		f, err := os.OpenFile(o.moov.filename, os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			return err
		}
		defer f.Close()

		pos := o.moov.posInfo.pos
		size := o.moov.root.Size()
		//moovSize := o.moov.root.Size()
		if !o.moov.posInfo.last {
			freeAtom := mov.NewAtom(mov.StrToFccOrPanic("free"))
			freeAtom.SetData(mov.NewAnyFree(o.moov.posInfo.size))

			_, err := f.Seek(o.moov.posInfo.pos, os.SEEK_SET)
			if err != nil {
				return err
			}

			wr := mov.NewStreamWriterBE(f)
			err = freeAtom.Write(wr)
			if err != nil {
				return err
			}
			wr.Flush()
			wr = nil
			pos = o.moov.posInfo.end
		}

		fmt.Printf("    pos: 0x%08x, size: 0x%08x\n", pos, size)

		_, err = f.Seek(pos, os.SEEK_SET)
		if err != nil {
			return err
		}

		wr := mov.NewStreamWriterBE(f)

		err = o.moov.root.Write(wr)
		if err != nil {
			return err
		}
		wr.Flush()
		f.Truncate(pos + size)

	case "-set":
		msg := ""
		if cmd.track >= 0 {
			msg = fmt.Sprintf("%v", cmd.track)
		}
		msg += "/" + cmd.path
		fmt.Printf("Setting data: %v\n", msg)

		root := o.moov.root
		if cmd.track >= len(o.moov.tracks) {
			return fmt.Errorf("invalid track index %v", cmd.track)
		} else if cmd.track < 0 {
			root = o.moov.root
		} else {
			root = o.moov.tracks[cmd.track].root
		}
		err = o.moov.SetUdta(root, cmd.path, cmd.typ, cmd.langCode, cmd.text)

	case "-remove":
		msg := ""
		if cmd.track >= 0 {
			msg = fmt.Sprintf("%v", cmd.track)
		}
		msg += "/" + cmd.path
		fmt.Printf("Removing data: %v\n", msg)

		root := o.moov.root
		if cmd.track >= len(o.moov.tracks) {
			return fmt.Errorf("invalid track index %v", cmd.track)
		} else {
			root = o.moov.tracks[cmd.track].root
		}
		err = o.moov.RemoveUdta(root, cmd.path, cmd.typ, cmd.langCode)

	case "-clean":
		fmt.Printf("Cleaning all metadata (udta)\n")

		err = o.moov.CleanAllUdta()

	} // switch cmd.cmd

	if err != nil {
		return err
	}

	return err
}

func (o *State) Print() {
	PrintMoov(o.moov)
}

type Command struct {
	cmd      string
	track    int
	path     string
	typ      mov.Fcc
	langCode mov.LangCode
	text     string
}

func takeArg(args *[]string, cmd string) (string, error) {
	if len(*args) == 0 {
		return "", fmt.Errorf("cmdline: have no argument for command %v", cmd)
	}
	ret := (*args)[0]
	*args = (*args)[1:]
	return ret, nil
}

func parseCommand(args *[]string) (Command, error) {
	ret := Command{}
	cmd, err := takeArg(args, "")
	if err != nil {
		panic("unreachable")
	}
	switch cmd {
	default:
		return ret, fmt.Errorf("cmdline: unknown command %v", cmd)

	case "-i", "-merge":
		arg, err := takeArg(args, cmd)
		if err != nil {
			return ret, err
		}

		if err := checkFile(arg); err != nil {
			return ret, err
		}
		ret.path = arg

	case "-export":
		arg, err := takeArg(args, cmd)
		if err != nil {
			return ret, err
		}

		ret.path = arg

	case "-set", "-remove":
		arg, err := takeArg(args, cmd)
		if err != nil {
			return ret, err
		}

		// 1/(c)nam,lng=text
		// 1/mdia/(c)nam,lng=text
		// /(c)nam,lng=text
		// 1/name=text

		// TODO: ugly stuff. It should be improved
		list := strings.SplitN(arg, "/", 2)
		track := -1
		list[0] = strings.TrimSpace(list[0])
		if len(list[0]) > 0 {
			v, err := strconv.ParseUint(list[0], 10, 16)
			if err != nil {
				return ret, fmt.Errorf("invalid specifier: %v in %v", err, arg)
			}
			track = int(v)
		}
		if len(list) < 2 {
			return ret, fmt.Errorf("incomplete specifier: have no '/' in %v", arg)
		}
		list = strings.SplitN(list[1], "=", 2)
		path := list[0]
		text := ""
		if len(list) == 2 {
			text = list[1]
		}
		list = strings.Split(path, "/")
		if len(list) < 1 {
			return ret, fmt.Errorf("invalid specifier: incorrect path in %v", arg)
		}
		temp := list[len(list)-1]

		list = list[:len(list)-1]
		path = strings.Join(list, "/")

		list = strings.SplitN(temp, ",", 2)

		typ := mov.Fcc(0)
		if len(list[0]) > 0 {
			typ, err = mov.StrToFcc(list[0])
			if err != nil {
				return ret, fmt.Errorf("invalid specifier: %v in %v", err, arg)
			}
		}
		langCode := mov.LangCode(0)
		if !typ.IsSimple() {
			if len(list) < 2 {
				return ret, fmt.Errorf("invalid specifier: have no langcode in %v", arg)
			}
			langCode, err = mov.StrToLangCode(list[1])
			if err != nil {
				return ret, fmt.Errorf("invalid specifier: %v in %v", err, arg)
			}
		}

		if path != "" && (track >= 0 && path != "mdia") {
			return ret, fmt.Errorf("unsafe atom path %v", path)
		}
		ret.track = track
		ret.path = path
		ret.typ = typ
		ret.langCode = langCode
		ret.text = text

	case "-write", "-clean":
	} // switch cmd

	ret.cmd = cmd
	return ret, nil
}

func checkFile(path string) error {
	// Lstat to follow the links
	stat, err := os.Lstat(path)
	if err == nil {
		if stat.IsDir() {
			err = fmt.Errorf("%q is a directory", path)
		}
	}
	return err
}

func printBanner() {
	if len(os.Args) > 1 {
		return
	}
	fmt.Println(`
-i <filename>       load file
-write              write changes to currently loaded file
-export <filename>  export metadata to file
-merge <filename>   merge metadata with currently loaded file
-clean              strip metadata
-set <specifier>    set metadata
-remove <specifier> remove metadata

*<specifier>        <trackNo>/<path>/<fourcc>=<text>
                    <trackNo>/<path>/(c)<rest of fourcc>,<lang>=<text>
                    /<path>/<fourcc>=<text>
                    /<path>/(c)<rest of fourcc>,<lang>=<text>
sample usage:
    -i filename -set 1/name=track-name
    -i filename -set /name=track-name -write
    -i filename -set /name=track-name -export metadata-file
    -i metadata-file -set "0/(c)nam,eng=localized-text" -write
    -i filename -import metadata-file -write
`)
	os.Exit(0)
}

func main() {
	errs := []string{}

	printBanner()

	commands := []Command{}
	args := os.Args[1:]
	for len(args) > 0 {
		cmd, err := parseCommand(&args)
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}
		commands = append(commands, cmd)
	}

	if len(errs) != 0 {
		fmt.Println("Error(s):")
		for i, e := range errs {
			fmt.Printf("  % 2v: %v\n", i+1, e)
		}
		os.Exit(-1)
	}

	state := NewState()
	for _, cmd := range commands {
		err := state.Process(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(-1)
		}
	}
	state.Print()
	fmt.Println("Ok.")
}
