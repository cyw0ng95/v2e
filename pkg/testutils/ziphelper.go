package testutils

import (
	"archive/zip"
	"bytes"
)

func MakeZip(entries map[string][]byte) ([]byte, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, data := range entries {
		f, err := zw.Create(name)
		if err != nil {
			return nil, err
		}
		if _, err := f.Write(data); err != nil {
			return nil, err
		}
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
