package carousel

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/pdfcpu/pdfcpu/pkg/api"
)

func RenderPDF(carouselDir string, outputPath string) (string, error) {
	config, err := LoadConfig(carouselDir)
	if err != nil {
		return "", err
	}

	slides, err := ListSlides(carouselDir)
	if err != nil {
		return "", err
	}
	if len(slides) == 0 {
		return "", fmt.Errorf("no slides found in %s", carouselDir)
	}

	if outputPath == "" {
		outputPath = filepath.Join(carouselDir, config.Name+".pdf")
	}

	// Start a local server to serve slides and assets
	addr, cleanup, err := startRenderServer(carouselDir)
	if err != nil {
		return "", fmt.Errorf("failed to start render server: %w", err)
	}
	defer cleanup()

	// Single slide: render directly to output
	if len(slides) == 1 {
		slideURL := fmt.Sprintf("%s/slide/%d", addr, slides[0])
		pdfData, err := renderSlideToPDF(slideURL, config)
		if err != nil {
			return "", fmt.Errorf("failed to render slide %d: %w", slides[0], err)
		}
		if err := os.WriteFile(outputPath, pdfData, 0644); err != nil {
			return "", fmt.Errorf("failed to write PDF: %w", err)
		}
		absPath, _ := filepath.Abs(outputPath)
		return absPath, nil
	}

	// Multiple slides: render each to temp PDF, then merge with pdfcpu
	tmpDir, err := os.MkdirTemp("", "fycr-render-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	var tmpPDFs []string
	for _, slideNum := range slides {
		slideURL := fmt.Sprintf("%s/slide/%d", addr, slideNum)
		pdfData, err := renderSlideToPDF(slideURL, config)
		if err != nil {
			return "", fmt.Errorf("failed to render slide %d: %w", slideNum, err)
		}
		tmpPath := filepath.Join(tmpDir, fmt.Sprintf("%03d.pdf", slideNum))
		if err := os.WriteFile(tmpPath, pdfData, 0644); err != nil {
			return "", err
		}
		tmpPDFs = append(tmpPDFs, tmpPath)
	}

	// Merge all single-page PDFs into one
	if err := api.MergeCreateFile(tmpPDFs, outputPath, false, nil); err != nil {
		return "", fmt.Errorf("failed to merge PDFs: %w", err)
	}

	absPath, _ := filepath.Abs(outputPath)
	return absPath, nil
}

func startRenderServer(carouselDir string) (string, func(), error) {
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

	// Serve local assets
	assetsDir := filepath.Join(carouselDir, "assets")
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(assetsDir))))

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", nil, err
	}

	server := &http.Server{Handler: mux}
	go server.Serve(listener)

	addr := fmt.Sprintf("http://127.0.0.1:%d", listener.Addr().(*net.TCPAddr).Port)
	cleanup := func() {
		server.Close()
	}

	return addr, cleanup, nil
}

func renderSlideToPDF(slideURL string, config *Config) ([]byte, error) {
	// Convert pixels to inches (96 DPI standard screen)
	widthInches := float64(config.Width) / 96.0
	heightInches := float64(config.Height) / 96.0

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var pdfData []byte
	err := chromedp.Run(ctx,
		chromedp.EmulateViewport(int64(config.Width), int64(config.Height)),
		chromedp.Navigate(slideURL),
		chromedp.WaitReady("body"),
		chromedp.Sleep(1*time.Second), // wait for Tailwind CDN
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().
				WithPaperWidth(widthInches).
				WithPaperHeight(heightInches).
				WithMarginTop(0).
				WithMarginBottom(0).
				WithMarginLeft(0).
				WithMarginRight(0).
				WithPrintBackground(true).
				WithPreferCSSPageSize(false).
				Do(ctx)
			if err != nil {
				return err
			}
			pdfData = buf
			return nil
		}),
	)

	return pdfData, err
}
