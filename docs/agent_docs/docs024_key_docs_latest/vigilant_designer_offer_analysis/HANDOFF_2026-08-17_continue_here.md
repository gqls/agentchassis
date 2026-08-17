# HANDOFF — vigilant designer + offer analyser (2026-08-17)

**COLD-START = this file + `features_open/030` (its new §10 v2 backlog) + `bugs_open/295`. NOTES
tail (08-16 → 08-17) has the evidence, the predictions-before-firing, and FOUR corrections of my
own claims — three of them made and caught inside this session. This supersedes
`HANDOFF_2026-08-15_continue_here.md`.**

**Re-run every liveness claim before acting on it, INCLUDING the ones in this file.** That warning
has been on every handoff in this lane and it earned itself again today: this session recorded a
site as lacking a premise field, and seven minutes later it had one — the site was being built
while I measured. See the WRONG_CALLS entries dated 2026-08-17; there are three.

## ⚠ UPDATE 16:15 UTC — A CHASSIS ROLLED AND DID NOT CARRY THE FIX; PREDICTION 5 IS CONFIRMED

**Read this before the rest of the file.** Pods restarted 14:42–14:43 UTC (`5bd56bdd9b`).
**`bugs_open/295`'s fix is NOT in the binary** — binary probe with two positive controls and a
negative control: the semicolon marker the fix introduces reads **0**, the pre-existing colon
variant reads **1**, the sibling guard reads **1**, a fabricated string reads **0**. The startup
`build provenance` line had scrolled on both pods, which means out of range, not unstamped.

**The likely reason, and the only action that helps: `IMAGE_TAG` was not bumped.** The fix is at
HEAD (12:12:01 UTC, `2a5798c4b`), but the fleet is deployed at **`v1.0.1305`** and the makefile's
`IMAGE_TAG` is **also `v1.0.1305`**. A same-tag rebuild ships the node's stale cached binary
(CLAUDE.md, build section). Whether it was rebuilt-at-the-same-tag or never rebuilt is
**undetermined** — I have no earlier digest to compare — but the remedy is the same either way and
it is the owner's: **bump `IMAGE_TAG`, rebuild, roll.** Re-rolling `v1.0.1305` cannot help.

**Prediction 5 CONFIRMED, and it is now a verified pre-fix baseline rather than an assumption.**
gamesdesign's `content_rewrite` on `tool-ttk-calculator` (owned) → **`failed` 13:02:18**,
orchestration `763b227b`, cause quoted verbatim as the owned-page guard, **0** `owned_page_review`
rows from `save_page_sections`. Third quoted instance, first predicted in advance, and the first
where the item was filed by the sweep and died with no session touching it. **When the fix ships,
re-test on that same site/page/item type: the bar is a row where there are provably zero.**

## The one-line state

> **B4 is enrolled, sweep-driven, and now PROVEN to grow the estate on its own** — two sites
> analysed today without a hand on them (gamesdesign.co.uk, robot-hands.com), taking
> `offer_ordering` from 3 sites to 5. **The sweep window that made that happen is CLOSED again**
> (it is the owner's cost control, not a default). `bugs_open/295` — a fleet-wide reporting gap
> this lane found in the drain — is fixed, council-APPROVED at round 1, and **inert until the next
> chassis roll**.

## What the owner decided today

1. **Fix 295 first** (over the claims-audit track, B4 v2, and the acceptance-test work). Done.
2. **Open a short sweep window** (over hand-firing named sites, or leaving it off). Done, run,
   measured, and **closed again** — see "The sweep window" below for the measured cost per site,
   which is the number he needs to decide whether to open a longer one.

## What is DONE and live

- **`bugs_open/295` — the owned-page guard in `save_page_sections` recorded nothing.** Three guards
  share the ownership policy; two file an `owned_page_review` row, the third returned a bare error.
  On `page-build-handler`'s route — which has **no `assemble_page` step** — that silent guard was
  the only one, so a content item died `failed` with its reason living only in the orchestration's
  `__step_error`, which ages out at ~24h. Two of this lane's own findings died that way on 08-15.
  **Fix `2a5798c4b`** (calls the existing `emitOwnedPageReviewItem`; the item still fails, which is
  honest). Test mutation-proven **both** ways. **Council APPROVED round 1, corr `d4f49ea5`**, 13
  reviewers, verdict read and three objections answered in the bug file rather than banked.
  Register **PBP-036** updated with the fourth producer and the shared `item_key` shape — which is
  the condition the 2026-08-02 ruling attaches to adding a producer without an RFC.
  ⚠ **INERT until a chassis roll.** Do not close it on the commit. The bar is a real
  `owned_page_review` row carrying `refused_by='save_page_sections'`; there were **zero** at filing.
- **`offer_ordering` is on 5 of 23 sites** (was 3). The two new ones were written by the sweep, not
  by a session.
- **LANDMINES**: new entry on the UTC/BST clock trap (synced, verifier armed).

## The sweep window — the number the owner needs

`improvement-sweep` is `enabled=false` again. **It is his cost control** (migration 389 quotes him:
*"it will be expensive so I am wary of costs"*), not a broken switch, and the 08-15 read-out that
said B4's findings would move "without a session firing anything" was **conditional on a window
being open** — corrected in the owner log today.

**The window as actually run: OPEN 12:14:53 → CLOSED 12:39:04 UTC — 24 minutes.** It produced
**2 sweeps, 2 B4 runs, `offer_ordering` 3 → 5 sites, and 9 new work items** (offer-analysis items
26 in total, from 17). Run 1 (gamesdesign): 5 findings → 5 items, 0 skipped. Run 2 (robot-hands):
5 findings → **4 created + 1 skipped**, the skip being dedup against an already-open item — note
the PAIR must account for ALL findings (created + skipped), not just the created ones.
⚠ **Closing the scheduler does not stop an in-flight sweep.** robot-hands' sweep was still at
`call_brief_fidelity` when the window closed and will finish on its own, dispatching its items.
That is expected, not a leak — but it means "window closed" and "no more work starting" are
different moments.

Measured this session, so it can be costed rather than guessed:
- **One site per ~15 minutes.** Sweep 1 fired 12:15:14 UTC, sweep 2 at 12:30:45 — the 900s interval
  held exactly, with `max_concurrent=2` on the `dispatch` group so sweeps overlap.
- **Every site is `audit_due`** (all 23, computed by running the gate's own `load_audit_state`
  query). So every firing runs the FULL audit chain plus B4 — the expensive shape, never the cheap
  one.
- **Each B4 run files ~4–5 dispatchable items** which the sweep then promotes and dispatches.
- **18 sites still lack an ordering ⇒ roughly 4.5 hours of open window to finish the estate.**

## What the next session should do

1. **Watch for the roll, then grade 295 at the artefact** — positive AND negative control in one
   run, because a zero is otherwise ambiguous: dispatch a content item at a known **owned** page
   (expect a new row with `refused_by='save_page_sections'`), and one at a known **generic** page
   (expect none). Running only the first cannot tell a working fix from an emit that fires
   unconditionally.
2. **PREDICTION 5 IS ARMED AND UNRESOLVED — it is free evidence, do not waste it.**
   > **UPDATE 12:59:44 UTC, minutes before this session ended:** the item moved `triaged` →
   > **`claimed`** by `build-dispatch-loop`. **The guard is being asked right now**, so the
   > `owned_page_review` count of 0 is no longer "the test has not run" — it is mid-flight, and the
   > next reader gets the answer for free. Its two siblings on the same site settled meanwhile
   > (`index` → `complete`, `tools-index` → `needs_human_review`), so dispatch is working normally.
   > **Check the item's final status AND the review count together**: `failed` + still 0 confirms
   > inertness; anything else needs the deploy state re-checked before trusting it. gamesdesign's
   `content_rewrite` on **`tool-ttk-calculator`** (`rebuild_policy='owned'`) was promoted but not
   yet claimed when this session ended. When it dispatches on the CURRENT (unfixed) binary it
   should die `failed` with **no** review row. If a row appears, the fix shipped earlier than I
   believed and my "inert until a roll" claim needs re-checking before anything else I said about
   the deploy state is trusted.
   ```sql
   SELECT w.status FROM site_work_items w JOIN pages p ON p.id=w.page_id
    WHERE p.name='tool-ttk-calculator' AND w.spec->>'audit_source'='offer-analysis';
   SELECT count(*) FROM site_work_items
    WHERE item_type='owned_page_review' AND spec->>'refused_by'='save_page_sections';
   ```
3. **The v2 batch — one migration, one re-proof** (`features_open/030` §10). (a) head-of-hero
   excerpt so page-level findings stop being hypotheses; (b) attribution in `why` clauses
   (intermittent — does not justify a migration alone); (c) `primary_model` outside the degraded
   arm's field list — **latent, no live instance, see the correction; do NOT let it motivate the
   batch**; (d) **the strongest of the four** — emit a machine-checkable predicate where one is
   expressible.
   ⚠ **(a) invalidates v1's truncation baseline** (`__truncated` absent at 104 pages). Re-run that
   check on webdesign.co.uk after (a) before trusting it anywhere.
4. **`features_open/034` — the claims audit over `site_specs` prose.** Still the owner-approved
   track from 08-14, still not designed. Unchanged by today.
5. **The four remaining offer-analysis casualties** on webdesign/gaswholesalers from the 08-15 drain
   are still `failed`/`needs_human_review`. Once 295 is live they will at least leave a record.

## The acceptance-test finding — read the correction before repeating it

I wrote on 08-16 that "nothing reads the test" and implied nobody had noticed. **The first half is
true; the second is false.** `complete_work_item_no_change.go:44-48` states the gap, costs it (the
field is free LLM prose; **10 of 15 live values name a computed property and 2 contain clauses no
probe can assess**) and names the blocker: **a producer-side contract change first.**

What survives, and is worth more than what I originally claimed:
- **A worked case beyond BOTH existing mechanisms.** On webdesign.co.uk the handler genuinely
  rewrote the hero — so the no-change gate would rightly have passed it — and the acceptance test
  was still unmet, because the new copy leads with *"Sixty-three browser tools"* and the test said
  *"before any count of tools or articles"*. **The repair reintroduced the exact fault the finding
  was filed to remove.**
- **B4 names that anti-pattern first, on every site.** `avoid_leading_with[0]` on gamesdesign is
  *"The number of tools or guides currently on the platform"* — the fourth site where it leads with
  that, and it is precisely what the downstream writer then did. **The analyser knows the rule and
  the writer breaks it**, which is the clearest argument for v2(d).
- **A census that makes v2(d) decidable:** of B4's 22 live acceptance tests, ~8 are expressible as a
  text/DB predicate, ~6 partly, ~8 judgement only — and **the one that failed is in the expressible
  set** (an ordering assertion over two positions in one string). `[CLASSIFIED BY ME]`, except that
  last point, which I read at the artefact.
  ⚠ **The trap: never emit a predicate for a judgement test.** It would grade confidently and
  wrongly, and carry a green tick while doing it.

## Watch-outs (new or changed today; the rest stand in HANDOFF_2026-08-15)

- **⚠ psql prints UTC; your shell prints BST.** Subtracting one from the other overstates every age
  by an hour, always toward alarm — a sweep idle 2m28s reads as a 72-minute stall, which is exactly
  what `bugs_open/294` looks like. **Make the DATABASE do the subtraction** (`now() - last_activity`).
  I made this error TWICE today, the second time *after* writing the landmine, and that one reached
  the owner. Full entry in LANDMINES.
- **⚠ A site with `created_at` = today, `0 pages`, or `status='active'` rather than `'deployed'` is
  UNDER CONSTRUCTION, and nothing about it is a fact about the estate.** This cost a recorded
  finding today. The `[MEASURED]` marker does not help — it certifies the number was taken, not
  that its subject had stopped moving.
- **⚠ The sweep selector is `ORDER BY s.updated_at ASC`, so ANY touch by another lane moves a site
  to the back of the queue.** A census of "which site is next" is stale within minutes; mine was
  wrong four minutes after I took it.
- **⚠ `improvement-sweep` must be DISABLED again after any window.** It is enabled by direct
  `UPDATE`, deliberately **not** a migration — a migration that enables it would re-enable the
  sweep for anyone who later runs the migration set.

## Who owns what nearby

Unchanged from 08-15, plus: **`bugs_open/295` is this lane's** and needs only its post-roll
verification. `bugs_open/208`'s lane last touched it 08-08 and is not active — 295 extends its
mechanism rather than competing. **`bugs_open/294`** (stalled orchestrations) is another lane's and
is the bug my UTC/BST error impersonated; do not file against it on a clock reading.
