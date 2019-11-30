package szip

import (
	"io"
	"archive/zip"
	"encoding/binary"
	"compress/flate"
	"errors"
)

type Reader struct {
	r io.Reader
	cur io.ReadCloser
}

func NewReader(ri io.Reader) (ro Reader, err error) {
	ro.r = ri
	return
}

func (r *Reader) Next() (h zip.FileHeader, err error) {
	var signature int32
	binary.Read(r.r, binary.LittleEndian, &signature)
	if signature != 0x04034b50 {
		err = errors.New("bad stuff happenin")
		return
	}
	binary.Read(r.r, binary.LittleEndian, &h.ReaderVersion)
	binary.Read(r.r, binary.LittleEndian, &h.Flags)
	binary.Read(r.r, binary.LittleEndian, &h.Method)
	binary.Read(r.r, binary.LittleEndian, &h.ModifiedTime)
	binary.Read(r.r, binary.LittleEndian, &h.ModifiedDate)
	binary.Read(r.r, binary.LittleEndian, &h.CRC32)
	binary.Read(r.r, binary.LittleEndian, &h.CompressedSize)
	binary.Read(r.r, binary.LittleEndian, &h.UncompressedSize)
	var namelen uint16
	binary.Read(r.r, binary.LittleEndian, &namelen)
	var extralen uint16
	binary.Read(r.r, binary.LittleEndian, &extralen)
	name := make([]byte, namelen)
	binary.Read(r.r, binary.LittleEndian, &name)
	h.Name = string(name)
	h.Extra = make([]byte, extralen)
	binary.Read(r.r, binary.LittleEndian, &h.Extra)
	r.cur = flate.NewReader(r.r)
	return
}

func (r Reader) Read(b []byte) (n int, err error) {
	return r.cur.Read(b)
}

func (r Reader) Close() error {
	return r.cur.Close()
}
