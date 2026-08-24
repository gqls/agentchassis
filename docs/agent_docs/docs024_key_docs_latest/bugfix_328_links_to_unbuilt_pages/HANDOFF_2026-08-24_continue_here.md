# HANDOFF 2026-08-24 — bug 328: LIVE and proven in the binary; ONE acceptance check outstanding

**Read this box, then §1. Everything else is background.**

> ## STATE
>
> **The fix is LIVE.** Chassis `v1.0.1334`, Go binary-proven on **both** replicas; migration `575`
> applied by hand 16:07Z and read back. Council **APPROVED at round 4**, corr
> `21c19c1f-e614-49bd-82ac-0bb5b58082e0`. Register **LNK-038**. `RFC_049` opened.
>
> **The bug is NOT closed, and exactly one thing is owed: the acceptance check at the served
> bytes.** A canary re-render is queued (item `b18a0287`, loanzy.uk `index`). When it completes,
> run §1. That is the whole remaining task.
>
> **⚠ BLOCKED ON ONE THING, AND IT IS NOT THIS FIX: the kubeconfig token expired at ~16:40Z
> 2026-08-24.** Every `kubectl` call returns `You must be logged in to the server (Unauthorized)`,
> fleet-wide (`kubectl get nodes` fails too, so it is expiry, not a query fault). **The owner
> refreshes it** — that is the standing arrangement; do not go looking for credentials.
>
> Until it is refreshed, §1 cannot run: it needs both a DB read and the item's status. The canary
> was still `triaged` at ~16:15Z with 3 items ahead of it and that site's dispatch loop running
> roughly every 30 minutes, so it has most likely dispatched in the meantime — **check its status
> first, do NOT re-file it.** Its `item_key` is
> `page_rerender_index_55213ded-03ec-40f7-8fc1-169de05e05c8_assemble`; a second row with that key
> would be refused by the dedup index anyway while the first is non-terminal.
>
> **Nothing else is blocked, and nothing needs a decision.**

## 1. THE ONE THING OWED — the acceptance check

The canary item re-renders loanzy.uk's home page, the bug's headline instance. When it reaches a
terminal status, fetch the page and assert **both halves in the same fetch**:

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At \
  -c "SELECT status, COALESCE(left(error,200),'-') FROM site_work_items WHERE left(id::text,8)='b18a0287';"

curl -s --max-time 20 "https://loanzy.uk/?cb=$RANDOM" | grep -o 'href="[^"]*"' | sort | uniq -c | sort -rn
```

| assertion | before (captured 2026-08-24 ~16:10Z) | required after |
|---|---|---|
| `href="/your-rights.html"` | 2 | **0** |
| `href="/guides/index.html"` | 1 | **0** |
| `href="/calculators.html"` | 5 | **still 5** ← THE POSITIVE CONTROL |

⚠ **The third row is not optional and is the whole test.** Without it, a fix that stopped emitting
internal links altogether passes — a state `bugs_open/313` (closed 2026-08-19) shows this platform
can reach and not notice.

Then confirm the suppression recorded itself:

```sql
SELECT error_message, context FROM agent_error_log
WHERE error_code = 'CONTENT_LINK_SUPPRESSED_UNSHIPPED'
ORDER BY created_at DESC LIMIT 5;
```

**If it passes:** append the proof to `bugs_open/328`, NOTES and `README_where_we_are`, and let the
fleet's own cadence carry the remaining pages (see §2). The bug then closes when the served
population is clean.

**If the canary FAILED to build** (not "suppressed nothing" — actually failed): read `error`. The
predicate was pre-flighted against loanzy's real rows and is correct (§4), so a failure is more
likely the re-render path than the suppression.

## 2. Why 28 dispatches were NOT fired — do not "finish the job" by firing them

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
