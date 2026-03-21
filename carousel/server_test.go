package carousel

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupCarousel creates a temp carousel with the given slide contents.
func setupCarousel(t *testing.T, slideContents ...string) string {
	t.Helper()
	tmp := t.TempDir()
	dir, err := Init(tmp, "test", DefaultWidth, DefaultHeight)
	if err != nil {
		t.Fatalf("Init() error: %v", err)
	}
	for i, content := range slideContents {
		if _, err := AddSlide(dir, content, i+1); err != nil {
			t.Fatalf("AddSlide(%d) error: %v", i+1, err)
		}
	}
	return dir
}

func TestViewerHTML(t *testing.T) {
	t.Run("single slide", func(t *testing.T) {
		cfg := &Config{Name: "test", Width: 1080, Height: 1350}
		html := viewerHTML(cfg, []int{1})

		for _, s := range []string{
			"FYCR Preview",
			"[1]",
			"1080",
			"1350",
		} {
			if !strings.Contains(html, s) {
				t.Errorf("viewerHTML output missing %q", s)
			}
		}
	})

	t.Run("multiple slides", func(t *testing.T) {
		cfg := &Config{Name: "test", Width: 800, Height: 600}
		html := viewerHTML(cfg, []int{1, 2, 3})

		if !strings.Contains(html, "[1,2,3]") {
			t.Error("viewerHTML missing slides array [1,2,3]")
		}
	})

	t.Run("non-contiguous slides", func(t *testing.T) {
		cfg := &Config{Name: "test", Width: 800, Height: 600}
		html := viewerHTML(cfg, []int{1, 3, 7})

		if !strings.Contains(html, "[1,3,7]") {
			t.Error("viewerHTML missing slides array [1,3,7]")
		}
	})
}

func TestStartPreviewServer_RootHandler(t *testing.T) {
	dir := setupCarousel(t, "<p>slide 1</p>", "<p>slide 2</p>")
	addr, err := StartPreviewServer(dir)
	if err != nil {
		t.Fatalf("StartPreviewServer() error: %v", err)
	}

	resp, err := http.Get(addr + "/")
	if err != nil {
		t.Fatalf("GET / error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("GET / status = %d, want 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	html := string(body)

	for _, s := range []string{"FYCR Preview", "[1,2]"} {
		if !strings.Contains(html, s) {
			t.Errorf("GET / response missing %q", s)
		}
	}
}

func TestStartPreviewServer_SlideHandler(t *testing.T) {
	dir := setupCarousel(t, "<p>hello world</p>")
	addr, err := StartPreviewServer(dir)
	if err != nil {
		t.Fatalf("StartPreviewServer() error: %v", err)
	}

	resp, err := http.Get(addr + "/slide/1")
	if err != nil {
		t.Fatalf("GET /slide/1 error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("GET /slide/1 status = %d, want 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "<p>hello world</p>") {
		t.Error("GET /slide/1 missing slide content")
	}
}

func TestStartPreviewServer_SlideNotFound(t *testing.T) {
	dir := setupCarousel(t, "<p>only slide</p>")
	addr, err := StartPreviewServer(dir)
	if err != nil {
		t.Fatalf("StartPreviewServer() error: %v", err)
	}

	for _, path := range []string{"/slide/999", "/slide/abc"} {
		resp, err := http.Get(addr + path)
		if err != nil {
			t.Fatalf("GET %s error: %v", path, err)
		}
		resp.Body.Close()

		if resp.StatusCode != 404 {
			t.Errorf("GET %s status = %d, want 404", path, resp.StatusCode)
		}
	}
}

func TestStartPreviewServer_UnknownPath(t *testing.T) {
	dir := setupCarousel(t, "<p>slide</p>")
	addr, err := StartPreviewServer(dir)
	if err != nil {
		t.Fatalf("StartPreviewServer() error: %v", err)
	}

	resp, err := http.Get(addr + "/unknown")
	if err != nil {
		t.Fatalf("GET /unknown error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 404 {
		t.Errorf("GET /unknown status = %d, want 404", resp.StatusCode)
	}
}

func TestStartPreviewServer_AssetsHandler(t *testing.T) {
	dir := setupCarousel(t, "<p>slide</p>")

	// Create an asset file
	assetsDir := filepath.Join(dir, "assets")
	os.WriteFile(filepath.Join(assetsDir, "test.txt"), []byte("hello asset"), 0644)

	addr, err := StartPreviewServer(dir)
	if err != nil {
		t.Fatalf("StartPreviewServer() error: %v", err)
	}

	resp, err := http.Get(addr + "/assets/test.txt")
	if err != nil {
		t.Fatalf("GET /assets/test.txt error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("GET /assets/test.txt status = %d, want 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "hello asset" {
		t.Errorf("GET /assets/test.txt body = %q, want %q", string(body), "hello asset")
	}
}

func TestStartPreviewServer_AssetsSubdir(t *testing.T) {
	dir := setupCarousel(t, "<p>slide</p>")

	// Create a nested asset
	subDir := filepath.Join(dir, "assets", "images")
	os.MkdirAll(subDir, 0755)
	os.WriteFile(filepath.Join(subDir, "logo.png"), []byte("fake png"), 0644)

	addr, err := StartPreviewServer(dir)
	if err != nil {
		t.Fatalf("StartPreviewServer() error: %v", err)
	}

	resp, err := http.Get(addr + "/assets/images/logo.png")
	if err != nil {
		t.Fatalf("GET /assets/images/logo.png error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("GET /assets/images/logo.png status = %d, want 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "fake png" {
		t.Errorf("body = %q, want %q", string(body), "fake png")
	}
}

func TestStartPreviewServer_AssetNotFound(t *testing.T) {
	dir := setupCarousel(t, "<p>slide</p>")

	addr, err := StartPreviewServer(dir)
	if err != nil {
		t.Fatalf("StartPreviewServer() error: %v", err)
	}

	resp, err := http.Get(addr + "/assets/nonexistent.png")
	if err != nil {
		t.Fatalf("GET /assets/nonexistent.png error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 404 {
		t.Errorf("GET /assets/nonexistent.png status = %d, want 404", resp.StatusCode)
	}
}

func TestStartPreviewServer_NoSlides(t *testing.T) {
	tmp := t.TempDir()
	dir, err := Init(tmp, "empty", DefaultWidth, DefaultHeight)
	if err != nil {
		t.Fatalf("Init() error: %v", err)
	}

	_, err = StartPreviewServer(dir)
	if err == nil {
		t.Fatal("expected error when no slides exist")
	}
	if !strings.Contains(err.Error(), "no slides") {
		t.Errorf("error = %q, want it to mention 'no slides'", err.Error())
	}
}
