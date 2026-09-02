# 423 — a chrome slot whose STORE fails is reported as a reasonless `false`: boxingonline's footer renders, Postgres refuses the bytes (invalid UTF-8, 0x80), and the action's own reason fields all stay empty

**Filed 2026-08-31 (~18:2xZ)** by the delivery-lane session. **Mechanism CAPTURED LIVE
at the artefact**, not inferred: a dedicated repro dispatch (rerender-pages, corr
`7fb750a3-f804-481e-8e8a-656c3290ecd9`) with a log monitor on every chassis pod. This
supersedes the iteration-capped 090 run `387c0a2d` (UNVERIFIABLE) on the same symptom;
the first-hand capture below is the substitute verification per the 2026-07-31 ruling.

## The captured mechanism (pod agent-chassis-6d6856d8d5-vswxv, 18:14:26Z)

```
error  render_site_components_action.go:1338  "Failed to store rendered component"
       slot=footer  error=ERROR: invalid byte sequence for encoding "UTF8": 0x80 (SQLSTATE 22021)
info   render_site_components_action.go:355   "RenderSiteComponentsAction: Complete"
       rendered={"footer":false,"head":true,...}
```

Two defects, one incident:

### Half 1 — the OBSERVABILITY defect (settled at the code + the artefact)

`renderAndStoreSiteComponent`'s store-failure branch
(render_site_components_action.go:1338-1343) logs the error and returns
`false, false, degraded, nil` — the **nil error** means the caller's
`chrome_render_failed` map never hears about it, `ineligible_chrome` and
`locked_slots_preserved` stay empty, and the step completes SUCCESS with
`rendered.footer=false` as the only trace. Measured consequence: THREE runs tonight
(15:39 wave `3f604312`, 17:47 nav-updater `07fed163`, 18:14 repro `7fb750a3`) each
declined the footer this way; two sessions spent ~an hour eliminating locks,
component eligibility, content_data emptiness (fleet control: 31/33 empty, 30
render — boxingonline session) and slot-set membership before a live log capture
named it. The action already has the exact surface for this (`chrome_render_failed`,
built for bugs_open/260) — the store-failure branch just doesn't use it.

### Half 2 — the DATA defect (mechanism named, source OPEN)

The bind that Postgres rejects is `renderedHTML` ($1 of the UPDATE at :1330), so the
invalid byte exists in the GO STRING the template pipeline produced. Every DB-sourced
input is valid UTF-8 by construction (Postgres cannot hold otherwise), so **0x80 — a
bare CONTINUATION byte, e.g. the middle of an em-dash E2 80 94 — is introduced
between template execution and the bind**: the classic signature of a byte-indexed
truncation or splice cutting a multi-byte character. `[UNVERIFIED]` which transform:
candidates include template helper funcs that slice by byte, DropDeadURLControls,
and the favicon/OG injection — none read, none accused. Site-specific: the same
statement stored 30 estate footers cleanly within the last week, and boxingonline's
own footer stored fine at its 13:31 first bake; failures began with the 15:39 wave,
i.e. after the day's data changes (email scrub, nav changes) altered the render
inputs — the corrupting input is in whatever changed for THIS site.

## Why this matters beyond one footer

- It is the third instance tonight of the 420-family shape: **a removal/refresh that
  reads as done** — here the enabling defect is half 1, which converts every store
  failure into silence.
- Consequence on the paid site: the served footer is a 16:05 hand-patch that is the
  ONLY definition of the site's footer (content_data empty by fleet norm), and the
  pipeline cannot replace it until half 2 is fixed. Recorded on the boxingonline
  pre-delivery list.

## Fix candidates

1. Half 1, small and surgical: the store-failure branch populates
   `chrome_render_failed[slot]` (return the error as `renderErr` like the execution-
   failure branch ~90 lines above — same disposition, same surface). Makes every
   future instance of half 2 loud.
2. Half 2: find the byte-slicer (start from the diff between this site's render
   inputs at 13:31 vs 15:39; or bisect by rendering the footer template against the
   live context in a test and hex-scanning the output), then make it rune-safe.
   Postgres's error does not say WHERE in the string the byte sits — a test harness
   hex-scan does.

## ADDENDA 2026-08-31 (~19:0xZ) — graders and a latent sibling, from the boxingonline session (attributed; evidentiary status preserved)

**Grader 1 — "the footer contains multi-byte characters" is NOT an explanation.**
32 of 32 stored footers fleet-wide contain multi-byte characters, 31 contain an
em-dash — all stored fine. What distinguishes this site is something that CUTS at a
byte offset, not something that contains one. Reject any proposed cause of that
shape unless it beats this census.

**Grader 2 — "empty sites.email" is NOT sufficient either** (and it killed that
session's own best timing-fit hypothesis, offered here so nobody re-derives it):
13 sites have an empty email and 12 have a rendered footer, one as recently as
today. The DropDeadURLControls-on-a-dead-mailto theory fits this site's TIMING
(failures start with the first render after the ~15:37 email scrub; the 13:31 bake
predates it) but is unsupported by timing alone. The sharp revival test, unrun:
whether those 12 sites' footer templates emit a mailto control at all (this site's
did; theirs may gate it out before it can go dead).

**Code read (theirs) — the obvious alignment bug is NOT there**: renderedHTML
reaches the bind unsliced; the only surgery between RenderTemplate (:1075) and the
store is DropDeadURLControls (:1227) and injectBrandHeadTags (:1261).
`ReplaceAllInMarkup` slices by offsets from a regex over a masked copy — the classic
mid-rune cut — but `maskNonMarkup` (markup_spans.go:108) masks PER BYTE, preserving
length and offsets. **One reading NOT discharged (a reading, not a finding — read it
properly, quote it never):** the span loop `for i := s.Start; i < s.End` can mask
PART of a multi-byte character if a span boundary lands mid-rune, putting a bare
continuation byte into the masked (matching-only) copy — it would have to move a
match boundary to matter.

**Latent sibling, fold into half 1's fix** — same file, same class, NOT this
footer's cause: lines 1439, 1689 and 1779 each do
`if len(summary) > 250 { summary = summary[:247] + "..." }` — a byte slice that cuts
mid-rune exactly as this file describes. They write work-item summaries — and
**:1689's summary is `fmt.Sprintf("Chrome %s failed to render: %v", slot, renderErr)`,
an arbitrary error string — so the surface that REPORTS a chrome failure can itself
mint invalid UTF-8 and fail its own insert.** Whoever makes the failure path
load-bearing (fix candidate 1) must make these three slices rune-safe in the same
pass, or the new reporting can die of the same disease it reports.

## How to verify

Re-run render_site_components for the site (the repro dispatch shape is in the
webdesign NOTES 2026-08-31 ~18:1x entry): half 1 fixed ⇒ the failure appears in
`chrome_render_failed` in collected_data; half 2 fixed ⇒ `rendered.footer=true`, the
footer row's updated_at moves, `rendered_html_digest = md5(rendered_html)`, and the
served footer still carries NO contact block (sites.email empty gates it —
component_library.go:1988) — that last probe is the pre-delivery check the
boxingonline session defined.

---

## ROOT CAUSE FOUND AND FIXED 2026-09-02 — half 2 is `w[:1]`, and it is a class, not a case

**By the bugfix_423 lane** (`docs/agent_docs/docs024_key_docs_latest/bugfix_423_chrome_utf8/`).
Council `Council-Submitted: dc62975f-9d38-4b3c-9174-330307b9df95`.
**Go, so INERT until the next chassis roll.**

### The cutter, named

`buildServicesHTML` (this file's own `render_site_components_action.go:1622`, called
at `:125` to build the **footer** "Our Services" column):

```go
words := strings.Fields(label)
for i, w := range words {
    if len(w) > 0 { words[i] = strings.ToUpper(w[:1]) + w[1:] }   // ← a BYTE slice
}
```

`strings.Fields` makes a **standalone em-dash its own word**. `w[:1]` then cuts a
3-byte rune after one byte: `ToUpper` of the lone lead byte decodes as U+FFFD, and
`w[1:]` re-attaches the orphaned continuation bytes.

`[MEASURED 2026-09-02, by execution]` — `go run`, not inference:

```
"—dash"  -> ef bf bd 80 94 64 61 73 68   valid=false
"“quote" -> ef bf bd 80 9c ...           valid=false
"École"  -> ef bf bd 89 ...              valid=false
```

**The first invalid byte is `0x80`** — verbatim the byte in the live pod capture at
the top of this file. The trigger on this site is
`pages.title` for `tool-boxing-trivia-quiz`: **"Boxing Quiz — Test Your Knowledge |
Tools"** (active, `in_header`, `nav_order` 200 — inside the query's `LIMIT 6`).

### It beats both graders, and the disconfirming result was available

**Grader 1 satisfied.** This CUTS at a byte offset; it does not merely contain a
multi-byte character. `[MEASURED 2026-09-02]` the discriminating census, run both
ways so it could have come out otherwise:

| census | result |
|---|---|
| sites with a services-column label (same predicate, within `LIMIT 6`) containing a word whose **first rune is multi-byte** | **exactly 2** — boxingonline.com, garden-tools.uk |
| sites whose `site_components` footer is **not** `build_status='rendered'` | **exactly the same 2** |

Zero false positives, zero false negatives; every other footer on the fleet is
`rendered` with `rendered_html_digest = md5(rendered_html)`.

**Grader 2's theory is not needed and is not implicated.** The empty-`sites.email`
/ dead-mailto timing fit was a coincidence of the same afternoon's data changes.
The `DropDeadURLControls` and `maskNonMarkup` readings the 08-31 addendum left
undischarged are **not the cause** — that addendum's code read was right that
`renderedHTML` reaches the bind unsliced; the cut happens **earlier**, in an input
built at `:125`, before `RenderTemplate` ever runs. The un-discharged mid-rune
masking reading remains un-discharged and is now **not urgent**: it can no longer
reach the database unnoticed, because of the gate below.

### A SECOND CASUALTY, older than this bug

**garden-tools.uk**'s footer has been failing since **2026-08-23** — ten days, on a
different trigger word ("How We Assess Garden Tools **—** Our Methodology | Garden
Tools UK") — and nobody knew, because of half 1. Its `rendered_html` is **NULL**,
not stale: nothing has ever been stored for that slot.

### What shipped

1. **`datahelpers.UpperFirst`** — one rune-safe primitive, and **all 8 call sites of
   the idiom converted** (census 2026-09-02, non-test Go: `assemble_from_library.go:211`,
   `render_site_components_action.go:1622`, `component_library.go:1379`,
   `multipage_actions.go:355` and `:1435`, `format_content_direction.go:119`,
   `data_helpers.go:1980`, `cmd/webdesignport/harvest.go:411`). ASCII parity with
   the old idiom is pinned by test — that is what made a one-pass conversion safe.
   ⚠ The estate had **already** fixed the TRUNCATION shape of this class on
   2026-07-20 (`SafeCut`, `bugs_open/027` §4b) and never found the CASING shape.
2. **Half 1** — the store-failure branch takes the same `chrome_render_failed`
   surface as the execution-failure branch, with the same disposition.
3. **The three `summary[:247]` truncations** → `datahelpers.SafeCut`, in the same
   pass, exactly as the 08-31 addendum required.
4. **A pre-store rune-safety gate** — the door-closer. Postgres names the offending
   BYTE and never its POSITION, so a refusal on a 40 KB document says "0x80" and
   nothing about where. `datahelpers.InvalidUTF8At` reports the **offset** and a
   `QuoteToASCII`'d window, so the **next** byte-indexed slice introduced anywhere
   upstream is attributable in one read rather than by bisection. It **refuses**
   rather than sanitising: this path has no gate downstream, so `ToValidUTF8` would
   ship silently mangled text over working chrome AND leave the cutter in place.
5. **`emitChromeRenderFailedItem` gained a `phase`** — its operator-facing text said
   "the template could not be executed", which is false for two of its three callers
   now. Reusing a surface means fixing its prose.

Registered as **STY-059**. Five tests, each **mutation-proven red** 2026-09-02 (the
`UpperFirst` revert's failure output reproduces `ef bf bd 80 94` verbatim).

### ⚠ BEHAVIOUR CHANGE, and it lands on a live site

A slot whose store fails **and which has nothing stored to serve** now fails the
step, via the existing `bugs_open/260` caller logic. **garden-tools.uk's footer is
exactly that case**, so its next build fails where it previously reported success.
Held to be correct — a site must not go live with a missing footer, and a build
failing loudly beats another ten days of silence — but flagged to the council as
the edit most wanting an argument.

### Still open

- **The fix is INERT until the next chassis roll.** Both casualties stay broken
  until then; boxingonline.com's served footer remains the 16:05 hand patch.
- **Scoped out deliberately:** the sibling `no row matched` branch (~`:1357`) still
  returns a nil error. That is the row-locked-or-gone case, it has its own lock arm
  above it, and widening it is a different blast radius. Named so the next reader
  sees it was considered, not missed.
- **Verification after the roll** is unchanged from §"How to verify" above, plus:
  `rendered.footer=true` for **both** sites, and a re-run of the two-way census
  returning **zero** rows on the left column.


## COUNCIL ROUND 1 → REVISE, and it changed the code (2026-09-02)

Verdict on `dc62975f`: **REVISE**, gated by **guardian (HIGH)**. Worth recording in
full, because the gating objection was right and my round-1 rationale was the failure.

**I claimed a store refusal failing the step was "bugs_open/260's ruling applied
consistently". That was an ANALOGY, not a measurement.** The seat asked for the blast
radius instead. `[MEASURED 2026-09-02]` **7** live workflows dispatch
`render_site_components` — nav-updater, nav-link-fixer, rerender-chrome, rerender-site,
rerender-pages, pageflow-builder, site-work-orchestrator — and **every one declares no
`error_step`, no `on_error` and no `continue_on_error`.** A hard step failure has no
handler anywhere; it takes the whole orchestration.

**So the escalation is now OPT-IN** (`escalate_chrome_store_failure`, unsafe default
OFF — the 2026-08-02 §2 remedy and the shape of the three sibling keys in
`mistyped_llm_fields_gate.go`), distinguished by a `chromeStoreRefusedError` sentinel so
260's execution-failure authority is untouched. **REPORTING stays unconditional** — the
silence *is* the bug, and gating a fix for silence behind a flag nobody sets reproduces
it. The gate costs nothing today: neither casualty needs the escalation, because both
footers are repaired at source by `UpperFirst`.

**The objection that most improved the work was bug_historian's** — "was a census run for
OTHER nil-swallow sites?" It had not been. Run now: five `degraded, nil` returns; :1446 is
success, :1407 a lock refusal, :1222 the 342 refusal (which reports), :1241 an
empty-render Warn, and **exactly one genuine remaining swallow — :1411, "no row matched"**.

**It stays unfixed, and the measurement is why — it also killed my own instinct to widen.**
`[MEASURED 2026-09-02]` 57 sites exist, only **34** have any `site_components` row, and
**23** are missing at least one of the three slots. Routing :1411 into the reporting
surface would file **up to ~69** `needs_human_review` items about sites that have simply
never been built. The honest fix distinguishes "never built" from "the row vanished" and
is a different bug. **Measured-and-deferred, not missed.**

Also answered with evidence rather than argument: nothing anywhere parses a
`site_work_items` summary (so the phase change breaks no consumer);
`renderAndStoreSiteComponent` has exactly one production caller; no existing UTF-8 utility
reports a byte offset, so `InvalidUTF8At` is not dormant machinery rebuilt; and the
fleet-wide slot sweep is clean apart from the two known footers (header 34/34, head 34/34,
footer 32/34).

**Not delivered, stated plainly:** guardian also asked that the disposition change be
split from the byte-safety fix so one could be reverted without the other. Forward-only
forbids an amend and the code was already committed when the verdict landed. The config
key is the isolation actually available — and arguably the better one, since it isolates
the risky half at RUNTIME, per step, with no revert and no roll.

Round 2 resubmitted on the same correlation.


## COUNCIL ROUND 3 → APPROVED — and one advisory reversed the round-2 design (2026-09-02)

**APPROVED**, 2 advisories, **none high**. But acting on `bug_historian`'s medium
advisory undid what rounds 1–2 argued about, so it is recorded here rather than filed
as a tidy win.

**The opt-in escalation gate is DELETED.** The advisory: with the key OFF, a slot with
nothing ever stored that the UTF-8 gate refuses does not fail the build — and
`bugs_closed/054` already ruled that filing an item without failing the step is
insufficient escalation for an unserved chrome slot. Checking it made things worse than
the advisory said: **the arm could never gate "store failures" at all.** Escalation is
reached only through `chromeUnserved`, which is appended to only when
`!chromeSlotHasStoredHTML`. So the single thing my key gated was **a slot with nothing to
serve** — exactly the state `bugs_open/260` and `054` had already ruled must fail,
greenfield builds explicitly included.

**And the enumeration the guardian demanded argues against the gate, not for it.**
`[MEASURED 2026-09-02]` across the whole fleet **exactly ONE** `site_components` row has
NULL or empty `rendered_html` — garden-tools.uk's footer, which this change repairs. The
flag protected seven pipelines from a population of **one**.

> **The transferable error, and it is mine:** the seven-workflow figure bounds **WHO is
> affected when the arm fires**. It says nothing about **HOW OFTEN it can fire**. I
> answered the second question with the first question's number, and did it for two
> rounds while believing I had measured the blast radius. A count of consumers is not a
> count of occasions.

Deleting it also dissolved four advisories rather than deferring them (the undeclared key
failing open on a typo, its invisibility to the RFC_022 counter, the untested armed path,
and bug_historian's own) — a fair sign the simpler design was right.

Two seats moved me in opposite directions and the **measurement**, not the seniority of
the objection, decided between them. Round 4 resubmitted so the trailer describes the code
that exists.

---

## ✅ VERIFIED AT THE ARTEFACT — 2026-09-02 16:21Z, half 2 PROVEN END TO END

Live on `agent-chassis:v1.0.1354`. A `rerender-chrome` run for **garden-tools.uk**
(correlation `af0857d2-61c5-4cf6-ab82-c7b5001134ad`) settled it:

| check | before | after |
|---|---|---|
| `garden-tools.uk` footer `build_status` | `pending` since **2026-08-23** | **`rendered`** 16:21:32Z |
| `rendered_html` | **NULL** (never stored, 10 days) | **2,427 bytes** |
| `rendered_html_digest = md5(rendered_html)` | NULL | **true** |
| footers fleet-wide not `rendered` | **2** | **1** (boxingonline only — untouched by choice) |
| rows with NULL/empty `rendered_html` | **1** | **0** |

**And the content check, which is the one that actually proves the mechanism** — the very
label that caused the bug renders intact, em-dash preserved rather than mangled or dropped:

```html
<li><a href="/how-we-assess.html">How We Assess Garden Tools — Our Methodology | Garden Tools UK</a></li>
```

Header (2,532 B) and head (54,482 B) stored in the same run, so the render was real work,
not an idle no-op. **This is the close condition for half 2.**

**boxingonline.com remains `pending` deliberately** — a paid site mid-delivery whose served
footer is a hand-patch; re-rendering replaces it, and that is the owning lane's call, not
mine. `bugs_open/423` closes when they run it.


---

## COUNCIL ROUND 4 → REVISE, and the OWNER broke the tie (2026-09-02)

Round 4 (the gate deletion) returned **REVISE**, gated by **bug_historian HIGH**, arguing
the deletion makes a store failure "return a hard error UNCONDITIONALLY" on 7 workflows
with no `error_step`, reproducing `bugs_closed/073` — an honest signal cascading into
failing the whole workflow.

**The factual premise is wrong, and it is corrected here so the trail does not mislead.**
A store failure returns a non-nil error, but that error does **not** fail the step. Quoting
the deciding arm rather than the function:

```go
// :333-338 — the error lands in a map, and only an UNSERVED slot is escalated
chromeRenderFailed[slot] = renderErr.Error()
if !chromeSlotHasStoredHTML(ctx, params.DB, siteID, slot) {
    chromeUnserved = append(chromeUnserved, slot)
}
// :349 — and only THAT fails the step
if len(chromeUnserved) > 0 { return nil, fmt.Errorf(...) }
```

A slot with previous bytes is a **degraded success**; the site keeps serving. `[MEASURED
2026-09-02]` rows fleet-wide with NULL/empty `rendered_html`: **0**.

**What was RIGHT in the objection, and is recorded rather than dismissed:** when a slot
genuinely has nothing to serve, one bad slot does fail the whole run — there is no per-slot
granularity. That is real, it is `073`'s shape, and **it is `bugs_open/260`'s design, not
this change's** — this change routes a second failure kind into a disposition 260 already
had reviewed and approved.

### ⚖ OWNER RULING 2026-09-02 — UNGATED, and the deletion stands

Three seats had now pushed in three directions across four rounds (guardian: do not
escalate · bug_historian r3: the gate withholds an approved protection · bug_historian r4:
the deletion cascades). CLAUDE.md's rule for exactly this — *"especially when seats
disagree with each other … let a human break it"* — was applied instead of a fifth round.

**The owner chose UNGATED.** A store refusal on a slot with nothing to serve fails the
step, exactly as an execution failure has since 260. Already carried by `badff59a9`; ships
on the next roll. **Not resubmitted for a round 5** — the council is advisory, it cannot
overrule the owner, and a fifth round would spend credits to be told something already
decided.

Two round-4 objections were craft rather than disposition and are accepted without action,
stated so they are not mistaken for oversights: *editquality* was right that the deleted
test file should have been a visible edit rather than a `grounded_in` assertion (my third
instance of that exact fault in this trail), and *prior_art* was right that it cannot
verify a `site_components` population count because that table is outside its accessible
schema — so that measurement is owner-verifiable, not council-verifiable, and is quoted
with its query wherever it is used.

---

## ✅ CLOSED 2026-09-02 — fixed, live, and BOTH casualties repaired at the artefact

Live on `agent-chassis:v1.0.1354` (probed at the binary with a removed-string control).

| | before | after |
|---|---|---|
| `garden-tools.uk` footer | NULL since **2026-08-23** | `rendered`, 2,427 B, digest ok, **16:21:32Z** |
| `boxingonline.com` footer | `pending`, hand-patch serving | `rendered`, 2,289 B, digest ok, **16:27:56Z** |
| footers fleet-wide not `rendered` | 2 | **0** |
| rows with NULL/empty `rendered_html` | 1 | **0** |

**The webdesign lane's own pre-delivery probe passes on boxingonline** — `sites.email` is
empty and the served footer carries **no mailto and no contact block**, gated at
`component_library.go:1988`, which was that lane's stated close condition.

And the content check that proves the mechanism rather than the absence of a crash: the
offending label renders **intact**, em-dash preserved —
`How We Assess Garden Tools — Our Methodology | Garden Tools UK`.

**Residual, tracked elsewhere, not left implicit:** `bugs_open/435` (the `:1411` no-row-
matched swallow, deferred on a measurement) and the follow-up to declare `ConfigKeys` for
`render_site_components` so the RFC_022 counter can see its three undeclared keys.
