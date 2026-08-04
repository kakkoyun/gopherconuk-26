# Blueprint

A Marp theme for GopherCon UK talks, with a quiet Datadog signature.

```
gophercon-datadog.css     the theme
deck.md                   sample deck exercising every layout
preview.html              all slide types rendered, open it in a browser
fonts/                    Space Grotesk + JetBrains Mono, self-hosted (247 KB)
assets/gophers/           six CC0 gophers
assets/placeholder-*.png  stand-ins for your meme and your QR code
```

## Use it

VS Code — add to `.vscode/settings.json`:

```json
{ "markdown.marp.themes": ["./gophercon-datadog.css"] }
```

CLI:

```sh
marp --theme ./gophercon-datadog.css --html deck.md -o deck.html
marp --theme ./gophercon-datadog.css --html --pdf deck.md -o deck.pdf
```

`--html` is only needed for the three layouts that use a `<div>` (chart, columns, QR).
Everything else is plain Markdown.

Front matter:

```yaml
---
marp: true
theme: gophercon-datadog
paginate: true
footer: '@handle · #gopherconuk'
html: true
---
```

Marp puts theme CSS into exported HTML but retains these relative URLs. If you
write HTML outside this directory, put `fonts/` and `assets/` beside it. This
repository's Makefile stages them before each build; those directories are ignored.

## The colour rule

**Go blue is the frame. Datadog purple is the measurement.**

Blue draws structure — the tick strip, the progress bar, section blocks, code
keywords, list markers. Purple only ever marks a reading: a chart series you're
pointing at, an improved number, a delta chip, a shell prompt, the live-demo dot.
It never becomes a background. On dark slides it lifts from `#632CA6` to `#A78BE8`
so it survives the projector.

If you want more or less of it, change `--purple` in `:root` — nothing else needs
to move.

## Slide classes

Set with `<!-- _class: … -->` above the slide's content.

| Class | What it does | Markdown it expects |
| --- | --- | --- |
| *(none)* | Content slide. Heading, body, code, tables. | anything |
| `title` | Full-bleed blue opener with a gopher. | `#####` kicker, `#` title, `###` name, `p` role |
| `section` | Dark divider, blue number block. | `######` number, `#` title, `p` subtitle |
| `stat` | One giant purple number. | `#` the number, `p` the sentence |
| `compare` | Two panels, before / after. | `######` delta chip, then **two** blockquotes |
| `terminal` | Purple demo marker, wide shell panel. | `#` title, ` ```console ` block |
| `chart` | Bar chart. | `#`, `#####` unit, `<div class="bars">` |
| `meme` | Full-bleed image, caption strip. | `![bg contain](x.png)` then one paragraph |
| `emoji` | Drops the square bullets so your emoji lead. | `- 🔥 …` |
| `end` | Blue closer with links, QR and a gopher. | `#`, `<div class="qr">`, a list |
| `agenda` | Recurring contents slide; the bold row is where you are. | `#`, an `1.` list, `**bold**` the current item |
| `speaker` | Round headshot, name, role, handles. | image, `#`, `###`, a list |
| `photo` | Full-bleed image with a legibility scrim. | `![bg …](x.jpg)`, `#####`, `#`, `p` |
| `punchline` | One line, no chrome, no pagination. | `#` with `*emphasis*` for the purple word |
| `dark` | Flips any slide to the night palette. | combine with anything |

Gopher modifiers, usable on `title` and `end`:
`gopher-hero`, `gopher-rocket`, `gopher-network`, `gopher-sage`, `gopher-hiking`,
`gopher-balloon`, `gopher-none`.

```md
<!-- _class: title gopher-sage -->
<!-- _class: end gopher-balloon -->
<!-- _class: dark -->
```

## Two conventions worth knowing

**`######` is the utility slot.** A level-6 heading means "small structural thing
here", and the theme decides what based on context:

- immediately before a code fence → the filename tab on the editor panel
- on a `section` slide → the big blue number
- on a `compare` slide → the delta chip between the panels

**Emoji get a hanging indent.** List items use `text-indent: -1.9em` with matching
padding, so a leading emoji sits in the margin and wrapped lines align under the
text, not under the emoji. Write `- 🔥 Flamegraphs, read wrong` and it lines up.

## Keeping a room awake

**Reveal a list one item at a time.** Use `*` instead of `-` and Marp's bespoke
output fragments it. Unrevealed items ghost to 18% rather than disappearing, so
the audience can see how much is coming:

```md
* 🔥 First you notice the wide bar
* 🧊 Then you notice it's been there for a year
```

Static HTML and PDF show every item, so your handout still reads.

**Make any photo look like it belongs.** Marpit's own image filters do a duotone
with no extra CSS. Blue:

```md
![bg grayscale:1 brightness:.45 sepia:1 hue-rotate:158deg saturate:3.2](venue.jpg)
```

Purple — swap the hue: `hue-rotate:232deg saturate:2.4`. Pair with
`<!-- _class: photo -->` and the theme adds a bottom-up scrim so your text stays
readable over anything.

**Transitions, once per section.** Marp CLI supports them in bespoke HTML:

```md
<!-- _transition: cover -->
```

Put it on section dividers only and leave the rest instant. One deliberate move
per section reads as intentional; movement everywhere reads as noise.

**Speaker notes** are HTML comments that aren't directives:

```md
<!-- Two minutes here. Do not open the pprof UI until the next slide. -->
```

## Snippets for the three HTML layouts

Bar chart — add `mark` to the row you're pointing at and it turns purple:

```html
<div class="bars">
  <div class="row"><span class="label">/checkout</span><span class="track"><span class="fill" style="width:92%"></span></span><span class="value">412</span></div>
  <div class="row mark"><span class="label">/checkout*</span><span class="track"><span class="fill" style="width:20%"></span></span><span class="value">88</span></div>
</div>
```

Two columns — `.cols`, optionally `.wide-left` / `.wide-right`. Leave blank lines
inside the divs so Markdown still parses:

```html
<div class="cols">
<div class="card">

### Read the flame

Wide bars are cost, not depth.

</div>
<div class="card measure">

### Then measure

Every frame is a decision someone made.

</div>
</div>
```

QR on the closing slide:

```html
<div class="qr"><img src="assets/my-qr.png" alt="slides"></div>
```

## Things to know before the talk

**The progress bar needs a recent Chromium.** It reads Marp's pagination
attributes and does the arithmetic in CSS, which needs Chromium 133 or newer
(early 2025). On anything older the bar renders full-width instead — deliberate
looking, just not informative. Everything else is plain CSS.

**Type floors.** Body text is 28px and no layout goes below 21px, which reads from
the back of a room the size of The Brewery. Code is 23px, terminal 25px. If you
find yourself shrinking code to fit, cut the code instead.

**PDF export needs `print-color-adjust`,** which the theme sets. If a backgrounds
come out white, it's the browser's "background graphics" print setting, not the
theme.

**Fonts are bundled, so nothing phones home.** Latin and Latin-Extended subsets
only. If you need Cyrillic or Greek, re-download those subsets from Google Fonts
into `fonts/` and add the `@font-face` blocks.

## Licences

| Thing | Licence |
| --- | --- |
| Space Grotesk, JetBrains Mono | SIL Open Font License 1.1 |
| Gophers in `assets/gophers/` | CC0 — [egonelbre/gophers](https://github.com/egonelbre/gophers) |
| Gopher character design | Renée French, the original — these are CC0 remixes of it |
| Datadog mark | Datadog trademark, used by a Datadog employee |
| This CSS | Do what you like with it |

If you swap in different gopher art, check the licence. Maria Letta's
[free-gophers-pack](https://github.com/MariaLetta/free-gophers-pack) is also CC0
and looks great as one hero image, but the files are heavy — better as a single
`![bg]` than as theme furniture. Renée French's original gophers are CC BY 3.0 and
need attribution on the slide.
