# 422 — transparency is NOT a promptable property, so the model paints the PICTURE of transparency: a grey-and-white checkerboard, served as the logo

**Filed 2026-09-02 by the `bugfix_417_420` lane. Artefact evidence and the diagnosis are the
`site_delivery_and_editor` lane's** (they downloaded, looked, and ran the chunk scan); **the code
evidence that it is a capability gap rather than a prompt defect is this lane's.** Diagnosis loop
NOT run — substituted equivalent first-hand verification per the 2026-07-31 ruling: every claim
below is a read of the served file, a live row, or the code, each quoted.

**Filed as its own file rather than folded into 417 or 421 — deliberately. Third mechanism, third
layer, third fix site:**
- `bugs_open/417` — the prompt **PERMITTED** something it should not (input licence). Fixed.
- `bugs_open/421` — the returned file's **STRUCTURE** was never examined (output acceptance). Open.
- **this** — the pipeline **CANNOT PRODUCE** the property being asked for (capability gap). No
  prompt and no config change can close it.

## What happened

OWNER RULING 2026-09-02: *"the background behind a logo shouldn't be part of the logo"* → the
regeneration prompt asked for *"a fully TRANSPARENT background (PNG alpha), no ground colour, no
panel, no backdrop of any kind"*.

The model returned an image of a **grey-and-white checkerboard** behind the mark — the
**UI REPRESENTATION** of transparency, painted as pixels.

Served asset, verified by the delivery lane: sha256 `1abcf69c08ab4462`, 243,080 B, **PNG colour
type 2 (RGB)**, and a chunk scan showing **no `tRNS` chunk**. There is no alpha mechanism in the
file at all, so the checkerboard is paint, not a viewer artefact. (Their note on how nearly this
was misread is worth keeping: a viewer draws a checkerboard *for* real alpha, so the visual is
ambiguous by construction — **only the chunk scan settles it**.)

## Why it is a capability gap, not a prompt defect `[MEASURED 2026-09-02, this lane]`

Alpha is a **file-format capability**, negotiated with the provider or added by post-processing.
It is not a property a model can be asked for in prose. Confirmed across the codebase:

- `internal/adapters/imagegenerator/banana/provider.go` — **no** `output_format`, no alpha, no RGBA
  and no mime negotiation anywhere; `MimeType` is only ever read back from what arrived.
- **No background-removal or post-processing step exists** anywhere in `platform/` or `internal/`
  (`background_remov|remove_background|rembg|alpha_channel|make_transparent` → zero hits).
- `kindDefaults["logo"]` carries only a negative prompt (*"…complex background"*) — **there is no
  ground/background knob** to set.

**So the owner's ruling cannot be satisfied by any prompt revision, and further regeneration
attempts asking for transparency will keep returning pictures of transparency.** This needs
imagery-pipeline work: request an alpha-capable output format from the provider, or add a
background-removal step after generation.

## The interim problem, and why there is no clean revert

The site now serves the checkerboard logo on every page — **visibly worse** than the 08-31 mark it
replaced, which had a baked dark ground and was invisible-safe against the dark header.

`[MEASURED 2026-09-02]` **There is no revert seam for a generated asset:**
- the `assets` table keeps **no version history** — the store path UPSERTs, so a regeneration
  overwrites `origin_prompt`/`url` in place (see the LANDMINES entry: `created_at` still reads the
  original generation);
- `retract_asset_files` **deletes** orphan files; it does not restore a prior version;
- nothing under `site-snapshots-and-revert` covers assets.

So the delivery lane holds the 08-31 bytes as a banked baseline with **no sanctioned path to put
them back**, and was right to refuse a hand-upload (owner ruling 2026-08-04: everything goes
through the framework).

**The framework-legal interim is a regeneration naming an explicit SOLID GROUND COLOUR matching the
header** — which *is* promptable, and is proven achievable because the 08-31 generation produced
exactly that. It is not a third blind attempt: we now know what to ask for and what not to.

## Fix candidates (ordered by what closes the door)

1. **Alpha at the provider** — request an alpha-capable format and carry it through the adapter,
   the store path (`mime_type`) and the deploy. Closes the class: "transparent" becomes a thing the
   pipeline *does*, not a thing it *asks for*.
2. **A post-generation background-removal step** for `kind=logo`. Independent of provider support;
   costs a processing step and a new failure mode (a mark clipped by an over-eager matte).
3. **Interim only — an explicit ground-colour clause** for logos, sourced from the site's header
   colour rather than left to the model. Does not satisfy the ruling; makes the artefact usable
   while 1 or 2 is built. ⚠ Note `composedPaletteDirection` deliberately **excludes logos** from
   palette direction (the 2026-05-20 contamination lesson), so this must be a narrow ground-colour
   clause, not a re-opening of that exclusion.
4. **A revert seam for generated assets** — the absence is a separate defect this incident exposed.
   Worth filing on its own if 1–3 are taken elsewhere.

## Verify
- The served file, not the row: `curl -s <logo url> | sha256sum`, then a PNG chunk scan for
  `tRNS` / colour type. **A visual check alone cannot distinguish painted checkerboard from real
  alpha** — that is this bug's whole trap.
- After any fix: colour type 6 (RGBA) or a `tRNS` chunk present, AND the mark still legible at
  favicon size.

## Related
- `bugs_open/417` (logo text policy — text-free HELD on this same asset, verified by eye).
- `bugs_open/421` (multi-panel design comp — single-composition HELD on this same asset).
- `bugs_open/235`, `bugs_open/322` (brand-asset handling).
- LANDMINES: the asset-upsert `created_at` trap, which bit this verification and was routed around.
