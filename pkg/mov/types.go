package mov

import (
	//"fmt"
)

const (
	boxUUID = "uuid"
)

type AtomFlags int

type IAtomData interface {
	String() string
	Size() int64
	Write(wr *StreamWriter) error
}

type Atom struct {
	typ   Fcc
	data  IAtomData
	atoms []*Atom
}

func InitAtom(typ Fcc) Atom {
	return Atom{typ: typ}
}

func NewAtom(typ Fcc) *Atom {
	return &Atom{typ: typ}
}

func (o *Atom) Type() Fcc {
	return o.typ
}

func (o *Atom) Data() IAtomData {
	return o.data
}

func (o *Atom) Atoms() []*Atom {
	t := make([]*Atom, 0, len(o.atoms))
	return append(t, o.atoms...)
}

func (o *Atom) SetData(data IAtomData) {
	o.data = data
}

func (o *Atom) SetAtoms(atoms []*Atom) {
	o.atoms = atoms
}

func (o *Atom) Size() int64 {
	size := int64(8)
	//fmt.Printf("\natom: %v, o.atoms: %v\n", o.typ, o.atoms)
	for _, a := range o.atoms {
		size += a.Size()
	}
	if o.data != nil {
		size += o.data.Size()
	}
	if size > int64(MaxInt32) {
		size += 8
	}
	return size
}

func (o *Atom) Write(wr *StreamWriter) error {
	wr.WriteU32(uint32(o.Size()))
	wr.WriteU32(uint32(o.typ))
	if o.data != nil {
		err := o.data.Write(wr)
		if err != nil {
			return err
		}
	}
	for _, a := range o.atoms {
		err := a.Write(wr)
		if err != nil {
			return err
		}
	}
	return wr.Err()
}
