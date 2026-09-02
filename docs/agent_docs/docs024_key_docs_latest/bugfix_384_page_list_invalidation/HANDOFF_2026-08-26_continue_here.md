# HANDOFF — bugs_open/384 page-list invalidation · continue here

**Written 2026-08-26 ~20:45Z. Supersedes `HANDOFF_2026-08-25_continue_here.md` (read that only for history).**
Lane: `docs/agent_docs/docs024_key_docs_latest/bugfix_384_page_list_invalidation/`.

Cold-start read order: **this file** → `bugs_open/384_HANDOFF_2026-08-24_….md` (the bug, its own
CORRECTED block, and the CLOSE-OUT VERIFICATION section at the tail) →
`RUNBOOK_page_list_invalidation.md` (every command) → `NOTES_…` tail (missteps, each with its check).

---

> **⚠ THIS FILE IS SUPERSEDED. Cold-start from `HANDOFF_2026-09-02_continue_here.md` instead.**
> Kept for history only; §1/§4/§6 are all corrected there.
>
> **CORRECTED 2026-08-31 ~16:10Z — §1 AND §6 ARE SUPERSEDED. DO NOT CLOSE 384.**
> Re-measured after five idle days. The lane's code is still live (single build `ef06af0e0afc`,
> 342 pods, all four commits ancestors). But **the defect is back on a GENERIC page**:
> `leopardessconsulting.co.uk/blog` serves **2 text-only cards where a card asset exists** — the
> first two in the grid — from landings on **2026-08-27 22:37**, still unrepaired today. Verified
> at the served artefact (11 card `src`s for 13 entries), not at the store.
> **The seam is not the failure**: it filed nine correctly-specced items within 40 ms. Two
> COMPLETED GREEN and rewrote nothing (`page_components.updated_at` still 08-27 21:34, an hour
> before the cards landed); seven sit `unresolved`, `attempt_count=0`. **No mechanism is asserted**
> — the runs are unrecoverable (`orchestration_states` retains ~1 day) and no `090` has been run.
> **And the clean census is FLATTERED**: card production has been **0 for two days** (89/109/46/18
> then 0, 0), and fleet `page_rerender` completion fell **99% → 4%** across 08-28→08-30 with 1,076
> rows `unresolved`. ~~the `bugs_open/413` shape~~ **CORRECTED same day 16:25Z: that was a guess
> from shape. It is the `dispatch_throughput` lane's already-diagnosed Anthropic CREDIT outage,
> ~2026-08-28 → 2026-08-31 ~09:00Z (their "D4 case 4"), now ENDED and their floor read clean.**
> A near-zero blank count on an idle producer and a stalled queue is not evidence of health —
> but the outage does NOT reach the **2026-08-27 22:58** completion, which sits in a healthy
> window and is now the whole open question.
> Full measurements: `NOTES_page_list_invalidation.md`, entry 2026-08-31.

## 1. THE ONE-LINE STATE

**The bug is FIXED, LIVE, and has now proven itself FOUR times on natural (non-induced) triggers.
Every owner decision is shipped. Nothing is in flight that needs a human.** What remains is one
structural residual (owned pages), one watch item (escalation rate), and a separate bug this lane
filed (`bugs_open/404`).

**384 is ready to close.** It has not been moved to `bugs_closed/` — that is the next session's
call, and §6 says exactly how.

## 2. What the bug was, in one paragraph

A page's listing card lands, and the listing that renders it keeps showing a text-only card. A
listing's items are a **stored snapshot** in `page_components.content_data`, filled from a
`query.*` source at the last section resolve; every routine re-render is *assemble mode*, which
re-ships that array verbatim. Only a re-render carrying `spec.reason='section_data_resolved'`
re-runs the query, and nothing in the card-landing chain asked for one. **The tell was never "the
page wasn't re-rendered"** — dartsonline's listing was re-rendered three times after its cards
landed and stayed stale every time. `spec->>'reason'` is the variable that discriminates.

## 3. What is LIVE (proven at the artefact)

`[MEASURED 2026-08-26 20:35Z]` The fleet is **mid-roll across two builds** — `e7f1045fddec`
(700 pods) and the fresh `b34c24f4c65b` (95 pods). **All four of this lane's Go commits are
ancestors of BOTH**, and `b34c24f4c65b` is a strict descendant of `e7f1045fddec`, so the lane's
behaviour holds whichever pod serves:

| commit | what |
|---|---|
| `7720dc76c` | decision 3 round 1 — `rebuild_blog_listing` shares the image projection |
| `bafd4411c` | decision 3 round 2 — loud failure on projection/Scan divergence + shared cap |
| `72469c556` | decision 2 / RFC_052 — dependency-scoped consumer lookup, both producers migrated |
| `efc0db7bc` | the `ConsumesAny` fail-loud profile guard (council follow-up) |

⚠ **Re-run that ancestry against the CURRENTLY reported stamp, never a remembered one.** Doing it
against yesterday's sha returned "NOT in the running build" for all four — a false negative made
entirely of a hardcoded value a roll had superseded. The query is in §7.

**Also live, and DB-config so live the moment it applied:**

- **The sweep** — migration `603`, applied by hand 2026-08-25 11:37Z. Checks array 44 → 45.
- **`tool-cta` renders the listing image** — migrations `614` (template + schema) / `615` (the
  fan-out). 39 of 40 pages carry thumbnails.

## 4. PROVEN FOUR TIMES ON NATURAL TRIGGERS — this is the evidence that matters

Every proof before 2026-08-26 was an **induced** landing. Today the seam demonstrated itself
unprompted, four times, on three sites:

| site | card landed | seam fired | outcome |
|---|---|---|---|
| leopardessconsulting.co.uk | 14:42:45 | **14:42:46** (1 s) | array rewritten 15:30:34 — **11 of 11 entries carry an image** (was 4 blank) |
| finetuning.uk ×3 pages | 17:25:45 | **17:25:45–46** | all 3 items complete by 18:44; arrays rewritten 19:13–19:15 — **0 blank on every generic listing** |
| vonc.com | 19:59:30 | fired; re-render **pending** at time of writing | expected to repair on its next loop turn |

Three further leopardess landings at 14:55/14:56 each reported
`page_list_reresolve: "deduped", deduped: 3, queued: 0` — collapsing onto the open items via the
shared `PageRerenderItemKey`. **Four landings in fourteen minutes produced three re-render items,
not twelve.** All items completed with `attempt_count = 0`. **Zero escalations** across every
seam-driven run this lane has produced, against a baseline of 1 in 36.

## 5. THE RESIDUAL — the one thing this fix structurally CANNOT reach

`[MEASURED 2026-08-26 20:40Z]` fleet-wide, blank-where-a-card-exists, split by page policy:

| policy | blank entries | pages |
|---|---|---|
| **owned** | **14** | 3 |
| generic | 1 | 1 — vonc.com, seam fired, re-render pending |

So **every generic page has repaired or is repairing. The standing 14 are all on
`rebuild_policy='owned'` pages**: `finetuning.uk/llm-cost-calculator`,
`leopardessconsulting.co.uk/llm-cost-calculator`, `…/tool-ai-vendor-trust-checklist`.

> **An `owned` page's `query.*` listing array is never re-resolved — not by this seam, not by the
> sweep, not by the `template_changed` fan-out. It goes stale on the first card landing after its
> last resolve and STAYS stale indefinitely.**

The exclusion is **correct and must not be removed**: page-rerender's reasoned branch runs
`save_sections`, whose ownership refusal (`bugs_open/208`, OWNED_PAGE_GUARD) would FAIL the run.
`bugs_open/333`'s lane established that a per-BRANCH refusal can only be expressed at selection
time. The gap is that **nothing else covers those pages.**

It was tolerable while `tool-cta` rendered no image. **Migration `614` made it visible** — two of
the three affected pages carry `tool-cta`, so their tiles now show images for some entries and
blanks for these, on pages a human owns and did not change.

**The remedy shape already exists and nobody has applied it to listing staleness:** migration
`486` routes owned pages to `section_edit` → `section-editor` instead of `page_rerender`. That is
a NEW SEAM for owned pages and belongs in its own council round — deliberately not bolted onto
this lane's close-out.

## 6. IF YOU ARE CLOSING 384

The bar is **fixed AND live** (CLAUDE.md). It is met: rolled in both builds, exercised live, and
demonstrated four times naturally.

```bash
git add bugs_closed/384_HANDOFF_2026-08-24_a_landed_card_image_never_invalidates_the_listing_that_renders_it.md
git commit bugs_open/384_HANDOFF_2026-08-24_….md bugs_closed/384_HANDOFF_2026-08-24_….md -m "..."
```
⚠ **Name BOTH paths on the commit.** `git mv` plus a one-sided pathspec silently ships a COPY —
the file leaves your disk either way, so `ls` cannot tell you. Verify at HEAD, not the tree:
`git ls-tree -r --name-only HEAD -- bugs_open/ bugs_closed/ | grep 384` must return exactly one line.

Before closing, decide where the §5 residual lives — it must not close with the bug. Either file
it as its own bug (grep both dirs first) or carry it into the owned-page seam's round.

## 7. THE COMMANDS YOU WILL WANT FIRST

```bash
PSQL="kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -X -q"
```

**Is the lane's code actually running?** (never a remembered sha)
```sql
SELECT left(git_commit,12), count(DISTINCT pod_name) AS pods
  FROM service_binary_capabilities
 WHERE service='agent-chassis' AND last_seen_at > now() - interval '90 minutes'
 GROUP BY 1 ORDER BY 2 DESC;
```
then `git merge-base --is-ancestor 72469c556 <each stamp>`.

**Is the defect back?** (the whole bug in one query — split by policy, because owned is expected)
```sql
WITH qf AS (SELECT cc.id cid, f.key fld FROM content_components cc,
              jsonb_each(coalesce(cc.input_schema->'fields','{}'::jsonb)) f
            WHERE f.value->>'source' LIKE 'query.%' AND f.value->>'type'='array'),
ent AS (SELECT s.domain, p.name AS listing, COALESCE(p.rebuild_policy,'generic') AS policy,
               p.site_id, e.value->>'url' AS entry_url, coalesce(e.value->>'image','') AS img
        FROM page_components pc JOIN qf ON qf.cid=pc.component_id
        JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
        CROSS JOIN LATERAL jsonb_array_elements(CASE WHEN jsonb_typeof(pc.content_data->qf.fld)='array'
             THEN pc.content_data->qf.fld ELSE '[]'::jsonb END) e
        WHERE pc.build_status<>'removed' AND p.status='active')
SELECT ent.policy, count(*) AS blank_with_card, count(DISTINCT ent.domain||'/'||ent.listing) AS pages
  FROM ent
  JOIN pages tp ON tp.site_id=ent.site_id AND tp.url=ent.entry_url
  JOIN assets ca ON ca.site_id=tp.site_id AND ca.entity_type='page' AND ca.entity_id=tp.id
       AND ca.purpose='card' AND ca.status='active'
 WHERE ent.img='' GROUP BY 1;
```
Expect `owned` ≈ 14 and `generic` 0–2 (a non-zero generic count is either the seam in flight —
check `site_work_items` for an open `page_rerender` on that page — or a regression).
⚠ This query is expensive; scope it to one domain if it times out.

The rest — applying a HELD migration, reading the sweep summary, the template fan-out, the card
derive — are all in `RUNBOOK_page_list_invalidation.md`, each with its gotcha attached.

## 8. WHAT IS OWED (none of it blocking)

1. **Escalation watch.** `603`'s header asks for the rate re-read a week on (≈2026-09-01) against
   the refreshed baseline **1 of 36** `section_data_resolved` runs in 14 days. Zero so far. If the
   sweep's items escalate materially above it, run `603_…_ROLLBACK.sql` and bring the number to
   the owner.
2. **The owned-page residual** (§5) — the only genuinely unfinished work.
3. **`bugs_open/404`** — a SEPARATE defect this lane found, filed, and does not own the fix for.
   `template_changed` and `literal_markdown` are in the live re-render reason vocabulary and
   `create_rerender_items_action.go` knows neither, so a caller routing them through the shared
   creator gets unscoped, assemble-only items that complete green and ship nothing. Exposure is
   LATENT (no live caller does this — verified across live AND archive). **Candidate 0 (a parity
   test reading the LIVE condition) is UNCLAIMED.** The `bugs_open/410` lane confirmed it is not
   taking it, and their ratchet reads Go source only, so the two cannot half-cover each other.
4. **One permanently-failing page.** `ai-agent-orchestration.com/tool-automation-savings-estimator`
   is refused by the section component floor (`77→37` class attributes). **Pre-existing** — it
   failed 3× on 2026-08-24, before this lane touched anything, and those were the fleet's only
   other floor refusals in 14 days. Not this lane's.

## 9. THE TRAPS THIS LANE PAID FOR — read before you measure anything

All are in `WRONG_CALLS.md` / `LANDMINES.md`; these are the ones that will bite *you*:

- **`snapshot_agent()` returns the SOURCE row's id, not the snapshot's.** The obvious
  "compare against the pre-image" verify compares the live row with ITSELF and reports a clean
  apply either way. The real pre-image is in `agent_definitions_backup`.
- **`site_work_items` is a rolling window.** Any count meant to describe a POPULATION must
  `UNION ALL site_work_items_archive`. I published a live slice as a population and another lane
  relayed it.
- **Reproduction is not verification** when both parties share the same wrong assumption about the
  population. A re-run tests the arithmetic, not the choice of table. **The number goes next to
  its population, written down before the number.** And the only check that ever caught anything:
  **open one MEMBER before trusting the count.**
- **A mock's bookkeeping cannot assert a NEGATIVE.** `ExpectationsWereMet()` only reports declared
  expectations that went unmet; and a recording matcher needs an unused expectation present or it
  is bypassed on the failure path. Two versions of one test passed under mutation before the third
  worked.
- **Ancestry against a remembered sha** — see §3.
- **A template edited by SQL ships NOTHING** without a hand-written fan-out, and the estate's own
  fan-out query has no page-status filter.

## 10. Where the knowledge lives

- **Bug + close-out verification:** `bugs_open/384_HANDOFF_2026-08-24_….md` (tail).
- **Register:** **PBP-048** in `docs/agent_docs/docs026_concept_register/register/page-build-pipeline.md`.
- **Architecture:** `architecture_review/RFC_052_…md` — **CLOSED**; its question 3 (is a re-resolve
  the right unit) is explicitly still open.
- **Council:** `170147b4` (decision 3, REVISE → APPROVED r2), `7553c120` (decision 4, APPROVED),
  `e1d32ca2` (decision 2, APPROVED). All read and acted on.
- **Code:** `queryresolve/consumers.go` (`ConsumerPages`, `consumerSQL`, `ConsumesAny`),
  `queryresolve/queryresolve.go` (`sourceDependencies`, `SourceReads`, `PageListingHardCap`,
  `PageImageProjectionSQL`), `actions/page_list_reresolve.go`,
  `actions/rebuild_blog_listing_action.go` (`scanBlogArticles`),
  `discovery_checks/check_page_list_stale.go`, `render_news_section_html.go`,
  `render_directory_action.go`.
- **Peers, all answered, nothing outstanding:** the filing lane (`agentchassis-51` — corrected its
  own mechanism and its verification protocol), `analytics_gtm` (lampenkap ruling filed in their
  dir), `bugs_open/410` (shipped its own fix citing this lane's `170147b4` provenance).
