package carousel

import (
	"os"
	"strings"
	"testing"
)

func TestRenderPNG_NoSlides(t *testing.T) {
	tmp := t.TempDir()
	dir, err := Init(tmp, "empty", DefaultWidth, DefaultHeight)
	if err != nil {
		t.Fatalf("Init() error: %v", err)
	}

	_, err = RenderPNG(dir, "")
	if err == nil {
		t.Fatal("expected error when no slides exist")
	}
	if !strings.Contains(err.Error(), "no slides") {
		t.Errorf("error = %q, want it to mention 'no slides'", err.Error())
	}
}

func TestRenderPNG_WritesNumberedFiles(t *testing.T) {
	dir := setupCarousel(t, "<p>slide one</p>", "<p>slide two</p>")
	outputDir := t.TempDir()

	result, err := RenderPNG(dir, outputDir)
	if err != nil {
		t.Fatalf("RenderPNG() error: %v", err)
	}

	// Verify output directory path returned
	if result == "" {
		t.Fatal("RenderPNG() returned empty path")
	}

	// Verify correct files exist
	for _, name := range []string{"1.png", "2.png"} {
		info, err := os.Stat(result + "/" + name)
		if err != nil {
			t.Errorf("expected file %s not found: %v", name, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("file %s is empty", name)
		}
	}

	// Verify no extra files
	entries, _ := os.ReadDir(result)
	if len(entries) != 2 {
		t.Errorf("expected 2 files, got %d", len(entries))
	}
}

func TestRenderPNG_DefaultOutputDir(t *testing.T) {
	dir := setupCarousel(t, "<p>slide</p>")

	result, err := RenderPNG(dir, "")
	if err != nil {
		t.Fatalf("RenderPNG() error: %v", err)
	}

	expectedSuffix := "/images"
	if !strings.HasSuffix(result, expectedSuffix) {
		t.Errorf("RenderPNG() = %q, want suffix %q", result, expectedSuffix)
	}

	// Verify file was written there
	_, err = os.Stat(result + "/1.png")
	if err != nil {
		t.Errorf("expected 1.png in default dir: %v", err)
	}
}

func TestRenderPNG_ClearsOutputDir(t *testing.T) {
	dir := setupCarousel(t, "<p>slide</p>")
	outputDir := t.TempDir()

	// Write a stale file
	if err := os.WriteFile(outputDir+"/stale.png", []byte("old"), 0644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	_, err := RenderPNG(dir, outputDir)
	if err != nil {
		t.Fatalf("RenderPNG() error: %v", err)
	}

	// Stale file should be gone
	if _, err := os.Stat(outputDir + "/stale.png"); !os.IsNotExist(err) {
		t.Error("stale.png should have been removed")
	}

	// New file should exist
	if _, err := os.Stat(outputDir + "/1.png"); err != nil {
		t.Errorf("expected 1.png: %v", err)
	}
}

func TestRenderPDF_StillWorks(t *testing.T) {
	dir := setupCarousel(t, "<p>pdf slide</p>")

	result, err := RenderPDF(dir, "")
	if err != nil {
		t.Fatalf("RenderPDF() error: %v", err)
	}

	if !strings.HasSuffix(result, ".pdf") {
		t.Errorf("RenderPDF() = %q, want .pdf suffix", result)
	}

	info, err := os.Stat(result)
	if err != nil {
		t.Fatalf("output PDF not found: %v", err)
	}
	if info.Size() == 0 {
		t.Error("output PDF is empty")
	}
}
