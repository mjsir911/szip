package szip

import (
	"testing"
	"os"
	"io"
	"archive/zip"
)

func TestReader(t *testing.T) {
	f, err := os.Open("test_files/corrupted_eocd.zip")
	if err != nil {
		t.Fatal("Could not open test file")
	}
	r := NewReader(f)

	var (
		hdr zip.FileHeader
		buf []byte
		n int
	)

	// file2 now
	hdr, err = r.Next()
	if err != nil {
		t.Fatal("Could not read first record")
	}
	if hdr.Method != zip.Deflate {
		t.Errorf("Expected first record to be flated (for testing), found: %v", hdr.Method)
	}
	if hdr.UncompressedSize64 != uint64(hdr.UncompressedSize) {
		t.Error("UncompressedSiz64 != UncompressedSize")
	}
	if hdr.UncompressedSize64 <= hdr.CompressedSize64 {
		t.Error("UncompressedSize is not greater than CompressedSize")
	}

	buf = make([]byte, 6)
	n, err = r.Read(buf)
	if err != nil {
		t.Errorf("Initial read on first record returned err: %v", err)
	}
	if n != 6 {
		t.Fatalf("Could not read first 6 bytes from first record, read: %v", n)
	}
	if string(buf) != "Lorem " {
		t.Errorf("Unexpected first 6 bytes of first record: %v", string(buf))
	}
	if r.Close() != nil { // for funsies
		t.Error("Close() in the middle of the first record failed")
	}

	// directory/file now, in the middle of the previous record's read
	hdr, err = r.Next()
	if err != nil {
		t.Fatal("Could not read second record, in the middle of the first record's Read()")
	}
	if hdr.Method != zip.Store {
		t.Errorf("Expected second record to be stored (for testing), found: %v", hdr.Method)
	}
	if hdr.Name != "directory/file" {
		t.Fatalf("Unexpected file name on second record: %v", hdr.Name)
	}
	if hdr.FileInfo().IsDir() {
		t.Fatal("second record is not file")
	}
	if hdr.FileInfo().Size() == 0 {
		t.Fatal("second record is empty")
	}

	devnull, _ := os.OpenFile(os.DevNull, os.O_WRONLY, 0644)
	_, err = io.Copy(devnull, r)
	if err != nil {
		t.Fatalf("io.Copy() on second record failed: %v", err)
	}
	if r.Close() != nil { // for funsies
		t.Error("Close() on second record failed")
	}


	// directory/ now
	hdr, err = r.Next()
	if err != nil {
		t.Fatal("Could not read third record")
	}
	if hdr.Name != "directory/" {
		t.Fatalf("Unexpected file name on third record: %v", hdr.Name)
	}
	if !hdr.FileInfo().IsDir() {
		t.Fatal("Third record is not directory")
	}
	if hdr.FileInfo().Size() != 0 {
		t.Fatal("Third record is not empty (expected directory)")
	}
	// this should work even though it doesn't make sense
	n, err = r.Read(buf);
	if n != 0 {
		t.Fatal("Third record Read is not empty")
	}
	if err != io.EOF {
		t.Fatalf("Third record Read returned err: %v (not io.EOF)", err)
	}
	// let's not close this()


	if _, err = r.Next(); err != io.EOF {
		t.Error("r.Next() didn't return EOF after three expected files")
	}

	if _, err = r.Next(); err == nil {
		t.Error("r.Next() again should return an error")
		
	}


}

func TestReaderCentralDirectory(t *testing.T) {
	f, err := os.Open("test_files/test.zip")
	if err != nil {
		t.Fatal("Could not open test file")
	}
	r := NewReader(f)

	// CentralDirectory() should be nondestructive
	_, err = r.CentralDirectory()
	if err == nil {
		t.Fatal("CentralDirectory() did not fail on premature read")
	}

	for err = nil; err != io.EOF; _, err = r.Next() {
		if err != nil {
			t.Fatalf("Unexpected error seeking through file: %v", err)
		}
	}
	// CentralDirectory() should work now, we are done
	files, err := r.CentralDirectory()
	if err != nil {
		t.Fatalf("CentralDirectory() returned error: %v", err)
	}

	if len(files) != 3 {
		t.Fatalf("Unexpected amount of records in CentralDirectory(): %v", len(files))
	}
	if files[0].Name != "file2" {
		t.Errorf("Unexpected name of first record in CentralDirectory()")
	}
	if files[1].Name != "directory/file" {
		t.Errorf("Unexpected name of second record in CentralDirectory()")
	}
	if files[2].Name != "directory/" {
		t.Errorf("Unexpected name of third record in CentralDirectory()")
	}
}

func TestReaderCorruptedEOCD(t *testing.T) {
	f, err := os.Open("test_files/corrupted_eocd.zip")
	if err != nil {
		t.Fatal("Could not open test file")
	}
	r := NewReader(f)

	for err = nil; err != io.EOF; _, err = r.Next() {
		if err != nil {
			t.Fatalf("Unexpected error seeking through file: %v", err)
		}
	}
	// CentralDirectory() should work now, we are done
	files, err := r.CentralDirectory()
	if err == nil {
		t.Error("CentralDirectory() did not return error on corrupted file")
	}
	if len(files) != 3 {
		t.Error("CentralDirectory() does not contain partial header data")
	}
}
