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

---

## Reading order note — the section directly above DUPLICATES the fix report, it does not supersede it

Two sessions worked this file within the same hour and appended in an order that
reads misleadingly. To be explicit:

- **"CONFIRMED at the artefact, and the fix candidates CORRECTED"** (immediately
  above) was written by the **filing** lane and landed in `4190da8b0`, but says
  nothing the **"Fix candidate 1 was NOT available"** section (§88) had not already
  established. Both sessions independently measured the same thing — B2 objects
  behind Cloudflare, so no `tools-api` header can reach the HTML page — and reached
  the same conclusion. **That is corroboration, not a later correction.** Read §88;
  the later section adds only the `pages.rendered_head` dead end.
- **The one thing the later section contributes:** `pages.rendered_head` is NULL
  here and is read *only* by two `discovery_checks`. Nothing writes it into served
  HTML, so setting it would satisfy a check and change nothing a crawler sees. Worth
  keeping because it is the obvious-looking non-fix.
- **The filing lane's recommendation is SUPERSEDED and the shipped fix is better.**
  That recommendation was an out-of-repo Cloudflare Transform Rule, on the reasoning
  that a renderer change is fleet-wide and needs an RFC. The shipped fix
  (`pages.noindex`, default `false`, gated at the call site) is the *correct* form of
  that same idea: an opt-in field with the unsafe default OFF, per owner ruling
  2026-08-02 §2 — in-repo, reviewable, rebuildable, and it needs no dashboard access.
  It is what candidate 2 should have said.
- **Nothing in the later section changes what the next session owes** — §"Not done"
  stands unmodified.

---

# COUNCIL VERDICT READ 2026-08-09 — **REVISE**, and it caught a real evidence gap plus one overstatement of mine

`1139cbbe-3173-4886-846b-c25daeeda93c`. Gated by **`editquality`** (high). 13 seats: 6
approve, 5 abstain, and objections from `editquality`, `bug_historian`, `guardian`,
`render_guardian`, `debug_historian`, `constitution`, `architecture`. **Every objection is
answered below with a measurement, not an argument.** The code is unchanged by this round
— what changed is what is *established*, plus a visible correction to a claim I made.

## 1. THE GATING OBJECTION WAS RIGHT TO ASK, AND THE FIX SURVIVES IT

> *"nothing in `grounded_in` establishes that this specific page is normally built via the
> page_rerender route rather than via `assemble_page`. If the real build path for this page
> is the other one, the whole edit is inert for the one page it targets."*

A fair hit: I proved the *general* divergence and never proved *this page's* route.
Three independent measurements, all [MEASURED 2026-08-09]:

```sql
-- (a) every work item EVER filed against this page — 3 of 3 route to page-rerender
SELECT item_type, status, handler_agent FROM site_work_items
 WHERE page_id='4629451e-e4f2-4fe2-b258-35107b5cb51e' ORDER BY created_at DESC;
--  page_rerender | complete | page-rerender     (x3, 2026-08-03).  ZERO via any assemble_page agent.

-- (b) who dispatches the action my fix is in (symmetric to the assemble_page census)
SELECT type FROM agent_definitions WHERE is_active AND COALESCE(is_snapshot,false)=false
  AND deleted_at IS NULL AND default_config::text ILIKE '%"action": "rerender_single_page"%';
--  page-rerender · report-builder

-- (c) the page's own policy
SELECT page_type, rebuild_policy FROM pages WHERE id='4629451e-…';   --  tool | owned
```

**(c) is the decisive one, and it also refutes something I wrote.** See §2.

## 2. ⚠ CORRECTION TO MY OWN §"THE FIX IS NARROWER THAN IT LOOKS" — I overstated it for this page

That section says *"a rebuild through that path regenerates the page **without** the tag
while `pages.noindex` still reads `true`"*, applied to this page. **For an `owned` page
that is wrong.** `AssemblePageAction` carries `bugs_open/208`'s ownership guard
(`multipage_actions.go:42-86`): a `rebuild_policy='owned'` page is **refused before
assembly**, returning the skip shape that `git_commit` already honours. It does not strip
the tag — it declines to touch the page at all.

So, precisely:

| page | what the other producer does | tag at risk? |
|---|---|---|
| `rebuild_policy='generic'`, `noindex=true` | assembles it, no injection | **YES — silently dropped.** The landmine stands |
| `rebuild_policy='owned'` (this page) | **refuses** at the 208 guard | No — page untouched |
| owned, but guard **fails open** | assembles anyway, logs `OWNED_PAGE_GUARD_UNCHECKED` (high) | Yes, but **countable** |

The fail-open window is real and named in 208's own code (`if !checked`). It makes the
residual exposure a **count rather than a hypothesis**:

```sql
SELECT count(*) FROM agent_error_log WHERE error_code='OWNED_PAGE_GUARD_UNCHECKED';
```
Non-zero ⇒ the window opened; re-read the served artefact.

Corrected in place at all three sites that carried the overstatement: this file, the
`LANDMINES.md` entry, and concept register SEO-003. **The landmine is still worth its
place** — the `generic` case is untouched by this correction, and that is the case a
future consumer of `pages.noindex` will hit first. Logged in `WRONG_CALLS.md`.

## 3. `constitution` (high) is a FALSE POSITIVE, and the cause is my own action

> *"the inspected schema already lists `pages.noindex boolean` as an existing column …
> the migration will fail outright."*

The column exists **because I applied migration 352 at ~14:2x, before submitting at
14:30**. The seat read the live schema *after* my own apply. This is not a schema-first
violation; it is the reverse — schema-first was followed, and the reviewer is seeing the
post-apply state. **Nothing to fix in the migration.** Worth recording as a gate artefact:
the DB-leads-commit ordering this estate requires on a shared tree makes any
"column already exists" objection structurally unavoidable for an additive migration
submitted after apply. A future submission should say so in the rationale up front — mine
did not, and that cost a high-severity objection.

## 4. `reuse_agent` and `guardian`'s missing checks — both now run, both clean

**Was an indexing/robots control already available anywhere?** No — nothing was
duplicated [MEASURED 2026-08-09], all six stores zero:
```
site_specs 0 · sites.settings 0 · sites.deploy_config 0 · content_components.html_template 0
· site_components.rendered_html 0 · page_components.rendered_html 0
```
plus zero pre-existing `noindex`/`X-Robots` in Go outside this change.

**Blast radius of the action I edited** (guardian asked for the symmetry I had only applied
to the other path): `rerender_single_page` is dispatched by **`page-rerender` and
`report-builder`** — two consumers, both of which gain the same default-off gate, and
`report-builder`'s pages are not flagged.

## 5. `debug_historian` (medium) — pod-grep, and it is a fair procedural hit

The *submission* never committed to proving the deploy at the pod. The **bug file and
RUNBOOK always did** (§"Not done" 1, RUNBOOK §8) — but a reviewer reads the submission, so
that is on me. Restating it here as the binding condition: **this fix is not "live" until
`strings /app/agent-chassis | grep -c injectRobotsNoindex` returns ≥1 on every replica**,
with `injectCanonicalLink` grepped in the same exec as a pipeline positive control (nothing
was removed, so no negative-string control exists).

Its low-severity point — the DO guard asserts an aggregate count, not the specific row — is
**correct and worth adopting next time**. Mitigated here: row *identity* was verified
separately (`WHERE noindex` returned the one expected id), so the aggregate was not the only
evidence. A `RETURNING`-based post-condition tied to the mutation would have been better.

## 6. `architecture` (medium) — ACCEPTED, and acted on

> *"This is the THIRD feature to land only in `assemblePage` … Worth an explicit tracking
> item (not this bug)."*

Agreed, and it is the one objection asking for work outside this fix. Recorded here as the
lane's stated next action rather than silently noted: **the two-head-producer divergence
should get its own `bugs_open/` file** (JSON-LD 07-28, canonical 08-02, noindex 08-09 —
three authors, three landmines, no convergence). Not filed in this round because the
correction in §2 changes its severity story materially, and a file asserting a structural
root cause should go through `090` or declare a substitute — see the ruling of 2026-07-31.
**Grep `bugs_open/`/`bugs_closed/` for it first: it may already exist.**

Its low-severity point about crawler directive-combination behaviour is **fairly put and
not yet answered**: I asserted "crawlers take the most restrictive of multiple robots
directives" to license the coexistence branch. That is unit-tested for coexistence, not
verified against real crawler behaviour. **It is not load-bearing today** — zero robots
metas exist fleet-wide, so the coexistence branch is unreachable until someone adds one —
but it should not be relied on later without a source.

## 7. Not resubmitted as a new round

The code is unchanged, so a resubmission would re-review an identical plan. What the
verdict actually demanded was **evidence and a correction**, both of which are now on
record here. If a later thread disagrees and wants the trail to accumulate, resubmit with
`RESUBMIT_CORR=1139cbbe-3173-4886-846b-c25daeeda93c`.

---

# ✅ LIVE AND BEHAVIOURALLY PROVEN 2026-08-10 — chassis v1.0.1277, both directions, at the artefact

The defect is **no longer reproducible on the live page.** Kept in `bugs_open/` per the
owner ruling of 2026-08-06 (a finished bug stays here). **One residual remains: tools-api
(§4 below).**

## 1. Deploy proven at the binary, every replica, with both controls

`-l app=agent-chassis` matches **2 of the 5** containers running this image — the
documented label trap — so the census was taken over all of them:

| pod | `injectRobotsNoindex` | `injectCanonicalLink` (positive control) | fabricated string (negative) |
|---|---|---|---|
| agent-chassis-…-lftkt | **2** | 4 | 0 |
| agent-chassis-…-v2b59 | **2** | 4 | 0 |
| business-intel-…-mdjpk | **2** | 4 | 0 |
| vet-intel-…-k59qv | **2** | 4 | 0 |

(A 5th, `agent-med-price-collector`, is a **completed job pod**, not a live replica —
`cannot exec … phase is Succeeded`.) Nothing was removed by this change, so no
negative-string control exists naturally; a fabricated symbol returning 0 supplies one.

## 2. THE PROOF IS THE PAIR, not the flagged page alone

Both re-rendered **on the same binary, through the same action**, minutes apart:

| page | `pages.noindex` | robots tag in rendered_html | served live | deployed |
|---|---|---|---|---|
| `/tools/gauntlet/round.html` | **true** | **present** | **1** | ✅ |
| `/about.html` | false | **absent** | **0** | ✅ (51,424 B) |

The unflagged control is the load-bearing half: a page rendered *before* the roll would
show no tag for the trivial reason, and would prove nothing about the gate. This one was
rendered *after* it, by the same code, and still correctly has no tag.

Served, cache-busted, and the tag is **inside** `<head>` (idx 9,779 vs `</head>` 9,822),
page intact at 30,327 B with its title:

```
<meta name="robots" content="noindex, nofollow">
```

Deploy leg read from the run rather than assumed —
`deploy_result.response.data`: `success: true`, `files: ["/tools/gauntlet/round.html"]`,
repo `gqls/sites`, 10:23:57Z.

## 3. ⚠ TWO TRAPS PAID FOR IN THIS VERIFICATION — read before repeating it

**(a) The spawn-wrapper rerender HUNG; the direct dispatch worked.**
`rerender_page_vonc.sh` (the RUNBOOK §18 route, a `spawn_agent`→`call_agent` wrapper)
parked at `spawn_rerender/AWAITING_RESPONSES` and **FAILED at 634s**, having spawned no
child at all (only the parent row exists for corr `04b6176c`). The dispatch lane was
**clear at the time — LAG 0**, so this was not queue latency. The bypass
(`cta_link_integrity/scripts/049b_deploy_single_page.sh <page_id> <site_id> <domain>`,
corr `1f3d125c`) dispatches `page-rerender` **directly, with no spawn step**, and
completed in ~25s. **Use 049b for a single page.** The hung row was deliberately **not
cancelled** — cancelling destroys the evidence, per the standing rule.

**(b) …and that failure is invisible in the table you are told to measure it in.**
The `spawn-call-handshake-races` account says to count these in `agent_error_log`, because
`orchestration_states` under-reports (it cited 166 COMPLETED / 0 FAILED against 79 logged
timeouts). **Here it is exactly inverted:** `orchestration_states` shows `FAILED`, and
`agent_error_log` has **zero** `%timed out after%` rows in the surrounding two hours.
So **neither table alone is a census of this failure** — the under-counting runs in both
directions, and a clean `agent_error_log` is not evidence the handshake is healthy. Worth
knowing for `bugs_open/029` and for the `page-rerender` "271/0, clean" figure, which this
run is a counter-example to.

## 4. STILL OWED — tools-api, and it is NOT live

`X-Robots-Tag` on `PublicRoundHandler` is committed but ships from the **island VM**
(docker compose), not the chassis image, so **the chassis roll did nothing for it**:

```bash
kubectl get deploy -n ai-persona-system | grep -i tools    # no rows — not in this cluster
```

Deploy is rebuild + `docker save|load` + `compose up -d` per RUNBOOK `gauntlet_dead_cta`
§5 — SSH to another host, owner-adjacent, deliberately not done unasked. Verify with a
**real published slug** (a 404 returns no header either, which reads identically):
```bash
curl -sI 'https://tools.apis.uk/api/v1/tools/gauntlet/round/<slug>' -H 'Origin: https://vonc.com' | grep -i x-robots
```
Note the same file now also carries another lane's RFC_020 §5.2 namecheck refusal, so that
deploy ships both changes — coordinate with whoever owns it.

## 5. Also still owed (unchanged)

- **Landmine verification re-dispatch.** Re-checked 2026-08-10: `code_symbols` has grown
  5,755 → **5,837** and still holds `injectRobotsNoindex` **0** against
  `injectCanonicalLink` **1** — the index has not yet reached `c3d7841f9`, so a verdict now
  would still be `bugs_open/223`'s false STALE. Hold.
- **The two-head-producer tracking item** (`architecture` seat's ask). Unchanged, and note
  §2's correction bounds its severity: `owned` pages are refused by the other path, so the
  exposure is `generic` pages plus 208's countable fail-open window.
