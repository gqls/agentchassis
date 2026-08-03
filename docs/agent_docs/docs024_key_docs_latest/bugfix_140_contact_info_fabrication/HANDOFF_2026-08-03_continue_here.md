# HANDOFF 2026-08-03 (evening, rechecked) — bugfix 140 / RFC_009 · **the 68 are gated; the class is not closed**

**Read this first.** Nothing is half-finished and nothing is broken. The work that was
asked for is done, live and proven. This doc hands over a **measurement taken after that
work which reframes it**, and an ordered plan. **This version supersedes the ~19:45Z one:
the owner asked for the plan to be rechecked, and the recheck corrected four of its
claims** — each correction is marked where it lands. History: `NOTES_…` (technical),
`README_where_we_are.md` (plain prose), the `SUMMARY_…` series.

---

## One-line state

`bugs_closed/140` **CLOSED**. RFC_009 **A not taken, B and C live** (re-verified on chassis
`v1.0.1243`, both replicas, compiled markers + negative control). **The 68 ungated
`skip_field` fields are GATED** — migration 295, commits `f2c8c6b41` `c7e135c54` `aacd6c5fd`
`d0ed9bc96`. The live lint reports **clean across 176 active components**.

**And 295 has a production-grade positive control nobody planned:** at 16:30Z and 16:55Z
another lane created two `featured_article` instances on finetuning.uk. Both rendered
through the post-295 template and **the gates worked** — image block gated out (no
`<img src="">`), meta row gated out (no empty spans). What leaked was a field 295 never
touched — see the finding below.

---

## THE FINDING — the lint checks ONE flavour of declaration, and the fleet has five

The 68 fields I gated were **the `skip_field`-declared subset**, not the broken set. The
lint's UNGATED class filters `on_missing == "skip_field"` — everything else is invisible
to it **by that filter, not by the schema's silence**:

| population (active components, live counts) | fields | checked by the lint? |
|---|---|---|
| declared `skip_field`, referenced, bare | **68 → 0** (migration 295) | **yes** |
| **no `on_missing` declared at all** | **1,991** | no |
| declared `skip_section` | 15 | **no — and this is what bit today** |
| declared `use_fallback` | 21 | no — and the fallback is never applied at render (see LANDMINES: `input_schema` fallback is never consulted) |
| declared `needs_human_review` | 8 | no |

Of the undeclared 1,991: **1,795 fields across 112 components are referenced by their
template and bare** — structural capability to render blank. *[APPROXIMATE in both
directions: it requires a schema entry to count at all, so template vars with NO schema
entry (the 287 desync class) are additional and uncounted; and a `{{range}}`-scoped
subfield sharing a top-level field's name over-counts. An upper bound on neither, a
census of exactly what it encodes.]*

### The demonstration, corrected

`featured_article` had zero live instances at 11:20Z; by 16:55Z it had two, on
finetuning.uk `/ai-guides.html` and `/insights.html`, both `build_status='deployed'`.
Both lack `featured_title` and both store an **empty `<h1>`**.

> **CORRECTED during the recheck (the ~19:45Z version of this file said "bare and
> undeclared"):** `featured_title` **is declared — `on_missing: "skip_section"`.** The
> schema says "without a title, skip the whole section"; instead the section rendered
> with an empty `<h1>`. That is a *stronger* finding than the one it replaces: **a
> declared contract of the wrong flavour is exactly as unchecked as no contract**, and
> nothing warns the author which flavours are enforced (answer today: `skip_field` only,
> and only via the lint's template-shape approximation). The cheap check that caught it:
> read the schema row before writing "undeclared" — one query, which the ~19:45Z version
> ran for `call-to-action` and skipped for `featured_article`.

### Realised damage, re-counted with the platform's own exemptions

The ~19:45Z version said "25 rows across 20 components". Two of those rows are
`data-runtime-fill` shells (vonc's `provocation-card`, `lobby-grid`) — **deliberately
empty at build time, filled by a browser-side loader**, the exact exemption
`check_empty_sections.go` documents as its first live catch. One more belongs to an
inactive component. Honest numbers, stored `rendered_html`, active components, runtime-fill
exempted:

| shape | rows | notes |
|---|---|---|
| `<h1..h4></h1..h4>` empty heading | 15 | + a handful on deleted/inactive components |
| `<a …></a>` dead control | 6 | contact-block 3, tool-list 2, Pricing Tiers 1 |
| `<img src="">` broken image | 3 | Ported Page 2, **case-studies-grid 1 — rendered TODAY**, post-295, on ai-agent-orchestration.com `/index.html`, via `card1..5_image_url`: referenced, bare, **no schema entry at all** |

Of the components I changed, the residue splits three ways — **the ~19:45Z sentence "all
through undeclared fields" was wrong**: `hero` and `Pricing Tiers` still store pre-295
renders of their now-gated **declared** fields (item 4); `featured_article` leaks through
a **`skip_section`-declared** field; `case-studies-grid` leaks through a **schema-absent**
field.

---

## THE PLAN — rechecked, ordered by what closes the door

### 1. A standing output-level check — as a COMPLEMENT to `check_empty_sections`, not a rival

**Correction from the recheck:** the ~19:45Z version repeated RFC_009's "nothing currently
detects it". **False at section granularity.** `check_empty_sections` exists, is one of
only four **enabled** discovery checks (`discovery_checks.go:97`), files `empty_section`
work items routed to `page-build-handler`, honours `data-runtime-fill`, and since
2026-08-03 can even close items it re-observes as fixed (the RFC_010 seam). The true gap
is one level down: **an empty ELEMENT inside a non-empty section** — the empty `<h1>` in
an otherwise-full hero, the dead `<a>` in a rendered contact block. That is what nothing
detects.

And the queue shows the second gap: **detection without resolution.** finetuning.uk's
`empty_section` on `insights` is *unresolved after 3 attempts*; another is stale-triaged
48h+. A new detector that files into the same stalled pipeline adds signal, not repair.

So item 1, precisely:

- **What:** for every active component, for every referenced field **regardless of
  declaration flavour or schema presence**, render through `actions.RenderTemplate` (the
  production entry point — this session's 20/20 harness is the seed, in NOTES) with the
  field absent, and flag `<h1..4></h1..4>`, `<a …></a>`, `<img src="">`, `<td></td>`,
  empty class-bearing blocks.
- **Must honour `data-runtime-fill`** — the same exemption `check_empty_sections` and
  `sectionHasVisibleContent` use, or its first pass re-finds vonc's shells, the
  documented false-positive of the sibling check.
- **Positive control per component** (field present ⇒ element renders): an over-firing
  gate passes any absence-only test.
- **Carrier: the `component-fallback-check` CronJob** — proven to fire unattended, direct
  Postgres. **Not** `quality-discovery-agent` (0 items in all history, `bugs_open/149`).
- **Calibrate against the live corpus first** and expect legitimate empties beyond
  runtime-fill; the must-allow arm is the larger half.

### 2. Reject a non-UUID `Council-Submitted:` — in the HOOK's gate half, not the nudge

Three sessions wrote `Council-Submitted: pending` on 2026-08-03 (the third was this lane,
`f2c8c6b41`); forward-only makes all three permanent false-shaped trailers. The fix is
~3 lines — **but placement matters, and the ~19:45Z version had it wrong**: 
`council-coverage-nudge.sh` is *wrapped by `.githooks/commit-msg` so it can NEVER block*
(deliberate, owner ruling "advisory, defer enforcement"). The rejection belongs in
`.githooks/commit-msg` itself, beside the D2 direction-doc gate that already blocks.
Rejecting a malformed value is not enforcing review — it refuses a false claim, which is
the coverage report's stated dishonesty surface — but say that in the commit message,
because it narrows an owner ruling and should be visibly deliberate.

### 3. The lint's exit code — decide AFTER 1

The "68 predate the check" rationale for exit-0 has expired (count is 0). But flipping it
today enforces a check satisfiable by a no-op — `<td>{{if .v}}{{.v}}{{end}}</td>` clears
the finding and renders the identical blank (residue of 295, mine, documented in
LANDMINES). Flip it once the output-level check exists, or accept that trade knowingly.

### 4. The residual rows — **route through the queue; do NOT fire rerenders**

**The ~19:45Z version said "queue 7-style scoped rerenders". The recheck withdrew that:**
open work items already exist on these exact pages, and firing parallel work at them is
the 2026-07-16 mistake the dispatch rule exists for.

| row | page | what the queue already holds |
|---|---|---|
| `hero` `4d3c6c61…` | finetuning.uk `/blog.html` | `needs_page` "Full rebuild of blog" — **status `failed`** |
| `Pricing Tiers` `25c73a1c…` | gaswholesalers.com `/how-pricing-works.html` | `save_refused_incomplete` on "pricing" — `needs_human_review` |
| (`featured_article` ×2) | finetuning.uk `/ai-guides.html`, `/insights.html` | `empty_section` on insights — **unresolved after 3 attempts** |

finetuning.uk is an **owned, active lane** (commits today; `docs024…/finetuning_uk_service/`).
The right actions: read why the blog rebuild **failed** and continue *that* item; leave the
gaswholesalers row to its human-review item (its `content_data` is an unparsed LLM envelope
— every field absent, so no rerender can populate it; and `hero`'s row has EMPTY
`content_data`, so even a reasoned rerender may just produce a different blank without
content regeneration). Nothing here is a quick win; all of it is coordination.

### 5. `pick-pod-marker` helper — a landmine that fired the same day it was written

A lane probed for RFC_009 B with a phrase that exists only in a Go `//` comment — 0 against
every binary ever built — hours *after* another lane wrote the landmine explaining exactly
this. Prose is not enough. A script: grep the commit's added lines for string literals
excluding comments; build and grep a real binary; print a negative control.

### 6. Retry inside the lint's `load()` — exit 2 is a flake, never a pass

The ~2 MB whole-library fetch through `kubectl exec` truncates intermittently
(`unexpected EOF`). The lint correctly exits 2; it hit twice in one session. Three
attempts with backoff removes the noise.

### 7. File the unparsed-envelope defect — with its prior art

`{"type":"text","result":"<markdown>"}` stored as `content_data`: 2 rows of 1,145, one of
them the gaswholesalers Pricing Tiers row above. Prior art to cite: **`bugs_closed/008`**
(stop_reason undecoded — the envelope-decode class) and `json_envelope.go`'s header. Grep
both bug dirs before filing; the writer that stored it un-decoded is the bug, the rows are
the evidence.

### 8. For the owner — the council gate cannot see config migrations

295 changed 20 shared components fleet-wide, live on apply, and **could not be submitted**:
the trigger refuses submissions touching no `platform/`/`internal/`/`pkg/` path
(`097_TRIGGER…v1.sh:127`). 287 was reviewed only because it rode with Go changes. Either
extend the scope to `sql_for_agents/*.sql` touching `content_components`/`agent_definitions`,
or accept the gap explicitly and rely on the lint + hindsight review.

---

## What is verified live RIGHT NOW (chassis `v1.0.1243`, RS `6cbdfdf4d4`, 19:05Z)

| grep (both replicas, per the roll-is-not-evidence rule) | mxjt7 | wxbbg |
|---|---|---|
| `template invents` — B's absolute refusal (compiled string, `:250`) | 1 | 1 |
| `replacement INTRODUCES` — B's comparative refusal | 1 | 1 |
| `fabricatedFallbackIssue` — the symbol | 2 | 2 |
| `library_fabricated_hours` — C's detector | 1 | 1 |
| `invented_string_xyzzy` — **negative control** | **0** | **0** |

⚠ No orchestration dispatch within ~300s of a chassis restart — the spawn silently drops.

## Health checks

```bash
python3 scripts/check_placeholder_fallbacks.py             # expect CLEAN across ~176. exit 2 = stream flake — retry, NOT a pass
python3 scripts/check_placeholder_fallbacks.py --selftest   # expect 10 must-refuse / 14 must-allow
go test ./platform/orchestration/actions/ -run TestFabricatedFallback
kubectl get cronjob component-fallback-check -n ai-persona-system   # LASTSUCCESS today
```
```sql
-- element-level blanks the lint cannot see (raw scan — exempt data-runtime-fill and
-- inactive components before quoting any of these as damage):
SELECT count(*) FROM page_components WHERE rendered_html ~ '<h[1-4][^>]*>\s*</h[1-4]>';
SELECT count(*) FROM page_components WHERE rendered_html ~ '<a [^>]*>\s*</a>';
SELECT count(*) FROM page_components WHERE rendered_html ~ '<img[^>]*src=""';
```

## The six things worth knowing before you touch any of it

1. **A rerender does NOT regenerate sections unless `spec.reason` is set** — reason-less
   re-staples stored HTML. And a rerender **cannot invent content**: a row with empty
   `content_data` needs content regeneration, not a redraw.
2. **`page_component_history.source` tells you what WROTE a section** — don't infer it
   from nearby work items.
3. **`RenderContext` supplies some fields itself** (json-tagged scalars reach the template
   contract) — "absent from `content_data`" is not "absent". Of the original 68, exactly
   one collided (`about-commercial-block.domain`).
4. **The write path does NOT close the door** — ~10 writers touch `html_template`, two are
   gated. Gate where sound, report everywhere.
5. **Go gate and Python lint are two implementations on purpose**, pinned to ONE fixture
   (24 cases). Change a pattern ⇒ change the fixture, or one side fails.
6. **Obeying `skip_field` is NOT "wrap the field in `{{if}}`"** — 62 of the 68 sat in
   fixed-arity rows where that edit is a no-op or malformed HTML, and both pass the lint.
   Full entry in LANDMINES; treatments in migration 295's header. **Corollary from the
   recheck: the lint checks only the `skip_field` flavour — a `skip_section` or
   `use_fallback` declaration is enforced by NOTHING at render time.**

## Explicitly NOT owed

- No pod-grep outstanding — table above, this evening, both replicas, with controls.
- No council verdict outstanding, and 295 could not have one (item 8).
- **No rerenders owed** — the residual rows route through existing queue items (item 4).
- The 09:20Z "B is not live" notice is REFUTED (comment-only marker; see `WRONG_CALLS.md`).
  It was raised correctly — observables only, routed to the owning lane.
