# maskoverlay-tui

A terminal UI for **maskoverlay**: clip a base image to the silhouette of a mask
PNG, blending the mask's own colours over it as a semi-transparent overlay.
Everything outside the mask's opaque pixels becomes transparent, the output is
sized to the mask, and the base is scaled to cover it.

It's a [Bubble Tea](https://github.com/charmbracelet/bubbletea) front-end that
walks you through picking the files and options, then shells out to ImageMagick
to do the compositing. The pipeline is a mirror of the standalone `maskoverlay`
zsh function it was extracted from.

## Requirements

- **[ImageMagick](https://imagemagick.org)** — the `magick` command must be on
  your `PATH`. This is a runtime dependency; the tool invokes `magick identify`
  and `magick` to render.
- **Go 1.26+** to build from source.

## Install

```sh
go install github.com/chrisleechdev/maskoverlay-tui@latest
```

Or build locally:

```sh
git clone git@github.com:chrisleechdev/maskoverlay-tui.git
cd maskoverlay-tui
go build        # produces ./maskoverlay-tui
```

## Usage

Run it with no arguments — everything is chosen interactively:

```sh
maskoverlay-tui
```

The flow has four steps:

1. **Choose base** — the content shown inside the shape. A GIF stays animated; a
   PNG/JPG/WebP is static. (Allowed: `.gif`, `.png`, `.jpg`, `.jpeg`, `.webp`.)
2. **Choose mask** — a PNG whose alpha channel is the shape and whose colours are
   the overlay tint.
3. **Options** — adjust the overlay opacity, pick the output format, and confirm
   the output path (pre-filled as `<base-dir>/<mask-name>-masked.<fmt>`).
4. **Render** — a spinner shows while `magick` runs, then the result path (or the
   error) is displayed.

### Options

| Option    | Details                                                              |
|-----------|----------------------------------------------------------------------|
| Opacity   | Overlay opacity `0.00`–`1.00` (default `0.75`). `←/→` adjust by 0.05. Lower shows more of the base. |
| Format    | `gif`, `webp`, or `png`. Changing it rewrites the output extension.  |
| Output    | Editable path. `~/` is expanded. Typing an extension syncs the format selector. |

### Output formats

- **`gif`** — hard 1-bit edge (alpha thresholded at 50%, background disposal).
  Best for animated GIF bases.
- **`webp`** — smooth alpha (anti-aliased edges).
- **`png`** — smooth alpha, written as APNG so animated bases are preserved.

## Keybindings

**File pickers (steps 1–2)**

| Key         | Action        |
|-------------|---------------|
| `↑` / `↓`   | Move          |
| `enter`     | Select        |
| `backspace` | Up a directory|
| `esc`       | Back a step (quits on step 1)|
| `q`         | Quit          |

**Options (step 3)**

| Key                 | Action                                    |
|---------------------|-------------------------------------------|
| `tab` / `↓`         | Next field                                |
| `shift+tab` / `↑`   | Previous field                            |
| `←` / `→`           | Adjust opacity or format                  |
| `enter`             | Render                                    |
| `esc`               | Back to mask picker                       |
| `q`                 | Quit (unless editing Output)              |

`ctrl+c` quits from anywhere. `esc` steps back to the previous screen — mask → base,
options → mask, and the final screen → options (to tweak and re-render) — except on the
first screen and while rendering, where it quits instead. On the final screen, `enter`
or `q` quits.
