package szip

import (
	"io"
	"archive/zip"
	"encoding/binary"
	"compress/flate"
	"errors"
	"fmt"
)

type Reader struct {
	r io.Reader
	cur io.ReadCloser
}

func NewReader(ri io.Reader) (ro Reader, err error) {
	ro.r = ri
	return
}

var decompressors map[uint16]zip.Decompressor;

func init() {
	decompressors = make(map[uint16]zip.Decompressor)
	decompressors[zip.Store] = io.NopCloser;
	decompressors[zip.Deflate] = flate.NewReader;
}

func (r *Reader) Next() (h zip.FileHeader, err error) {
	var signature uint32
	binary.Read(r.r, binary.LittleEndian, &signature)
	if signature == 0x2014b50 {
		err = io.EOF;
		return;
	} else if signature != 0x04034b50 {
		err = errors.New(fmt.Sprintf("szip: Invalid signature: %x", signature))
		return
	}
	binary.Read(r.r, binary.LittleEndian, &h.ReaderVersion)
	binary.Read(r.r, binary.LittleEndian, &h.Flags)
	binary.Read(r.r, binary.LittleEndian, &h.Method)
	binary.Read(r.r, binary.LittleEndian, &h.ModifiedTime)
	binary.Read(r.r, binary.LittleEndian, &h.ModifiedDate)
	binary.Read(r.r, binary.LittleEndian, &h.CRC32)
	binary.Read(r.r, binary.LittleEndian, &h.CompressedSize)
	h.CompressedSize64 = uint64(h.CompressedSize)
	binary.Read(r.r, binary.LittleEndian, &h.UncompressedSize)
	h.UncompressedSize64 = uint64(h.UncompressedSize)
	h.Modified = h.ModTime()
	var namelen uint16
	binary.Read(r.r, binary.LittleEndian, &namelen)
	var extralen uint16
	binary.Read(r.r, binary.LittleEndian, &extralen)
	name := make([]byte, namelen)
	binary.Read(r.r, binary.LittleEndian, &name)
	h.Name = string(name)
	h.Extra = make([]byte, extralen)
	binary.Read(r.r, binary.LittleEndian, &h.Extra)
	r.cur = decompressors[h.Method](io.LimitReader(r.r, int64(h.CompressedSize64)));
	return
}

func (r Reader) Read(b []byte) (n int, err error) {
	return r.cur.Read(b)
}

func (r Reader) WriteTo(w io.Writer) (n int64, err error) {
	return io.Copy(w, r.cur)
}

func (r Reader) Close() error {
	return r.cur.Close()
}
