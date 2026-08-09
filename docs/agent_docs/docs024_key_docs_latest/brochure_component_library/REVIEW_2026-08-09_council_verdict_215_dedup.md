# Council verdict — bug 215 dedup — APPROVED with 3 advisory objections

**Correlation `8ab18991-ee83-4048-8965-4f7990baa188`** · commit `14b1cff28` ·
decided 2026-08-09 15:23:49Z · round 1 · `gated_by_truncation: false` ·
6 seats abstained (relevance filter), 11 reported.

**Pinned here on purpose:** `diagnosis_artifacts` carries `expires_at`, and this
front has already lost two pieces of evidence to the ~24h prune. The verdict
JSON is quoted below rather than linked. Verdict: **approved**, "approved with
3 advisory objection(s) — none high-severity".

Seat verdicts: `object` — editquality, tooling_provenance, guardian.
`approve` — reuse_agent, guidelines, compliance, debug_historian, constitution,
mission, prior_art_librarian, architecture (the last four with low-severity
notes attached).

`Council-Reviewed: 8ab18991-ee83-4048-8965-4f7990baa188` is **not** written on
`14b1cff28`, and cannot be: that commit predates the verdict and forward-only
forbids an amend. It carries `Council-Submitted:` instead, which `098` resolves
to this approval at report time. That is the designed path, not a gap.

---

## Disposition of every objection

### 1. editquality (medium) — the quiet mode has no covering edit
> "dedupePlanPageRows only dedupes within one planRows batch — it cannot see
> rows already persisted from a prior plan or the live `pages` table, so the
> cited quiet-mode mechanism is untouched. … council should confirm the
> quiet-mode follow-up is tracked, not dropped."

**Accepted, and the seat is exactly right about the mechanism.** Tracked, not
dropped: `bugs_open/215` stays **OPEN** for this reason and its 2026-08-09
section states the scope boundary and names the owner (the reconciler, against
realised pages under either spelling), with the standing phantom census as its
check — **20 candidates fleet-wide today**. Nothing further owed here beyond
keeping 215 open, which is also the owner's 08-06 rule for finished bugs.

### 2. tooling_provenance (medium) — no NOTES entry recording the merge-rule decision
> "doesn't commit to leaving a NOTES entry recording the merge-rule decision
> (richer-wins/tie-keeps-first, backfill-blank-only) for the next fix on this
> file to build on instead of re-deriving it."

**Accepted; the seat could not see it because the NOTES entry was written after
submission.** Now explicit — the decision record is below, and reproduced in
`NOTES_brochure_component_library.md`:

| decision | chosen | rejected, and why |
|---|---|---|
| which entry survives | more sections wins | *keep-first* (the bug file's own wording): discards the composed page whenever the stub is emitted first, which is the live incident's shape |
| tie (equal section counts) | keep first | *fail the write*: restores exactly the whole-replan loss this bug is about |
| discarded metadata | backfill blank fields only | *no backfill* (loses a title the stub alone carried); *unconditional* (overwrites authored text) |
| lossy merge | log at Warn, proceed | *fail the write* — see guardian below; this is the one open policy question |
| where | after canonicalisation, before the tx | *ON CONFLICT DO NOTHING*: silently picks a winner by insert order, no log, no counter |

### 3. guardian (medium) — silent data loss is a policy choice, not a reviewer's call
> "Merge rule can silently discard an authored composed page's sections at
> Warn-only severity (no hard failure, no work-item filed) … this is a policy
> choice about data loss that the owning pipeline should ratify, not just this
> reviewer."

**Accepted and NOT resolved by this thread — it is an owner call, raised in the
handoff.** The narrow question: when two *composed* pages collide (both carry
authored sections), the fix keeps the richer and logs the other at Warn. The
alternative is to fail the plan write, which is today's behaviour and loses the
entire replan. My position is that proceeding is right — losing 1 page's
sections beats losing 25 pages' plan — but the seat is correct that "how much
silent loss is acceptable" is not mine to settle. Note the branch is rare
squared: it needs a collision **and** both entries composed; the observed shape
is composed-plus-stub, which loses nothing.

### 4. guardian (low) — does the insertion brush the fragile imagery/lock block?
> "confirm the insertion doesn't shift or interact with that adjacent, unrelated
> fragile block."

**CHECKED, and it does not.** The dedup call is at `write_site_plan_action.go:437-450`,
before `tx, err := params.DB.BeginTx` — it is outside the transaction entirely.
`transferDirectiveLocks` is at `:786` and `transferImageryLocks` at `:1123`,
both separate functions, neither called from nor textually adjacent to the
insertion point, and neither edited. Verified by grep at HEAD, not by eye.

### 5. reuse_agent + prior_art_librarian (low) — was existing collision-handling prior art considered?
> "unclear whether the pages-table upsert helpers (UpsertPageForRole /
> resolveNewPageConflict …) were considered as a reusable collision-handling
> pattern"; "Whether any existing generic dedup/merge helper … was considered
> and rejected for this table, vs. simply not looked at."

**Honest answer: not looked at before submitting. Looked at after — and the
seats were right to ask, because there IS a close precedent, though not the one
they named.** `platform/orchestration/actions/save_sections_dedup.go`
(`dedupSectionsBeforePersist`, from `bugs_open/156`) is the *same fix one layer
down*: a duplicate collapse at the save choke point, and its header states the
same insight in almost the same words — "SavePageSectionsAction has seven
guards … Every one of them compares the incoming set against EXISTING rows, or
against a floor. **None of them compares the incoming set against ITSELF.**"
Its signature shape is mine (`([]T, int)` — kept set plus collapsed count), it
logs loudly, and it never refuses the save.

**Not reusable as a function** — different type, and a different discriminator:
156 keys on *content identity* (and its own census proved a unique index would
be wrong there, because 11 of 12 duplicate slot-names are legitimate), whereas
this keys on *canonical name*, where a duplicate is never legitimate because
the index forbids it. So a fresh implementation was correct, but the precedent
should have been cited, and its `writeSectionDedupLog` durable-record writer is
a **better observability answer than my counter** — logged as a follow-up below.

The named `UpsertPageForRole` / `resolveNewPageConflict` pair resolves a new
name against **persisted** rows on the `pages` table; it cannot see an
intra-batch duplicate, so it is not a substitute here. Its own header already
records that these two are a known fork ("refusal sites — this helper and
applyNewPage's `resolveNewPageConflict`").

Incidental, not acted on: there are **five** separate string-dedup helpers in
this one package (`deduplicateStrings`, `dedupeStrings`, `dedupe`, `fpDedup`,
`dedupCodeChecks`). That is the fork pattern the reuse seat exists to catch,
but it is not this bug's to fix.

### 6. guidelines (low) — `output_contract` parity for the new key
> "the convention is for output_contract to enumerate output_field names
> regardless of whether anything reads them yet … Recommend a follow-up …
> that should not gate this fix."

**Accepted as a follow-up, deliberately not done here.** It is a live-config
edit to `build-site-planner`, immediate and fleet-wide, for a field with zero
measured consumers; bundling it into a code fix's tail is how config changes
arrive unreviewed. Listed below.

### 7. architecture (approve, with a class observation worth more than the objection)
> "ARCHITECTURE_SIGNAL: insufficient … approve the fix, and record that the
> architecture underneath (no centralized page-identity/canonicalisation
> authority) is not sufficient for what's coming."

The seat notes page identity is resolved independently in **at least five**
places (`page_canonical.go`, `storage.DeployedWebPath` vs
`pageFilenameFromIdentifiers`/`PageDeployFilename`, four nav-label functions of
which one is live, the adoption crawl index keyed by URL string) and that every
one will need its own local collision patch, discovered only when it crashes or
silently duplicates. It explicitly says the RFC trigger did **not** fire on this
contained fix and that forcing one would be the over-design failure mode.

**Recorded, not actioned.** This is the strongest argument yet for an
identity-resolution RFC, and it now has three independent exhibits: this bug's
three collision families, `bugs_open/214` (imagery `scope_ref` keyed on the raw
name), and the quiet mode. Raised in the handoff as a candidate for
`architecture_review/`; not opened by this thread, because the seat's own
judgement is that the point fix should not carry it.

---

## Follow-ups this verdict created

1. **OWNER CALL — ratify or reject Warn-and-proceed on a lossy merge** (guardian,
   objection 3). The only question: two *composed* pages colliding — keep the
   richer and log, or fail the write?
2. Consider a durable merge record on the 156 model (`writeSectionDedupLog`)
   rather than log-line-plus-counter, if the counter proves too thin once the
   fix is live.
3. Add `duplicate_pages_merged` to `build-site-planner`'s `output_contract`
   (guidelines, low) — live config, separate change.
4. Candidate `architecture_review/` RFC: one owned notion of canonical page
   identity. Exhibits: 215's three families, 214, the quiet mode.
