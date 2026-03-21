package carousel

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
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

	// Set up file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return "", fmt.Errorf("failed to create file watcher: %w", err)
	}
	if err := watcher.Add(carouselDir); err != nil {
		watcher.Close()
		return "", fmt.Errorf("failed to watch directory: %w", err)
	}
	assetsDir := filepath.Join(carouselDir, "assets")
	if _, err := os.Stat(assetsDir); err == nil {
		watcher.Add(assetsDir)
	}

	// SSE clients
	var mu sync.Mutex
	clients := make(map[chan struct{}]struct{})

	// Watch for file changes and notify SSE clients
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
					mu.Lock()
					for ch := range clients {
						select {
						case ch <- struct{}{}:
						default:
						}
					}
					mu.Unlock()
				}
			case _, ok := <-watcher.Errors:
				if !ok {
					return
				}
			}
		}
	}()

	mux := http.NewServeMux()

	// SSE endpoint for live reload
	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming not supported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		ch := make(chan struct{}, 1)
		mu.Lock()
		clients[ch] = struct{}{}
		mu.Unlock()

		defer func() {
			mu.Lock()
			delete(clients, ch)
			mu.Unlock()
		}()

		for {
			select {
			case <-ch:
				fmt.Fprintf(w, "data: reload\n\n")
				flusher.Flush()
			case <-r.Context().Done():
				return
			}
		}
	})

	// JSON endpoint returning current slide list
	mux.HandleFunc("/slides", func(w http.ResponseWriter, r *http.Request) {
		currentSlides, err := ListSlides(carouselDir)
		if err != nil {
			http.Error(w, "failed to list slides", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(currentSlides)
	})

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

	// Serve local assets
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(assetsDir))))

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
		watcher.Close()
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
        .live-dot {
            width: 8px;
            height: 8px;
            border-radius: 50%%;
            background: #4ade80;
            display: inline-block;
            margin-right: 4px;
            animation: pulse 2s infinite;
        }
        @keyframes pulse {
            0%%, 100%% { opacity: 1; }
            50%% { opacity: 0.4; }
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
    <div class="hint"><span class="live-dot"></span>Live reload active &mdash; Use arrow keys to navigate</div>

    <script>
        let slides = %s;
        let currentIndex = 0;

        function navigate(delta) {
            const newIndex = currentIndex + delta;
            if (newIndex < 0 || newIndex >= slides.length) return;
            currentIndex = newIndex;
            render();
        }

        function render() {
            document.getElementById('slide').src = '/slide/' + slides[currentIndex] + '?t=' + Date.now();
            document.getElementById('info').textContent = (currentIndex + 1) + ' / ' + slides.length;
            document.getElementById('prev').disabled = currentIndex === 0;
            document.getElementById('next').disabled = currentIndex === slides.length - 1;
        }

        document.addEventListener('keydown', function(e) {
            if (e.key === 'ArrowLeft') navigate(-1);
            if (e.key === 'ArrowRight') navigate(1);
        });

        // Live reload via SSE
        function connectSSE() {
            const evtSource = new EventSource('/events');
            evtSource.onmessage = function(event) {
                if (event.data === 'reload') {
                    // Refresh slide list and current slide
                    fetch('/slides')
                        .then(r => r.json())
                        .then(newSlides => {
                            const currentSlideNum = slides[currentIndex];
                            slides = newSlides;
                            // Try to stay on the same slide
                            const newIndex = slides.indexOf(currentSlideNum);
                            currentIndex = newIndex >= 0 ? newIndex : Math.min(currentIndex, slides.length - 1);
                            render();
                        })
                        .catch(() => render());
                }
            };
            evtSource.onerror = function() {
                evtSource.close();
                setTimeout(connectSSE, 1000);
            };
        }

        render();
        connectSSE();
    </script>
</body>
</html>`, config.Width, config.Height, slidesJS)
}
