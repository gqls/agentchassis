# How to rebuild the pitch PDF

Source for `../PITCH_travelling_docs_self_verifying_tools.pdf` — a 12-page A4-landscape
colour deck describing the travelling-docs / self-verifying-tools mechanism, written for
an outside audience (a CV/pitch attachment or a LinkedIn link).

## Rebuild

```bash
chromium --headless --disable-gpu --no-pdf-header-footer \
  --virtual-time-budget=6000 \
  --print-to-pdf=travelling-docs.pdf deck.html
```

`--virtual-time-budget` matters: the diagrams are drawn by `sketch.js` at load time, so the
render must wait for JavaScript. Page size comes from `@page { size: 297mm 210mm }`.

## What is in here

| File | Notes |
|---|---|
| `deck.html` | The whole deck. One `<section class="slide">` per page. |
| `sketch.js` | Small deterministic hand-drawn SVG helper (seeded PRNG → wobbly rects, ellipses, arrows). Vector, so it stays crisp in print. |
| `fonts-local.css` + `fonts/` | Caveat (annotations), Inter (body), Kalam. Self-hosted so the render needs no network. |
| `shots/` | Real screenshots of the live pages, captured with headless Chromium. |
| `evidence_from_live_db.txt` | The raw `doc_plans` / `doc_notes` rows quoted in the deck, as pulled from production on 2026-07-16. |
| `REVERSE_ENGINEERED_STYLE_PROMPT_v1/v2/v3.md` | A reusable "de-AI-ify this copy" instruction prompt, reverse-engineered from a hand-edited rewrite of this deck's own text. Each file is immutable once written — corrections land as a new `_vN` file, never an edit in place, so a prior round's reasoning stays inspectable. **`_v3.md` is current.** |

## Rules the content follows

- **Nothing is illustrative.** Every screenshot is a live page; every record quoted on
  slides 3 and 8 is copied from the running database. If you refresh the deck, re-pull the
  evidence rather than editing the quotes.
- **No personal details** (owner's decision) — the deck carries the project and date only.
- Real site names (gamesdesign.co.uk, vonc.com) are shown deliberately: the evidence is
  verifiable because a reader can visit the pages.

## Re-capturing screenshots

```bash
chromium --headless --disable-gpu --window-size=1366,760 --hide-scrollbars \
  --virtual-time-budget=9000 \
  --screenshot=shots/economy-sim-desktop.png \
  "https://gamesdesign.co.uk/games/economy-simulator/index.html"
```

Window heights are chosen to crop each page at a sensible point (the tools sit on dark
backgrounds, so a 900px-tall shot leaves a dead band). The quiz is captured at 390×844 —
the mobile profile the browser tier itself uses.

## Checking the layout after an edit

Text overflowing the footer is the failure mode to watch. Rather than eyeballing it,
append a probe script to a copy of `deck.html` that walks every `.card`/`figcaption`/`pre`
and reports any element whose bottom passes `.foot`'s top, then read it back with
`chromium --headless --dump-dom`. A `<figure>` must keep `flex: none`, or a flex parent
will shrink it below its content height and silently clip the caption.
