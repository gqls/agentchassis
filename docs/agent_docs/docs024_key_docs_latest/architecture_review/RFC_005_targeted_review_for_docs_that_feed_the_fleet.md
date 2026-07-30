# RFC 005 — targeted review for the two doc classes that actually feed the fleet back, plus a periodic staleness sweep for `bugs_open/`

**Status: PROPOSED, awaiting owner ruling.** Raised 2026-07-30 out of a live
discussion (dartsonline.com session), prompted by the owner's question: should
the council — "especially the architect member" — be in scope for changes to
`bugs_open/`, `016b_debugging_guide`, and `LANDMINES.md`, since they're "core
architecture"? This is not retrospective in RFC_002/004's sense — nothing has
shipped yet. It exists because the owner and the session agreed a full
comprehensive council was the wrong shape, but a *targeted* one might not be,
and per the 2026-07-29 seam ruling, a change to what the review mechanism
*covers* is itself the kind of decision that goes to a ruling, not something a
session should just wire in because it seems obviously good.

---

## 1. What prompted this

Filing `bugs_open/155` (`deploy_image_asset` resolving a source image by
`purpose` instead of `asset_id` — a real, silent, `success:true` defect) surfaced
two things in the same session:

1. `LANDMINES.md` is not inert. Migration 271 (live 2026-07-30) syncs it into
   `doc_notes` and feeds it into every council seat's `schema_hint`. A wrong or
   misleading entry there doesn't just cost a future reader a wasted check — it
   shapes how the architecture/guardian seats judge *future* platform-code
   submissions. That is a real feedback loop from documentation back into
   review.
2. `155` is exactly the class of claim CLAUDE.md's own "Diagnosis before
   debugging" section says should go through `090_TRIGGER_needs_diagnosis`
   before being asserted as a filed root cause — cross-cutting, cause not
   where the symptom is, a fix that will change behaviour for every
   multi-same-purpose-asset site. The session that filed it did rigorous
   first-hand verification (reproduced the identical hashes, read the exact
   function, confirmed the fix), but did not run the platform's own designated
   mechanism for this. That gap is what makes "should there be *some* review"
   a fair question — the answer the discussion reached was **yes, but not the
   full multi-seat council**, because that council is tuned to judge diffs
   against code invariants, not to fact-check prose.

## 2. What this RFC does NOT propose

- **No full multi-seat council over `bugs_open/`, `016b`, or general docs.**
  These are edited many times a day across concurrent sessions by design (the
  "standing five" convention); routing every append through council-run cost
  (~30 min turnaround, real credits, per CLAUDE.md's own numbers) either goes
  unfollowed — CLAUDE.md already names "an unadopted ledger" as the failure
  worse than no mechanism — or measurably throttles the fast diagnostic work
  this whole apparatus exists to enable.
- **No change to `016b`'s review status.** It has no wire into the council's
  own inputs (checked directly: `schema_hint`/`doc_notes` consumers touch
  `LANDMINES.md` via `landmines-sync.py`; nothing pulls `016b` or `bugs_open/`
  content into `doc_notes` at all). It stays governed the way it is today —
  corrections in place, dated, visible.

## 3. What this RFC does propose

### 3.1 Apply the *existing* diagnosis-loop discipline to durable `bugs_open/` claims — no new mechanism, a practice reminder

`090_TRIGGER_needs_diagnosis_v1.sh` already exists and already covers this
case; CLAUDE.md already says to use it before committing to a durable
root-cause claim. The gap is adherence, not tooling. Proposal: **a bug file
that asserts a cross-cutting or structural root cause is not "filed" until
either (a) it has been through `090`, or (b) the filing session states plainly
why it skipped it** (e.g. "verified first-hand: reproduced the failure,
confirmed the fix resolves it, read the exact code path" — a real substitute,
but a stated one, not a silent omission). This is a documentation-of-practice
change, not a new pipeline.

### 3.2 A narrow, content-focused check for `LANDMINES.md` entries — NOT the multi-seat council

Given §1's feedback loop, new/changed entries deserve *some* verification
before they reach `schema_hint` — but the check is a fact-check, not a code
review:
- does the **footprint** actually resolve (grep the file/table/symbol)?
- does the **check** query, run for real, produce what **the tell** claims?
- is the entry internally consistent (footprint actually related to the
  fires-when clause)?

This is closer to `landmines-sync.py`'s own job than to a council seat's. Open
design question for the ruling, not decided here: extend the sync script
itself to spot-check the fresh entries it's about to sync (mechanical,
cheapest, no LLM), versus a small dedicated single-pass verifier agent
(catches subtler mismatches a grep can't, costs an LLM call per sync). Either
way: **advisory, non-blocking** (matching every other review in this repo),
and **not** routed through the existing multi-seat council-gate machinery —
reusing that pipeline for a fact-check task risks off-target verdicts (a
reviewer built to find diff-level problems has nothing to object to in
markdown) and dilutes the architecture seat's signal for the platform-code
case that most needs it.

### 3.3 A periodic staleness sweep for `bugs_open/` — the owner's addition to this discussion

The bar for closing a bug is already "fixed AND live" (016b §10), and this
repo already has the pattern "a closed bug's scope-out EXPIRES" as a *lesson*,
not a *check*. Proposal: a scheduled task, same family as
`build-pipeline-trigger`, that periodically re-examines open `bugs_open/*.md`
entries and flags (does not auto-close) ones whose cited evidence has moved:

- **Cheap first pass:** does the cited `file:line` / symbol still exist,
  roughly where the bug says it does? Catches refactors and fixes-elsewhere
  mechanically, no LLM.
- **Deeper pass (sampled, not every run):** re-run the bug's own "How to
  verify a fix" query/command against live HEAD/DB and diff the result against
  what the bug file recorded. A changed result is the flag — a human or thread
  still decides whether that means fixed, regressed differently, or unrelated
  drift.
- Output is a worklist (mirroring how discovery checks create work items
  rather than silently mutating state), not an auto-close — closing still
  needs the same fixed-AND-live judgment call it needs today.

**A live operational wrinkle worth planning for, found while drafting this
RFC:** any sweep that diagnoses against "current HEAD" needs a real answer for
*which* HEAD. This session's own branch is currently 547 commits ahead of
`origin/087_towards_multiple_domains`, and the function `155` describes was
itself not yet an ancestor of `origin` at time of writing. `090`'s own script
refuses to run against a ref it can't resolve on the remote, and warns rather
than silently diagnosing a stale tree — exactly the trap a periodic sweep must
avoid by design, not by convention. Whatever cadence/ref policy the sweep uses,
it needs the same "REF=current branch, refuse rather than fall back to a stale
`main`" discipline `090` already learned the hard way (2026-07-19 entry in that
script's own header).

## 4. What's needed before any of §3 is built

Per the 2026-07-29 seam ruling: a change to what a shared review mechanism
covers is architecture-scope even when small and well-reasoned. This RFC is
the registration; it is not asking to be self-ratified by the session that
wrote it. Concretely, before code:

1. Owner decides yes/no on 3.1 (practice-only, near-zero cost to adopt).
2. Owner decides the shape of 3.2 (sync-script spot-check vs. dedicated
   verifier agent) — or declines it, accepting the feedback-loop risk as-is.
3. Owner decides whether 3.3 is wanted at all, and at what cadence, before a
   thread builds and registers it as a new scheduled mechanism in the concept
   register.

## 5. Consumers to tell, per the 2026-07-29 ruling's third answer

Anyone who currently reads `LANDMINES.md`'s `doc_notes` sync (council seats via
`schema_hint`) or the `bugs_open/` index (016b §10, `who-owns.py`,
`098_REPORT_unreviewed_commits`) should know this RFC exists before 3.2/3.3
change what those artefacts mean — not because a collision is measured, but
because their owners should get to object on the same terms the guardian seat
asked for in RFC_002.
