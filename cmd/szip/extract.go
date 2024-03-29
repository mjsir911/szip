package main

import (
	"code.sirabella.org/szip"
	"archive/zip"
	"io"
	"path/filepath"
	"os"
	"strings"
	"log"
)

func extractZip(target string, r *szip.Reader) error {
	// https://golangdocs.com/tar-gzip-in-golang
	for {
		hdr, err := r.Next();
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if strings.Contains(hdr.Name, "..") {
			log.Printf("Record name contains \"..\", not extracting: %v", hdr.Name)
			continue
		}
		path := filepath.Join(target, hdr.Name)
		// path = hdr.Name
		// path := filepath.Join(target, hdr.Name)
		info := hdr.FileInfo()
		if info.IsDir() {
			// permission gets narrowed with umask
			if err := os.MkdirAll(path, 0777); err != nil {
				return err;
			}
		} else {
			if _, err := os.Stat(path); err == nil {
				continue
			}
			file, err := os.Create(path);
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(file, r)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func extractPermissions(target string, hdrs []zip.FileHeader) error {
	for _, hdr := range hdrs {
		if strings.Contains(hdr.Name, "..") {
			log.Printf("Record name contains \"..\", not updating permissions: %v", hdr.Name)
			continue
		}
		path := filepath.Join(target, hdr.Name)
		info := hdr.FileInfo()
		if err := os.Chmod(path, info.Mode()); err != nil {
			return err
		}
	}
	return nil
}

