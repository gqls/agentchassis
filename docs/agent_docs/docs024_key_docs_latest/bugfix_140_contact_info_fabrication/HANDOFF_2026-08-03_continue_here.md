# HANDOFF 2026-08-03 (evening) — bugfix 140 / RFC_009 · **the 68 are gated; the class is not closed**

**Read this first.** Nothing is half-finished and nothing is broken. The work that was
asked for is done, live and proven. What this doc exists to hand over is a **measurement
taken after that work, which reframes it** — and an ordered plan for what to do about it.

Supersedes the 10:02Z version of this file. History lives in `NOTES_…` (technical,
append-only), `README_where_we_are.md` (plain prose) and the `SUMMARY_…` series.

---

## One-line state

`bugs_closed/140` **CLOSED**. RFC_009 **A not taken, B and C live**, and **the 68 ungated
`skip_field` fields are GATED** (migration 295, commits `f2c8c6b41` + `c7e135c54` +
`aacd6c5fd`). The live lint reports **clean across 176 active components**.

**And that clean result is narrower than it sounds.** See the next section — it is the
reason this handoff exists rather than a closing note.

---

## THE FINDING THAT CHANGES THE PICTURE

The 68 fields I gated were **the fields that happened to be DECLARED**, not the fields that
are broken. `on_missing` is what made them *visible to the lint*; it is not what made them
*wrong*. Measured this evening, live:

| population | size | can the lint see it? |
|---|---|---|
| declared `skip_field`, referenced, **bare** | **68 → 0** (migration 295) | yes — this is its UNGATED class |
| **undeclared**, referenced, **bare** | **1,795 fields across 112 components** | **NO — invisible by construction** |

*[MEASURED 2026-08-03 19:2xZ, live. This is structural CAPABILITY to render blank — a
field that is always supplied never renders empty — so it is an upper bound on exposure,
not a defect count. The realised damage is the next table.]*

**Realised today, in stored `rendered_html`, fleet-wide — 25 rows across 20 components:**

| shape | rows | why it is not "a mild blank" |
|---|---|---|
| `<h1..h4></h1..h4>` empty heading | 21 | an empty heading is a WCAG failure and a broken document outline |
| `<a …></a>` dead control | 7 | invisible, unclickable, announced as nothing by a screen reader |
| `<img src="">` | 4 | resolves to the page URL and re-requests it — a broken image |
| `<td></td>` | 0 | (this is the shape 295 removed) |

**Only 4 of those 20 components are among the 20 I changed**, and their remaining defects
are all through **undeclared** fields.

### The single fact that makes the case

`featured_article` had **zero live instances when I measured at 11:20Z**, so I recorded its
6 gated fields as "purely prophylactic". At **16:30Z and 16:55Z** another lane created two
instances of it, on `finetuning.uk/ai-guides.html` and `/insights.html`. **Both lack
`featured_title`. It is bare, it declares nothing, and both pages now render an empty
`<h1>`** — the two empty headings attributed to `featured_article` above.

So within five hours of gating that component's declared fields, it began serving the same
defect through an undeclared one. That is not bad luck; it is the boundary being in the
wrong place.

> **CORRECTION to my own figures, and what caught it.** I wrote "20 of the 68 fields have
> zero live instances" and "3 stored rows carry the empty element" into migration 295's
> header, `f2c8c6b41`'s message and RFC_009. Re-measured this evening: **three** of those
> four components are still at zero (`product-specs`, `archetype-result-card`,
> `bayesian-ranking-hero-tool_pre_037`) but `featured_article` is not, and the "3 rows" is
> **2** for my 20 components — the third (`<h2></h2>` on `finetuning.uk/blog.html`) belongs
> to `call-to-action`, which I never touched and whose `headline` is *undeclared and bare*.
> Neither figure was wrong when taken; the first went stale in five hours. **A census of
> live instances on this fleet has a half-life of hours — date it, or re-run it.**

---

## THE PLAN — ordered by what closes the door, not by effort

### 1. Promote the render harness to a standing check ← **the one that matters**

**The problem it solves is twofold, and the second half is the dangerous one.**

*First*, the lint cannot see the 1,795 undeclared fields, because its UNGATED class keys on
`on_missing`. A check that looks at **rendered output** instead of template shape needs no
declaration and would have caught `featured_article`, `call-to-action`, `contact-block` and
the other 17 on day one.

*Second — and this is residue I created* — the lint tests for a gate **anywhere** in the
template and **cannot see what the gate encloses**. So the next person handed an ungated
finding can "fix" it as `<td>{{if .v}}{{.v}}{{end}}</td>`, which renders the identical empty
cell **and clears the finding for ever**. I have removed the detector's ability to complain
about those 68 without any guarantee the blank is gone. An output-level check is immune to
this by construction: it measures the artefact, not the shape.

**It already exists and it already passed 20/20.** The harness in this session's scratchpad
renders a component through `actions.RenderTemplate` — the production entry point, not a
replica of its `text/template` config — with a field present and absent, and asserts the
element vanishes when absent and **still renders when present** (the positive control is
the half that catches an over-firing gate). Promote it:

- **Where:** a Go CLI under `cmd/`, or a test in `platform/orchestration/actions/`. It must
  call `RenderTemplate`, not reimplement `executeGoTemplate`'s options and FuncMap.
- **What it asserts:** for every active component, for every referenced field
  **regardless of declaration**, render with that field absent and fail on
  `<h1..4></h1..4>`, `<a …></a>`, `<img src="">`, `<td></td>`, or an empty
  class-bearing block element.
- **Calibrate first**, per `component_write_guard.go`'s standing instruction. Expect real
  false positives: some empty elements are legitimate JS mount points. The must-allow arm
  is the larger half.
- **Carrier:** the `component-fallback-check` CronJob already exists, is proven to fire
  unattended, and has direct-Postgres access. Do **not** wire it to
  `quality-discovery-agent` — that carrier has raised 0 items in all history
  (`bugs_open/149` Group B/C), which is how the original 140 defect survived from birth.

**Do this before 3.**

### 2. Make `Council-Submitted:` un-fakeable — ~3 lines, no judgement

`scripts/council-coverage-nudge.sh:52` **already greps for the trailer**. Have it reject a
value that is not a UUID. **Three independent sessions wrote `Council-Submitted: pending`
on 2026-08-03** — the third was me, on `f2c8c6b41` — each having read the CLAUDE.md
paragraph that explains it, none stopped by anything. Forward-only forbids the amend, so
all three commits carry a correlation resolving to nothing, permanently. The recommendation
is logged three times in `WRONG_CALLS.md`. **A rule three careful sessions break in one day
is missing its enforcement, not its explanation.**

### 3. Decide the lint's exit code — but only after 1

C's UNGATED class deliberately does not fail the exit code, because "68 predate this check
and a permanently-red gate is one everybody learns to ignore". **That reasoning has expired:
the count is 0.** Flipping it to exit 1 closes the door behind 295. The cost is that a new
component from another lane turns the daily CronJob red until someone adds one `{{if}}`.

**Owner's call, and worth taking it after step 1** — on its own, exit 1 enforces the check
that can be satisfied by a no-op (see 1), which is the worst of both.

### 4. Rerender the two residual rows — small, visible, finishes the job

Templates do not retro-apply. Two stored rows still serve an empty element from a component
295 fixed:

| component | site / page | row |
|---|---|---|
| `hero` | finetuning.uk `/blog.html` | `4d3c6c61-3575-4860-97ec-8f4da3057b0b` |
| `Pricing Tiers` | gaswholesalers.com `/how-pricing-works.html` | `25c73a1c-b3af-48af-978b-95f7e500e8fa` |

⚠ **A reason-less rerender will not fix them** — it re-staples stored section HTML.
`check_rerender_mode` routes only `image_landed|section_data_resolved|cta_links_stale` to
`rerender_page_sections`. Pattern to copy: `SQL_2026-08-02_scoped_rerender_seven_contact_pages.sql`
(this lane, `395246bb5`). This thread got that wrong in the middle of 140 and watched six
pages come back unchanged.

⚠ **The gaswholesalers row will not repair by rerendering alone.** Its `content_data` is an
**unparsed LLM envelope** — `{"type":"text","result":"{\n  \"section_title\": …\n}\n\n---\n\n**⚠ HUMAN REVIEW REQUIRED…"}`
— so every tier field is absent and the section is an empty shell. Post-295 it renders
*less* rather than *wrong*. See 7.

### 5. A helper for picking a pod-grep marker — the check exists only as prose

On 2026-08-03 a lane reported RFC_009 B missing from the running binary, on
`strings … | grep -c "declared skip_field but never gated"` returning 0. That phrase lives
**only in a `//` comment** (`component_fallback_guard.go:78`), so it returns 0 against every
binary ever built. **B was and is live.** The landmine was already written that morning by
another lane (`663b063ef`) and it still happened hours later, which says prose is not
enough. A `scripts/pick-pod-marker.sh <commit>` that (a) greps the commit's added lines for
string literals **excluding comments**, (b) builds and greps a real binary to prove the
marker compiles in, (c) prints a negative control, converts a landmine into a command.

### 6. Retry the DB fetch inside the lint — a recurring flake, correctly handled but noisy

`kubectl exec … psql -tAc` on the ~2 MB whole-library payload intermittently truncates
mid-stream (`"Copying stdout failed" … unexpected EOF`). The lint handles it correctly —
**exit 2, "could not reach the database"**, never a false clean — but it hit twice in one
session. Three attempts with backoff inside `load()` removes it. **Until then: exit 2 is a
flake, not a pass; retry.**

### 7. File the unparsed-envelope defect — 2 rows, but a real class

`{"type":"text","result":"<markdown>"}` stored as `content_data` where flat fields were
expected. **2 rows of 1,145**, 2 components. Small, but every field of those components is
absent, so it defeats any per-field gate. Grep `bugs_open/` and `bugs_closed/` for
`json_envelope` / `raw-text envelope` before filing — `json_envelope.go:17` documents a
sibling case and this may already be owned.

### 8. For the owner — a gap in the council gate's scope, not a task

Migration 295 changed **20 shared components fleet-wide and went live the moment it was
applied**, and it **could not be submitted for review**: the trigger refuses any submission
touching no `platform/`, `internal/` or `pkg/` path (`097_TRIGGER…v1.sh:127`). Migration 287
was reviewed only because it rode alongside Go changes in the same submission. **Config that
ships instantly is arguably the class most worth reviewing, and it is the class the
path-based rule excludes.** Either extend the scope to `sql_for_agents/*.sql` that touch
`content_components`/`agent_definitions`, or accept it explicitly and rely on the lint.

---

## What is verified live RIGHT NOW (re-checked after the new chassis roll)

**Chassis `v1.0.1243`**, ReplicaSet `6cbdfdf4d4`, pods started 19:05–19:06Z. RFC_009 B is in
this binary — pod-grepped on **both** replicas with compiled markers and a negative control,
per the standing rule that a roll is not evidence your fix shipped:

| grep | mxjt7 | wxbbg |
|---|---|---|
| `template invents` — absolute refusal (`component_fallback_guard.go:250`) | 1 | 1 |
| `replacement INTRODUCES` — comparative refusal | 1 | 1 |
| `fabricatedFallbackIssue` — the symbol | 2 | 2 |
| `library_fabricated_hours` — C's detector | 1 | 1 |
| `invented_string_xyzzy` — **negative control** | **0** | **0** |

⚠ **Do not dispatch orchestration within ~300s of a chassis restart** — the spawn is
silently dropped.

## Health checks (all fast, all read-only)

```bash
python3 scripts/check_placeholder_fallbacks.py             # expect CLEAN. exit 2 = stream flake, retry — NOT a pass
python3 scripts/check_placeholder_fallbacks.py --selftest   # expect 10 must-refuse / 14 must-allow
go test ./platform/orchestration/actions/ -run TestFabricatedFallback
kubectl get cronjob component-fallback-check -n ai-persona-system   # LASTSUCCESS should be today
```
```sql
-- the class this lane is really about, and what the lint cannot see:
SELECT count(*) FROM page_components WHERE rendered_html ~ '<h[1-4][^>]*>\s*</h[1-4]>';  -- 21 today
SELECT count(*) FROM page_components WHERE rendered_html ~ '<a [^>]*>\s*</a>';           --  7 today
SELECT count(*) FROM page_components WHERE rendered_html ~ '<img[^>]*src=""';            --  4 today
```

## The five things worth knowing before you touch any of it

1. **A rerender does NOT regenerate sections unless `spec.reason` is set.** A reason-less
   item re-staples stored HTML. *Do not read `create_rerender_items_action.go:219`'s
   `&& componentIDStr != ""` as the consumer's rule — it is producer-side.*
2. **`page_component_history.source` tells you what WROTE a section.** Do not infer it from
   whatever work item completed nearby.
3. **The roster-free detector is UNSOUND.** `RenderContext` carries json-tagged scalars that
   reach the template contract, so a component can legitimately render a value its
   `content_data` lacks. Of the 68, exactly one (`about-commercial-block.domain`) collides.
   Check yours before assuming absence.
4. **The write path does NOT close the door.** ~10 writers touch `html_template`; two are
   gated. **Gate where it is sound, report everywhere.**
5. **The rule has TWO implementations on purpose** (Go gate, Python lint) — the drift class
   the rule itself detects — pinned to ONE shared fixture, 24 cases. Change a pattern,
   change the fixture; one side will fail if you don't.

**Sixth, added by this session:** obeying `skip_field` is **not** "wrap the field in
`{{if}}`". 62 of the 68 sat in a cell of a fixed-arity row, where that edit is either a
**no-op** or emits **malformed HTML** — and both pass the lint. Full entry in `LANDMINES.md`;
the four treatments and per-component reasoning are in migration 295's header.

## Explicitly NOT owed

- No pod-grep outstanding — re-verified above on `v1.0.1243`, both replicas, with controls.
- No council verdict outstanding, and **295 could not have one** (see 8).
- No rerenders outstanding **except the two named in 4**.
- The 2026-08-03 09:20Z notice in `NOTES` claiming B is not live is **REFUTED** — see 5 and
  `WRONG_CALLS.md`. It was raised correctly and that is why it cost twenty minutes.
