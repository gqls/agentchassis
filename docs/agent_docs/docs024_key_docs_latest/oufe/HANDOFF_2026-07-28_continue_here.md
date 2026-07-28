# HANDOFF — oufe.com / oxenunity.com, 2026-07-28 evening

**Cold-start entry point. Supersedes `HANDOFF_2026-07-27_continue_here.md`**, which
is now stale in several places (it still says "there is no chart renderer", which
is wrong and corrected in place there).

Read this, then `SUMMARY_2026-07-28b_oufe.md` for the plain-prose read-out, then
`NOTES_oufe.md` from the bottom up — the missteps are the part you cannot
rederive. `DESIGN_2026-07-28_premise_branching_and_deepthink.md` holds the next
big design decision.

---

## 1. State, verified 2026-07-28 evening

Chassis **`v1.0.1192`, TWO replicas** (was one — the replica-scaling work from
another thread has landed; if you see odd dispatch behaviour, that is new and not
yours). Series code pod-verified present in the running binary. Council lane
consuming at lag 0.

All eight pages 200:

| url | what is on it |
|---|---|
| `/` | banner, hero, brief-explanation, info-card-grid, CTA |
| `/about.html` | |
| `/cases/index.html` | |
| `/cases/thames-water.html` | **prose + mechanism diagram + TWO charts** |
| `/tools/tool-recovery-waterfall.html` | the tool, Tier-4 verified 13/13 |
| `/disclaimer.html` | sections E + F, live and locked |
| `/privacy.html` | live and locked |
| `/contact.html` | `mailto:` form |

Site id `a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39`. Evidence register: **32 facts**.
`claimscan` **0 findings across 17 components**. `contrastscan` 14 pairs all pass.

`oxenunity.com` unchanged, still a hand-authored one-pager.

## 2. What shipped today (all live, all verified)

1. **Council lane** — `bugs_open/096` **CLOSED** with a measured confirmation.
2. **Tier-4 acceptance** — failed 2/11, then **13/13 with the tool source
   untouched**. The fence was the defect. Filed `bugs_open/126`.
3. **Legal pages** — disclaimer + privacy, `rebuild_policy='owned'` +
   `lock_type='permanent'`, linked from the footer on every page.
4. **`mechanism-flow` component** (mig 247) + Part 26A cram-down on the Thames
   page (248). Seven steps, two branches, every step from a registered fact.
5. **Series facts** (`claims_series.go`) — `Observation{as_of,value,source}`.
6. **`evidence-timeseries` component** (mig 250) — **BUILT, STILL UNUSED.**
7. **`cmd/contrastscan`** — browser-based contrast audit.
8. **Two charts on the Thames page** — capital structure (251/252) and the Ofwat
   determination (254/255).
9. **Scanner fix** — written-out British dates ("28 July 2026") were read as
   business figures fleet-wide.

## 3. NEXT — in priority order

### 3a. Wire `scripts/render_audit.py` — the tool EXISTS, nothing runs it

> **CORRECTED, same evening.** This section used to say "wire `contrastscan`".
> `cmd/contrastscan` was a **duplicate I built without finding
> `scripts/render_audit.py`** (brochure workstream, 2026-07-27), because my
> prior-art grep was `--include=*.go` and the prior art is Python. **The Go tool
> has been deleted.** Do not rebuild it.

What is **already wired**: `check_palette_contrast` — a discovery check, DB-only,
microseconds, reading the *composed* palette. It is good and it states in its own
header what it cannot see.

What is **not wired**: `scripts/render_audit.py`, which renders every element in
headless Chromium and is the only thing catching 026 **family 3** — a component
hard-coding an ink over a themed fill. That is precisely the class that put an
unreadable chart on this site. Shape: `bugs_open/122` candidate 2. It needs a
browser, so it is not a DB-only discovery check; the natural home is the
**browser-runner-adapter, which already runs Playwright in-cluster**.

**Fleet audit run 2026-07-28 evening — 65 live contrast failures on 7 homepages:**

| site | failures | note |
|---|---|---|
| gamesdesign.co.uk | **30** | stat band, cyan-on-cyan |
| **idea.uk** | **18** | the site that took the first sale |
| vetcomparison.uk | 10 | |
| relojistas.com | 2 | |
| oufe.com | ~~5~~ **0** | fixed, see below |
| webdesign.co.uk | 0 | |
| leopardessconsulting.co.uk | 0 | |

Those belong to other workstreams — route, do not fix over the top.

### 3b. More charts, tools and guides — the owner's standing ask
His words on seeing the first chart: *"this is starting to look great! We'll need
a lot more of this of course, and tools as is already in the plan."*
Reusable now: `evidence-chart` (magnitudes), `evidence-timeseries` (dated series,
unused), `mechanism-flow` (process), `tool-guide-intro` (**exists** — the answer
to "guides at each point", never used here).

### 3c. First real use of `evidence-timeseries`
It needs a measure that genuinely moves over time. Ofwat's performance commitments
(leakage, spills, sewer flooding) are stated as 2024-25 baseline → 2029-30 target
and may yield one. **Do not force it** — twice today the honest answer was that
the data was a comparison, not a trend.

### 3d. `bugs_open/126` has an inbound contribution from another thread
Another workstream audited `tool-improver.note_refusal` and found its error
handler disabled (`error_step_disabled_086`) as containment from bug 086. They
handed the decision here because `who-owns.py 126` names oufe. **Unread by this
thread** — read the CONTRIBUTION section at the foot of 126.

### 3e. Owed elsewhere
- `bugs_open/122` candidate 1: header chrome hardcodes `color: white` on the
  accent, failing **5 of 6 sites**. Needs a generator fix; chrome is a stored
  artefact so existing sites need `nav_drift` → `nav-updater` (`bugs_open/117`).
- robot-hands' primary CTA is **white on white, invisible**; vonc's Gauntlet
  buttons fail. Both owned by other threads — evidence is in 122, do not fix over
  the top of them.

## 4. Traps that will cost you time

**Rendering an OWNED page.** `save_page_sections` REFUSES `rebuild_policy='owned'`
pages, so `section_data_resolved` fails at `save_sections`. You need assemble-only
— and `TRIGGER_rerender_page.sh` **cannot request it**: `REASON="${3:-...}"` treats
an empty string as unset. Publish the envelope directly with `spec.reason` absent.
Recipe in `RUNBOOK_oufe.md` §9–10.

**Do NOT lock a component at insert time.** `save_page_sections` *preserves* locked
rows rather than rendering them, so a row locked with empty `rendered_html` renders
as **nothing, for ever**, and reports success. Write `rendered_html` in the same
statement (migration-182 pattern), by executing the component's OWN template.

**`$facts` is declared by the template**, not supplied by the engine
(`{{- $facts := .facts -}}`). An undeclared variable is a **parse error**, so the
component renders nothing. Same class: **`inc`/`add` are not in the funcmap** —
use a CSS counter.

**SVG text is invisible to the claims gate.** Charts must use real HTML text with
CSS-drawn furniture. Verified by decoding the payload claimscan received.

**`context_terms` must match the LABEL, not the prose.** A registered value gets
flagged as unregistered when the chart's terse label contains none of its terms.
Fix the label, not the guard.

**Contrast — ROOT CAUSE NOW FIXED.** `--color-primary` was `#1B2A3B`, *identical to
`--color-surface`*, so all 7 of its foreground uses were invisible or near it.
Three components were patched individually before anyone traced it to the palette.
Its 3 background uses all pair it with `--color-primary-text`, so the pair flips
together: now `#86ADDE` / `#0F1820`. oufe went **5 → 0** real failures.
`--color-card-bg` was also white on a dark site (fixed). Chart furniture needs the
**3.0** non-text threshold and `--color-border` scores 1.66 — use `--color-accent`.

**A `render_audit.py` reading marked "over an image — ratio approximate" is a
heuristic, not a verdict.** It assumes mid-grey for an unknown backdrop. oufe's
last "failure" was a white button over a near-black hero and is genuinely fine —
confirmed by screenshot.

**Supersede-then-insert must be SEQUENTIAL.** A CTE doing both fails: every branch
sees one snapshot, so the partial unique index still sees the old row (23505).

**403s are bot protection, not authentication.** A browser gets 200 where curl and
WebFetch get 403. `RUNBOOK_oufe.md` §12 has the recipe, including
`pdftotext -layout` and why a column mapping must be corroborated arithmetically.

**Catastrophic regex:** `grep -oE "[^.]{0,180}x[^.]{0,180}\."` over 68KB was
OOM-killed. Split in Python.

## 5. The council submission that is still unresolved

Series facts, correlation **`da40ddf0-7acd-40f6-9826-d9161f5601be`**.

- **Round 1 → REVISE**, and the objection was correct and valuable: `ValidateSeries`
  enforced the per-observation source rule but `numberSupported` never called it.
  **Fixed** — `observationHasResolvableSource` is now shared by both.
- **Round 2 → `complete_invalid` three times**, including a control run **without**
  `RESUBMIT_CORR`, which refutes the resubmission path as the cause. The `fix_plan`
  artifact persists each time, so structural validation passes and the failure is
  in a review or decide step. Evidence written into `bugs_open/119`.

**The code is committed, tested and live, and carries NO `Council-Reviewed:`
trailer.** Do not add one — it has not been approved.

**When you read a verdict, query by YOUR correlation.** The most recent
`council-gate` note was a different submission's APPROVED; reading the latest note
would have attached someone else's verdict to this plan.

## 6. Rails that must not be relaxed

- **No figure in any brief, spec, identity or content_direction.** Only the
  evidence register, with a source.
- **Never publish a figure about a real company without a source URL and capture
  date.** Today's additions each had their quote checked *literally present* in the
  fetched page — not taken from a model's summary of it, and not from a search
  result.
- **The grounded lane must keep its inability to publish.**
- **Section G (liability cap) stays parked** until something is for sale. It is the
  only item genuinely needing a solicitor.

## 7. The thing worth carrying forward

Nearly everything that went wrong today was **something this workstream had already
written down and then not checked**: a warning in a handoff ("no chart renderer" —
two were live), a hazard in a bug file (the white card, described as "waiting"), a
rule in a validator nothing called, and a checker whose scope we chose ourselves
and then trusted.

The sharpest form: **a measuring instrument you built yourself is the hardest to
distrust, because when it agrees with you it may only be agreeing with your blind
spot.** `contrastscan` reported the page clean because it only measured links and
buttons — a scope I chose, for reasons that felt sufficient.

The checks that actually caught things were dull: run the test, open a screenshot,
query the component library rather than the Go source, read `features_open/`, ask
by your own correlation. All of them are cheap. None of them are interesting. That
is why they get skipped.
