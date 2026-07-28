# 102 — the claims layer is page_type-blind: a guide's worked example is indistinguishable from a business claim

**Filed 2026-07-27** by the fabricated-stats lane (`bugs_closed/043`, resolve **by slug**),
while doing owed item (b) — building evidence registers for the remaining publishing sites.

**Status: FIXED 2026-07-28** (`3ddb4ed2d`, session "bugsearch 7") — code committed, **inert
until the next chassis roll**. Workstream:
`docs/agent_docs/docs024_key_docs_latest/bugfix_102_page_type_claims/`. See
§ "The fix, 2026-07-28" at the bottom, which also **corrects two claims in this
file's own opening**.

~~**Live exposure today: nil**, because the affected sites have no `evidence_base` row, so the
scans are off.~~ > **CORRECTED 2026-07-28 — true of webdesign.co.uk, false fleet-wide.**
Nine sites are armed and carry **124** live unregistered-number findings, **61 of them
editorial-page false positives**, at a severity that BLOCKS a rebuild. Measured, below.
**This bug is what stops that being fixed**, which is why it is worth filing
rather than absorbing.

---

## The finding

The claims layer's only precision control on prose is `businessClaimContextRe`
(`datahelpers/claims.go`) — a **lexical** gate asking whether the words around a number
sound like business. Nothing in the layer consults **what kind of page the number is on**:
`check_unverified_claims.go` joins `pages` but reads only `p.name`, and neither
`ScanUnregisteredNumbers` nor `ScanStatClaims` receives a page type at all.

On a marketing site that is fine — nearly every page asserts something about the business.
On a site whose main body is **teaching content**, it is wrong in the expensive direction:
an explainer's worked example is lexically identical to a sales claim.

## The measurement that found it

Survey run 2026-07-27 with the real scanner (`ExtractAssertionText` +
`ScanUnregisteredNumbers` against an **empty** register, so every business-shaped number
surfaces) over the four sites that owe registers:

| site | components | number claims | verdict |
|---|---|---|---|
| gaswholesalers.com | 102 | 0 | needs no register |
| dartsonline.com | 17 | 1 | a 30-day returns window — a policy term |
| finetuning.uk | 139 | 5 | 4 are audience descriptors / a privacy age limit; **1 is real** |
| **webdesign.co.uk** | **101** | **15** | **all 15 are false positives** |

Every one of webdesign.co.uk's 15 is a worked example inside teaching content, and all of
them sit on `page_type = 'guide'`:

```
10,000  | Once your site hits 10,000 monthly visitors and is generating revenue…
100     | …bringing the server to a halt at just 100 concurrent users.
502     | …random 502 Bad Gateway errors become inevitable.
3.5     | …let's say the global average for all products on your site is 3.5 stars.
0 / 1   | …getting a site from 0 to 1 is no longer the hard part.
100,000 | …getting that site from 1 to 100,000 users.
5 / 1   | * Try rating Sci-Fi '5' and Rom-Com '1' to match User A.
```
```sql
SELECT DISTINCT p.name, p.page_type FROM page_components pc
JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE s.domain='webdesign.co.uk' AND (pc.rendered_html LIKE '%concurrent users%' OR …);
-- learn-ai-builders-content-first        | guide
-- learn-algorithms-bayesian-theory       | guide
-- learn-algorithms-p-values-explained    | guide
-- learn-algorithms-recommendation-math   | guide
-- learn-communities-graph-vs-relational  | guide
-- learn-operations-scaling               | guide
```

**Six pages, one page_type, 100% of the false positives.** The signal that separates them
from a real claim already exists in the schema and is simply not read.

## Why this matters beyond noise

1. **It blocks owner-requested work.** Item (b) of the fabricated-stats lane is "evidence
   registers for all the publishing sites". For webdesign.co.uk that currently cannot be
   done: switching the register on would raise 15 human-review items that are all correct
   copy, and the queue (`bugs_open/033`) has no working surface to dismiss them through.
   So the largest site in the estate stays unprotected — not because we lack evidence, but
   because the checker cannot be trusted on it.
2. **False positives are the expensive failure here.** A missed claim is one wrong number.
   A checker that cries wolf on a teaching site trains a human to dismiss its findings,
   which silently disables it on the same site's real claims. That is worse than off.
3. **It is the mirror of the `043` design decision, and the boundary is now visible.**
   `ScanStatClaims` deliberately does NOT apply `businessClaimContextRe`, because a figure
   in a `stat*_value` field is a claim *by construction* — structural position replacing a
   lexical gate. That reasoning is sound and it cuts both ways: **page type is structural
   position too**, one level up, and the layer uses it nowhere.

## Fix candidates (unranked)

1. **Pass page context into the scan and let the caller scope it.** `scanDeployedClaims`
   already selects from `pages`; add `p.page_type` and hand it to the scan. A `guide` /
   `blog-post` page's *prose* scan is either skipped or graded a rung lower, while its
   **stat fields keep being scanned** — a stat card on a guide is still a published claim.
   Makes the bad state unrepresentable at the point of decision rather than asking
   operators to remember which sites are educational.
2. **A per-site or per-page-type opt-out in the `evidence_base` row** (e.g.
   `"scan_prose_on_page_types": ["landing","about","content"]`). Cheaper, no Go change —
   but it is "operators must remember X" in a configuration costume, and it will drift the
   moment a new page type appears.
3. **Tighten `businessClaimContextRe` to exclude tutorial framing** ("let's say", "imagine",
   "for example", "Try rating"). Narrowest, needs no plumbing, and it would have caught most
   of the seven above — but it is a lexical patch on a lexical gate's failure, so the next
   phrasing escapes it again.

**Prefer (1).** The structural signal exists and is free; (3) is worth doing anyway as
defence in depth, and it is the only one that helps a guide-shaped section on a
non-guide page.

## How to verify a fix

Do **not** grade it on a count going down — a scan that got quieter may just be off. Seed
webdesign.co.uk a real register, run the audit, and assert **both** directions:

- the six `learn-*` guide pages raise **zero** prose findings; and
- a deliberately false figure planted in a **non-guide** page's copy on the same site is
  still raised (the negative control — without it, "quiet" and "broken" look identical).

## Related

- `bugs_closed/043` — the parent lane; its `claims_stats.go` header is where the
  structural-position-replaces-lexical-gate argument is made for stat fields.
- `bugs_open/093` — the second call site for the stat audit, which is what made this survey
  cheap enough to run fleet-wide.
- `bugs_open/033` — the review queue with no working surface, which is what turns 15 false
  positives from an annoyance into a reason not to enable the check at all.

---

## Triage 2026-07-27, post-roll (v1.0.1174) — unchanged, and the blocking claim is confirmed live

Verification sweep, not a fix. Nothing to diagnose here and nothing has moved.

**The structural signal is still unread.** `page_type` / `PageType` appears **nowhere** in
`datahelpers/claims.go` or `discovery_checks/check_unverified_claims.go` — grep returns no
matches. Confirmed against the running image: no Go commits after `e96d42226`, which is in
`v1.0.1174`.

**The premise "live exposure today is nil" is confirmed, and so is the blockage.**
webdesign.co.uk still carries **zero** `banned_claims` and has no `evidence_base` row:

```sql
SELECT s.domain, jsonb_array_length(COALESCE(ss.data->'banned_claims','[]'::jsonb))
FROM sites s LEFT JOIN site_specs ss ON ss.site_id=s.id AND ss.aspect='evidence_base' AND ss.is_current
WHERE s.domain='webdesign.co.uk';   -- webdesign.co.uk | 0
```

So the estate's largest site is unprotected for exactly the reason this file gives, and the
owed item (b) of the fabricated-stats lane still cannot be done for it. That makes 102 a
**precondition of `bugs_open/104`'s coverage work**, not a parallel concern: arming
webdesign.co.uk before this lands would fire 15 correct-copy findings into
`needs_human_review`, which has no surface (`bugs_open/033`) — the exact "trains a human to
dismiss its findings" failure § "Why this matters" warns about.

One sizing note for candidate 1: the prose scan runs at the **build gate**, which is live, so
the fix has real effect immediately on roll — unlike the post-deploy half, which does not run
at all (`bugs_open/083`). The § "How to verify a fix" negative control (a deliberately false
figure on a non-guide page must still be raised) is therefore exercisable on the build path
without waiting for any sweep.

---

## The fix, 2026-07-28 — candidate 1, with the boundary drawn one level finer

Session "bugsearch 7". Commit `3ddb4ed2d`. Council round 1 submitted before committing:
`de4a19f5-8f03-4e74-92cb-c23c10ab829d`. Workstream (plan, runbook, notes, prose log):
`docs/agent_docs/docs024_key_docs_latest/bugfix_102_page_type_claims/`.

### What the fleet measurement found, and the two corrections it forces

Method: `cmd/claimscan` — the same shared engine as the gate and the audit — run against
**each opted-in site's own live register** over live `page_components.rendered_html`, with
`pages.page_type` carried through the export. Nine sites have a current `evidence_base` row.

| page_type | before | after the fix |
|---|---|---|
| blog-post | 46 | 0 |
| content | 38 | 38 |
| report | 14 | 14 |
| adoption-tracker | 8 | 8 |
| tool | 7 | 0 |
| game | 4 | 0 |
| protocol-tracker | 3 | 3 |
| section-index | 2 | 0 |
| news-index | 1 | 0 |
| guide | 1 | 0 |
| **total** | **124** | **63** |

`comm` over the sorted finding lists: **61 suppressed, 0 newly appearing**.

> **CORRECTION 1 — "live exposure today: nil" was scoped to webdesign.co.uk.** The nine armed
> sites carry 124 findings today. I read all 61 on editorial page types and **all 61 are false
> positives**: `0.40 × 0.0833 = 3.33%` in a probability explainer, "expose an endpoint that
> returns **200**", "Set to **0** to disable" in tool help text, a quoted third-party market
> share ("cobot tending cells to hold **38%** share") on a news index. The 63 on
> content/report/tracker pages include the real ones the layer exists for (leopardess's
> `90,790`, ai-agent-orchestration's `170`). **Measured precision on that class: 0%.**

> **CORRECTION 2 — the cost is blocking, not noise.** § "Why this matters" argues from
> false-positive fatigue. Also true, but `unregistered_number` is **`error`** severity in the
> build gate and `valid := blockerCount == 0 && errorCount == 0` — so an armed site's blog
> post carrying a worked example **cannot be rebuilt today**. gamesdesign.co.uk has 40 such
> findings across four posts.

### The boundary this file's candidate 1 did not draw

Candidate 1 says a guide's "**prose** scan is either skipped or graded a rung lower". Taken
literally that would have regressed the case that motivated the whole check:
`check_unverified_claims.go`'s header records its first live run (2026-07-16) finding
"70+ agents across eight functional departments" **on a guide**.

**So only the NUMBER scan reads the surface.** `ScanBannedClaims` runs on every page type — a
human-authored pattern for a known falsehood has no false-positive problem to protect
against. `ScanStatClaims` likewise (`claims_stats.go`: a figure in a `stat*_value` field is a
claim by construction, and that argument does not weaken on a guide). Both pinned by tests.

### What shipped

- `datahelpers.ClaimSurface{PageType}` + `ProseNumbersAreClaims()` — the policy in one place,
  next to the lexical gate it corrects. **Zero value = UNKNOWN = scanned**: a scanner that has
  gone quiet and one that is broken look identical from outside.
- `ScanUnregisteredNumbers(blocks, surface)` — a **signature change, not a variant**. An
  optional-safe second entry point is `bugs_open/093`'s shape exactly; requiring the parameter
  makes the compiler visit every caller.
- The three callers pass what they already hold: the gate via `resolvePageType`
  (`page_record.page_type`, which `load_page_record` populates and `page-build-handler` runs
  before `validate_content`); the audit via one column added to a `JOIN pages` it already did;
  `cmd/claimscan` via an optional 4th TSV field, warning on stderr when it is absent.
- Editorial set: `guide`, `blog-post`, `blog-index`, `news-index`, `section-index`, `tool`,
  `game`. Fleet-wide, hard-coded, no per-site override — candidate 2 is "operators must
  remember X in a configuration costume", and an unknown type defaults to SCANNED.

### Verification (both directions, per § "How to verify a fix")

Positive control: the fleet re-run above, 124 → 63, suppressing exactly the 61.
Negative control: the same blocks on `content`/`landing`/`report`/unknown surfaces still flag
(`TestSameTextOnBusinessSurfaceIsStillScanned`). Regression control: a banned claim on a guide
still flags while the number scan on the same block does not. Plumbing: `resolvePageType` is
pinned against the live workflow shape, because a policy that never receives a page type is
inert and looks identical to a working one.

### Still open, filed here rather than fixed (each is a DIFFERENT mechanism)

1. **`report` pages carry 14 false positives of another class** — model numbers inside product
   names ("Schunk EGP **40**-N-S-B — manufacturer specification") tripping on `verified` in the
   context window. Deliberately not fixed by calling `report` editorial: that would be a
   coincidence, not a mechanism, and a report page's figures genuinely can be business claims.
2. **Candidate 3 (tutorial-framing lexical exclusion) is still worth doing**, and is the only
   thing that helps a guide-shaped section on a business page. One live instance measured:
   finetuning.uk's "particularly not in the **5** to **50** employee businesses" on a `content`
   hero. Deferred as a second mechanism rather than accreted onto this change.
3. **webdesign.co.uk can now be armed** — the blocker this file describes is gone. Its 15
   false positives were all on `page_type = 'guide'`. That is `bugs_open/104`'s coverage work
   and this file was its precondition.
4. **After the next chassis roll, pod-grep before believing it is live** (the marker and its
   negative control are in the workstream RUNBOOK § 2).
