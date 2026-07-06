# Slice 4b — 003_contracts_and_standards: the producer-side rewrite

Two edits to the repo doc. The consumer-side narrative (the three-layer defence, the
base-CSS chains at ~477–512) is correct and stays; only the producer rule keyed on the
flag and the literal example change.

## Edit 1 — list item 6 (line ~309)
OLD:
```
6. **Dark section contract** — if `is_dark_section = true`, template MUST set `--section-*` CSS variables on container
```
NEW:
```
6. **Section painting contract** — a template's appearance derives from what its own CSS paints; `is_dark_section` is catalogue metadata and MUST NOT key styling. A painting section chooses exactly one model and re-exports `--section-*` on its container AS REFERENCES ONLY: (a) a pair band (`background: var(--color-cta-bg)` or the header/footer pair) re-exporting the pair text (`--section-text: var(--color-cta-text);` muted/border/surface via `color-mix`); (b) a palette band (`background: var(--color-primary)`) re-exporting the on-colour family (`--section-text: var(--color-primary-text, var(--color-background));`); (c) an image/layered background defining `--hero-ink` per branch and re-exporting from the ink (the hero is the model); or (d) ambient — no background of its own and NO `--section-*` declarations at all. Literal colours in `--section-*` declarations are forbidden; `fix_forced_text_colors` enforces this mechanically.
6b. **Image fields** — any `site_assets.*`-sourced field MUST be `required: false` with `"on_missing": "skip_field"`, and its markup gated with a template conditional (brief-explanation's image wrapper is the model). Imagery arrives asynchronously and must never block or defer a section.
```

## Edit 2 — the literal example (lines ~490–496)
OLD:
```
Dark section component sets --section-* on container:

    --section-heading: #ffffff;
    --section-text: rgba(255,255,255,0.9);

  h2 gets var(--section-heading) → #ffffff
  p gets var(--section-text) → rgba(255,255,255,0.9)
```
NEW:
```
A pair-band component re-exports --section-* on its container as references:

    --section-heading: var(--color-cta-text);
    --section-text: var(--color-cta-text);

  h2 gets var(--section-heading) → the pair's text colour (per scheme)
  p gets var(--section-text) → the pair's text colour (per scheme)

The values flip with the scheme because the pair flips; no literal ever appears.
```
(Adjust the OLD blocks to the doc's exact bytes if line wrapping differs — needle
discipline applies to docs too.)
