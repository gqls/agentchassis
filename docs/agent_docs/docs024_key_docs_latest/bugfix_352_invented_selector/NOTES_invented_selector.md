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
