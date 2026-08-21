package middleware_test

import (
	"bytes"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"testing"

	"github.com/gbplantwiki/gbplantwiki/internal/util"
)

func r007TestFile(t *testing.T, extraSize int64) *multipart.FileHeader {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="a.png"`)
	h.Set("Content-Type", "image/png")
	pw, err := mw.CreatePart(h)
	if err != nil {
		t.Fatalf("create part: %v", err)
	}
	pw.Write([]byte("abc"))
	mw.Close()
	form, err := multipart.NewReader(&buf, mw.Boundary()).ReadForm(1 << 20)
	if err != nil {
		t.Fatalf("multipart form: %v", err)
	}
	f := form.File["file"][0]
	f.Size += extraSize
	return f
}

func TestUploadErrorNotSwallowed(t *testing.T) {
	dir := t.TempDir()
	f := r007TestFile(t, 100)
	path, err := util.SaveUpload(dir, f)
	if err == nil {
		t.Fatalf("SaveUpload must surface the copy error, got path=%q err=nil", path)
	}
}

func TestUploadNoPartialFileOnError(t *testing.T) {
	dir := t.TempDir()
	f := r007TestFile(t, 100)
	_, _ = util.SaveUpload(dir, f)
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".png" {
			t.Fatalf("partial file %s left behind after failed upload", e.Name())
		}
	}
}

func TestUploadRejectsEmptyFile(t *testing.T) {
	dir := t.TempDir()
	f := r007TestFile(t, -3) // shrink size below actual -> 0 bytes copied? no: size 0
	f.Size = 0
	_, err := util.SaveUpload(dir, f)
	if err == nil {
		t.Fatalf("empty file must be rejected")
	}
}
