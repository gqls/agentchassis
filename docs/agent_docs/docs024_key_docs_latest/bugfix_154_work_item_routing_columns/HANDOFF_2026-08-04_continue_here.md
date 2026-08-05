# HANDOFF 2026-08-04→05 (rev 6) — council APPROVED, fix re-verified on a second build, a stale trailer fixed forward; only watch-items remain

**This revision replaces rev 5.** Rev 5's one open item (read the council
round-2 verdict) is resolved: **APPROVED**. This revision also re-verifies
the fix against a second chassis build and fixes a trailer bookkeeping gap
found while doing so. Read `bugs_open/178`'s own file in full first — it is
still the authoritative account, eight updates deep now.

## What happened since rev 5

1. **Council round 2 (`56f9a5a2-4d37-4114-9442-239861acd36e`): APPROVED**,
   decided 2026-08-04 20:27:31. The resubmission answering `bug_historian`'s
   gating objection and `prior_art_librarian`'s pointer with code evidence
   (rev 5) landed clean.
2. **A second fresh chassis build was rolled** (owner-initiated). Re-verified
   at the pod: `v1.0.1252`, both replicas (`agent-chassis-5b64b888f5-4j2bc`,
   `agent-chassis-5b64b888f5-fs4dq`), `single-unmatched-prose-slot` count = 1
   and `SECTION SHRINK` count = 2 on each. Fix survives the rebuild.
3. **Found and fixed forward a trailer mismatch**: the commit that shipped
   the fix (`4b3f9f89b`) carries `Council-Submitted: 8a3e0315-…` — the
   FIRST submission, which silently dropped and never got a verdict. The
   verdict that actually landed is under `56f9a5a2-…` (the resubmission).
   `098`'s coverage report resolves a trailer against its OWN correlation, so
   `4b3f9f89b` as committed would read as permanently unresolved. Forward-only
   forbids amending it — corrected with this revision's own commit, carrying
   `Council-Reviewed: 56f9a5a2-4d37-4114-9442-239861acd36e`.
4. **Checked for the fallback's first natural firing**: still 0 rows. Not
   informative alone given `orchestration_states`' ~24h retention.
5. Confirmed `bugs_open/154` and `bugs_open/192` — this workstream's other two
   named bugs — are both already **CLOSED, fixed, live, witnessed**. Nothing
   outstanding there; `178` is this workstream's only open thread.

## State

**The fallback fix (candidate 1 of 3) is fully closed out**: live across two
separate builds (`v1.0.1251`, `v1.0.1252`), council-approved on a clean
resubmission trail. `bugs_open/178` **stays OPEN** for two remaining items
only, neither of which has an actionable next step right now:

1. **Root-cause mechanism (candidate 3)** — investigated (090 diagnosis,
   `167d2cc2-…`) and returned genuinely **UNVERIFIABLE**, not merely unread.
   No live lead. The only thing that would move this is a fresh occurrence
   caught with an intact `page_component_history` trail (the one witnessed
   occurrence's trail didn't predate its own overwrite).
2. **The ambiguous case** (two-or-more unmatched sections or candidate prose
   slots) — unhandled by design (refuses to guess), unknown severity, not yet
   observed anywhere in the fleet.

Plus the unchanged watch-list item (five revisions running): the shrink guard
doesn't fire on a whole-slot rename — separate, narrower gap, still open,
still unobserved.

## OPEN — in priority order

1. **Watch for the fallback's first natural firing** — don't force a repro:
   ```sql
   SELECT orchestration_id, created_at, collected_data->'section_plan'->'edit_live_meta'
   FROM orchestration_states
   WHERE collected_data->'section_plan'->'edit_live_meta'->>'fallback_matched' = '1'
   ORDER BY created_at DESC LIMIT 10;
   ```
   (Note: `orchestration_states`' primary key column is `orchestration_id`,
   not `id` — a bare `id` in this query errors.)
2. **Root cause (candidate 3)** — no action until a fresh occurrence with an
   intact history trail appears. Do not re-run the 090 diagnosis against the
   same stale evidence; it already established the trail doesn't exist for
   the one known case.
3. **The ambiguous case** — unobserved; nothing to do until it is.
4. **Shrink-guard whole-slot-rename gap** — unchanged, unactioned, separate
   from everything above.

This bug is now in a genuinely low-activity state: everything actionable has
been done, and what remains needs either the passage of time (a fleet
occurrence) or a deliberate decision to manufacture a test case on production
data (previously judged out of scope). A session picking this up should
check the watch-query above first and, finding nothing new, likely has
nothing to do on `178` beyond confirming that.

## Landmines specific to this lane (carry-forward + one addition)

- All landmines from the 2026-08-03 through rev-5 handoffs still apply
  unchanged (shrink guard, dependency release rules, dispatch quiet-spell
  reading, `orchestration_states` retention, `output_field`
  shape-preservation, `bugs_open/087` vs `192` error-string collision,
  `content_components` has no site_id/page_id, "competing candidates"
  theories need a COUNT not a plausibility check, a council-trigger's printed
  correlation is not proof of dispatch, a council objection needs checking
  against code before conceding or resubmitting, a diagnosis artifact's prose
  can assert a wrong claim even when its main verdict is UNVERIFIABLE).
- **A dropped-then-resubmitted council round leaves the ORIGINAL commit's
  `Council-Submitted:` trailer pointing at a correlation that can never
  resolve, even once the resubmission is approved.** The fix is a follow-up
  commit with `Council-Reviewed: <the correlation that actually got the
  verdict>` — not an attempt to relabel the first commit, which forward-only
  forbids anyway. Check this any time a submission was resubmitted under a
  new correlation after a drop: the shipping commit's trailer is stale by
  construction, not just possibly stale.

## Cold-start pointers

- `bugs_open/178`'s own file — still the authoritative account, eight
  updates deep now.
- `NOTES_work_item_routing_columns.md`'s 2026-08-05 tail entry.
- `README_where_we_are.md`'s 2026-08-05 entries (plain-prose version of the
  same, for the owner).
- `bugs_open/154` and `bugs_open/192` are both CLOSED — this workstream's
  directory is retained for `178` only now.
- Commits: `08d0515f3` (178's original fix), `2b9d84072`/`71ecbb013` (192's
  fix + 178's correction), `4b3f9f89b` (the fallback fix — carries a stale
  `Council-Submitted:` trailer, see above), and this revision's own commit
  (carries the corrected `Council-Reviewed: 56f9a5a2-…` trailer — check
  `git log` for its hash).
