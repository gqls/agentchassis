# 211 — the renderer's LAST appended CSS block is absent from one site's served stylesheet, and it takes the headings out

**Filed** 2026-08-06 by the `bugfix_122_contrast_ink_slots` lane.
**Site:** ai-agent-orchestration.com — **30 firm contrast failures, the worst on the
fleet**, and until this file it appeared in **no** bug file, including `bugs_open/122`.

> **This file asserts a MECHANISM as measured and leaves the CAUSE open.** Read §4
> before quoting any of it as a root cause. Two `090` diagnosis runs were spent on
> this symptom and neither returned one; §5 says why the first was doomed.

---

## 1. The symptom, browser-measured

`python3 scripts/render_audit.py https://ai-agent-orchestration.com/`, 2026-08-06:

```
FAIL https://ai-agent-orchestration.com/     contrast=30
   1.00:1 need 4.5  rgb(13, 17, 23) on rgb(13,17,23)  .H3  'Deployed on Kubernetes, Kafka, and Postgres'
   1.00:1 ... x6 .H3
   1.04:1 need 3.0  rgb(13, 17, 23) on rgb(8,11,16)   .H2
   1.10:1 ...                                         .section-heading
```

Six `.H3` headings painted in **exactly their own ground colour**. Not "low
contrast" — literally invisible.

## 2. What the served stylesheet actually says [MEASURED 2026-08-06]

Fetched with a browser UA (a bare `curl` gets 403 on every site — that is a
user-agent rejection, not a routing fault):

```bash
curl -fsS -A "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 Chrome/126 Safari/537.36" \
  https://ai-agent-orchestration.com/assets/css/styles.css -o css_aiao.css
```

| fact | value |
|---|---|
| `--color-primary` | `#0D1117` = **rgb(13,17,23)** — the measured heading colour |
| `--color-surface` | `#0D1117` — **byte-identical to primary** |
| `--color-background` | `#080B10` = rgb(8,11,16) — the measured `.H2` ground |
| `--color-text` | `#E6EDF3` |
| heading rule | `h1,h2,h3,h4,h5,h6 { color: var(--section-heading, var(--color-primary)); }` |
| `--color-heading` definitions in the file | **0** |
| compatibility-alias block present | **NO** |
| `--hero-ink` defined | **NO** |

**The file ends at the closing brace of the renderer's section-defaults block.**
Step 11's output is simply not there.

### The comparison that makes it a defect rather than a quirk

Four other sites sampled the same way, same day:

| site | step-11 alias block | step-10 section defaults |
|---|---|---|
| gamesdesign.co.uk | **present** | present |
| gaswholesalers.com | **present** | absent |
| finetuning.uk | **present** | absent |
| idea.uk | **present** | absent |
| **ai-agent-orchestration.com** | **ABSENT** | present |

It is the only one of five missing it, and the only one where step 10 ran and
step 11 did not.

## 3. The causal chain, for the hero path [MEASURED]

`buildTokenAliases` maps `--hero-ink` → `var(--color-text)`. With the alias block
absent, `--hero-ink` is **undefined**. The `hero` component's stored markup sets:

```css
.hero-section { --section-heading: var(--hero-ink); }
```

A custom property set to an undefined `var()` is **guaranteed-invalid at
computed-value time**, so `var(--section-heading, var(--color-primary))` falls
through to its fallback — `--color-primary` — which on this site is
`#0D1117`, identical to `--color-surface`. Heading painted in its own ground.

Verified by query, ai-agent-orchestration.com `/index.html`:

```
hero              -> --section-heading: var(--hero-ink);
system-stats      -> --section-heading: #ffffff;
case-studies-grid -> --section-heading: #ffffff;
call-to-action    -> --section-heading: var(--color-cta-text, var(--color-primary-text));
```

## 4. What is NOT established — read this before quoting the file

- **WHY the alias block is missing. [UNMEASURED]** Step 11 runs unconditionally
  after step 10 in `RenderCSSFromSpecAction`, and step 10's output *is* in the
  file, so "the renderer never ran" does not fit. Candidates, none tested:
  truncation at persist or deploy; `deploy_css` shipping a stale artefact; a
  post-processor; a second write path. **Do not pick one from this list without
  measuring it.**
- **Ruled out:** staleness relative to the feature. `buildTokenAliases` landed
  `568205c31`, **2026-07-06**; the site's pages last deployed **2026-08-06 02:07**.
  A month apart, so the stylesheet does not predate the mechanism.
- **The six `.H3`s specifically. [UNRESOLVED]** They are in
  `differentiators-section`, which does **not** set `--section-heading`
  (measured). It is in the renderer's own section-defaults selector list, which
  sets `--section-heading: #8B949E` at `body` scope — so by the cascade those
  H3s should be `#8B949E`, and they measure `#0D1117`. **The hero chain in §3 is
  demonstrated; it is NOT demonstrated that it is what colours these six.**
  Whoever picks this up should start here, in a browser with devtools, not in
  the stylesheet.
- **A palette whose `primary` and `surface` are byte-identical is its own
  latent fault** and would make any fallback-to-primary invisible. Worth its own
  check fleet-wide; not done.

## 5. Diagnosis history — two runs, no verdict, and the first was unanswerable

| run | correlation | outcome |
|---|---|---|
| 1 | `5853ee07-a49c-4571-8ea0-3eb660e43dfd` | **UNVERIFIABLE** — "Diagnosis NOT confirmed (stopped: iteration-cap)". 5 bundles, no cause. |
| 2 | `750e162e-2b3e-4f96-89e1-5486197942cd` | same shape — 5 bundles, work item `complete`, no verdict artefact |

**Run 1's symptom statement was built on a false premise and could not have
succeeded.** It asserted the site "does define `--color-heading: var(--color-text)`"
and sent the loop at `tokenAliases`' `--color-heading` entry and
`darkSchemeDerivations`' heading-from-text rule. The stylesheet defines
`--color-heading` **zero** times and headings never consult it — the path is
`--section-heading`. The loop spent five iterations reading two mechanisms that
do not participate.

> **The transferable bit:** a `090` symptom that names the wrong mechanism does not
> come back REFUTED, it comes back **UNVERIFIABLE** — and an iteration-cap stop
> reads like "hard bug" when it may mean "wrong question". Run 2 was refiled with
> the corrected symptom (the absent block) and also capped, so this is a genuine
> gap in what the loop can reach, not only an authoring fault.

## 6. Fix candidates, ordered by what makes the bad state unrepresentable

1. **Find and close the loss of the last appended block.** Everything else is a
   symptom patch. Start by re-rendering this site's CSS and comparing the action's
   returned `result` length against what lands in the DB and then the bucket — the
   `css_length` / `token_alias_length` fields are already in the step-12 summary
   log as of `1d2c93a87`.
2. **Make the heading fallback safe.** `var(--section-heading, var(--color-primary))`
   picks the one palette slot most likely to equal a ground (`warnUnusablePrimary`
   exists because 4 of 31 palettes score primary below 1.25:1 on their own
   background). `var(--color-primary-ink, var(--color-primary))` — shipped
   `1d2c93a87`, register VIZ-014 — is the legible-by-construction version.
   **Inert until an image roll.**
3. **Refuse a palette where `primary` and `surface` are byte-identical**, or at
   least warn. Cheap, and it removes the coincidence this failure depends on.
4. Re-point this one site's palette. Rejected as a class fix — it is what produced
   the state.

## 7. How to verify a fix

Not on the stylesheet and not on a palette row (both gave wrong answers in
`bugs_open/122`'s own history). `python3 scripts/render_audit.py
https://ai-agent-orchestration.com/` against the banked pre-fix baseline at
`docs/agent_docs/docs024_key_docs_latest/bugfix_122_contrast_ink_slots/BASELINE_2026-08-06_render_audit.txt`
(15 sites, 109 failures). **Per selector, not by count** — a falling count is
content-dependent.

## 8. Relations

`bugs_open/122` (the parent contrast bug; this is its sub-shape B, which 122 does
not name), `bugs_open/212` (components overriding renderer-owned `--section-*`
tokens — adjacent and possibly the same family), register VIZ-014, VIZ-012,
LANDMINE "`--color-<x>-text` and `--color-<x>-ink` are OPPOSITE questions".
