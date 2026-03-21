# fycr — Carousel Generator for AI Agents

You are using `fycr`, a CLI tool to generate LinkedIn carousel PDFs from HTML+Tailwind CSS slides.

## Commands

### Initialize a carousel
```
fycr init <name> [--format square|vertical] [--width N] [--height N]
```
Creates a `./<name>/` directory with a `carousel.json` config file.

Preset formats:
- `square` (default): 1080×1080 px (1:1)
- `vertical`: 1080×1350 px (4:5)

Use `--width` and `--height` to set custom dimensions (overrides `--format`).

### Add a slide
```
fycr add <name> --html '<div class="...">content</div>'
```
You only pass the **inner body content** — the tool wraps it in a full HTML document with Tailwind CSS CDN included. Slides are auto-numbered (1.html, 2.html, ...). Use `--position N` to overwrite a specific slide.

### Render to PDF
```
fycr render <name>
```
Generates a single multi-page PDF at `./<name>/<name>.pdf` and prints the absolute path. Use `--output path` to change the destination.

### Preview slides
```
fycr preview <name>
```
Opens all slides in the default browser.

## Local assets

Each carousel has an `assets/` folder (created automatically by `fycr init`) where you can place images, fonts, or any other files. Reference them in slide HTML using absolute paths:

```html
<img src="/assets/logo.png" />
<div style="background-image: url('/assets/photo.jpg')">...</div>
```

Subdirectories are supported: `<img src="/assets/icons/arrow.svg" />`.

Assets are served automatically during both preview and PDF rendering.

## Slide authoring guidelines

- Each slide is a single `<div>` using Tailwind CSS classes.
- The slide container matches the carousel dimensions (default 1080x1080px). Use `w-full h-full` on your root div to fill it.
- Use large text sizes (`text-5xl` to `text-8xl`) — these are high-resolution slides.
- Backgrounds: use Tailwind gradients (`bg-gradient-to-br from-blue-600 to-purple-700`) or solid colors.
- Layout: use flexbox (`flex items-center justify-center`) for centering content.
- Padding: use generous padding (`p-12` to `p-20`) to keep content away from edges.
- Keep text concise — carousels work best with short, punchy messages per slide.

## Typical workflow

```
fycr init my-post
fycr add my-post --html '<div class="w-full h-full bg-blue-900 flex items-center justify-center p-16"><h1 class="text-8xl font-bold text-white text-center">Title Slide</h1></div>'
fycr add my-post --html '<div class="w-full h-full bg-white flex flex-col justify-center p-20"><h2 class="text-5xl font-bold text-gray-900 mb-8">Key Point</h2><p class="text-3xl text-gray-600">Supporting detail goes here.</p></div>'
fycr add my-post --html '<div class="w-full h-full bg-blue-900 flex items-center justify-center p-16"><p class="text-5xl font-bold text-white text-center">Follow me for more!</p></div>'
fycr render my-post
```
