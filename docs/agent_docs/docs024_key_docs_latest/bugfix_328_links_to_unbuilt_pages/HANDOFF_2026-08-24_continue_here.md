# HANDOFF 2026-08-24 — bug 328: LIVE, and PROVEN ON THE WIRE. Nothing owed but time.

> # ⚠ SUPERSEDED 2026-08-25 — READ THIS FIRST, THEN IGNORE THE BOX BELOW
>
> **The "wait, then close" instruction and the "do NOT dispatch them" ruling in §2 are both
> RETIRED.** Re-measured 2026-08-25 ~09:50Z against the live web:
>
> - Of **21** public referring pages, the **13** that re-rendered AFTER the flag (16:07Z 08-24)
>   serve **0** dead anchors; the **8** that last deployed BEFORE it serve all **12** that remain.
>   **21 of 21, no exceptions** — the flag time predicts the served result exactly, across six
>   domains, on pages nobody dispatched. Positive control held (15–49 internal anchors survive).
> - **The cadence stopped carrying.** The fleet ran 1,671 `page_rerender` items in 36 h, but per
>   PAGE, not per site: `remortgagecalculator.uk` had **zero** queued in 36 h, `loanzy.uk`'s newest
>   was 08-24 16:15, and **none of the 8 was queued for anything**. "24 of 25 touched within 7 days"
>   is a claim about a population and cannot retire a TAIL risk — which is exactly the
>   `bug_historian` MEDIUM this lane recorded in council round 3 and then answered with a statistic.
> - Owner ruled **dispatch**. Fired 11:03:10Z as **7 inserts + 1 re-arm**. The anti-dispatch case
>   inverts at this size: it was 28 pages of which 26 were unnecessary; this is **8 of 8 necessary**,
>   at the least accumulated drift they will ever carry (19 h – 2 days).
>
> **Current state, docs and the two closure queries: `NOTES_328_links_to_unbuilt_pages.md`
> (bottom) and `RUNBOOK_328_links_to_unbuilt_pages.md` (the two queries were in NO document until
> today and had to be reconstructed from the Go predicates).**

**Read this box. Everything else is background — §1 is now the record of the passing test, not a task.**

> ## STATE — PROVEN AT THE ARTEFACT 2026-08-24 18:5xZ
>
> **The fix is LIVE and PROVEN on the wire, on two sites, with both arms and the positive control.**
> Chassis `v1.0.1334` (binary-proven, both replicas); migration `575` applied 16:07Z; council
> **APPROVED r4** `21c19c1f-e614-49bd-82ac-0bb5b58082e0`; register **LNK-038**; `RFC_049` opened.
>
> The acceptance test passed both halves. loanzy.uk: `/your-rights.html` 2 → **0**,
> `/guides/index.html` 1 → **0**, `/calculators.html` **still 5**, and every other href count
> byte-identical. The audit row (`CONTENT_LINK_SUPPRESSED_UNSHIPPED`, 17:21:10) names the three
> hrefs and shows **both arms firing on one page**. Both targets are still unbuilt, so the links
> were removed, not validated. Second, unprompted confirmation on `remortgagecalculator.uk` (0 dead,
> 17 and 15 internal anchors surviving) — the fleet's own cadence carrying it, exactly as predicted.
>
> **Nothing is blocked. Nothing needs a decision. No dispatch is owed.**
>
> ## THE ONLY THING LEFT: wait, then close
>
> **5 of the 28 referring pages have re-rendered since the flag; 23 have not**, and their stored
> HTML still holds all 48 anchors *by design* (outbound-only, so the authored href survives and the
> link returns when a target ships). They go clean as they re-render on their own cadence —
> measured, 24 of 25 within 7 days. **Do NOT dispatch them** (§2).
>
> **To close the bug:** re-run the census in §3 until the *served* population is clean, then move
> `bugs_open/328` → `bugs_closed/` per the fixed-AND-live bar. Re-measure; do not trust the numbers
> in these docs (§3).

## 1. THE ACCEPTANCE CHECK — ✅ PASSED 2026-08-24 (kept as the recipe, and as the record)

The canary (`b18a0287`) completed 17:21:23; the page deployed 17:21:17. This is the check that was
run, and it is the one to re-run on any further page. **Assert both halves in the same fetch:**

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At \
  -c "SELECT status, COALESCE(left(error,200),'-') FROM site_work_items WHERE left(id::text,8)='b18a0287';"

curl -s --max-time 20 "https://loanzy.uk/?cb=$RANDOM" | grep -o 'href="[^"]*"' | sort | uniq -c | sort -rn
```

| assertion | before (16:10Z) | required | ACTUAL |
|---|---|---|---|
| `href="/your-rights.html"` | 2 | **0** | **0** ✅ |
| `href="/guides/index.html"` | 1 | **0** | **0** ✅ |
| `href="/calculators.html"` | 5 | **still 5** ← THE POSITIVE CONTROL | **5** ✅ |
| every other href (9 distinct) | — | unchanged | **byte-identical** ✅ |

⚠ **The third row is not optional and is the whole test.** Without it, a fix that stopped emitting
internal links altogether passes — a state `bugs_open/313` (closed 2026-08-19) shows this platform
can reach and not notice.

Then confirm the suppression recorded itself:

```sql
SELECT error_message, context FROM agent_error_log
WHERE error_code = 'CONTENT_LINK_SUPPRESSED_UNSHIPPED'
ORDER BY created_at DESC LIMIT 5;
```

It returned, at **17:21:10**, six seconds before the page deployed:

```
[{"href": "/your-rights.html",  "action": "drop_control_unshipped"},
 {"href": "/guides/index.html", "action": "suppress_unshipped"},
 {"href": "/your-rights.html",  "action": "suppress_unshipped"}]
```

**Both arms fired on one page**, discriminating correctly within a single document — one
`/your-rights.html` was a classed control (dropped whole), the other prose (unlinked, words kept).

**The alternatives are ruled out, not argued away:** both targets are STILL unbuilt (so the links
were removed, not validated); every other href count is byte-identical (so the page was not
rewritten); and 9 distinct internal hrefs survived (so it did not stop emitting internal links —
`bugs_open/313`'s failure mode).

**Second, unprompted confirmation:** `remortgagecalculator.uk` re-rendered on its own cadence —
`/index.html` and `/mortgage-lenders.html` both serve **0** dead anchors while keeping **17** and
**15** internal anchors respectively, and both targets return 404 on the wire.

## 2. ~~Why 28 dispatches were NOT fired — do not "finish the job" by firing them~~

> **RETIRED 2026-08-25.** The reasoning below was sound for 28 pages of which 26 were unnecessary.
> It does not survive contact with the tail: 24 hours later the cadence had carried 13 of 21 and
> then stopped, and none of the remaining 8 was queued for anything. **8 of 8 were necessary, and
> they were dispatched.** What still stands from this section is the *cost* it names — a re-render
> carries every platform change since that page last rendered — and that is an argument for acting
> while the accumulated drift is smallest, not for waiting.

The plan said "file `page_rerender` for the 24 affected pages". **Check staleness before you do**:
**26 of the 28 had re-rendered THAT DAY**, all before the flag went live at 16:07Z. So:

- The fleet's own cadence carries almost all of this within a day, with no dispatch at all.
- **A re-render carries every platform change since that page last rendered, not just this one.**
  Firing 28 is 26 unnecessary chances to ship unrelated drift onto customer sites.

Only two are genuinely stale and worth a deliberate look later:
`leopardessconsulting/case-study-data-pipeline-companies-house` (114 days, `needs_rebuild`) and
`pool-ai-agents.internal/about` (never deployed, so not serving anyway).

Re-measure before acting — the query is in `RUNBOOK_328_links_to_unbuilt_pages.md`.

## 3. ⚠ THE CENSUS IN EVERY DOC HERE IS STALE BY ADDITION. RE-RUN IT.

| | 08-23 | 08-24 |
|---|---|---|
| dead anchors | 36 | **48** |
| referring pages | 24 | **28** |
| unservable targets | 14 | **16** |

**+12 in one day**, including `garden-tools.uk` which did not exist in the first census and arrived
with **9**. This is the bug's own self-fuelling property. Never quote a number from these docs
without re-running it.

## 4. What was already proven — do NOT re-derive any of this

- **Go live on both replicas**, binary probe with a **present-control and an absent-control** in the
  same run. ⚠ Never `strings` (absent from the images) and never a discovery grep. The
  `build provenance` startup line had already scrolled — an empty result there means "not in range",
  **not** "unstamped".
- **575 applied**: 5 runtime rows, 5 snapshots *before* any UPDATE, `$post$` passed, keys read back
  with a query the migration does not contain. All five seam steps `true`;
  `page-rerender.rerender_sections` correctly ABSENT (not a seam — it flows on to `render_page`).
- **The predicate pre-flighted against loanzy's real rows**: `/guides/index.html` (planned, stale)
  → refused; `/your-rights.html` (`needs_rebuild`, zero components) → refused — *the case the
  estate's existing `PageMayBeLinkedPredicateFor` misses*; the three deployed pages → not refused.
- **`bugs_closed/049` reconciled** (council asked): same harm, different surface, no overlap. 049 is
  chrome/nav (live `v1.0.1171`); 328 is page body. 049's closure is why chrome is out of scope here.
- **Council round 4 advisories checked, not accepted**: assemble-mode DOES apply suppression
  (`repairOutboundPageLinks` runs post-assembly, unconditionally past the skip guard); and the
  ≤10-visible-char section floor runs *before* suppression, so a dropped control cannot drop a
  section — measured, all 8 control anchors sit in components of 2,300–4,826 visible chars.

## 5. Advisories left open (none blocking)

- **N+1**: one policy query per page per build. A per-run cache is the named follow-up.
- **Optional-key blind spot**: `assemble_page` / `rerender_single_page` have no `ActionInputSpec`, so
  `575` arms a key nothing audits fleet-wide.
- **`RFC_049`** — "is this internal URL a legitimate link target?" is now hand-rolled three times
  (CLC-013, LNK-030, LNK-038). Read it before writing a fourth.

## 6. Two corrections this lane made that outlive it

- **A parked domain 200s EVERY path.** An uncontrolled 56-URL census said "19 planned pages serve
  200" — the opposite of the truth. Always fetch an invented URL **per domain, in the same run**.
- **`site_work_items` is a ROLLING WINDOW.** "72 rows, ZERO ever closed" was false; terminal rows go
  to `site_work_items_archive`, so **closing a row is what removes it from the table the census
  reads**. True: 99 rows, 26 completed. **Three of the four earlier revalidators carry the same false
  claim.** Both are LANDMINES with the one-UNION check; both in `WRONG_CALLS.md`.

## 7. Where everything lives

- **Bug**: `bugs_open/328_HANDOFF_2026-08-19_...md` (the shared account; three corrections written in)
- **This dir**: PLAN · NOTES (append-only, newest at the bottom — the cold-start read) · RUNBOOK
  (every query with its gotcha) · README_where_we_are (owner prose) · SUMMARY_2026-08-23 ·
  `submission_328_r1..r4.json`
- **Register**: LNK-038 + its index row; LNK-030 amended with the third-instance count
- **Code**: `datahelpers/links.go` (`PageLinkRefusedPredicateFor`), `datahelpers/link_suppress.go`,
  `actions/refused_link_targets.go`, `actions/revalidate_unbuilt_link.go`; seams at
  `rerender_link_repair.go` and `multipage_actions.go`
- **Migration**: `sql_for_agents/575_..._HOLD.sql` (+`_ROLLBACK`), applied, `record-only`
