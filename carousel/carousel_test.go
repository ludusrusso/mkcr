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
			"/style.css",
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

func TestInit_CreatesAssetsDir(t *testing.T) {
	tmp := t.TempDir()
	dir, err := Init(tmp, "test", DefaultWidth, DefaultHeight)
	if err != nil {
		t.Fatalf("Init() error: %v", err)
	}

	assetsDir := filepath.Join(dir, "assets")
	info, err := os.Stat(assetsDir)
	if err != nil {
		t.Fatalf("assets dir not found: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("assets is not a directory")
	}
}

func TestInit_CreatesStyleCSS(t *testing.T) {
	tmp := t.TempDir()
	dir, err := Init(tmp, "test", DefaultWidth, DefaultHeight)
	if err != nil {
		t.Fatalf("Init() error: %v", err)
	}

	stylePath := filepath.Join(dir, "style.css")
	data, err := os.ReadFile(stylePath)
	if err != nil {
		t.Fatalf("style.css not found: %v", err)
	}
	if len(data) == 0 {
		t.Error("style.css is empty")
	}
}

func TestLoadConfig(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		tmp := t.TempDir()
		want := Config{Name: "test", Width: 500, Height: 700}
		data, _ := json.Marshal(want)
		err := os.WriteFile(filepath.Join(tmp, "carousel.json"), data, 0644)
		if err != nil {
			t.Fatalf("failed to write carousel.json: %v", err)
		}

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
			if err := os.WriteFile(filepath.Join(tmp, name), []byte("x"), 0644); err != nil {
				t.Fatalf("WriteFile(%s) error: %v", name, err)
			}
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
