package carousel

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Config struct {
	Name   string `json:"name"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

const DefaultWidth = 1080
const DefaultHeight = 1080

type Format struct {
	Width  int
	Height int
}

var Formats = map[string]Format{
	"square":   {Width: 1080, Height: 1080},
	"vertical": {Width: 1080, Height: 1350},
}

func Init(dir string, name string, width int, height int) (string, error) {
	carouselDir := filepath.Join(dir, name)
	if err := os.MkdirAll(carouselDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	config := Config{
		Name:   name,
		Width:  width,
		Height: height,
	}

	configPath := filepath.Join(carouselDir, "carousel.json")
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write config: %w", err)
	}

	assetsDir := filepath.Join(carouselDir, "assets")
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create assets directory: %w", err)
	}

	absPath, _ := filepath.Abs(carouselDir)
	return absPath, nil
}

func LoadConfig(carouselDir string) (*Config, error) {
	configPath := filepath.Join(carouselDir, "carousel.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read carousel.json: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse carousel.json: %w", err)
	}

	return &config, nil
}

func NextSlideNumber(carouselDir string) (int, error) {
	slides, err := ListSlides(carouselDir)
	if err != nil {
		return 1, nil
	}
	if len(slides) == 0 {
		return 1, nil
	}
	return slides[len(slides)-1] + 1, nil
}

func ListSlides(carouselDir string) ([]int, error) {
	entries, err := os.ReadDir(carouselDir)
	if err != nil {
		return nil, err
	}

	var slides []int
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".html") {
			continue
		}
		numStr := strings.TrimSuffix(name, ".html")
		num, err := strconv.Atoi(numStr)
		if err != nil {
			continue
		}
		slides = append(slides, num)
	}

	sort.Ints(slides)
	return slides, nil
}

func SlidePath(carouselDir string, position int) string {
	return filepath.Join(carouselDir, fmt.Sprintf("%d.html", position))
}

func WrapHTML(content string, config *Config) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=%d, initial-scale=1.0">
    <script src="https://cdn.tailwindcss.com"></script>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        html, body {
            width: %dpx;
            height: %dpx;
            overflow: hidden;
        }
    </style>
</head>
<body>
    %s
</body>
</html>`, config.Width, config.Width, config.Height, content)
}

func AddSlide(carouselDir string, content string, position int) (string, error) {
	config, err := LoadConfig(carouselDir)
	if err != nil {
		return "", err
	}

	html := WrapHTML(content, config)
	slidePath := SlidePath(carouselDir, position)

	if err := os.WriteFile(slidePath, []byte(html), 0644); err != nil {
		return "", fmt.Errorf("failed to write slide: %w", err)
	}

	absPath, _ := filepath.Abs(slidePath)
	return absPath, nil
}
