package openrouterapigo

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildImageContent(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})

	parts, err := buildImageContent("hello", []image.Image{img})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(parts) != 2 {
		t.Fatalf("expected 2 parts, got %d", len(parts))
	}
	if parts[0].Type != ContentTypeText || parts[0].Text != "hello" {
		t.Fatalf("unexpected text part: %+v", parts[0])
	}
	if parts[1].Type != ContentTypeImage || parts[1].ImageURL == nil {
		t.Fatalf("expected image part, got %+v", parts[1])
	}
	const prefix = "data:image/png;base64,"
	if !strings.HasPrefix(parts[1].ImageURL.URL, prefix) {
		t.Fatalf("unexpected image data URL prefix: %s", parts[1].ImageURL.URL)
	}
	data, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(parts[1].ImageURL.URL, prefix))
	if err != nil {
		t.Fatalf("failed to decode base64: %v", err)
	}
	if _, err := png.Decode(bytes.NewReader(data)); err != nil {
		t.Fatalf("failed to decode png data: %v", err)
	}
}

func TestBuildPDFContent(t *testing.T) {
	dir := t.TempDir()
	pdfPath := filepath.Join(dir, "test.pdf")
	content := []byte("%PDF-1.4\n%mock\n")
	if err := os.WriteFile(pdfPath, content, 0o644); err != nil {
		t.Fatalf("failed to write temp pdf: %v", err)
	}

	parts, err := buildPDFContent("hello", []string{pdfPath})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(parts) != 2 {
		t.Fatalf("expected 2 parts, got %d", len(parts))
	}
	if parts[1].Type != ContentTypePDF || parts[1].File == nil {
		t.Fatalf("expected pdf part, got %+v", parts[1])
	}
	const prefix = "data:application/pdf;base64,"
	if !strings.HasPrefix(parts[1].File.FileData, prefix) {
		t.Fatalf("unexpected pdf data URL prefix: %s", parts[1].File.FileData)
	}
	data, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(parts[1].File.FileData, prefix))
	if err != nil {
		t.Fatalf("failed to decode base64: %v", err)
	}
	if !bytes.Equal(data, content) {
		t.Fatalf("decoded pdf data mismatch")
	}
}
