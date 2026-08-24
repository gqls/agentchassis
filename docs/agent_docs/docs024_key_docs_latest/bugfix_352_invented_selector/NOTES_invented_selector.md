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

---

## 2026-08-24 (evening) — the canary never ran, the rotation proved it anyway, and 587 is APPLIED

### The canary was dead and it did not matter

The correlation the last session dispatched at ~16:52 UTC — `c2fce02e-2fe7-489f-bc22-edcfa75b0761`,
site `2a8ebf9c-…` — has **no orchestration row of any kind**, checked by correlation, by site_id and
by a substring sweep of `initial_request_data` and `collected_data`, **2 h 15 min** after publish.
The handoff's warning ("a missing row is LATENCY, not a dropped dispatch, ~29 min under load") is
sound and I did **not** re-dispatch, for a better reason than patience: by then the proof already
existed and a second audit would only have cost credits.

Context that makes the dead canary unsurprising rather than alarming: `render-audit-agent`'s only
run **ever** against that site was `16781a84-…` at 02:23 UTC today, and it ended `complete_error`.

### What replaced it — a before/after pair nobody staged

`render-audit-agent` runs roughly **hourly** across the estate (13:28, 14:29, 15:29, 16:30, 17:30,
18:31 today), cycling sites; *per site* it is the ~fortnightly rotation this lane measured. Two of
those runs straddle the 15:39 UTC roll, and that is the experiment:

| | run `6dc00a26`, **15:29→15:31:50** (old image) | run `e0bd33d0`, **17:30→17:33:16** (new image) |
|---|---|---|
| site | `55213ded` (loanzy.uk) | `ee4a8199` (loancash.co.uk) |
| rows filed | 47 | 10 |
| invented `TAG.TAG` | **3** | **0** |
| `spec.selector_scheme` | absent, all 47 | `verified/v1`, all 10 |
| `spec.matches` | absent | present, all 10 |

⚠ **Different sites, so this is not a controlled A/B** — it is the same producer and the same code
path two hours apart, and the pre-roll arm is what makes the post-roll zero mean something. A
`still_invented = 0` on its own could not have come out otherwise if no class-less element happened
to be measured; the post-roll selectors are `.ported-page-content A` — an ancestor anchor with a
bare-tag leaf, i.e. **exactly** the class-less case the old code turned into `A.A`. The new path
fired on the very population at issue.

### Then settled in the page, because the producer's own `matches` cannot vouch for the producer

Quoting `spec.matches` back as proof is circular. So: fetched the live pages over HTTPS and counted
independently, with a stdlib `HTMLParser` that tracks the open-element stack (so "descendant" means
descendant, not "appears later in the file"). Script kept at
`scratchpad/sel_check.py` for the next reader; it is not a CSS engine, it handles the one
`.ANCESTOR TAG` shape this producer emits.

⚠ **Domain control first** (the parked-domain landmine): an invented path on `loancash.co.uk` and on
`loanzy.uk` both returned **404** with a different body, so a 200 on the real page is a real page.

| selector | page | producer said | I measured | class-less among them |
|---|---|---|---|---|
| `.ported-page-content A` | `loancash.co.uk/guides/index.html` | 15 | **15** | 15 of 15 |
| `.ported-page-content A` | `loancash.co.uk/guides/jargon-buster.html` | 8 | **8** | 8 of 8 |
| `SPAN.SPAN` (pre-roll) | `loanzy.uk/tools/loan-repayment-calculator/` | — | **0** | 22 real `<span>`s exist |
| `LABEL.LABEL` (pre-roll) | `loanzy.uk/tools/loan-comparison-calculator/` | — | **0** | 6 real `<label>`s exist |

Parser controls, all discriminating: a class that does not exist → 0; the same selector against the
404 body → 0; `class="A"` and `class="H3"` occur **nowhere** in the markup.

Those two pre-roll rows are both already `complete`. **Two more false repairs were recorded today,
eight minutes before the fix rolled** — which is why the damage figure moved (below).

### A figure in my predecessor's handoff carried the wrong date, and the arithmetic pins it

The handoff's §5 table says `contrast_failure` total **452** `[MEASURED 2026-08-24 ~16:55 UTC]`.
The total now is **509**, and only **10** rows have been created since 15:39. That does not add up
until you notice the 47 rows the 15:31:50 audit filed: **509 − 10 − 47 = 452**. So 452 was the total
as of *before* 15:31:50, and by 16:55 it was already 499.

Nothing was measured wrongly. **The label carried the time the sentence was written, not the time
the query was run**, and a `[MEASURED <date>]` marker makes that indistinguishable from a fresh
figure. This lane's own rule — mark the unverified ones too — does not cover it, because the claim
*was* verified, just earlier than it says. → `WRONG_CALLS.md`. **Date a figure when you measure it.**

### ⚠ The RUNBOOK's own §2(b) `sites` column is not the number this lane quotes

`count(DISTINCT site_id)` in that query ranges over the **whole open still-failing population**, not
over the invented subset in the `FILTER` beside it. Run today it returns `open 181 / invented 73 /
sites **16**` — and this lane has always quoted **13**, which is the invented subset's site count and
needs its own query. Two counts in one row, only one of them filtered, and the column name says
neither. Fixed in the RUNBOOK.

### Fresh census immediately before applying 587 [MEASURED 2026-08-24 19:10 UTC]

| status | rows | of which invented |
|---|---|---|
| complete | 327 | **111** |
| deferred | 145 | 58 |
| unresolved | 26 | 15 |
| needs_human_review | 10 | 0 |
| cancelled | 1 | 0 |

**111, not 108** — the three the 15:31 audit filed and closed. The 73 open (58 + 15, 13 sites) is
unchanged and matched 587's premise exactly.

### 587 applied

`_VERIFY` arms 1–3 re-run read-only first, all fresh, all passing: 66 distinct keys, **every one**
of the form `<path>#TAG.TAG` with identical uppercase halves (eyeballed, not sampled); false-positive
arm **0 of 31** tag tokens present as real classes with the positive control at **166 of 173** (it
was 154/161 in the morning — the control still discriminates).

```
NOTICE:  587: withdrawing 73 open invented-selector contrast_failure row(s)
BEGIN / DO / UPDATE 73 / INSERT 0 1 / COMMIT
```

Applied **2026-08-24 19:11:22 UTC**. Arms 4–5 after: `open_invented = 0`, `withdrawn = 73`,
`withdrawn_without_prior_status = 0`, `falsely_completed = 0`. The recovery query returns
`deferred 58 / unresolved 15`, 13 sites — the figure that keeps returning 73 for ever.

### Two things found on the way that are NOT this lane's, recorded so they are not lost

- **The render audit times out more often than it succeeds.** [MEASURED 2026-08-24 19:08 UTC]
  Over 7 days, `render-audit-agent`: **11 of 20 pre-roll runs** ended `complete_error`, all on
  `{"message": "Request timed out (code: TIMEOUT)", "failed_step": "audit"}` at almost exactly
  **3 minutes**. That is a **55% failure rate before this lane touched anything**, and it is the
  clock on 587's re-detection window. Not filed as a bug here — it wants its own grep of
  `/bugs_open/` first.
- **No evidence either way that my change affected that rate, and the sample cannot tell.**
  Post-roll: **2 of 3** runs errored. Three runs cannot distinguish 55% from 67%; naming the
  detectable effect size is the honest version of "no regression observed". Re-check after ~20
  post-roll runs (≈ a day at the current cadence).

### The fleet re-rolled underneath me at 18:32 and the fix survived — checked, not assumed

Chassis pods restarted **18:32 UTC** on `v1.0.1335`, built from `48f55f218…` (19:01:51 BST). Which
is why the 17:30 run's `write_render_audit_findings: complete` log line — and its two new counters —
were already gone when I looked: **605 lines of chassis log in five hours is a restarted pod, not a
quiet service.**

Timestamps printed side by side **before** interpreting anything, which is the correction this lane
wrote up this morning:

```
ffa6e1c3d (fix)          2026-08-24T14:45:03+01:00
48f55f218 (v1.0.1335)    2026-08-24T19:01:51+01:00
HEAD                     2026-08-24T20:12:28+01:00
```

`merge-base --is-ancestor ffa6e1c3d 48f55f218` → **YES**; control `HEAD` → correctly **NOT** an
ancestor. Capability probed on the running chassis binary too: `skipped_unverified_selector`,
`skipped_unanchored_selector`, `selector_scheme` all **present**, an invented control string
**absent**, the build sha **present** and a nonsense sha **absent**.

### ⚠ Misstep: I ran the binary probe BEFORE grepping LANDMINES for the instrument I was using

The probe above (`kubectl exec … grep -aq "<literal>" /proc/1/exe`) is the estate's prescribed
check, and I ran it with the prescribed discipline — a present control and an absent control in the
same breath. **Then**, while correcting a different landmine entry, I found this one, added earlier
today by the `bugfix_283_component_instance_scope` lane:

> **BusyBox `grep` over `/proc/1/exe` reports FALSE ABSENCES — and your present/absent controls PASS
> while it does it.** The fleet images are BusyBox v1.37; its grep is line-oriented and a "line" of a
> Go binary can be enormous, so a literal inside an over-long line reads as NOT FOUND, exit 1, no
> error. Measured on **`agent-chassis-855587d4dc-pn2t8`** — a pod of the same replicaset I probed.

**My result survives, but not because I was careful — because of which way the instrument fails.**
Everything load-bearing in my probe was a **presence** (three capability symbols, the build sha);
the only absences were nonsense controls whose absence is true either way. A false-*presence* would
have broken the conclusion and that is not the described failure mode. Note what that means about my
controls, though: **on this instrument a nonsense-string control cannot discriminate**, because it
reads ABSENT whether it is genuinely absent or merely inside a long line. I recorded two passing
controls that could not have failed — the exact family this lane wrote up this morning about a
mis-chosen commit, one rung along.

Re-run on the instrument the entry prescribes (NUL-split so every line is short, **both** controls
through the same pipeline), and the result set is coherent:

```
tr '\0' '\n' < /proc/1/exe | grep -Fc '<literal>'      # on agent-chassis-855587d4dc-h4hcg
skipped_unverified_selector   1     InstanceID (positive control)          25
skipped_unanchored_selector   1     ZZZ_not_a_real_symbol_352_control       0
selector_scheme               4     0123…4567 (nonsense sha)                0
48f55f21834ac3e2d95aa43716f6e63e40ac12ee   3
```

**The transferable bit is the ordering, not the grep.** MEMORY says *grep LANDMINES for the SYMBOL
you are about to trust* — and the SessionStart hook cannot help here, because it matches entries
against files **already dirty in the tree**, and `/proc/1/exe` is not a file in the tree. An
instrument has a footprint too. I grepped the entry an hour after I had already banked its answer.

### A peer's narrow question turned up a trap neither lane was looking for (`bugs_open/384`, 19:30 UTC)

The 384 lane sent an FYI: its Phase 2 adds a discovery check `page_list_stale` that files
`page_rerender` / `spec.reason='section_data_resolved'` keyed with
`PageRerenderItemKey(name, site, 'section_data_resolved')`, closing its own rows via
`CheckResult.Resolved`. It framed this as "a detector filing a remedy — exactly the class you own"
and asked only whether it conflicts with 352.

**It does not, and the reason is worth stating rather than waving through.** 352's failure mode is a
remedy that is *inert*: the instruction was an address (a CSS selector) naming nothing, so the fix
was authored, deployed and completed while touching nothing. Their remedy is "rerender this page" —
**no address to get wrong**, so 352's invariant is satisfied by construction. Different `item_type`,
disjoint key namespace, no shared code path.

**But answering it properly meant reading the index rather than recalling it, and that found two
things.** ⚠ Neither is in 384's code — their check is not in the tree yet, so everything below is
measured off the live DB and `pg_indexes`, not off their source.

1. **They are the FOURTH producer on `section_data_resolved`, not the second.** [MEASURED
   2026-08-24 19:30 UTC] `render_news_section` 182 · `bugfix-238-contact-block-2026-08-21` 6 ·
   `rerender-pages` 3. That triggers the owner ruling of 2026-08-02 §1: converging producers onto
   one `item_type`/`item_key` needs no RFC **provided** the producer set and the key shape are
   stated in the register entry. Told them, and told them to date the count — this lane published
   "108" this morning and it was 111 by evening for exactly that reason.

2. **`idx_swi_dedup` is `(site_id, item_key)` — `item_type` is NOT a column**, so the key space is
   global across every type on a site. Every existing LANDMINES entry about this index discusses
   its *status* predicate; **none mentions its columns**, and the natural mental model ("it dedups
   within my type") is wrong. Measured: `item_key LIKE 'page_rerender%'` is already carried by two
   types (4,809 `page_rerender` + 150 `needs_page`), four prefixes estate-wide are shared by two
   types, and **zero** `(site_id, item_key)` pairs have ever carried two — the suffixes have always
   differed. **A clean history is not a guard; it is a coincidence that has held.** Filed as a
   landmine with the estate-wide invariant query as its keepable form (it returns nothing today, so
   a row appearing in it *is* the incident, already live, with no other symptom).

**Verifier armed:** `./scripts/landmines-verify-dispatch.sh` dispatched correlation
`97e80561-1094-4822-aa70-a22122256c33` for the new entry. ⚠ Only **2** entries needed verification,
not the eight I had braced for — the other lanes' recent entries had already been synced by their
own lanes, so my earlier worry about spending their dispatches was unfounded. ⚠ And read any
`NEEDS_HUMAN_REVIEW` verdict against the known stale-INDEX false negative (commit `b2e1006d3`)
before believing it.

**The transferable bit:** the useful answer to *"does this conflict with your thing?"* was not about
my thing at all. Checking the disjointness claim honestly meant reading a shared index definition,
and that is where the finding was. **A narrow question from a peer is a cheap prompt to re-read a
shared mechanism you have been quoting from memory.**

### ⚠ The landmine I filed at 19:30 was wrong by 19:50 — and I had already shipped the wrong number to a peer

`bugs_open/384` came back having acted on all three points inside twenty minutes: a register entry
(PBP-048) naming the producers with a dated count, a unit test on their key shape, and my
"all-history: 0 pairs with two types" **quoted in their helper's doc comment**. Reading their reply
closely enough to check *their* shape claim is what turned up mine.

**`site_work_items_archive` exists** — 25,281 rows, 25,070 keyed, 2026-02-22 → 2026-08-17. I had
queried only the live table and written **"ever"**.

| claim I gave them | true value (live ∪ archive) |
|---|---|
| 0 `(site_id,item_key)` pairs have ever carried two `item_type`s | **20** — incl. `needs_page` + `page_rerender` on `page_rerender:llm-cost-calculator`, their exact pair |
| they are the **4th** producer on `section_data_resolved` | **53** distinct `created_by` over 1,289 rows (live alone: 10 over 198) |

The first came from reading the wrong **table**; the second from a `LIMIT 12` ordered by count across
all reasons, so the small producers fell off the bottom. Both were reported as censuses.

⚠ **RUNBOOK §2 of this lane carries that exact warning, in bold, in my own words** — *"`site_work_items`
is a ROLLING WINDOW … a figure for 'how many were ever X' cannot be taken from here"* — and I had
re-read it hours earlier while applying 587. Knowing the trap did nothing. What was missing was the
one question: **which table would hold a counter-example, and am I reading it?**
`information_schema.tables WHERE table_name LIKE '%work_item%'` found the archive in one query.

**Not over-corrected, because the swing is its own error.** The 20 are **not** proven index
violations: `idx_swi_dedup` constrains only live rows with non-terminal status, so an archived row
has freed its slot and a later row of another type may take that key legitimately. What the 20
establish is that the key space is **shared across types in practice**. The entry's check is now a
**ratchet against a dated baseline of 20**, not a zero-invariant, and it unions the archive — read
against the live table alone it returns 0, which is the failure the entry's own first draft made.

**One thing got BETTER on re-measurement, and it belongs to them.** I was ready to tell 384 their
unit test was one-sided. It is not: `needs_page` has used the colon shape **exclusively** — 337
archive + 154 live — and underscore **zero** times in six months, and the one real collision was
colon-on-colon. Their underscore key is safe by evidence, not by convention. The residual hazard is
the **46 `page_rerender`-typed rows that use the colon shape** (36 archive + 10 live, all lane
one-offs) sitting in `needs_page`'s namespace. **The convention is not the hazard; the minority that
opted out of it is, and nobody reviews a lane one-off.**

**Then the correction had its own trap.** I struck the false line through and abbreviated it with an
ellipsis, which silently swallowed two figures that were still correct (the prefix-sharing census and
its per-producer breakdown). The count-gate this file's own landmine prescribes — `git diff --numstat`
— catches lines *leaving*, and I ran it and read it; it says nothing about content vanishing *inside*
a line you rewrite. **An ellipsis in a correction is a deletion the diff renders as an edit.** Caught
by the pre-commit pattern check, not by me. Restored verbatim.

**The transferable one:** a wrong figure you keep is an error; a wrong figure you **relay** is a
supply-chain defect, and it compounds at the speed of a competent peer — register entry, doc comment
and unit test were downstream inside twenty minutes. **Hand over the query with the number.** I did
not the first time and did the second, so their re-check costs one query instead of an excavation.

### Closing the loop: 384 re-measured independently and we reconcile — plus two housekeeping notes

**384 caught it in time.** Their doc comment, `PBP-048` and the council rationale were corrected
before commit and dispatch. They re-ran the union query themselves rather than taking my second
number on trust, and every figure matches: **20** cross-type pairs live ∪ archive; `needs_page`
**491 colon / 0 underscore**; `page_rerender` **16,097 underscore / 46 colon**. Their helper's
comment now records the 20 as a dated ratchet and names the hazard as colon-shaped hand dispatches.
Their producer census: **1,289** rows / **53** `created_by`, split by a stated rule (an agent or
action name = standing producer) rather than a bare count.

⚠ **One off-by-one worth flagging back, measured here 2026-08-24 ~20:05 UTC.** Their split reads
"`render_news_section` 795, `rerender-pages` 203, `completeness-discovery-agent` 2 … the other **49**
`created_by` / 289 rows are hand dispatches". The rows reconcile exactly (795+203+2 = 1,000; +289 =
1,289) but the producer count does not: excluding those three names leaves **50** distinct
`created_by`, not 49. `3 + 50 = 53`. Trivial in itself — and this whole exchange is about numbers
that get enshrined with a date, so it goes back.

**Housekeeping: my WRONG_CALLS commit carried a same-file passenger, and the pre-commit pattern check
is what told me.** It flagged *"3 lines removed from WRONG_CALLS.md, a fleet-wide append-only
ledger"* — I had only appended. Checked: another session had **rewritten** an existing entry in
place (the `orchestration_states` runs-vs-items one), expanding *"What was true"* into *"What was
true — in three steps, because the correction was itself wrong once"* and *"The cheap check"* into a
two-part *"cheap checks"*. So the three lines were **replaced, not lost**, and my pathspec commit
carried their improvement. Nothing to undo. **The check earned its place here**: appending only, I
would never have looked, and the one shape it cannot distinguish — a rewrite from a deletion — is
exactly the one where looking is cheap and being wrong is unrecoverable.

### ⚠ The landmine verifier returned STILL_VALID — and it CANNOT have checked the number I got wrong

Verdicts on `LANDMINES.md#idxswidedup-…` landed 2026-08-24 19:52:40 UTC, both **STILL_VALID**. Read
what they actually cover before quoting them:

- **What was verified, and it is the useful half:** the footprint resolves (`work_items_common.go`,
  `insertWorkItem` at `load_work_item_actions.go:1491`, `workItemTerminalStatuses:42`,
  `discovery_checks/`, `pageRerenderItemKey` in `create_rerender_items_action.go`) and **21+ Go
  files reference `idx_swi_dedup` as `(site_id, item_key)` without `item_type`** — an independent
  confirmation of the entry's structural claim from code comments, arrived at without my query.
- **What was NOT and could not be:** every data claim. The corpus is **Go only — 8,700 symbols, no
  `.sql`, no database access** — and it says so itself (`1 NOT ANSWERABLE by this index`, and
  `site_work_items.item_key` column existence "not verifiable"). **The 20 cross-type pairs, the
  491/0 and 16,097/46 shape counts, the 53 producers: none of them were checked by anything here.**
- ⚠ **And it ran against indexed commit `e347c5ad`, 2026-08-23 12:21 UTC — a day old, and PRIOR to
  the correction.** So the STILL_VALID is stamped on the entry as it stood *before* I discovered the
  "zero, ever" was 20.

**So a green verdict beside this entry means "the symbols still exist and the index shape claim
matches the code", not "the entry is right".** That distinction matters more than usual here,
because the headline of the first draft was false and the verdict cannot tell the difference — the
one claim a reader would most want vouched for is precisely the one outside the instrument's scope.
Same family as everything else in this file today: **an instrument's PASS is scoped to what it can
see, and nothing in the pass says what that was** unless, as here, the tool prints its own corpus.
Credit where due: this one does.

### ⚠ CORRECTION to my own two entries above: 384's `page_list_stale` has NO Resolved arm

Both earlier entries in this file describe that check as *"closing its own rows via
`CheckResult.Resolved`"*, and my reply endorsed the arm's shape. **The 384 lane reversed it and told
me so, precisely so this file would not carry the stale claim.** Their reason: a `page_rerender` is
an action request that completes on its own, and its key is **shared with every other
`section_data_resolved` producer for that page** — so a retraction keyed on "the images match now"
could close `render_news_section`'s legitimate request on a page that also carries a listing. The
"positive observation only" rule survives, moved to the **filing** side: unknown pages are counted
in a per-run summary finding rather than silently filing nothing.

**Measuring their reasoning gave a stronger version of it, and nearly gave a false alarm.**
[MEASURED 2026-08-24 ~20:15 UTC, live ∪ archive — the union, this time]

| `item_type` | rows | filing producers | rows carrying `result.resolved_at` |
|---|---|---|---|
| `page_rerender` | 18,360 | **122** | **0** |
| `needs_page` | 1,418 | 46 | 4 |
| `contrast_failure` | 513 | **1** | 79 |

**Nothing has ever retracted a `page_rerender`, in 18,360 rows.** So the arm they declined would not
have been one retractor among several — it would have been the **first**, deciding on behalf of 121
other producers whose requests it cannot recognise.

⚠ **And the near-miss, which is the part worth keeping.** Four types (`empty_section`,
`literal_markdown`, `needs_rerender`, `canonical_mismatch`) have rows filed under **two** distinct
`created_by` values *and* rows retracted — which reads instantly as "the hazard 384 avoided is
already live in three places". **It is not.** One more query before sending: each of those four has
exactly **one** distinct `result->>'resolved_by'`. Multiple filers, single retractor, every time. No
type in this estate has two competing retraction authorities.

**So the rule is not the one I was about to write.** It is not *"only a sole producer may retract"* —
`contrast_failure` is merely the degenerate case of it. It is:

> **A producer may close work items on a key only by being the SOLE RETRACTION AUTHORITY for that
> `item_type`, and only if it can recognise every other producer's rows as legitimately its own.**

That is the condition WII-016 satisfies by accident (`contrast_failure` has exactly one producer,
and VIZ-016 states so) and the condition `page_rerender` cannot satisfy at all. **Declining the arm
is the correct move there, not the cautious one.**

⚠ **Note what saved me: I have now had `created_by` mislead me twice in one hour** — once as a
producer count truncated by a `LIMIT`, once as a proxy for retraction authority. It is a free-text
label, written by whoever filed the row, and it answers *"what wrote this"*, never *"what may close
this"*. The column that answers the second question is `result->>'resolved_by'`, and it took one
query.

### Their re-run caught my third wrong number of the evening — and three of the four were right by luck

384 re-ran the retraction census independently rather than accepting mine, and reported
`needs_rerender 635 / 21 / 17`. **I had published that filer count as 2.**

The landmine's table has a `filers` column. The top three rows are whole-type `created_by` counts.
The bottom four I filled from a *different* query — one that grouped `created_by` **among retracted
rows only**, because that is the query I happened to have on screen from checking whether any type
had two retractors. Two populations, one column, one header.

⚠ **Three of the four were correct anyway** — `empty_section`, `literal_markdown` and
`canonical_mismatch` each have 2 filers by either measure, because every filer of those types files
rows that get retracted. **That is worse than four wrong.** A wholly wrong column gets noticed; a
column that is 75% right by coincidence reads as verified, and the one wrong cell was the one
carrying the most weight (21 producers is a far stronger illustration than 2).

**And the correction supplied what the entry had been missing.** With `needs_rerender` at **2**
filers, every row in the table pointed the same way and the rule read as *"do not retract on a
shared type"*. At **21** — 21 producers, 17 retractions, **one** authority — it is the estate's
clearest SAFE case, and the rule becomes discriminating: retraction on a many-producer type is fine
*precisely when* there is a single authority. Passed back to them for the Phase 2 council rationale,
which had only the STOP case and would have invited "so is this just blanket caution?".

**Tally for the evening, since the point of this file is the missteps and not the wins:** three
figures published wrong, all three to a peer, none of them caught by me first —
`0 collisions ever` (was 20, found by re-reading their reply), `fourth producer` (was 53rd-ish,
found the same way), and `needs_rerender 2 filers` (was 21, found by **their** re-run). Every one
came from running a query I already had against a question I had just changed. **The common cause is
not carelessness with SQL; it is reusing a result whose population was chosen for a different
question** — and the tell is always the same, that the query was already on screen.
