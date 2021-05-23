package szip2tar
// https://blog.golang.org/package-names
// "packages that are frequently used together should have distinct names."

import (
	"archive/tar"
	"archive/zip"
	"io"
	"io/fs"
	"code.sirabella.org/szip"
)

type RecordHeader interface {
	FileInfo() fs.FileInfo
}

func header(hdr RecordHeader, fullPath string) *tar.Header {
	thdr, _ := tar.FileInfoHeader(hdr.FileInfo(), "");
	thdr.Name = fullPath;
	return thdr
}

// unused because .Name is a field :(
type ArchiveReader interface {
	Next() (RecordHeader, error)
	io.ReadCloser
}

func record(tw *tar.Writer, r szip.Reader) (n int64, err error) {
	var hdr zip.FileHeader
	if hdr, err = r.Next(); err != nil {
		return
	}
	tw.WriteHeader(header(&hdr, hdr.Name));
	n, err = r.WriteTo(tw)
	if err == io.EOF {
		return n, nil
	}
	return
}

func Write(w io.Writer, r szip.Reader) (err error) {
	var tw *tar.Writer
	defer func() { // https://www.joeshaw.org/dont-defer-close-on-writable-files/
		cerr := tw.Close()
		if err == nil {
			err = cerr
		}
	}()

	for tw = tar.NewWriter(w); err != io.EOF; _, err = record(tw, r) {
		if err != nil {
			return
		}
	}
	return nil
}
