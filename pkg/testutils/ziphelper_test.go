package testutils

import (
	"archive/zip"
	"bytes"
	"io"
	"testing"
)

func TestMakeZip_MultipleEntries(t *testing.T) {
	entries := map[string][]byte{
		"file1.txt":     []byte("hello"),
		"dir/file2.txt": []byte("world"),
	}

	zipped, err := MakeZip(entries)
	if err != nil {
		t.Fatalf("MakeZip returned error: %v", err)
	}

	r, err := zip.NewReader(bytes.NewReader(zipped), int64(len(zipped)))
	if err != nil {
		t.Fatalf("zip.NewReader error: %v", err)
	}

	if len(r.File) != len(entries) {
		t.Fatalf("Expected %d entries, got %d", len(entries), len(r.File))
	}

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("Open failed for %s: %v", f.Name, err)
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			t.Fatalf("Read failed for %s: %v", f.Name, err)
		}
		if want := entries[f.Name]; !bytes.Equal(data, want) {
			t.Fatalf("Content mismatch for %s: got %q want %q", f.Name, data, want)
		}
	}
}

func TestMakeZip_Empty(t *testing.T) {
	zipped, err := MakeZip(map[string][]byte{})
	if err != nil {
		t.Fatalf("MakeZip returned error: %v", err)
	}

	r, err := zip.NewReader(bytes.NewReader(zipped), int64(len(zipped)))
	if err != nil {
		t.Fatalf("zip.NewReader error: %v", err)
	}

	if len(r.File) != 0 {
		t.Fatalf("Expected no files in archive, got %d", len(r.File))
	}
}
