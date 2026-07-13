# Replicating this workstream inside the chassis (no external tools)

**Why this document exists.** This thread was driven by an interactive agent with tools
the running chassis does not have: a general web browser, direct `curl`/`python` on an
operator workstation, an image-review renderer, and ad-hoc web search. The owner's
requirement: **everything done here must be reproducible by the chassis itself, or
documented so it can be.** This file maps each off-platform action to its on-platform
equivalent, and flags the few things that genuinely need a human or a new capability.

Legend: **[chassis]** already possible inside the platform · **[human]** needs a person ·
**[gap]** would need new platform work (noted, not silently assumed).

---

## 1. Reading the truth (audit) — [chassis], already how the platform works

Everything in `AUDIT_verified_facts.md` was established with SQL against the live DB and
`grep` over the repo. The chassis does this natively:
- The DB is the same `clients_db` its own agents read and write.
- The diagnose/analyser agents already read the repo via `internal/reposource` and index
  code (`code-indexer`, `118_code_indexer_for_analyser.sql`).
So an in-chassis "site truth audit" is a normal agent workflow: query specs, query
`page_components`, grep the indexed code, write findings to a spec or a work item. No
external tool was essential — the interactive tools were just faster for a one-off.

## 2. Rewriting the specs — [chassis], this is the platform's core loop

The four spec rewrites (`identity`, `voice`, `design_intent`, `portfolio`) were applied
with the same `is_current` supersede-and-insert that `WriteSiteSpec`
(`platform/orchestration/actions/site_spec_actions.go`) performs. An agent doing this
uses that action directly. **Backups**: the owner requires a row/table backup before any
spec UPDATE. In-chassis that maps to the existing snapshot mechanism
(`snapshot_agent()` / site snapshots, `031_site_snapshots.sql`); ad-hoc, we used
`CREATE TABLE bak_… AS SELECT` (the repo's own `bak_*` / `*_backup_*` convention).

Standing note for the chassis: `WriteSiteSpec` **does not honour `pinned`**. If the
platform is ever meant to protect operator-corrected specs from `content-gap-planner`
re-writes, that is a **[gap]** — a `pinned` check in the write path. Today the only thing
protecting these specs is that `improvement-sweep` is disabled.

## 3. Verifying deployed pages / assets — [chassis]

`curl`-ing the live site and asset URLs (which found the expired-presigned-URL logo bug,
D1) is exactly what the platform's discovery checks do against artifacts. The
`unfulfilled_imagery_plan` check and the deploy verification already inspect asset rows;
checking an asset URL resolves is in scope for a discovery check. No browser needed —
these are HTTP HEAD/GET checks an adapter can make.

## 4. The logo generation — [chassis] for the real asset, [human] for the choice

**What I did off-platform:** called the Gemini image API directly with the cluster's own
`BANANA_API_KEY` to make six candidates quickly, and rendered a review page to look at
them.

**Why it is still reproducible:** I used the *same model and key the chassis uses*, and
saved every prompt (`PROMPTS.md`). The permanent asset is **not** kept from this step —
it is regenerated through the pipeline:
1. **[chassis]** the routing change (`dynamic_adapter.go`: `logo/illustration/infographic`
   → Banana) is committed; deploy it.
2. **[chassis]** insert a `site_plan_imagery` row (`kind=logo`, `asset_key=logo`, the
   chosen prompt) — RUNBOOK O5.
3. **[chassis]** `image-build-handler` generates, `asset-deployer` commits to
   `/assets/images/logo.png`, set `sites.logo_url`.
4. **[human]** *choosing* which candidate is the permanent mark is a judgement call the
   owner makes (H4). The platform can generate; it should not self-approve a
   once-for-the-life-of-the-site brand decision.
5. **[gap]** favicon + OG-card derivation from the logo is not a pipeline feature today
   (no favicon/OG generator in the codebase). Documented in the PLAN; small, but real.

**The review page itself** (`review.html`) is a throwaway operator convenience, not a
site artifact. Nothing about the delivered site depends on it. It is plain self-contained
HTML openable in any browser — no server, no external tool. If the chassis needed to
present candidates to a human, that is the existing human-in-the-loop review surface
(`checkpoint_for_review`), not this page.

## 5. Web research (worldsoccernews.com) — [human], correctly

The Wayback/press research was genuine external web research. The chassis *has* web
search and scrape adapters (Firecrawl, news search) and could gather similar material,
but **the decision about how to state an unprovable personal-history claim is the
owner's**, not the platform's — and was made by the owner (H7). This is the right
division: the platform can fetch evidence; a person decides what to claim about their own
past. No attempt should be made to automate that judgement.

## 6. Positioning / engagement shape — [human]

H6/H8/H9 (engagement ladder, data-sovereignty framing, startup angle) are commercial
judgement. The platform can draft copy from a decided position; it cannot decide the
position. Recorded in `RUNNING_NOTES.md` and the specs so a drafting agent has the
decided inputs.

---

## Summary: what genuinely needs something the chassis lacks

| Item | Status |
|---|---|
| Site audit, spec rewrite, artifact verification, imagery generation & deploy | **[chassis]** — all normal platform operations |
| Choosing the permanent logo; stating personal-history claims; setting engagement terms | **[human]** — judgement, correctly not automated |
| `pinned` respected in the spec write path | **[gap]** — small, optional; only matters if agents might overwrite operator corrections |
| Favicon + OG-card derivation from the logo | **[gap]** — small; currently manual |
| Code-rendered data charts (L7) | **[gap]** — the one substantial new capability, already scoped (D1/D3/I4 + PLAN §5) |

Nothing in this thread's *outcome* depends on a tool the chassis cannot replicate. The
external tools bought speed on a one-off; the durable artifacts (SQL applied, specs in
the DB, code change committed, prompts saved) are all first-class platform objects.
