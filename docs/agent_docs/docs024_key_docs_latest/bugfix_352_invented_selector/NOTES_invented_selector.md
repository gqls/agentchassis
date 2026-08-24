# NOTES — bugs_open/352, the invented selector

Append-only, newest at the bottom. Technical log: evidence, commands, what the system actually
said, and every misstep. The missteps are the point, not an appendix.

---

## 2026-08-24 — session opens, lane created

Picked up `bugs_open/352` cold. It was filed 2026-08-22 by the `bugfix_198_roundtrip_writers`
lane as candidate (6) spun out of 198 at close-out.

### Ownership — resolved by ASKING, not by the tool

`scripts/who-owns.py 352` returned **"OWNED or recently active … Do not start a competing fix."**
That verdict is **wrong**, and the reason is structural: the only evidence it has is commit
`6475c3eea`, which is the commit that **FILED** 352, plus three cross-references in the filing
lane's own directory. **A filing commit is indistinguishable from an ownership commit to a tool
that reads `git log`.**

Settled by messaging the `bugs_open/198` session directly. Its reply: *"352 is yours. Not working
it, not planning to."* It also confirmed first-hand that both files I intended to touch were
clean in the tree, and volunteered the same caveat about its own who-owns verdict independently.
→ Logged in `WRONG_CALLS.md`.

### The bug is still valid at source [MEASURED 2026-08-24]

`internal/adapters/browserrunner/render_audit_action.go:202` — unchanged since filing:

```js
var cls=(typeof el.className==='string'?el.className:'')||el.tagName;
```

and `:310` still files that value as `Class:`. Last commits to the file are `b32aa9cd9` (the 131
lane's `contrast_ratio` check) and earlier; **nothing has touched the fallback.**

### The thing the bug file gets slightly wrong, in our favour

`contrastSelector` (`write_render_audit_findings_action.go:780`) **already handles an empty class
correctly**:

```go
tokens := strings.Fields(class)
if len(tokens) > 0 { return tag + "." + tokens[0] }
if tag != "" { return tag }
return "*"
```

So the consumer is not the defect and needs no new branch. The producer simply never sends an
empty class. This makes the minimal producer fix genuinely small — which is what made me look
harder for what the small fix *breaks*, below.

### A correct precedent exists IN THE SAME PACKAGE

`internal/adapters/browserrunner/contrast_check.go:86`, `describe`:

```js
var describe = function(el){ return el.tagName.toLowerCase()
  + (el.id ? '#' + el.id : '')
  + (el.className && typeof el.className === 'string' && el.className.trim()
     ? '.' + el.className.trim().split(/\s+/).join('.') : ''); };
```

No fallback, adds `#id`, joins **all** class tokens. So the render audit's version is the odd one
out in its own package, not a house style. That is the shape to converge on.

Four sites carry a `tagName` fallback (`grep -n tagName internal/adapters/browserrunner/*.go
scripts/render_audit.py`):

| site | verdict |
|---|---|
| `render_audit_action.go:202` | **the bug** — feeds a filed work item |
| `scripts/render_audit.py:139` | same defect, local probe only — but it MISLED `bugs_open/211` (below) |
| `contrast_check.go:86` | correct, no fallback — the model |
| `run_checks_action.go:1140` | names a **component**, not a selector — out of scope |

### Live damage [MEASURED 2026-08-24, `site_work_items`]

Of **452** `contrast_failure` rows, **181** carry a selector matching `^([A-Za-z0-9]+)\.\1$`:

| status | rows | of which `X.X` |
|---|---|---|
| `complete` | 280 | **108** |
| `deferred` | 145 | 58 |
| `unresolved` | 26 | 15 |
| `cancelled` | 1 | 0 |

Commonest: `P.P` ×77, `A.A` ×44, `H2.H2` ×16, `H3.H3` ×16, `LEGEND.LEGEND` ×7, `H1.H1` ×6, then
`EM/LABEL/BUTTON/STRONG/SPAN/CODE/TH/H4`. First seen **2026-08-10**, last **2026-08-24**, with
**92 filed in the last 7 days** — so this is actively producing, not a historical artefact.

**108 rows marked `complete` were "fixed" with a rule that selects nothing.**

### ⚠ THE THING I DID NOT EXPECT — the naive fix is a REGRESSION, not a fix

Today `p.P { color:#fff }` matches nothing, so it is **inert and harmless**. Correct the selector
to `p` and css-patch-agent appends `p { color:#fff }` to the **site** stylesheet — recolouring
every paragraph on the site. The two commonest fallbacks are `P.P` (77) and `A.A` (44), i.e.
**precisely the two most dangerous bare selectors available.**

So "omit the class component so the selector is `h3`" — the bug file's own candidate (1) — is
right about the producer and **incomplete about the consequence**. The fix must emit a *scoped*
selector, not merely a lowercase one. This is the single most important finding of the session and
it is the reason this is not a one-line change.

### The second unexpected hazard — the key shape change FALSELY closes live rows

`item_key` is `workItemKey("contrast_failure", pagePath+"#"+selector)` (`:327`), and
`retractResolvedContrastFindings` (`:529`) builds its `stillFailing` set with the **same**
`contrastSelector`. Change the producer and a legacy row keyed `…#H3.H3` is absent from a
new-shape `stillFailing`, its page *was* audited, so it is retracted as **resolved** and stamped
*"render audit re-measured %s and this pairing is no longer below its contrast threshold"* —
which is **false**. That is the exact false completion the function's own header says it exists to
prevent.

Exposure [MEASURED]: `workItemClosedStatuses` = complete/verified/rejected/wont_fix/cancelled
(`work_items_common.go:85-91`), so `deferred` (58) and `unresolved` (15) are **not settled** and
**are** retraction candidates — **73** rows, including part of the 226-row migration-389 park.

> **CORRECTED 2026-08-24, same session:** I first wrote this as "~84 rows". It is **73**. The
> error was pure arithmetic — I added the 26 rows that are `unresolved` *in total* instead of the
> 15 that are `unresolved` **and** `X.X`, having just printed both numbers in the same table.
> **Caught by the `bugs_open/198` session**, which reproduced all four of my figures against the
> live DB before entering them in its own record and checked the addition I had not. Nothing
> downstream had used the wrong figure yet. → `WRONG_CALLS.md`.

**And the bigger population is the one I nearly under-weighted** (the 198 session's framing, and
it is right): of the 181, **108 are already `complete`** — the false completion has *already*
happened on three-fifths of them, against **73** that are merely *at risk* from a key-shape
change. 198's own file recorded exactly one instance (the dartsonline `H3` row, `bugs_closed/
198_…md:562-571`, `complete` on 2026-08-18); it generalises to 108 fleet-wide.

One asymmetry not to "tidy" while in there: `unresolved` **is** in `workItemTerminalStatuses`
(`:42-48`) but **not** in the closed set, and that is deliberate and documented at `:97`. It has a
real consequence for the rekey migration: a terminal row does **not** hold the dedup slot, so an
`unresolved` row cannot collide, while a `deferred` row **can**.

**Ordering alone cannot fix this**, and that is worth stating because it is the obvious wrong
answer: rekey first and the still-running OLD binary falsely retracts the rekeyed rows instead.
The window is symmetric, so the transition needs shape-tolerance in the code, not a careful
sequence.

### A third mechanism, stated WITHOUT overclaiming

`htmlCorpusContainsClass` (`:691`) is a raw `strings.Contains` over locked components' HTML, fed
the tag name today. For a class-less `<p>` it searches locked markup for `"P"` — one uppercase
letter — so a class-less finding *can* be silently dropped as `skipped_locked`.

**[UNPROVEN] I have NOT established that this ever dropped a real finding.** The counter-evidence
is in the data: the 6 sites that have locked components hold 108 contrast items of which 51 *are*
`X.X`, so the drop is plainly not universal. The mechanism permits it; that is all I claim.

The direction that *is* certain is the opposite one: once the class is truthfully empty,
`strings.Fields("")` is empty and the locked check becomes a **no-op** for every class-less
finding. So the fix **opens** a hole — a class-less element inside a locked component would now be
filed and patched, and a lock is a human's "hands off".

### Two operational facts that change how this ships [MEASURED 2026-08-24]

1. **The fix spans TWO images, and the chassis is only half of it.** `internal/adapters/
   browserrunner` compiles into `cmd/browser-runner-adapter` and **nothing else** (checked every
   `cmd/*` for the import). `platform/orchestration/actions/*` is the chassis. `render-audit-
   adapter` runs the **browser-runner image** (makefile:107). So a chassis roll ships the
   consumer half and **not the producer fix**. All three overlays sit at `v1.0.1332`;
   `IMAGE_TAG` is `v1.0.1333`.
2. **The council gate is WORKING again.** The 131 lane's handoff (2026-08-22) says it is down —
   `claude-sonnet-5` capped until 2026-09-01, all 17 seats on that model. **That is now stale:**
   47 `fix_correlation_id` runs COMPLETED in the last 3 days, and today alone produced
   `complete_approved` at 13:01, 13:00, 11:55 and 11:30 plus two `complete_revise`. So submitting
   is available and the 131 lane's blocked round-3 resubmit is now unblocked — told them.

---

## 2026-08-24 (later) — three things the LANDMINES grep gave me that I would not have found

Grepped `LANDMINES.md` for my symbols before writing any code (the SessionStart hook only matches
files already DIRTY, so a shared helper is never shown). Four entries sit on this footprint. Two
changed the design.

### (a) The retraction path is NOT theoretical — it has already fired 79 times

LANDMINES entry *"A PARKED (`deferred`) work item is retractable"* names my exact hazard from the
other end, and made me measure the path's actual traffic rather than reason about it:

```sql
SELECT status, count(*) AS n, count(batch_id) AS with_batch_id,
       count(*) FILTER (WHERE result ? 'resolved_at') AS retracted
FROM site_work_items WHERE item_type='contrast_failure' GROUP BY status ORDER BY status;
```

| status | n | with batch_id | retracted |
|---|---|---|---|
| `cancelled` | 1 | 0 | 0 |
| `complete` | 280 | 200 | **79** |
| `deferred` | 145 | **0** | 0 |
| `unresolved` | 26 | 26 | 0 |

**79 contrast rows have been closed by a retraction.** This is live, exercised machinery — so the
false-retraction hazard is a change to a path that demonstrably fires, not a latent one. **This
measurement could have come out zero** (the path never fires, hazard is theoretical) and it did
not, which is the only reason it is worth writing down.

### (b) That landmine's own figure is now STALE, and the staleness is by ADDITION

It records **`contrast_failure` 0 of 226 rows carried a `batch_id`** [MEASURED 2026-08-12]. Today
**226 of 452 carry one** — the filer sets `batchID` on the item it builds
(`write_render_audit_findings_action.go:~331`), so every row filed since carries one, and only the
older population does not. Exactly the failure mode CLAUDE.md's dated-count rule exists for: the
census was right when taken and wrong by arrival.

**But the entry's conclusion still holds where it matters, and this is the sharp bit:** the
**145 `deferred` rows — the migration-389 park, the population my transition endangers — carry 0
batch_ids**, so for precisely those rows `resolveWorkItems`' self-protection guard
(`batch_id IS DISTINCT FROM $6`, and `NULL IS DISTINCT FROM <uuid>` is TRUE) is **inoperative**.
The overall figure improved; the guard's coverage of the at-risk set did not. → I owe LANDMINES a
dated correction that says both halves, because "226 of 452 now carry one" read alone would retire
a live trap.

### (c) The wire contract is MIRRORED, not imported — so the fix is four structs, not one

`platform/` does not depend on `internal/adapters`, so `renderAuditContrast`
(`write_render_audit_findings_action.go:133`) is a hand-kept copy of `ContrastFinding`
(`render_audit_action.go:83`), and the JSON tags are the coupling point. Adding a field means all
of: the in-page `out.contrast.push({…})`, the adapter's anonymous `pageAudit.Contrast` struct, the
adapter's exported `ContrastFinding`, and the orchestrator's `renderAuditContrast`. Miss the last
one and the field arrives, unmarshals into nothing, and reads as absent — **with no error**, which
on the version-skew branch is indistinguishable from an un-rolled adapter.

This is also why the old-shape reply must stay *inert rather than wrong*: the two halves are
separate images (see §6 of the RUNBOOK) and will be un-rolled relative to each other for a window.

### (d) Two more entries on this footprint, noted so I do not measure into them

- **`render_audit.py`'s total understates ~100× without `--sitemap`** (`dartsonline.com` 1 → 125),
  and **~8.5% of rows are the probe's own `rgb(128,128,128)` guess** (`overImage`). If I quote any
  per-site contrast figure to prove this fix worked, it must say whether `--sitemap` was passed and
  must discount `overImage`. The filer already excludes `over_image` from filing, so the *work
  item* population is clean — the trap is in the probe output I might paste as evidence.
- **A `090` run on a symbol in a file over ~60 KB returns bundles and no verdict**, and that looks
  exactly like a run still in progress. `render_audit_action.go` is over that. So no 090 on this
  file — which is consistent with 352 already having stated its 090 substitution.

---

## 2026-08-24 (implementation) — what the council asked that I had not, and the guard that guarded nothing

### The misstep worth reading: my test matched my own COMMENT

I wrote a confinement assertion pinning the legacy `cls` echo —
`strings.Contains(newRegion, "||el.tagName")` — then mutated the echo line out to prove the guard
fired. **It passed.** The line I was protecting has an explanatory comment directly above it saying
*"do not 'clean up' the `||el.tagName` fallback"*, **which I had written myself minutes earlier to
protect that very line.** So the source contained the needle twice — once in prose, once in code —
and the prose occurrence survives deleting the code. **The better my comment, the more reliably the
test lied.**

Fixed by asserting the whole expression, which no comment would ever spell out. The general check:
`grep -c` the needle before trusting a source-scanning assertion — **if the count is above one, the
second hit is usually your own explanation of why the first one matters.**

The part that nearly hid it: the other two mutations (removing the in-page verification, removing
`selector:sel`) failed correctly, so the set *looked* mutation-proven. **One vacuous guard in a set
of three is invisible if you report the set.** → `WRONG_CALLS.md`.

### Mutation results, in full — all seven guards

| mutation | test that must fail | result |
|---|---|---|
| in-page verification → `true` | `TestAuditJSComposition` | FAIL ✓ |
| legacy `cls` echo removed | `TestAuditJSComposition` | FAIL ✓ (only after the fix above) |
| `selector:sel` dropped from the push | `TestAuditJSComposition` | FAIL ✓ |
| legacy alias key dropped from `stillFailing` | `…LegacyKeyedRowSurvivesStillFailingClasslessFinding` | FAIL ✓ |
| scheme skew guard dropped | `…SchemeStampedRowNotRetractedByOldShapeReply` | FAIL ✓ |
| `selectorLockTokens` reverted to `c.Class` | `…LockedAnchorSkips` | FAIL ✓ |
| bare-tag refusal dropped | `…UnanchoredBareTagIsSkippedAndCounted` | FAIL ✓ |

⚠ A `sed` mutation with `|` as the delimiter **silently did not apply** against `||el.tagName`, and
the resulting "ok" read exactly like a passing guard. Mutations are now applied with a python
snippet that **asserts the anchor was found and the text actually changed** before running the test.
A mutation you did not apply is a guard you did not test.

### Two council objections answered by measurement, one of which corrected me

- **`bug_historian`, medium — "the plan ASSUMES an old chassis drops unknown JSON keys."** Fair: I
  had asserted it. Now verified — `write_render_audit_findings_action.go:785` is a plain
  `json.Unmarshal` (lenient), and the tree's only three `DisallowUnknownFields` calls are in
  `provocation_gate_action.go:549`, `provocation_generator_action.go:238` and
  `internal/tools-api/gripper/prompt.go:146`, none on this path.
- **`prior_art_librarian`, medium — "no evidence the audit runs on a live schedule."** This one
  **changed a number I had published.** I had checked that audits *run* (8 distinct days in the last
  21, 67 rows today) and let that stand in for *per-site coverage*, which is a different question.
  Measured properly: of the 13 affected sites, **all 13** audited within 14 days, **3** within 7,
  **2** within 3, oldest last-audit 2026-08-10. So the honest window is **a fortnight**, not "the
  next weekly audit" as I had written in three places. Corrected in the plan, the migration and the
  bug-file banner. → the seat asked for the check I had skipped, not for reassurance.

### The migration's premise, measured with a control

Before withdrawing 73 rows on the grounds that their selectors match nothing, I tested whether any
affected site *genuinely* uses `class="H3"` on an `<h3>` — 352's own candidate-4 caution, and the
one theoretical false positive.

- **0 of 31** (site, tag-token) pairs have that class anywhere in `page_components` +
  `site_components` rendered_html.
- **POSITIVE CONTROL, because a zero with no control could not have come out otherwise:** the same
  predicate against **real** class tokens from non-`X.X` findings finds **154 of 161**.

Both arms are in `587_..._VERIFY.sql` so the next reader re-runs them rather than trusting this
paragraph, and the VERIFY file has been **executed against the live DB**, not merely written: 66
distinct keys listed, `open_invented=73` matching the census exactly, `falsely_completed=0` as the
pre-ship baseline.

### Things that went wrong operationally, recorded so the RUNBOOK stays true

- **Migration 586 was taken by another lane between my writing the submission and my writing the
  file.** Shipped as **587**. The RUNBOOK already said to re-check the number immediately before
  writing; it went stale inside a single session, which is a sharper version of the same warning.
- **`go test` failed mid-session on `refresh_evidence_fact_drift.go:608`** — another session's
  uncommitted WIP arriving in the shared tree between my build and my test. Settled the right way:
  `scripts/verify-head-builds.sh ./cmd/agent-chassis/... ./cmd/browser-runner-adapter/...` →
  **OK**, so HEAD is clean and the breakage was never mine. ⚠ That script takes **package paths**;
  bare service names give `package agent-chassis is not in std`, which reads like a build failure
  and is a usage error.
- **`doc_notes.categories` is `jsonb`, not `text[]`**, and `subject_type`/`subject_key` are NOT
  NULL. My first INSERT used an `ARRAY[...]` literal and would have failed at apply time. Schema
  first, every time — the shape came from live rows in the end, not from memory.

---

## 2026-08-24 (post-roll) — BOTH halves are live, and my first negative control was worthless

### The roll [MEASURED 2026-08-24 ~16:50 UTC]

`v1.0.1334`, all three overlays. Pods started **15:39 UTC**:

| service | evidence | verdict |
|---|---|---|
| `browser-runner-adapter` | own startup line, `git_commit 70fd163c2` | **producer half LIVE** |
| `render-audit-adapter` | same line, same commit (shares the image, makefile:107) | **LIVE** |
| `agent-chassis` | startup line SCROLLED (busy service, exactly as the landmine warns); binary probe found `70fd163c2` | **consumer half LIVE** |

`git merge-base --is-ancestor ffa6e1c3d 70fd163c2` → **YES**. My fix (13:45 UTC) predates the build
commit (15:11 UTC), which predates the pod start (15:39 UTC). All three clocks agree.

**Capability probed, not just the commit** — the stronger check, since a commit being an ancestor
does not prove the code path shipped:

- chassis: `skipped_unverified_selector` ✓, `skipped_unanchored_selector` ✓, `selector_scheme` ✓
- browser-runner: `verified/v1` ✓, `indexOf.call(nodes,el)` ✓, `selectorVerified` ✓
- both: a deliberately invented string absent ✓ (so the probe is not over-matching)

**So migration 587's ordering gate is SATISFIED.** Both images confirmed at the artefact.

### ⚠ MY FIRST NEGATIVE CONTROL COULD NOT HAVE FAILED, AND I ALMOST BANKED IT

I ran the ancestor check with what I called a negative control — `ffdca67fd`, "a commit made AFTER,
must NOT be an ancestor". It returned **YES**, which I first read as the control failing.

It was not the control failing. **`ffdca67fd` was committed at 14:08 UTC and the build commit at
15:11 UTC, so it genuinely IS an ancestor.** I had picked a control on the assumption that a commit
made later in my *session* was made later than the *build*, and the build had happened in between.
The control was not discriminating; it was simply mis-chosen, and had it happened to return NO I
would have recorded a passing control and moved on with no idea it proved nothing.

Redone with a real one: **HEAD** (16:47 UTC, genuinely after the build) → correctly **NOT** an
ancestor. Only then is the positive result worth anything.

**The lesson is the family this lane keeps hitting** — *a measurement that could not have come out
otherwise is not evidence*. Here it wears a new coat: **the control's validity depended on a
timestamp I never looked at.** Choosing a control is itself a measurement, and mine was an
assumption. The cheap check is one line: print the timestamps of the control, the subject and the
build side by side **before** interpreting either result. That is what I did the second time and it
took seconds.

### The canary, dispatched

No audit had run since the roll (`rows_since_roll = 0`), which is expected on a ~fortnightly
rotation — so the proof has to be driven, not waited for.

Dispatched a render audit on **ai-agent-orchestration.com** (`2a8ebf9c-…`), correlation
**`c2fce02e-2fe7-489f-bc22-edcfa75b0761`**, via `kafka_publish_checked` (a receipt, not a hopeful
`kcat -P`). Chosen deliberately: it is **`bugs_open/211`'s own site**, the one whose `[UNRESOLVED]`
§4 item is written around the six class-less `<h3>`s, and it holds 8 invented rows. Queue checked
first — no render-audit work in flight there.

Expect ~30 min publish→start under fleet load. **587 is NOT applied and should not be until this
canary proves the producer files a verified, anchored selector** — the gate in the file is "both
images rolled", which is now met, but seeing one good row first costs nothing and is the difference
between an argument and a measurement.
