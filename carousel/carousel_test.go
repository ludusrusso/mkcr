package carousel

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestSlidePath(t *testing.T) {
	tests := []struct {
		dir      string
		position int
		want     string
	}{
		{"dir", 1, filepath.Join("dir", "1.html")},
		{"dir", 10, filepath.Join("dir", "10.html")},
		{"/abs/path", 3, filepath.Join("/abs/path", "3.html")},
	}
	for _, tt := range tests {
		got := SlidePath(tt.dir, tt.position)
		if got != tt.want {
			t.Errorf("SlidePath(%q, %d) = %q, want %q", tt.dir, tt.position, got, tt.want)
		}
	}
}

func TestWrapHTML(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		cfg := &Config{Name: "test", Width: 1080, Height: 1350}
		html := WrapHTML("<p>hello</p>", cfg)

		for _, s := range []string{
			"width: 1080px",
			"height: 1350px",
			"<p>hello</p>",
			"cdn.tailwindcss.com",
		} {
			if !strings.Contains(html, s) {
				t.Errorf("WrapHTML output missing %q", s)
			}
		}
	})

	t.Run("custom config", func(t *testing.T) {
		cfg := &Config{Name: "custom", Width: 800, Height: 600}
		html := WrapHTML("<div>custom</div>", cfg)

		for _, s := range []string{
			"width: 800px",
			"height: 600px",
			"<div>custom</div>",
		} {
			if !strings.Contains(html, s) {
				t.Errorf("WrapHTML output missing %q", s)
			}
		}
	})
}

func TestInit(t *testing.T) {
	t.Run("square format", func(t *testing.T) {
		tmp := t.TempDir()
		dir, err := Init(tmp, "mycarousel", 1080, 1080)
		if err != nil {
			t.Fatalf("Init() error: %v", err)
		}

		info, err := os.Stat(dir)
		if err != nil {
			t.Fatalf("created dir not found: %v", err)
		}
		if !info.IsDir() {
			t.Fatal("expected directory")
		}

		data, err := os.ReadFile(filepath.Join(dir, "carousel.json"))
		if err != nil {
			t.Fatalf("carousel.json not found: %v", err)
		}

		var cfg Config
		if err := json.Unmarshal(data, &cfg); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		if cfg.Name != "mycarousel" || cfg.Width != 1080 || cfg.Height != 1080 {
			t.Errorf("unexpected config: %+v", cfg)
		}
	})

	t.Run("vertical format", func(t *testing.T) {
		tmp := t.TempDir()
		dir, err := Init(tmp, "mycarousel", 1080, 1350)
		if err != nil {
			t.Fatalf("Init() error: %v", err)
		}

		data, err := os.ReadFile(filepath.Join(dir, "carousel.json"))
		if err != nil {
			t.Fatalf("carousel.json not found: %v", err)
		}

		var cfg Config
		if err := json.Unmarshal(data, &cfg); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		if cfg.Width != 1080 || cfg.Height != 1350 {
			t.Errorf("unexpected config: %+v", cfg)
		}
	})

	t.Run("custom dimensions", func(t *testing.T) {
		tmp := t.TempDir()
		dir, err := Init(tmp, "custom", 800, 600)
		if err != nil {
			t.Fatalf("Init() error: %v", err)
		}

		data, err := os.ReadFile(filepath.Join(dir, "carousel.json"))
		if err != nil {
			t.Fatalf("carousel.json not found: %v", err)
		}

		var cfg Config
		if err := json.Unmarshal(data, &cfg); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		if cfg.Width != 800 || cfg.Height != 600 {
			t.Errorf("unexpected config: %+v", cfg)
		}
	})
}

func TestLoadConfig(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		tmp := t.TempDir()
		want := Config{Name: "test", Width: 500, Height: 700}
		data, _ := json.Marshal(want)
		os.WriteFile(filepath.Join(tmp, "carousel.json"), data, 0644)

		got, err := LoadConfig(tmp)
		if err != nil {
			t.Fatalf("LoadConfig() error: %v", err)
		}
		if *got != want {
			t.Errorf("LoadConfig() = %+v, want %+v", *got, want)
		}
	})

	t.Run("missing file", func(t *testing.T) {
		tmp := t.TempDir()
		_, err := LoadConfig(tmp)
		if err == nil {
			t.Fatal("expected error for missing carousel.json")
		}
	})
}

func TestListSlides(t *testing.T) {
	t.Run("mixed files", func(t *testing.T) {
		tmp := t.TempDir()
		for _, name := range []string{"1.html", "3.html", "10.html", "notes.txt", "abc.html"} {
			os.WriteFile(filepath.Join(tmp, name), []byte("x"), 0644)
		}

		got, err := ListSlides(tmp)
		if err != nil {
			t.Fatalf("ListSlides() error: %v", err)
		}
		want := []int{1, 3, 10}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("ListSlides() = %v, want %v", got, want)
		}
	})

	t.Run("empty dir", func(t *testing.T) {
		tmp := t.TempDir()
		got, err := ListSlides(tmp)
		if err != nil {
			t.Fatalf("ListSlides() error: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("ListSlides() = %v, want empty", got)
		}
	})
}

func TestNextSlideNumber(t *testing.T) {
	t.Run("empty dir", func(t *testing.T) {
		tmp := t.TempDir()
		got, err := NextSlideNumber(tmp)
		if err != nil {
			t.Fatalf("NextSlideNumber() error: %v", err)
		}
		if got != 1 {
			t.Errorf("NextSlideNumber() = %d, want 1", got)
		}
	})

	t.Run("existing slides", func(t *testing.T) {
		tmp := t.TempDir()
		for _, name := range []string{"1.html", "3.html", "5.html"} {
			os.WriteFile(filepath.Join(tmp, name), []byte("x"), 0644)
		}
		got, err := NextSlideNumber(tmp)
		if err != nil {
			t.Fatalf("NextSlideNumber() error: %v", err)
		}
		if got != 6 {
			t.Errorf("NextSlideNumber() = %d, want 6", got)
		}
	})
}

func TestAddSlide(t *testing.T) {
	tmp := t.TempDir()
	_, err := Init(tmp, "deck", DefaultWidth, DefaultHeight)
	if err != nil {
		t.Fatalf("Init() error: %v", err)
	}

	carouselDir := filepath.Join(tmp, "deck")
	path, err := AddSlide(carouselDir, "<p>slide one</p>", 1)
	if err != nil {
		t.Fatalf("AddSlide() error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("slide file not found: %v", err)
	}

	html := string(data)
	if !strings.Contains(html, "<p>slide one</p>") {
		t.Error("slide missing content")
	}
	if !strings.Contains(html, "cdn.tailwindcss.com") {
		t.Error("slide missing Tailwind CDN")
	}
}
