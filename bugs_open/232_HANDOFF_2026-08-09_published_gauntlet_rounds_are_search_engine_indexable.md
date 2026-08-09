# 232 — Published Gauntlet rounds are search-engine indexable: nothing sets `noindex`

**Filed 2026-08-09 by the `provocation_pipeline` lane. Owner: `gauntlet_dead_cta`**
(they own `internal/tools-api` and the round-record page).

**Severity: low today, high the moment a stranger publishes a round about a named
person.** Filed separately from `architecture_review/RFC_020` because it is true
*now*, is a two-line fix, and does not depend on any of RFC_020's open questions.

## Symptom

`vonc.com/tools/gauntlet/round.html?r=<slug>` serves a visitor's own prose and an
AI verdict at a permanent public URL, and **nothing instructs search engines not to
index it**. A published round is therefore discoverable by searching its text —
including, if a visitor writes about a real person, by searching that person's name.

## Evidence [MEASURED 2026-08-09]

```
grep -rniE "noindex|robots" internal/tools-api/          → 0 hits
grep -rniE "noindex|robots" .../gauntlet_dead_cta/round_record/  → 0 hits
```

Both excluding tests. **First-hand verification, stated per the owner ruling of
2026-07-31:** this is an absence check over the two locations that could carry the
header or the meta tag, not a structural root-cause claim, so it is
self-evidencing and no `090` run was made. If the tag is set somewhere neither grep
covers — a Caddy header, a CDN rule, the page's stored `html_template` — **that
would refute this file and should be recorded here**; the check is
`curl -sI https://vonc.com/tools/gauntlet/round.html | grep -i x-robots` plus a
`curl -s … | grep -i noindex` on the served body.

## Why it matters more than it looks

Discoverability is the largest single multiplier on harm from user-generated
content. Something reachable only by a link you were handed is a contained problem;
the same words returned by a name search are not. This is the cheapest available
reduction in that exposure and it costs **nothing** in reach — a shared link works
exactly as before, which is the point: it removes the harm multiplier without
touching the viral mechanism the owner wants to keep.

## Fix candidates, ordered by what closes the door

1. **`X-Robots-Tag: noindex` response header on the published-round route**
   (`GET /round/:slug` and whatever serves `round.html`). Best: it cannot be lost by
   a page re-render, and it covers the JSON endpoint as well as the page.
2. `<meta name="robots" content="noindex">` in the round-record component's
   `html_template`. Works, but lives in stored component content, which this estate
   has repeatedly found can be regenerated away — see the chrome/rerender family.
3. `robots.txt` disallow. **Weakest — do not rely on it alone**: it asks crawlers
   not to fetch, does not prevent indexing of a URL discovered elsewhere, and is
   itself a public list of the paths you did not want looked at.

Recommend **1**, optionally with 2 as belt-and-braces.

## How to verify the fix

At the artefact, not the tag or the commit:

```sh
curl -sI 'https://vonc.com/tools/gauntlet/round.html?r=<a published slug>' | grep -i x-robots
curl -sI 'https://tools.apis.uk/api/v1/tools/gauntlet/round/<slug>' -H 'Origin: https://vonc.com' | grep -i x-robots
```

Both should carry `noindex`. **Check a real published slug, not a 404** — a missing
route returns no header either, which reads identically to a fix that is not there.

## Related

- `architecture_review/RFC_020_third_party_harm_in_the_gauntlet_before_and_after_publish.md`
  — the wider question this was found inside. RFC_020 §5.1 recommends this fix be
  made **independently of** its own open questions.
- `bugs_open/139` — poster identity is a constant, so published rounds are
  effectively anonymous. Relevant because anonymity plus indexability is the
  combination that makes a takedown request the only remedy.

---

# FIX BUILT + COMMITTED 2026-08-09 — `c3d7841f9`. Migration LIVE; Go half INERT until the next chassis roll

Taken up by a separate session from the filing lane (`provocation_pipeline` /
`gauntlet_dead_cta`), on this file's own invitation that it is independent of RFC_020.
**Still OPEN**, and it must stay open: the defect is reproducible on the live page
until the chassis rolls and the page is re-rendered.
Council `Council-Submitted: 1139cbbe-3173-4886-846b-c25daeeda93c` — **verdict not yet
read.** Lane docs: `docs/agent_docs/docs024_key_docs_latest/bugfix_232_gauntlet_noindex/`.

## Fix candidate 1 was NOT available, and candidate 2's stated weakness is not the one that mattered

This file recommends **1** (`X-Robots-Tag` on the route). **For the HTML page that is
impossible in this repo, and the reason is worth recording because it also rules out
any future variant of the same idea.** Measured 2026-08-09:

```
curl -sI https://vonc.com/tools/gauntlet/round.html
  HTTP/2 200 · server: cloudflare · x-amz-request-id: … · x-amz-version-id: …
```

The static site is **B2 objects behind Cloudflare**. There is no server-side request
path for static assets that this repo controls — the only Caddy configs here
(`gauntlet_dead_cta/infra/…`) front **tools-api**, a different service on a different
host. A response header on the page would have to be a Cloudflare dashboard rule:
outside the repo, outside review, unrebuildable. So candidate 1 survives **only for
the JSON endpoint**, which is where it has been applied.

Candidate 2 (a meta tag) is this file's second choice, warned against because stored
component content "can be regenerated away". **That warning is right and its stated
mechanism is the wrong one here.** The tag is not in stored component content at all —
the round-record component's `html_template` is a body/style/script **fragment with no
`<head>` element in it** (`SELECT count(*) FROM page_components WHERE
component_id='71a54cc2-…'` = 1; grep of that template for `<head` = 0 hits), so it
*cannot* carry a head-scoped tag. The tag is injected at **assembly** instead, from a
DB flag — which is regeneration-proof on the path that honours it, and completely
absent on the path that does not. See the landmine below: that, not component
regeneration, is the real decay path.

## What shipped

| piece | where | state |
|---|---|---|
| `pages.noindex boolean NOT NULL DEFAULT false` | `sql_for_agents/352_pages_noindex_flag.sql` (+ `_ROLLBACK`) | **APPLIED 2026-08-09**, row-verified |
| `PageInfo.Noindex` + `getPageInfo` reads `p.noindex` | `helpers.go`, `rerender_single_page_action.go` | committed, **inert until roll** |
| `injectRobotsNoindex` + call-site gate (assemblePage step 5a3) | `rerender_single_page_action.go` | committed, **inert until roll** |
| 5 unit tests, mutation-proven both ways | `inject_robots_noindex_test.go` | green at committed HEAD |
| `X-Robots-Tag: noindex, nofollow` | tools-api `PublicRoundHandler` | committed, **NOT live** — ships from the island VM, not the chassis |
| concept register SEO-003 + index row | `docs026_concept_register/register/seo.md` | committed (index row swept into `e42000bb0`) |

The gate is at the **call site** (`if page.Noindex`), not inside the helper — owner
ruling 2026-08-02 §2: new authority on a shared seam ships as an opt-in field with the
unsafe default OFF, where a reviewer of the *caller* can see it.

**Measured, so the blast-radius claim is a query and not an argument** (at apply time):

```sql
SELECT count(*) FILTER (WHERE noindex) AS is_true, count(*) FILTER (WHERE NOT noindex) AS is_false,
       count(*) FILTER (WHERE noindex IS NULL) AS is_null, count(*) AS total FROM pages;
--  1 | 629 | 0 | 630
SELECT p.id, s.domain, p.url, p.noindex FROM pages p JOIN sites s ON s.id=p.site_id WHERE p.noindex;
--  4629451e-e4f2-4fe2-b258-35107b5cb51e | vonc.com | /tools/gauntlet/round.html | t   (row IDENTITY, not a count)
```

## ⚠ THE FIX IS NARROWER THAN IT LOOKS — two head producers, one honours the flag

`assemblePage` (`rerender_single_page_action.go`, the `page-rerender` path) injects the
tag. **`AssemblePageAction` (`multipage_actions.go`) does not, and it is live:**

```sql
SELECT type FROM agent_definitions WHERE is_active AND COALESCE(is_snapshot,false)=false
  AND deleted_at IS NULL AND default_config::text ILIKE '%"action": "assemble_page"%';
--  site-work-orchestrator · pageflow-builder · page-rebuild        [MEASURED 2026-08-09]
```
```bash
grep -n "injectRobotsNoindex\|injectCanonicalLink\|injectPageJSONLD\|spliceMetaDescription" \
  platform/orchestration/actions/multipage_actions.go     # no hits at all
```

So a rebuild through that path regenerates the page **without** the tag while
`pages.noindex` still reads `true` — the wrong result looking exactly like the right
one. **`noindex` is the THIRD head fix to land on one path only** (JSON-LD 2026-07-28,
canonical 2026-08-02). Deliberately not widened into this bug: it is a pre-existing
divergence and an architecture-scope convergence question. Recorded as a `LANDMINES.md`
entry (*"There are TWO page-head producers…"*, synced to `doc_notes`, 14 footprints)
and as SEO-003's open review question. **A future thread may reasonably promote it to
its own bug file — the evidence above is sufficient; grep first, it may already exist.**

**And a stale comment argues the opposite.** `inject_canonical_link_test.go`'s header
asserts *"assemblePage is the single live assembly path … no live agent uses
assemble_page"*. The query above refutes it. Left in place rather than edited, so the
contradiction stays visible where the landmine can be found from it.

## Not done — what the next session owes

1. **The roll.** Go half is inert. After it: pod-grep **every replica**
   (`strings /app/agent-chassis | grep -c injectRobotsNoindex`, expect ≥1; run
   `grep -c injectCanonicalLink` in the same exec as a **pipeline positive control**,
   since nothing was removed there is no negative-string control available).
2. **Then re-render the page** — RUNBOOK `gauntlet_dead_cta` §18, rerender+verify steps
   only (the component is untouched, so skip regenerate/deliver):
   `scripts/rerender_page_vonc.sh 4629451e-e4f2-4fe2-b258-35107b5cb51e`.
   Wait ≥300s after any chassis pod restart or the dispatch is silently dropped.
3. **Verify at the SERVED ARTEFACT, cache-busted** — never at the flag, the tag or the
   commit, because Cloudflare fronts the origin:
   `curl -s "https://vonc.com/tools/gauntlet/round.html?cb=$(date +%s)" | grep -io '<meta name="robots"[^>]*>'`
   Then the **no-op control on the new binary**: re-render some other vonc page and
   assert the tag is ABSENT. A page rendered *before* the roll proves nothing.
4. **tools-api is a separate deploy** and is NOT live from this commit — island VM,
   rebuild + `docker save|load` + `compose up -d` (RUNBOOK `gauntlet_dead_cta` §5).
   Owner-adjacent (SSH to another host); deliberately not done unasked. Verify:
   `curl -sI 'https://tools.apis.uk/api/v1/tools/gauntlet/round/<real-slug>' -H 'Origin: https://vonc.com' | grep -i x-robots`
   — **use a real published slug**, a 404 returns no header either and reads identically.
5. **Read the council verdict** (`1139cbbe-…`) and act on a REVISE/REJECTED. The three
   things reviewers were explicitly asked to check are in the submission's `risks`: the
   migration-before-commit ordering, the idempotency-marker judgement, and whether the
   two-producer divergence should have been in scope after all.
6. **Landmine verification deliberately NOT dispatched, and this is the reason.** The
   entry's own `injectRobotsNoindex` footprint is **not in `code_symbols` yet** —
   measured with positive controls in one query: `injectRobotsNoindex` 0,
   `injectCanonicalLink` 1, `AssemblePageAction` 1, total 5755. The index predates this
   commit, so a verdict now would be `bugs_open/223`'s documented false-STALE against a
   correct entry. Re-dispatch after the index refreshes. Note also that the entry's
   NEEDS_VERIFICATION signal was **already consumed** by a concurrent session's
   `landmines-sync.py --apply` (223's second trap) — 0 verification rows exist for it,
   so it will not surface on its own.

## §"How to verify the fix" above is right and one line of it needs adding

Its two `curl -sI` checks cover the API endpoint. For the **page**, `-I` is not enough —
the meta tag is in the body, so use the cache-busted body grep in step 3 above.

---

## CONFIRMED at the artefact, and the fix candidates CORRECTED [2026-08-09]

**The defect is real.** `curl -sI https://vonc.com/tools/gauntlet/round.html` →
`HTTP/2 200`, **no `x-robots-tag`**; and the served body carries **no `<meta
robots>`**. The page is indexable. (`robots.txt` exists but carries only Cloudflare
content-signal boilerplate — no `Disallow` for this path, and a `Disallow` would be
the weakest remedy anyway, per candidate 3 above.)

> **CORRECTED 2026-08-09 — fix candidate 1 as originally written was wrong about
> WHERE the page comes from, and would not have worked.** It said to put
> `X-Robots-Tag` on "the published-round route … and whatever serves `round.html`",
> on the assumption that `tools-api` serves both. It does not.

**What the response headers actually say:** `server: cloudflare`, `cf-ray`, plus
`x-amz-id-2` / `x-amz-version-id` — the page is **static HTML in a Backblaze B2
bucket behind Cloudflare**, `last-modified: 2026-08-03`. `tools-api` serves only the
JSON at `tools.apis.uk/api/v1/tools/gauntlet/round/:slug`. **So no Go change in
`tools-api` can put a header on the indexable surface.**

**And the obvious platform route is a dead end:** `pages.rendered_head` is **NULL**
for this page (`4629451e-e4f2-4fe2-b258-35107b5cb51e`, `gauntlet-round-record`,
`/tools/gauntlet/round.html`, `page_type='tool'`, `build_status='deployed'`), and
`grep -rn "rendered_head\|RenderedHead"` over `platform/` and `internal/` finds it
read **only by two `discovery_checks`** — `check_missing_structure.go:96` and
`check_decision_guards.go:95`. **Nothing writes it into the served HTML.** Setting
it would satisfy a discovery check and change nothing a crawler sees, which is a
worse outcome than doing nothing because it looks like a fix.

### Corrected fix candidates

1. **An edge response header — `X-Robots-Tag: noindex` on `/tools/gauntlet/round.html*`
   via a Cloudflare Transform Rule.** Now the recommended fix. Immediate, needs no
   re-render, no code, no deploy, and it cannot be regenerated away by the renderer.
   **Owner action** — it is dashboard/API access, not in this repo.
2. **A renderer change emitting `<meta name="robots" content="noindex">` into the
   head** for pages that opt in. Correct long-term home, but it is a **fleet-wide**
   change to shared head assembly — most pages SHOULD be indexed — so it needs a
   per-page or per-`page_type` flag and belongs in the architecture track, not in
   this bug. **Do not bolt it on unconditionally.**
3. `X-Robots-Tag` on `tools-api`'s `GET /round/:slug` — still worth doing for the
   JSON surface, still small, but it is **not** the indexable surface and must not
   be mistaken for closing this bug. Deploys to the island VM.

### ⚠ A hazard that makes candidate 1 more attractive than it looks

The page was last built **2026-08-03**. Any fix routed through a re-render pulls in
**every** change to the layout, chrome and component library since then — the
standing "a stale page holds every improvement since it rendered" trap. Candidate 1
avoids touching the artefact at all, which on a page that is currently correct is
the safer trade.

### How to verify (unchanged, and it must be a REAL published slug)

```sh
curl -sI 'https://vonc.com/tools/gauntlet/round.html?r=<a real published slug>' | grep -i x-robots
```

A 404 returns no header either, which reads identically to a fix that is not there.
