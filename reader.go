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

	FileHeader []zip.FileHeader
	n int
}

func NewReader(ri io.Reader) (ro Reader, err error) {
	ro.r = ri
	ro.n = 0
	return
}

var decompressors map[uint16]zip.Decompressor;

func init() {
	decompressors = make(map[uint16]zip.Decompressor)
	decompressors[zip.Store] = io.NopCloser;
	decompressors[zip.Deflate] = flate.NewReader;
}

const (
	LOCALRECORD uint32 = 0x04034b50
	CENTRALRECORD uint32 = 0x2014b50
	EOCD uint32 = 0x06054b50
)

func readHeader(signature uint32, r io.Reader) (h zip.FileHeader, err error) {
	if signature != LOCALRECORD && signature != CENTRALRECORD {
		// h is empty & should be unused here
		return h, errors.New(fmt.Sprintf("szip: Invalid signature: %x", signature))
	}
	if signature == CENTRALRECORD {
		binary.Read(r, binary.LittleEndian, &h.CreatorVersion)
	}
	binary.Read(r, binary.LittleEndian, &h.ReaderVersion)
	binary.Read(r, binary.LittleEndian, &h.Flags)
	binary.Read(r, binary.LittleEndian, &h.Method)
	binary.Read(r, binary.LittleEndian, &h.ModifiedTime)
	binary.Read(r, binary.LittleEndian, &h.ModifiedDate)
	h.Modified = h.ModTime()
	binary.Read(r, binary.LittleEndian, &h.CRC32)
	binary.Read(r, binary.LittleEndian, &h.CompressedSize)
	h.CompressedSize64 = uint64(h.CompressedSize)
	binary.Read(r, binary.LittleEndian, &h.UncompressedSize)
	h.UncompressedSize64 = uint64(h.UncompressedSize)
	var namelen uint16
	binary.Read(r, binary.LittleEndian, &namelen)
	name := make([]byte, namelen)
	var extralen uint16
	binary.Read(r, binary.LittleEndian, &extralen)
	h.Extra = make([]byte, extralen)

	var comment []byte
	if signature == CENTRALRECORD {
		var commentlen uint16
		binary.Read(r, binary.LittleEndian, &commentlen)
		comment = make([]byte, commentlen)
		var diskNbr uint16
		binary.Read(r, binary.LittleEndian, &diskNbr)
		var internalAttrs uint16
		binary.Read(r, binary.LittleEndian, &internalAttrs)
		binary.Read(r, binary.LittleEndian, &h.ExternalAttrs)
		var offset uint32
		binary.Read(r, binary.LittleEndian, &offset)
	}

	binary.Read(r, binary.LittleEndian, &name)
	h.Name = string(name)
	binary.Read(r, binary.LittleEndian, &h.Extra)

	if signature == CENTRALRECORD {
		binary.Read(r, binary.LittleEndian, &comment)
		h.Comment = string(comment)
	}
	return
}

func (r Reader) fillFiles() (files []zip.FileHeader, err error) {
	files = make([]zip.FileHeader, 0, r.n) // r.n is just a suggestion
	for signature := CENTRALRECORD; signature != EOCD; binary.Read(r.r, binary.LittleEndian, &signature) {
		var chdr zip.FileHeader
		if chdr, err = readHeader(signature, r.r); err != nil {
			return
		}
		files = append(files, chdr)
	}
	return
}

func (r *Reader) Next() (h zip.FileHeader, err error) {
	if r.cur != nil {
		io.ReadAll(r.cur)
	}
	var signature uint32
	binary.Read(r.r, binary.LittleEndian, &signature)
	if signature == CENTRALRECORD { // collect all central directory headers
		if r.FileHeader, err = r.fillFiles(); err != nil {
			return // with err
		}
		return h, io.EOF; // h is empty here
	}
	if h, err = readHeader(signature, r.r); err != nil {
		return
	}
	r.n += 1
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
