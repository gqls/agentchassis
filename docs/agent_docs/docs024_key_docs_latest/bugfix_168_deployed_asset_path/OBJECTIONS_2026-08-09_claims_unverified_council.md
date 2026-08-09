# Council round 1 on `claims_unverified` — **REVISE**, and the gating objection caught a real error of mine

**Correlation `b67eb26a-14ef-45d7-b755-3e489fd57ef0`** · 15 reviewers, 2 abstained,
`gated_by_truncation: false` · decided by **gating objection from `editquality`** ·
verdict written 2026-08-09 09:51:59Z.

Eleven of the fifteen seats approved. Four objected or flagged: `editquality` (gating),
`compliance` (HIGH, policy), `guardian` (containment), `debug_historian` (rollout), plus
`bug_historian`, `prior_art_librarian` and `guidelines` with checks attached.

> ⚠ **How to read this verdict, and a trap I hit getting it.** The documented command
> `SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY created_at DESC LIMIT 1`
> returned **another lane's verdict** (`46f87e4c-…`, `bugs_open/228`) — on a tree this concurrent,
> "most recent" is almost never yours. Read it by **your** correlation:
> `SELECT body FROM diagnosis_artifacts WHERE correlation_id='<CORR>' AND kind='council_report';`
> Note `metadata` carries only the summary (`decision`, `reviewers`, `abstained`); the objections
> are in **`body`**.

---

## 1. GATING — `editquality`, HIGH, edit 3: "two producers" was asserted, scan-logic parity was not

> *"The diagnosis names TWO converging producers … but `ScanDeployedClaims` is extracted only from
> `check_unverified_claims.go`'s logic. If `check_unverified_claims_stats.go` emits
> `claims_unverified` findings via different logic than what `ScanDeployedClaims` re-scans, the
> revalidator will judge those items against the wrong producer's predicate … the plan's own
> item_key-shape argument does not establish that the underlying scan is shared, only that the
> item_key format matches."*

**The seat is right about my reasoning and the feared consequence is refuted. Both halves matter.**

**My error, stated plainly: there is ONE producer, not two.** I wrote "TWO converging producers"
in the submission, the code header, the register entry, the map comment and the coverage test —
and I invoked the owner ruling of **2026-08-02 §1** (converging producers need no RFC provided both
are named in the register entry) as the authority for shipping without an RFC. **That ruling never
applied to this change.** What misled me was the stats file's own header line, *"reuses the existing
`claims_unverified` item type"*, which describes contributing findings to a type — not filing items.

```bash
grep -n "WorkItemSpec\|ItemKey\|ItemType\|func init\|Register(\|CheckResult" \
  platform/orchestration/actions/discovery_checks/check_unverified_claims_stats.go
#   (no output — it registers no check, has no init(), emits no work item)

grep -rn "scanStoredStatClaims(" --include=*.go platform/ | grep -v "func scanStoredStatClaims"
#   check_unverified_claims.go:385   <- inside ScanDeployedClaims, pages
#   check_unverified_claims.go:427   <- inside ScanDeployedClaims, site chrome
#   (everything else is its own _test.go)
```

**And that same call graph refutes the consequence.** The seat asked the sharper question — is the
*scan logic* shared, not just the key shape? It is, structurally: `ScanDeployedClaims` calls
`scanComponentClaims` **and** `scanStoredStatClaims` on every component, so there is exactly ONE
scan and both halves live inside it. The revalidator cannot judge by "the wrong producer's
predicate" because there is no second predicate to judge by. Key-shape parity was indeed the weaker
claim I made; scan-logic parity is the stronger one that happens to be true.

**Actioned:** corrected in `revalidate_unverified_claims.go`'s header (with the citation),
`revalidate_review_queue_action.go`'s map comment, `revalidate_review_queue_test.go`, the CQ-021
register entry (visibly, strike-through style — a register entry is read as ground truth by the
next council round) and the index row. The change is **simpler** than I described it, and it no
longer leans on a ruling that does not cover it.

## 2. HIGH and OWNER-FACING — `compliance`: this is a policy change dressed as a mechanical one

> *"`claims_unverified` was designed HITL-terminal precisely because it is a factual-claim-
> correctness surface … Evidence-base registration proves provenance, not correctness (the plan's
> own landmine note says so) — so a wrong or sloppy register entry can retract a live
> claims-integrity finding with no human ever looking at it. This is a policy-level narrowing of
> the human-in-the-loop guarantee on the platform's highest-stakes content-integrity item type …
> it should get explicit sign-off from the two bug-file owners (033, 083) before merge, not just a
> notification."*

**Deliberately unactioned, and escalated to the owner.** Three seats independently reached this
from different mandates and all three routed it to a human rather than blocking:

- `compliance`: *"Route to a human sign-off from the two named bug owners before shipping, not just
  a notification line."*
- `bug_historian`: *"ARCHITECTURE-LEVEL CONCERN FOR A HUMAN … that precedent was a style/tone gate;
  this one closes findings about **factual claims**. Worth a human policy sign-off independent of
  this council's bug-pattern lens."*
- `architecture`: *"the guardian seat, not this one, should weigh whether auto-closing a
  factual-claim review queue is the right policy call"* — while recording `point_fix` on the
  mechanism itself.

**Seats agreeing that it needs a human IS the signal, and it is not answerable by resubmitting with
better measurements** — the same shape as the `bugs_closed/124` scope veto. The mechanism is sound
(every seat that judged the *engineering* approved it); the question is whether the platform wants
a machine closing factual-claim review rows at all. That is the owner's call, and the distinction
`compliance` draws is the sharp one: **the register proves provenance, not correctness.**

Recorded here and surfaced to the owner in `README_where_we_are`. Not resubmitted around.

## 3. Answered with measurements — the checks the seats asked for

### `guardian` (MEDIUM): confirm nothing else already closes this type

```sql
SELECT count(*) FILTER (WHERE status IN ('complete','verified'))       AS ever_closed,
       count(*) FILTER (WHERE result ? 'deploy_result')                AS deploy_payloads,
       count(*) FILTER (WHERE handler_agent <> '' AND handler_agent IS NOT NULL) AS with_handler,
       count(DISTINCT resolution_path)                                 AS distinct_resolution_paths
FROM site_work_items WHERE item_type='claims_unverified';
--  0 | 0 | 0 | 0
```

Decisive on all four axes: nothing has ever closed one, no fix pipeline stamps one, no row carries
a handler, and no row has any `resolution_path` at all. No race, no duplicate closer.

### `bug_historian` (LOW): the no-page-status-filter choice is unverified against actual archived pages

**The seat was right to ask, and the exposure is real: 2 of 23.**

```sql
-- page status across the live claims_unverified population
--   active 19 · archived 2 · page deleted 2
```

The two deleted pages refuse via the nil-scan arm and are not exposed. **The two archived ones
can close on a page nobody serves.** I keep the emit-side parity — the revalidator must judge by
the check's exact predicate or the two ends disagree, which is the whole point of the shared scan —
but the landmine now carries the number instead of the argument alone. Changing it means changing
what the CHECK reports, which is a separate decision.

### `guidelines` (missing): does adding a type starve other types' revalidation budget?

No. `max_items` is **1500**; the last run scanned **186** with `cap_binding: false`. Adding ~23
takes it to ~209 — **14% of the cap**. The starvation shape this lane fixed is not reachable at
this volume, and `cap_binding` is the standing tripwire if it ever is.

### `debug_historian` (MEDIUM): no deploy-verification step was stated

Fair — I had one and did not put it in the submission. It is in `HANDOFF_2026-08-09` §0c and the
**baseline is already spent**: `2026-08-09T09:36:10Z`, `v1.0.1270` (predates the commit), both
replicas **0 / 0 / 0** with positive control `2`, three ASCII-only needles each verified to return 0
on a pre-change binary. Post-roll requires ≥1 on every replica. Pod-grep against the running
binary, never git and never the tag.

### `prior_art_librarian` (MEDIUM, ×2): second producer unconfirmed; precedent not re-verified

First is answered by §1 — and the answer is that the second producer does not exist. Second:
the voice_tells precedent (`4d430ca8-7e34-479a-95f3-71fdc12fdef6`) is APPROVED r1, and it is now
more than a citation — it went LIVE and **closed a real item unattended on 2026-08-09 08:38:54Z**
(`voice:ecfd0bfd-…`, page `ai-readiness-quiz`). The analogy is behaviourally, not just
documentarily, grounded.

### `guardian` (LOW): does anything outside this package reach the newly-exported symbols?

`ScanDeployedClaims`, `LoadEvidenceBase`, `ClaimsPageScan` and `ClaimFinding` are new exports —
nothing could have depended on them before they existed. The previously-unexported spellings
(`scanDeployedClaims`, `loadEvidenceBaseForCheck`, `pageClaimFindings`) were package-private and had
no callers outside `discovery_checks`; `unverifiedClaimFinding` is untouched and aliased, not
renamed, so its ~20 in-package use sites including the tests are unaffected. `go build ./...` and
the package tests are green.

## 4. What was NOT changed, and why

- **The ladder, the arms and their order.** Every seat that judged the mechanism approved it;
  `constitution`, `reuse_agent` and `architecture` each singled out the shared-scan extraction and
  the read-nothing-vs-clean accounting as the right shape.
- **The no-page-status-filter asymmetry** — kept, now with its 2-of-23 exposure measured (§3).
- **The HITL policy question** — owner's, not mine to resolve by resubmission (§2).

## 5. Resubmission

Resubmitted with `RESUBMIT_CORR=b67eb26a-14ef-45d7-b755-3e489fd57ef0` so the trail accumulates.
The substantive delta is §1 (the producer-count correction, which *narrows* the change and drops a
misapplied ruling) plus the four measurements in §3 folded into `grounded_in`. §2 is carried as a
stated, deliberately-unactioned owner escalation rather than argued away.
