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

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/pdfcpu/pdfcpu/pkg/api"
)

type slideCapture struct {
	SlideNum int
	PNGData  []byte
}

// captureSlides starts a render server and captures screenshots of all slides.
func captureSlides(carouselDir string, config *Config, slides []int) ([]slideCapture, error) {
	addr, cleanup, err := startRenderServer(carouselDir)
	if err != nil {
		return nil, fmt.Errorf("failed to start render server: %w", err)
	}
	defer cleanup()

	var captures []slideCapture
	for _, slideNum := range slides {
		slideURL := fmt.Sprintf("%s/slide/%d", addr, slideNum)
		pngData, err := renderSlideToScreenshot(slideURL, config)
		if err != nil {
			return nil, fmt.Errorf("failed to render slide %d: %w", slideNum, err)
		}
		captures = append(captures, slideCapture{SlideNum: slideNum, PNGData: pngData})
	}

	return captures, nil
}

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

	addr, cleanup, err := startRenderServer(carouselDir)
	if err != nil {
		return "", fmt.Errorf("failed to start render server: %w", err)
	}
	defer cleanup()

	tmpDir, err := os.MkdirTemp("", "mkcr-render-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	var pdfFiles []string
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
		pdfFiles = append(pdfFiles, tmpPath)
	}

	// Remove output file if it exists
	_ = os.Remove(outputPath)

	if err := api.MergeCreateFile(pdfFiles, outputPath, false, nil); err != nil {
		return "", fmt.Errorf("failed to merge PDFs: %w", err)
	}

	absPath, _ := filepath.Abs(outputPath)
	return absPath, nil
}

func RenderPNG(carouselDir string, outputDir string) (string, error) {
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

	if outputDir == "" {
		outputDir = filepath.Join(carouselDir, "images")
	}

	captures, err := captureSlides(carouselDir, config, slides)
	if err != nil {
		return "", err
	}

	// Clear output directory
	if err := os.RemoveAll(outputDir); err != nil {
		return "", fmt.Errorf("failed to clear output directory: %w", err)
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	for _, cap := range captures {
		outPath := filepath.Join(outputDir, fmt.Sprintf("%d.png", cap.SlideNum))
		if err := os.WriteFile(outPath, cap.PNGData, 0644); err != nil {
			return "", fmt.Errorf("failed to write %s: %w", outPath, err)
		}
	}

	absPath, _ := filepath.Abs(outputDir)
	return absPath, nil
}

func startRenderServer(carouselDir string) (string, func(), error) {
	config, err := LoadConfig(carouselDir)
	if err != nil {
		return "", nil, fmt.Errorf("failed to load config: %w", err)
	}

	mux := http.NewServeMux()

	// Serve individual slide HTML files (wrap raw content at serve time)
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
		wrapped := WrapHTML(string(data), config)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(wrapped))
	})

	// Serve style.css
	mux.HandleFunc("/style.css", func(w http.ResponseWriter, r *http.Request) {
		stylePath := filepath.Join(carouselDir, "style.css")
		data, err := os.ReadFile(stylePath)
		if err != nil {
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
			_, _ = w.Write([]byte(""))
			return
		}
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		_, _ = w.Write(data)
	})

	// Serve local assets
	assetsDir := filepath.Join(carouselDir, "assets")
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(assetsDir))))

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", nil, err
	}

	server := &http.Server{Handler: mux}
	go func() { _ = server.Serve(listener) }()

	addr := fmt.Sprintf("http://127.0.0.1:%d", listener.Addr().(*net.TCPAddr).Port)
	cleanup := func() {
		_ = server.Close()
	}

	return addr, cleanup, nil
}

func renderSlideToScreenshot(slideURL string, config *Config) ([]byte, error) {
	opts := chromedp.DefaultExecAllocatorOptions[:]
	if os.Getenv("CHROMEDP_NO_SANDBOX") != "" {
		opts = append(opts, chromedp.Flag("no-sandbox", true))
	}
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var pngData []byte
	err := chromedp.Run(ctx,
		chromedp.EmulateViewport(int64(config.Width), int64(config.Height), chromedp.EmulateScale(2)),
		chromedp.Navigate(slideURL),
		chromedp.WaitReady("body"),
		chromedp.Sleep(1*time.Second), // wait for Tailwind CDN
		chromedp.FullScreenshot(&pngData, 100),
	)

	return pngData, err
}

func renderSlideToPDF(slideURL string, config *Config) ([]byte, error) {
	opts := chromedp.DefaultExecAllocatorOptions[:]
	if os.Getenv("CHROMEDP_NO_SANDBOX") != "" {
		opts = append(opts, chromedp.Flag("no-sandbox", true))
	}
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
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
			// Force screen media so print-specific CSS doesn't apply
			if err := emulation.SetEmulatedMedia().WithMedia("screen").Do(ctx); err != nil {
				return err
			}
			buf, _, err := page.PrintToPDF().
				WithPrintBackground(true).
				WithPaperWidth(float64(config.Width) / 96.0).
				WithPaperHeight(float64(config.Height) / 96.0).
				WithMarginTop(0).
				WithMarginBottom(0).
				WithMarginLeft(0).
				WithMarginRight(0).
				WithPreferCSSPageSize(false).
				WithDisplayHeaderFooter(false).
				WithScale(1.0).
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
