# mkcr — Make Carousel

[![CI](https://github.com/ludusrusso/mkcr/actions/workflows/ci.yml/badge.svg)](https://github.com/ludusrusso/mkcr/actions/workflows/ci.yml)
[![Release](https://github.com/ludusrusso/mkcr/actions/workflows/release.yml/badge.svg)](https://github.com/ludusrusso/mkcr/actions/workflows/release.yml)

A Go CLI tool designed for AI agents to generate LinkedIn carousel PDFs from HTML+Tailwind CSS slides.

## Overview

`mkcr` lets you build carousel slide decks using HTML and Tailwind CSS. Each slide is a simple HTML fragment file that gets wrapped automatically with Tailwind CDN. Slides can be previewed in a browser with live reload, then rendered into a single multi-page PDF ready for LinkedIn upload.

## Installation

### Download prebuilt binaries (recommended)

Grab the latest release for your platform from [GitHub Releases](https://github.com/ludusrusso/mkcr/releases/latest).

Available for: **linux/amd64**, **linux/arm64**, **darwin/amd64**, **darwin/arm64**.

Extract the archive and move the binary to a directory in your `PATH`:

```bash
tar xzf mkcr_*.tar.gz
sudo mv mkcr /usr/local/bin/
```

### Install with Go

If you have Go installed, you can install directly:

```bash
go install github.com/ludusrusso/mkcr@latest
```

### Requirements

- **Google Chrome or Chromium** must be installed for PDF rendering.

## Usage

### Initialize a carousel

```bash
mkcr init my-post
mkcr init my-post --format vertical
mkcr init my-post --width 1200 --height 1200
```

Creates a `./my-post/` folder with:
- `carousel.json` — config file with name, width, height
- `assets/` — directory for images, fonts, and other static files
- `style.css` — custom stylesheet loaded in every slide (supports Tailwind `@apply`)

Preset formats:
- `square` (default): 1080×1080 px (1:1)
- `vertical`: 1080×1350 px (4:5)

Use `--width` and `--height` for custom dimensions (overrides `--format`).

### Create slides

Slides are HTML fragment files named `1.html`, `2.html`, `3.html`, etc. inside the carousel directory. Each file contains **only the inner HTML content** — no `<!DOCTYPE>`, `<html>`, `<head>`, or `<body>` tags. The tool wraps them automatically with Tailwind CSS CDN and the correct dimensions.

```html
<!-- my-post/1.html -->
<div class="w-full h-full bg-blue-900 flex items-center justify-center p-16">
  <h1 class="text-8xl font-bold text-white text-center">Title Slide</h1>
</div>
```

Slide numbering must be sequential integers starting from 1. Non-numeric `.html` files are ignored.

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

Starts a local server with live reload — slides update automatically when files change. Use arrow keys to navigate between slides.

### Install binary

```bash
mkcr install
mkcr install --dir /usr/local/bin
```

Copies the `mkcr` binary to a directory in your PATH.

### Print LLM documentation

```bash
mkcr llm
```

Prints detailed documentation designed for AI agents to use this tool.

## Assets

Each carousel has an `assets/` folder for images, fonts, or any other static files. Reference them in slide HTML using absolute paths:

```html
<img src="/assets/logo.png" />
<div style="background-image: url('/assets/photo.jpg')">...</div>
```

Subdirectories are supported: `<img src="/assets/icons/arrow.svg" />`.

Assets are served automatically during both preview and PDF rendering.

## Custom styles (style.css)

Each carousel has a `style.css` file at its root, loaded in every slide. It supports standard CSS and Tailwind `@apply` directives (via CDN play mode):

```css
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;700;900&display=swap');

body {
  font-family: 'Inter', sans-serif;
}

.card {
  @apply rounded-2xl shadow-lg p-8;
}
```

## Slide dimensions

Default: **1080×1080px** (square, 1:1 ratio).

Configurable via `--format` flag on `init`, or directly in `carousel.json`:

```json
{
  "name": "my-post",
  "width": 1080,
  "height": 1080
}
```

## Examples

See the [`examples/`](./examples/) directory for complete carousel projects you can use as a reference.

- **[mkcr-presentation](./examples/mkcr-presentation/)** — A presentation about mkcr itself, built with mkcr. Demonstrates custom fonts, glow effects, dot grids, and multi-slide storytelling. ([PDF output](./examples/mkcr-presentation/mkcr-presentation-en.pdf))

## How it works

1. Slides are stored as numbered HTML fragment files (`1.html`, `2.html`, ...) in the carousel folder
2. At preview/render time, each fragment is wrapped with Tailwind CDN, `style.css`, and the correct viewport dimensions
3. `render` uses headless Chrome (via chromedp) to screenshot each slide at 2x resolution
4. Screenshots are assembled into a single multi-page PDF
