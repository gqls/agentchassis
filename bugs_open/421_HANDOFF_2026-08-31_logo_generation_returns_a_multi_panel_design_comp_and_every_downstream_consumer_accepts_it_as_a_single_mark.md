# 421 — logo generation returns a multi-panel DESIGN COMP, and every downstream consumer accepts it as a single mark

**Filed 2026-08-31 by the bugfix 417/420 lane, from a finding by the boxingonline.com session.**
Diagnosis loop NOT run — substituted equivalent first-hand verification per the 2026-07-31 owner
ruling: the asset was downloaded and **looked at** by the finding session, and every other link
below is a read of a live row or the served URL. The chain has no inferential hop.

**Split out of `bugs_open/417` deliberately.** 417 is an INPUT-LICENCE defect (the prompt
permitted lettering without naming it, so the model invented a brand). This is an
OUTPUT-ACCEPTANCE defect, and **417's fix cannot catch it**: a model can disobey "one
composition" while obeying "no text". Different mechanism, different fix layer (store time, not
prompt time), different owning family (imagery/designer, beside 235 and 322).

## What is being served, right now

`https://boxingonline.ugg2.com/assets/images/logo.png` — a **400×218 PNG** which is not a logo.
It is a **two-panel presentation board**: left half the mark on a dark navy ground, right half
the same mark on a light grey ground with lettering beside it. Two artboards side by side, the
way a designer presents options — scaled whole into the site's header slot on all 19 pages, and
cropped into the favicon and og_card by `derive_brand_head_assets`.

`assets` row `20ce80fb-53a9-490c-95a7-97a5a6a33097` (`asset_key='logo'`, `purpose='logo'`,
`status='active'`, created 2026-08-31 12:56:10Z), site boxingonline.com — **the first paid
customer site**.

**Even with 417's fix applied, this asset is unusable.** That is the point of filing separately:
neutralising the invented wordmark leaves a contact sheet in the header.

## The class

**Nothing between the provider's response and the served page ever checks that a brand asset is
ONE composition.** The store path takes whatever bytes arrive (subject only to the lock guard),
the deployer publishes them, the head-asset derivation crops them, and the header scales the
whole board into a logo slot. Every step is individually correct and none of them is looking.

This is the estate's `trust the rendered artefact, not the status` shape (bugs 012, 028) at the
**pixel layer**: the generation reported success, the row says `active`, the page returns 200,
and the only thing that noticed was a human opening the image. `kindDefaults` asks for 1024×1024;
what came back is ~2:1. **Nobody compared them.**

## Fix candidates (ordered by what closes the door)

1. **A dimensional envelope at STORE time.** A `kind=logo` asset whose aspect ratio falls far
   outside a sane envelope is refused or flagged before it is ever `active`. Deterministic, no
   vision model, no classifier, and it catches the whole two-panel family (a comp is
   ~2:1 where a mark is ~1:1). This is the cheap first cut and should ship alone.
   ⚠ Design it to FLAG-and-review rather than hard-refuse, per bugs_open/210's
   unhandleable-item lesson — a refused logo with no fallback is a queue that never drains.
2. **A prompt-side belt only.** 417's new `LogoTextFreeClause` already says "one composition on
   one plain background". That is an instruction, not a control, and the model has already
   demonstrated it will disobey an instruction in the same prompt that permits the opposite.
   **Do not count this as a fix** — it is cross-referenced here so the two files do not each
   assume the other covered it.
3. **The two-background-fields signature** (detecting the comp by its distinct grounds). This is
   a CLASSIFIER, and inherits classifier gaps — a stylised or low-contrast board returns a clean
   pass, and a clean pass from a blind check outlives the blindness in every document that later
   cites it. Cost it as its own step; do not smuggle it into the envelope check.

## Verify
- The artefact: `curl -s https://boxingonline.ugg2.com/assets/images/logo.png | file -` → dimensions.
- The row: `SELECT asset_key, purpose, status, created_at FROM assets WHERE id='20ce80fb-53a9-490c-95a7-97a5a6a33097';`
- The expectation it violated: `kindDefaults` in `generate_image_actions.go` (logo → 1024×1024).
- **Any fix must be verified by LOOKING at the image**, not by a row or a status. That is how this
  was found and it is the only check that has ever caught it.

## Related
- `bugs_open/417` (the invented wordmark in this same asset — input licence, fixed separately).
- `bugs_open/235` (logo stored as hero), `bugs_open/322` (brand-head block page-blind) — the same
  family: brand assets accepted without examination.
- `bugs_closed/012`, `bugs_closed/028` — the trust-the-status shape this instantiates.
- Owed on this site regardless: the logo needs regenerating before handover (delivery lane owns
  the dispatch), and favicon + og_card must be re-derived AFTER it lands.
