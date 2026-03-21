package carousel

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func StartPreviewServer(carouselDir string) (string, error) {
	config, err := LoadConfig(carouselDir)
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}

	slides, err := ListSlides(carouselDir)
	if err != nil {
		return "", fmt.Errorf("failed to list slides: %w", err)
	}
	if len(slides) == 0 {
		return "", fmt.Errorf("no slides found in %s", carouselDir)
	}

	mux := http.NewServeMux()

	// Serve individual slide HTML files
	mux.HandleFunc("/slide/", func(w http.ResponseWriter, r *http.Request) {
		numStr := r.URL.Path[len("/slide/"):]
		num, err := strconv.Atoi(numStr)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		slidePath := SlidePath(carouselDir, num)
		data, err := os.ReadFile(slidePath)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	})

	// Serve the main viewer page
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		io.WriteString(w, viewerHTML(config, slides))
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", fmt.Errorf("failed to start server: %w", err)
	}

	addr := fmt.Sprintf("http://127.0.0.1:%d", listener.Addr().(*net.TCPAddr).Port)

	go http.Serve(listener, mux)

	return addr, nil
}

func viewerHTML(config *Config, slides []int) string {
	// Build JS array of slide numbers
	parts := make([]string, len(slides))
	for i, s := range slides {
		parts[i] = strconv.Itoa(s)
	}
	slidesJS := "[" + strings.Join(parts, ",") + "]"

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>FYCR Preview</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            background: #1a1a1a;
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            min-height: 100vh;
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            color: #fff;
        }
        .controls {
            display: flex;
            align-items: center;
            gap: 16px;
            margin-bottom: 16px;
        }
        .controls button {
            background: #333;
            color: #fff;
            border: 1px solid #555;
            border-radius: 6px;
            padding: 8px 16px;
            font-size: 16px;
            cursor: pointer;
            transition: background 0.15s;
        }
        .controls button:hover { background: #444; }
        .controls button:disabled { opacity: 0.3; cursor: default; }
        .slide-info {
            font-size: 14px;
            color: #aaa;
            min-width: 80px;
            text-align: center;
        }
        .slide-frame {
            border: 1px solid #333;
            box-shadow: 0 4px 24px rgba(0,0,0,0.5);
        }
        iframe {
            display: block;
            border: none;
            width: %dpx;
            height: %dpx;
        }
        .hint {
            margin-top: 12px;
            font-size: 12px;
            color: #666;
        }
    </style>
</head>
<body>
    <div class="controls">
        <button id="prev" onclick="navigate(-1)">&larr; Prev</button>
        <span class="slide-info" id="info">1 / 1</span>
        <button id="next" onclick="navigate(1)">Next &rarr;</button>
    </div>
    <div class="slide-frame">
        <iframe id="slide"></iframe>
    </div>
    <div class="hint">Use arrow keys to navigate</div>

    <script>
        const slides = %s;
        let currentIndex = 0;

        function navigate(delta) {
            const newIndex = currentIndex + delta;
            if (newIndex < 0 || newIndex >= slides.length) return;
            currentIndex = newIndex;
            render();
        }

        function render() {
            document.getElementById('slide').src = '/slide/' + slides[currentIndex];
            document.getElementById('info').textContent = (currentIndex + 1) + ' / ' + slides.length;
            document.getElementById('prev').disabled = currentIndex === 0;
            document.getElementById('next').disabled = currentIndex === slides.length - 1;
        }

        document.addEventListener('keydown', function(e) {
            if (e.key === 'ArrowLeft') navigate(-1);
            if (e.key === 'ArrowRight') navigate(1);
        });

        render();
    </script>
</body>
</html>`, config.Width, config.Height, slidesJS)
}
