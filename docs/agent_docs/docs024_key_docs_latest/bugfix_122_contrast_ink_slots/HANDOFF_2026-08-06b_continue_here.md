# HANDOFF 2026-08-06b — bug 122, contrast / ink slots. START HERE.

Supersedes `HANDOFF_2026-08-06_continue_here.md` (that file's "two things in flight"
are both resolved; its "what NOT to do" list still stands and is repeated below).

## The one-paragraph state

The **engine half is written, council-APPROVED, tested and committed** (`1d2c93a87`,
gofmt follow-up `5506d35c8`) — but it is **NOT LIVE**: it needs an image roll, and
v1.0.1261 predates it. The **config half — which is what actually changes a visitor's
page — is NOT WRITTEN**: two migrations, both fully specified in the approved
submission. Round 1 of the council was REVISEd and it was **right**: my sub-shape C
diagnosis was wrong, none of those rules hard-codes anything, and correcting it changed
the fix from two variables to three. Two new bugs came out of the checking
(`bugs_open/211`, `bugs_open/212`) and are filed, not folded in.

## Nothing is in flight. Both prior runs are closed.

| what | correlation | outcome |
|---|---|---|
| council gate | `c4d9c841-3658-4742-85b5-961e062ecad2` | **APPROVED** (round 2), 3 advisory objections, none high |
| 090 diagnosis, run 1 | `5853ee07-a49c-4571-8ea0-3eb660e43dfd` | UNVERIFIABLE (iteration-cap) — **my symptom named the wrong mechanism** |
| 090 diagnosis, run 2 | `750e162e-2b3e-4f96-89e1-5486197942cd` | UNVERIFIABLE, corrected symptom, also capped → `bugs_open/211` |

Use `Council-Reviewed: c4d9c841-3658-4742-85b5-961e062ecad2` on further commits in this
lane — the verdict is read and approved.

## What was shipped, in one paragraph you should read before touching the code

A palette colour used as a **fill** and the same colour used as an **ink** are different
questions, and the platform only answered the first:

- `--color-<x>-text` — the ink that goes **ON** an `<x>` fill. `accent_text` has been
  derived since 2026-07-27 and **0 of 18 layouts declare it**, so it never reached one
  stylesheet.
- `--color-<x>-ink` — `<x>` **itself, made legible as an ink**. New. `--color-primary-ink`,
  `--color-accent-ink`.

`buildLegibleInkDefaults` emits all three from the renderer's own `:root` block as
**step 12** of `RenderCSSFromSpecAction`, after the token aliases. Register **VIZ-014**.

## Next actions, in order

### 1. Roll an image and prove it at the pod — NOT at git, the tag, or a roll

`make build-agent-chassis` (builds from committed HEAD, so this is safe), bump
`IMAGE_TAG` in the makefile (currently `v1.0.1261`), push, deploy.

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
for s in buildLegibleInkDefaults legibleInkFor fillDarkSchemeSpecialisedSlots zzzInventedControlXyz; do
  printf "%-34s " "$s"; kubectl -n ai-persona-system exec $POD -- sh -c "strings /app/agent-chassis | grep -c '$s'"
done
# want: >0 / >0 / 4 (positive control) / 0 (proves the grep discriminates)
```

**A roll is not evidence your fix shipped** — the image may predate your commit and
carries no provenance. Check **every replica**.

### 2. Write and apply migration `324` — the consumers

**This is the half that changes what a visitor sees.** Nothing above it does.
Full sketch, needle-gate discipline and rationale: `SUBMISSION_2026-08-06b_ink_slots_round2.json`,
edit 7. Summary — every replacement is `var(<new>, <EXACTLY today's value>)`:

| target | from | to | closes |
|---|---|---|---|
| `case-studies-grid` `.csg-cta-btn`, `.csg-filter-btn.active` | `var(--color-primary-text, #fff)` | `var(--color-accent-text, var(--color-primary-text, #fff))` | finetuning 2 |
| `system-stats` `.stats-eyebrow` | `var(--color-accent, #7dd3fc)` | `var(--color-accent-ink, var(--color-accent, #7dd3fc))` | gamesdesign 1, vonc 1 |
| `image-hover-card-grid` `__eyebrow` | `var(--color-primary)` | `var(--color-primary-ink, var(--color-primary))` | dartsonline 1 |
| `tool-list` `.tl-eyebrow`, `.tl-card-link` | `var(--color-primary)` | `var(--color-primary-ink, var(--color-primary))` | robot-hands 2 |
| 5 of 18 `layouts` base `a {}` | `var(--color-accent)` | `var(--color-accent-ink, var(--color-accent))` | gaswholesalers 6 |

The 5 layouts: `brochure-formal`, `high-energy`, `technical-precise`,
`tool-portal-dark`, `tool-portal-light`.

**Non-negotiables the council attached to this edit:**
- **On the ledger.** `docs/agent_docs/sql_for_agents/NNN_*.sql`, applied by
  `scripts/migration/run-migrations.sh`, registered in `schema_migrations`. **324 was
  free on 2026-08-06 — re-check immediately before applying**, a number is not yours
  because you named a file.
- **Needle gate.** Mechanically derived pre-count, `position()` on the exact substring
  (not `LIKE` — no `%` to be tripped by), `\copy` backup first, inline rollback.
- **Every check `DO`/`RAISE`.** A verify block of bare `SELECT`s **cannot stop the
  `COMMIT`** — `ON_ERROR_STOP` ignores a non-empty result. Induce each RAISE once
  against a scratch copy to prove it fires.
- **Propagation is a separate step** (`editquality`, medium, still open on the approved
  verdict): the UPDATE changes the SOURCE. Live pages keep their old `rendered_html`
  until a scoped rerender. **Enqueue it** — the plan as approved only *listed* the
  affected rows, and the seat was right that this is not enough. 16 placements across
  ≤4 sites each.

### 3. Write and apply migration `325` — the render-audit cadence

**One row.** `\d scheduled_tasks` first — it is `interval_seconds`, not a cron string,
plus `target_topic`, `input_data`, `pre_query`, `fire_message`.

The whole detection chain is live in v1.0.1257 and **nothing dispatches it**: 46
`scheduled_tasks` rows, 29 enabled, **0 targeting `render-audit-agent` enabled or
disabled** (re-measured unfiltered on the council's objection). Total `contrast_failure`
items ever raised: 4, all relojistas.com, all 2026-08-04 — one hand-run.

Weekly, not daily: findings dedupe on `contrast_failure:<page-path>#<selector>` and a
falling count is content-dependent, so a daily cadence multiplies noise without signal.

### 4. Verify against the banked baseline — per selector, never by count

`BASELINE_2026-08-06_render_audit.txt` — **15 sites, 109 failures**, complete.

```bash
python3 scripts/render_audit.py <the 15 urls> > after_$(date +%F).txt
```

The strong check is that the **named selectors at the named ratios are gone and no new
selector has appeared**. A falling total is weak evidence: 122's own dartsonline round
found the same defect reporting 1 or 2 depending which cards a page rendered.

Expected close: **12 failures** (dartsonline 1, robot-hands 2, finetuning 2,
gaswholesalers 6, gamesdesign 1) plus `.stats-eyebrow` on vonc.

### 5. Advisory objections still open on the APPROVED verdict

Approved means you may proceed; these were not withdrawn.

- `editquality` **medium** — no rerender/propagation enqueue (see §2). **Act on this.**
- `editquality` low — the `warnUnusablePrimary` remedy edit closes no failure; don't
  count it as a mechanism addressed. (Shipped anyway; it is honest logging.)
- `guardian` low — `--color-accent-ink` is genuinely new surface beyond round 1; read it
  as a fresh item, not a detail of an approved plan.
- `guardian` low — the ledger discipline is an assertion the seat cannot verify. **You**
  verify it.
- `tooling_provenance` low — leave a `doc_notes` record for `render-audit-agent` when you
  schedule it live.
- `prior_art_librarian` low — `scripts/render_audit.py` already does multi-ground
  computed-style contrast in Python; it was not checked as prior art for
  `worstRatioAgainst`. (Different layer — Go renderer vs Python witness — but unchecked.)

## What NOT to do, each for a measured reason

Carried forward from the first handoff, plus what this session added:

- **Do not add the slot to `darkSchemeDerivations`.** It compiles, logs success and
  changes nothing: a palette slot reaches a stylesheet only via `{{palette "X" "lit"}}`
  in a layout, and `accent_text` is declared by **0 of 18**. That is the recorded
  LANDMINE and it is why the fix emits from the renderer.
- **Do not simplify `legibleInkFor`'s `grounds` to a single ground.** dartsonline places
  one ink on the page (1.04) and on a card (1.07). It is a silent half-regression that
  reads as a working fix on whichever page you open.
- **Do not re-narrow the emission behind an `isDarkHex` guard.** Two of the three
  accent-direction sites are LIGHT. Deliberate fork, recorded in VIZ-014.
- **Do not move the step-12 call.** It skips names already in the assembled CSS.
- **Do not grade any of this on a stylesheet or a palette row.** A stylesheet cannot
  resolve the cascade; a palette cannot see a literal that is in no palette. Both
  produced wrong answers in this bug's own history.
- **Do not repoint a palette to fix sub-shape A.** No value satisfies both the fill and
  the ink role — the whole point is a *second* variable.
- **Do not touch vonc.com's Gauntlet buttons** (22 failures). The `gauntlet_dead_cta`
  lane owns that surface. Repointing the shared `system-stats` component is not the same
  thing and is in scope.
- **Do not read a falling failure count as repair.** Content-dependent.
- **Do not conclude "never ran" from an empty `orchestration_states`.** Terminal rows
  reap at ~24h. Ask `scheduled_tasks`, which has no reaper.
- **NEW — do not say "hard-coded" from the audit's output.** It reports a *computed*
  colour and cannot tell you which declaration chose it. Open the template. This error
  cost a council round; the check is in the RUNBOOK.
- **NEW — do not `go build` in the working tree and believe the result.** Another
  session's WIP breaks the package today. Use `git archive HEAD` + your files.

## Facts worth not re-deriving (measured 2026-08-06 unless noted)

- `--color-accent-text` / `--color-primary-ink` / `--color-accent-ink`: **0 uses** across
  `content_components`, `layouts`, `css_snippets`, `site_components`, `page_components`,
  and **0 definitions** in the served stylesheets of finetuning / gamesdesign /
  gaswholesalers. Checked before naming.
- `accent_text` declared by **0 of 18** layouts; `primary_text`/`cta_text`/`header_text`/
  `footer_text` by **18 of 18**; `card_bg` 18; `surface_alt` 3; `icon_chip_bg` 0.
- **17 of 18** layouts use `color: var(--color-primary)` as an ink. **5 of 18** colour
  the base `a {}` with accent.
- Component blast radius: `tool-list` 6 placements/4 sites, `system-stats` 5/4,
  `case-studies-grid` 4/3, `image-hover-card-grid` 1/1 = **16 placements**.
- `scheduled_tasks`: 46 rows, 29 enabled, **0** for `render-audit-agent` either way.
- finetuning `--color-primary-text` **is** `#ffffff` and is **correct for its own slot**
  (primary `#1A1A2E`). The value is right; the slot is wrong.
- ai-agent-orchestration `--color-primary` `#0D1117` is **byte-identical** to
  `--color-surface`.
- `buildTokenAliases` landed `568205c31`, 2026-07-06 — so a missing alias block is not
  staleness.
- Clean sites that 122 lists as failing: relojistas.com, vetcomparison.uk. Also clean:
  fundamentallyai.com, leopardessconsulting.co.uk.

## Where things are

- `PLAN_2026-08-06_contrast_ink_slots.md` — **three corrections marked in place**, plus a
  corrections log at the foot. §2B and §2C are both wrong as originally written; read the
  correction blocks, not the prose under them.
- `RUNBOOK_contrast_ink_slots.md` — five new sections: which declaration chose the
  colour, building against a clean HEAD, proving a test is load-bearing, satisfiable
  contrast fixtures, psql regex quoting.
- `NOTES_contrast_ink_slots.md` — missteps 5 and 6, the council round, the fourth shape,
  the incomplete baseline.
- `README_where_we_are.md` — the owner's plain-prose account, appended.
- `SUBMISSION_2026-08-06b_ink_slots_round2.json` — **the approved plan. The migration
  sketches are real; use them.**
- `BASELINE_2026-08-06_render_audit.txt` — 15 sites, 109 failures, complete.
- `bugs_open/211` — the missing alias block (mechanism measured, cause open).
- `bugs_open/212` — 47 components overriding renderer-owned `--section-*` tokens.
  **Has a decision in it that wants a human**: the only class fix is also the only
  option that can break something that currently works.

## No SUMMARY yet, still deliberately

The engine is committed but inert and the pages are unchanged, so the five headings
would still read "we measured, planned and built". Write one when the first page
measures clean — that is a genuine inflection and this would not be.
