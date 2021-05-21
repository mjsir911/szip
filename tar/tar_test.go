package szip2tar

import (
	"testing"
	"os"
	"io"
	"sirabella.org/code/szip"
	"archive/tar"
)

func TestWrite(t *testing.T) {
	f, err := os.Open("../test_files/test.zip")
	if err != nil {
		t.Fatal("Could not open test file")
	}
	rp, wp := io.Pipe()
	go func(){
		if err := Write(wp, szip.NewReader(f)); err != nil {
			t.Errorf("Write() returned error: %v", err)
		}
	}()
	r := tar.NewReader(rp)
	hdr, err := r.Next()
	if err != nil {
		t.Fatalf("Next() returned error: %v", err)
	}
	if hdr.Name != "file2" {
		t.Errorf("Unexpected record 1 name: %v", err)
	}
	_, err = r.Next()
	if err != nil {
		t.Fatalf("Next() returned error: %v", err)
	}
	_, err = r.Next()
	if err != nil {
		t.Fatalf("Next() returned error: %v", err)
	}
	_, err = r.Next()
	if err != io.EOF {
		t.Fatalf("Next() returned an unexpected record")
	}
}

func TestWriteCorrupted(t *testing.T) {
	// this should work just as well, corrupted central dir
	f, err := os.Open("../test_files/corrupted.zip")
	if err != nil {
		t.Fatal("Could not open test file")
	}
	rp, wp := io.Pipe()
	go func(){
		if err := Write(wp, szip.NewReader(f)); err == nil {
			t.Error("Write() didn't return error on corrupted file")
		}
	}()
	r := tar.NewReader(rp)
	_, err = r.Next()
	if err != nil {
		t.Fatalf("Next() returned error: %v", err)
	}
	_, err = r.Next()
	if err != nil {
		t.Fatalf("Next() returned error: %v", err)
	}
	_, err = r.Next()
	if err != nil {
		t.Fatalf("Next() returned error: %v", err)
	}
	_, err = r.Next()
	if err != io.EOF {
		t.Fatalf("Next() returned an unexpected record")
	}
}
