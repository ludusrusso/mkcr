# mkcr — Make Carousel

A Go CLI tool designed for AI agents to generate LinkedIn carousel PDFs from HTML+Tailwind CSS slides.

## Overview

`mkcr` lets you build carousel slide decks one slide at a time using HTML and Tailwind CSS. Each slide is stored as a self-contained HTML file that can be previewed in a browser, then rendered into a single multi-page PDF ready for LinkedIn upload.

## Installation

```bash
go install github.com/ludusrusso/mkcr@latest
```

Requires Google Chrome or Chromium installed for PDF rendering.

## Usage

### Initialize a carousel

```bash
mkcr init my-post
```

Creates a `./my-post/` folder with a `carousel.json` config file.

### Add slides

```bash
# Via --html flag
mkcr add my-post --html '<div class="text-4xl font-bold">Hello World</div>'

# Via stdin
echo '<div class="text-4xl font-bold">Hello World</div>' | mkcr add my-post

# From a file
mkcr add my-post --file slide-content.html

# Insert/overwrite at a specific position
mkcr add my-post --position 2 --html '<div>Slide 2 updated</div>'
```

You only pass the inner content — the tool wraps it in a full HTML template with Tailwind CSS.

### Render to PDF

```bash
mkcr render my-post
# Outputs: /absolute/path/to/my-post/my-post.pdf

mkcr render my-post --output ./output.pdf
```

### Preview in browser

```bash
mkcr preview my-post
```

## Slide dimensions

Default: **1080x1350px** (portrait, 4:5 ratio for LinkedIn).

Configurable in `carousel.json`:

```json
{
  "name": "my-post",
  "width": 1080,
  "height": 1350
}
```

## How it works

1. Slides are stored as numbered HTML files (`1.html`, `2.html`, ...) in the carousel folder
2. Each HTML file is self-contained with Tailwind CDN — open in any browser to preview
3. `render` uses headless Chrome (via chromedp) to convert each slide to a PDF page
4. Pages are merged into a single multi-page PDF
