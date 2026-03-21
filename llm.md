# fycr — Carousel Generator for AI Agents

You are using `fycr`, a CLI tool to generate LinkedIn carousel PDFs from HTML+Tailwind CSS slides.

## Commands

### Initialize a carousel
```
fycr init <name> [--format square|vertical] [--width N] [--height N]
```
Creates a `./<name>/` directory with:
- `carousel.json` — config file with name, width, height
- `assets/` — directory for images, fonts, etc.
- `style.css` — global stylesheet (supports Tailwind `@apply`)

Preset formats:
- `square` (default): 1080×1080 px (1:1)
- `vertical`: 1080×1350 px (4:5)

Use `--width` and `--height` to set custom dimensions (overrides `--format`).

### Render to PDF
```
fycr render <name>
```
Generates a single multi-page PDF at `./<name>/<name>.pdf` and prints the absolute path. Use `--output path` to change the destination.

### Preview slides
```
fycr preview <name>
```
Opens all slides in the default browser with live reload.

## Creating slides

Slides are simple HTML files named `1.html`, `2.html`, `3.html`, etc. inside the carousel directory. Each file contains **only the inner HTML content** — no `<!DOCTYPE>`, `<html>`, `<head>`, or `<body>` tags. The tool wraps them automatically at preview/render time with Tailwind CSS CDN and the correct dimensions.

**Write slide files directly** using the Write tool. Do not use shell commands to create files.

Example slide file (`1.html`):
```html
<div class="w-full h-full bg-blue-900 flex items-center justify-center p-16">
  <h1 class="text-8xl font-bold text-white text-center">Title Slide</h1>
</div>
```

Slide numbering must be sequential integers starting from 1. Non-numeric `.html` files are ignored.

## Local assets

Each carousel has an `assets/` folder where you can place images, fonts, or any other files. Reference them in slide HTML using absolute paths:

```html
<img src="/assets/logo.png" />
<div style="background-image: url('/assets/photo.jpg')">...</div>
```

Subdirectories are supported: `<img src="/assets/icons/arrow.svg" />`.

Assets are served automatically during both preview and PDF rendering.

## Custom styles (style.css)

Each carousel has a `style.css` file at its root. This file is loaded in every slide and supports:
- Standard CSS rules
- Tailwind `@apply` directives (via CDN play mode)

Example:
```css
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;700;900&display=swap');

body {
  font-family: 'Inter', sans-serif;
}

.card {
  @apply rounded-2xl shadow-lg p-8;
}
```

## Slide authoring guidelines

- Each slide is a single `<div>` using Tailwind CSS classes.
- The slide container matches the carousel dimensions (default 1080x1080px). Use `w-full h-full` on your root div to fill it.
- Use large text sizes (`text-5xl` to `text-8xl`) — these are high-resolution slides.
- Backgrounds: use Tailwind gradients (`bg-gradient-to-br from-blue-600 to-purple-700`) or solid colors.
- Layout: use flexbox (`flex items-center justify-center`) for centering content.
- Padding: use generous padding (`p-12` to `p-20`) to keep content away from edges.
- Keep text concise — carousels work best with short, punchy messages per slide.

## Typical workflow

```bash
# 1. Initialize a carousel
fycr init my-post

# 2. Write slide files directly (use the Write tool)
# my-post/1.html:
# <div class="w-full h-full bg-blue-900 flex items-center justify-center p-16">
#   <h1 class="text-8xl font-bold text-white text-center">Title Slide</h1>
# </div>

# my-post/2.html:
# <div class="w-full h-full bg-white flex flex-col justify-center p-20">
#   <h2 class="text-5xl font-bold text-gray-900 mb-8">Key Point</h2>
#   <p class="text-3xl text-gray-600">Supporting detail goes here.</p>
# </div>

# my-post/3.html:
# <div class="w-full h-full bg-blue-900 flex items-center justify-center p-16">
#   <p class="text-5xl font-bold text-white text-center">Follow me for more!</p>
# </div>

# 3. Preview in browser (with live reload)
fycr preview my-post

# 4. Render to PDF
fycr render my-post
```
