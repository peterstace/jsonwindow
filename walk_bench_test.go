package jsonwindow_test

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/peterstace/jsonwindow"
)

func BenchmarkWalkObject(b *testing.B) {
	const fname = "./testdata/big.zip"
	zipped, err := os.ReadFile(fname)
	assertNoError(b, err)
	zr, err := zip.NewReader(bytes.NewReader(zipped), int64(len(zipped)))
	assertNoError(b, err)
	assertEq(b, 1, len(zr.File))
	zf, err := zr.File[0].Open()
	assertNoError(b, err)
	raw, err := io.ReadAll(zf)
	zf.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := jsonwindow.WalkObject(raw, nil)
		assertNoError(b, err)
	}
}