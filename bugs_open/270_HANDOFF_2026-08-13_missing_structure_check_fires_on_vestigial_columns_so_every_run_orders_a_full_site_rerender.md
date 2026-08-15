# 270 — `missing_structure` tests columns that are empty on every page, so every run of it orders a full-site rerender that fixes nothing

**Filed 2026-08-13** by the `portfolio_positioning` build-out session, found while
answering a council objection about an ADJACENT check (round 2 of
`check_site_structural_validity.go`'s review, corr `51cb66fb…`, whose reviewers
correctly flagged that a comment describing this check's predicate collided with the
standing vestigial-columns landmine — verifying THAT surfaced THIS).

> **On the 2026-07-31 ruling (a cross-cutting root-cause claim goes through `090`, or
> the filer states why first-hand verification substitutes): substituted.** The
> mechanism is four lines of SQL quoted below from the check's own source; the false
> premise is a single fleet-wide query re-run today (683/0/0); the consequence is a
> `site_work_items` census by `item_key`, also below, showing repeat firings and
> completed dispatches per site. There is no not-where-you-are-looking cause left for
> a loop to find — every link in the chain is read or measured directly, and the
> standing landmine had already established the premise half on 2026-08-03.

## The defect

`platform/orchestration/actions/discovery_checks/check_missing_structure.go:94-100`:

```sql
SELECT p.id::text, p.name, p.url,
       p.rendered_header IS NULL as missing_header,
       p.rendered_footer IS NULL as missing_footer,
       p.rendered_head IS NULL as missing_head
FROM pages p
WHERE p.site_id = $1
  AND p.status IN ('active', 'deployed')
  AND (p.rendered_header IS NULL OR p.rendered_footer IS NULL)
```

Both halves of the predicate are broken, each by an already-documented landmine:

1. **`rendered_header/rendered_footer IS NULL` is TRUE for every page in the fleet.**
   Chrome lives in `site_components`, not these `pages` columns; the columns are
   vestigial (LANDMINES.md, "`pages.rendered_header` / `rendered_footer` /
   `rendered_head` are VESTIGIAL", 2026-08-03). Re-verified live 2026-08-13:
   ```sql
   SELECT count(*), count(*) FILTER (WHERE coalesce(length(rendered_header),0)>0),
          count(*) FILTER (WHERE coalesce(length(rendered_footer),0)>0) FROM pages;
   -- 683 | 0 | 0
   ```
2. **`p.status IN ('active','deployed')` matches everything not archived**, because
   `pages.status` never takes the value `'deployed'` (LANDMINES.md,
   "`include_statuses: [\"deployed\",\"active\"]` filters `pages.status`, where
   `'deployed'` NEVER OCCURS", 2026-08-03 — different caller, same wrong column).

So `findPagesWithMissingStructure` returns **every non-archived page of any site it
is pointed at, always**, and the check's response to a non-empty result
(`check_missing_structure.go:53-77`) is one `needs_rerender` work item for the whole
site — `HandlerAgent: "rerender-pages"`, `severity: "high"`, `priority: 30`,
`refresh_site_components: true`, reason string *"Pages deployed without header/footer
— likely built before site_components were rendered"* — a full-site reassembly, on a
diagnosis that is false on its face.

## It is live and firing — this is not a latent defect

The check is named in `completeness-discovery-agent`'s live `checks` config
(`agent_definitions`, verified 2026-08-13), and:

```sql
SELECT status, count(*) FROM site_work_items
WHERE item_key='missing_structure:rerender' GROUP BY 1;
--  cancelled 2 · complete 25 · deferred 2 · detected 2 · unresolved 12   (43 total)

SELECT min(created_at), max(created_at) FROM site_work_items
WHERE item_key='missing_structure:rerender';
--  2026-04-24 → 2026-08-12 17:39  (fired again YESTERDAY)
```

Per-site recurrence (top of the census): `robot-hands.com` filed 5×,
`webdesign.co.uk` 4× (twice in one day, 2026-08-11), **`dartsonline.com` filed 3× and
completed 3×** — three full-site rerender cycles executed on the false premise. 25 of
43 items reached `complete`, i.e. ~25 whole-site reassemblies fleet-wide were
dispatched and done for nothing. The dedup key (`missing_structure:rerender`, per
site) holds exactly one item open at a time — and the moment one completes, the next
discovery pass re-files it, which is what a permanently-true predicate guarantees.
That the churn is periodic rather than continuous is the dedup key working, not the
check being right.

## Why it looked fine for four months

The item's summary reads like a plausible, even virtuous, finding ("N pages missing
header/footer — need reassembly"), the rerender it orders genuinely completes, and an
assemble-mode rerender is near-idempotent — it redeploys what is stored, so nothing
visibly breaks. A check that is wrong in the "always fire" direction on a repair
that is harmless-looking produces no symptom anyone chases. `bugs_closed/018`
(2026-07-19) even recorded the mismatch in passing — *"claims 9 pages deployed
without header/footer"* against a site whose chrome demonstrably served — and moved
on, because its own bug was elsewhere.

## Costs, stated without inflation

- ~25 completed full-site rerender dispatches (and their LLM-adjacent pipeline load)
  since April, each `severity: high, priority: 30` — jumping ahead of real work in
  every queue it sat in.
- 12 `unresolved` + 2 `detected` items currently polluting `site_work_items` with a
  false diagnosis whose reason string actively misleads whoever reads it
  (`bugs_open/235` and `bugs_open/232` both had to reason around one).
- Every future site (the ~145-domain build-out this session is preparing) inherits
  one spurious high-priority rerender cycle per discovery pass on day one.
- **The subtler cost**: any rerender-churn investigation (e.g. the `stale_chrome`
  wave class, `bugfix 117`) has this as a standing confounder — a periodic,
  unexplained whole-site rerender source that no one has been attributing.

## Fix candidates, ordered by what closes the door

1. **Retype the predicate to the real chrome store** — a site's chrome health is
   `site_components` (`slot_name` in header/footer/head, written by
   `render_site_components`): flag when a site has ZERO such rows (the landmine's own
   "real no-chrome signature") or when every row is `build_status='pending'`. This
   keeps the check's purpose (its 2026-04-24 origin predates the `site_components`
   migration, when the `pages` columns were presumably real) and makes it able to be
   false. The status filter should also move to the column it means
   (`build_status`), or be dropped — a page's lifecycle status is irrelevant to
   whether the SITE's chrome exists.
2. **If nobody can name what the check should now assert, retire it** — deregister
   from `completeness-discovery-agent`'s checks array (config, live immediately) and
   delete the file. A check that cannot discriminate is not a safety net, it is a
   scheduled false alarm with a dispatch arm.
3. **Either way, cancel the 14 open items** (`unresolved`/`detected`) — they carry a
   false reason string and a live dispatch route to `rerender-pages`.
4. **Reject: narrowing the predicate to `length(...)>0` checks on the same columns.**
   The columns are not "sometimes stale"; they are structurally unwritten. Any fix
   that keeps reading them preserves the defect with better manners.

## How to verify a fix

- Unit: the retyped predicate must return ZERO findings for a site with healthy
  `site_components` rows (e.g. any currently-serving site), and findings for a
  synthetic site with none — the current check cannot pass the first half.
- Fleet: after one full discovery rotation, `item_key='missing_structure:rerender'`
  (or its successor key) must stop accruing new rows on sites whose chrome serves.
- The 016b §9 rule applies: a rerender-churn measurement before/after is the
  behavioural proof — expect the periodic per-site `needs_rerender` cycle to stop.

## Relations

- LANDMINES.md "`pages.rendered_header` … are VESTIGIAL" (2026-08-03) — established
  the premise half; this file adds the live consequence (the check FIRES on it, with
  a dispatch arm) which the landmine's "read by exactly one caller left in the tree"
  line pointed at but did not measure.
- LANDMINES.md "`include_statuses` … `'deployed'` NEVER OCCURS" — the same wrong
  column, in this check's OTHER filter.
- `bugs_closed/018` — recorded the false claim in passing, 2026-07-19.
- `bugs_open/232` (`check_missing_structure.go:96` cross-ref), `bugs_open/235`
  (webdesign.uk's `missing_structure` item) — both had to reason around items this
  check filed.
- `check_site_structural_validity.go` (`head_essentials_missing`'s header) — the
  council-review thread whose round-2 objection surfaced this; its comment now
  points here. That comment's cited counts (43/25) are now stale as of the fix
  below; not edited here — it belongs to a different, actively council-reviewed
  workstream (`portfolio_positioning`), flagged for them rather than hand-patched.
- `bugs_open/280` — a second, previously-undocumented caller of the same
  vestigial `pages.rendered_header/rendered_footer` columns, found while
  fixing this bug (`check_decision_guards.go`'s stored-assembly SQL). Filed
  separately — different failure shape, out of scope for this fix.

## UPDATE 2026-08-15 — fix committed, council-submitted, NOT YET SHIPPED

Picked up by the `bugfix_270_missing_structure` session (this bug had been
explicitly left "unowned" by the filing session). Re-verified live before
starting: still firing, worse than at filing (50 items / 31 complete vs. 43 /
25 two days earlier).

**Fix:** retyped the predicate to `site_components` (real chrome store),
keyed on non-empty `rendered_html` per slot, not `build_status`. Kept the
check's name/item_type/item_key unchanged so the ~17 open false-positive
items self-close via the framework's own `CheckResult.Resolved` mechanism
(RFC_010) on each site's next discovery pass — no manual cleanup needed.
Full design, the weighing against retiring the check outright, and a marked
correction to this bug file's own original fix sketch (`build_status=
'pending'` is NOT a safe "missing" signal): `docs/agent_docs/docs024_key_docs_latest/bugfix_270_missing_structure/PLAN_2026-08-15_missing_structure_check.md`.

- Committed: `fdc5daec1` (`check_missing_structure.go` +
  `check_missing_structure_test.go`, 5 new tests, all pass).
  `Council-Submitted: 524ff897-b697-4c5c-a66f-8939b0457049`.
- **UPDATE 2026-08-15, later same day — council verdict: APPROVED.** Read
  directly (`doc_notes`, categories `? 'council-gate'`, body confirmed
  against the submission correlation, not assumed from a `LIMIT 1` on a busy
  shared queue): *"COUNCIL GATE — APPROVED — approved with 5 advisory
  objection(s) — none high-severity (round 1)"*. 13 reviewers, 4 abstained, 0
  unreadable, not truncation-gated. Per CLAUDE.md this means the commit is
  credited automatically by the `098` coverage report once it resolves the
  correlation — **no amend, no new commit needed**; `Council-Submitted:` was
  the correct trailer to have used and stays as-is (forward-only forbids
  rewriting it to `Council-Reviewed:` after the fact). The 5 advisory
  objections' individual text was not retrievable from `diagnosis_artifacts`
  for this correlation (only `council_report` summary + the echoed
  `fix_plan` exist there) — not chased further given none were high-severity
  and the gate is advisory only.
- **Still NOT shipped.** This bug stays OPEN (not moved to `bugs_closed/`)
  until an image builds from this commit, the fleet rolls, and one full
  discovery rotation is observed closing the stale items and going quiet on
  healthy sites — CLAUDE.md's fixed-AND-live bar, not "committed" or
  "approved" alone.
- Continue from: `docs/agent_docs/docs024_key_docs_latest/bugfix_270_missing_structure/HANDOFF_2026-08-15_continue_here.md`.
