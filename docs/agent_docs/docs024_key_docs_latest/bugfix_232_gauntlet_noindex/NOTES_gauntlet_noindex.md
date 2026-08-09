# NOTES — `bugs_open/232` gauntlet noindex

Append-only, newest at the bottom. Evidence, commands, what the system actually said,
and every misstep.

---

## 2026-08-09 — picking the bug, and why the ownership check was nearly useless

Swept `bugs_open/` for something unowned. `scripts/who-owns.py` returned **"OWNED or
recently active"** for every one of the fifteen numbers I tried — 211, 212, 213, 214,
218, 219, 220, 221, 222, 223, 226, 227, 228, 232, 233. That verdict is not
discriminating on a tree this busy: it fires on any workstream dir that *mentions* the
number, and nearly every bug is mentioned somewhere.

What actually discriminated, and is the check I would use next time:

- **Is there a commit whose SUBJECT is about fixing it?** 218/222 had many (both fixed).
  232 had exactly one, `54cd29e57`, and it is the *filing* commit.
- **Does the owning lane's own handoff say "come and take this"?** For 232 the
  `gauntlet_dead_cta` handoff says it was "filed separately … because it is true today,
  is a two-line fix, and depends on none of RFC_020's open questions" — an explicit
  invitation.
- The other genuinely-open candidate was `223` (landmine verifier / Go-only index),
  marked **"OPEN, UNOWNED"** in its own header. Passed over because its own fix
  candidate 3 is flagged architecture-scope and its cheapest candidate is a change to
  a mechanism two other lanes were actively using that hour. 232 was self-contained.

## The premise re-verified before any work (it is a bug file from this morning, but still)

```
curl -s https://vonc.com/tools/gauntlet/round.html | sed -n '/<head/,/<\/head>/p'
```
Head contains: charset, a GTM block, viewport, meta description, title, a large inline
`<style>`. **No robots meta of any kind.** Bug still real.

## MISSTEP 1 (avoided, but only just) — I assumed the fix would be a response header

The bug file recommends `X-Robots-Tag` on the route, and I started looking for the
route that serves `round.html`. There isn't one. `curl -sI` shows `server: cloudflare`
with `x-amz-request-id`/`x-amz-version-id`: **B2 objects behind Cloudflare**, no
server-side path in this repo. The Caddyfiles I found under `gauntlet_dead_cta/infra/`
look like the answer and are not — they front **tools-api**, a different service on a
different host. Had I pattern-matched "Caddyfile exists ⇒ I can add a header" I would
have written a fix that never executes for the page.

**Cheap check, recorded:** before proposing a response header, ask **who serves the
bytes** (`curl -sI` for `x-amz-*` / `server:`), and when you find a proxy config, read
**what it fronts** rather than assuming it fronts the thing you care about.

## MISSTEP 2 (avoided) — the component cannot hold the tag either

Next instinct: put `<meta name="robots">` in the round-record component's template.
The component (`content_components` `71a54cc2-…`, 1 page_components row) is a
**body/style/script fragment** — `grep '<head' ` over its `html_template` returns
nothing. So the bug file's candidate 2 is not merely fragile here (its stated reason
was "stored component content gets regenerated away"); it is **impossible**. The
`<head>` comes from somewhere else entirely.

## The real mechanism, read rather than inferred

`assemblePage()` (`rerender_single_page_action.go:532+`) pulls the **site-level** head
from `site_components` (`slot_name='head'` → "Document Head", shared by **14 sites**),
then does all per-page customisation by regex/string surgery on that already-rendered
blob:

- title replace, `spliceMetaDescription`
- `injectPageJSONLD` (added 2026-07-28, comment: "ZERO of 14 live sites emitted any ld+json")
- `injectCanonicalLink` (added 2026-08-02, comment: "zero canonicals fleet-wide")
- `injectComponentCSS` — the simplest insert-before-`</head>` shape, and the one I copied

So the shared "Document Head" component must **not** be edited (14 sites), and the
per-page tag belongs in this sequence as a fifth injection. That is the whole design.

## ⚠ FINDING THAT CHANGED THE DELIVERABLE — a test-file comment is false, and three live agents prove it

`inject_canonical_link_test.go`'s header states:

> *"assemblePage is the single live assembly path (page-build-handler deploys through
> the page-rerender agent; no live agent uses assemble_page)"*

I nearly took that at face value — it is exactly the reassurance you want when you are
about to add a fifth injection to one function. It is **wrong**:

```sql
SELECT type FROM agent_definitions WHERE is_active AND COALESCE(is_snapshot,false)=false
  AND deleted_at IS NULL AND default_config::text ILIKE '%"action": "assemble_page"%';
--  site-work-orchestrator | pageflow-builder | page-rebuild
```
and `registry.go:537` maps `assemble_page` → `AssemblePageAction`
(`multipage_actions.go`), which calls **none** of the four injections:
```bash
grep -n "injectPageJSONLD\|injectCanonicalLink\|injectRobotsNoindex\|spliceMetaDescription" \
  platform/orchestration/actions/multipage_actions.go     # no hits at all
```

**Consequence, and it bounds the fix honestly:** JSON-LD (07-28), canonical (08-02) and
now noindex (08-09) have each landed on **one of two live head producers**. For noindex
the consequence is worse than a missing SEO nicety — `pages.noindex` can read `true`
while the served page has no tag, which is the wrong result looking exactly like the
right one. Landmined + registered as SEO-003's open question; deliberately **not**
widened into this bug (architecture scope, pre-existing, and the guardian seat vetoes
exactly that kind of ride-along).

**Left the false comment in place** rather than editing it, so the contradiction stays
visible at the spot where a reader would otherwise be reassured.

## Applying the migration — the guard was induced before it was trusted

```
$ ... < 352_induce.sql          # same file, mangled uuid
ALTER TABLE
ERROR:  bugs_open/232: expected exactly 1 noindex row after flip, found 0
$ SELECT count(*) ... column_name='noindex';
0                                # whole txn rolled back, ALTER included
```
Then the real apply: `UPDATE 1`, and verified by **row identity** —
`4629451e-… | vonc.com | /tools/gauntlet/round.html | t` — plus the no-op census
`is_true 1 / is_false 629 / is_null 0 / total 630`.

Pre-checks that adding a column is safe: no `SELECT * FROM pages` anywhere in
`platform/` or `internal/`; all 8 `INSERT INTO pages` name their columns;
`site_snapshots.pages_snapshot` is jsonb, not a fixed shape.

## Mutation-testing the new tests, because green is not evidence

- **Mutation 1** — injection returns `headHTML` unchanged → **5 cases fail**.
- **Mutation 2** — idempotency marker swapped from the exact tag to any
  `<meta name="robots"` → **the coexistence case fails**, and only that one.

Mutation 2 is the one that earns its keep: it is the only thing separating my judgement
(emit alongside a foreign permissive robots meta; crawlers take the most restrictive)
from the other plausible one (defer to it — which fails **silent**).

## MISSTEP 3 (real, ~90 seconds) — I read "file is clean" as "my edit is gone"

Appended to `LANDMINES.md`; `git diff --numstat` said `41 added, **1 deleted**`. I only
appended. Investigated (correct) — the deleted line was **another session's**
uncommitted edit that my diff was showing alongside mine. Then minutes later
`git status` on the file returned **nothing**, because that session **committed it**
(`9a9fef332`), taking my entry with it. Same again for `000_concept_index.md`, swept
into the loancalculator lane's `e42000bb0`.

Nothing lost; forward-only holds. **The near-miss was re-appending** a 40-line entry to
an append-only file that syncs to `doc_notes`. The one-command check, now in
WRONG_CALLS: ask **HEAD for your content**, not the tree for your file —
`git show HEAD:<file> | grep -c "<your distinctive phrase>"`. `git status` cannot
distinguish "clean because unwritten" from "clean because someone committed it".

## Consequence of that sweep — 223's trap fired on my entry, from a session that never heard of it

Because `9a9fef332` ran `landmines-sync.py --apply`, my entry reached `doc_notes`
(**14 footprint rows**, 14:32:53Z) **and its NEEDS_VERIFICATION signal was consumed in
the same motion** — `bugs_open/223`'s second documented trap. Confirmed: **0** rows in
`landmine-verification` for it.

I then deliberately did **not** dispatch verification, and the reason is measured rather
than cautious — the entry's own new symbol is not in the index yet, with positive
controls in the same query:

```sql
SELECT count(*) FILTER (WHERE symbol='injectRobotsNoindex') AS new_symbol,      -- 0
       count(*) FILTER (WHERE symbol='injectCanonicalLink') AS control_existing, -- 1
       count(*) FILTER (WHERE symbol='AssemblePageAction')  AS control_other,    -- 1
       count(*) AS total FROM code_symbols;                                      -- 5755
```

A verdict now would be 223's textbook false STALE against a correct entry. Re-dispatch
once the index refreshes past `c3d7841f9`.

## Committed

`c3d7841f9`, 7 files, `Council-Submitted: 1139cbbe-3173-4886-846b-c25daeeda93c`.
The 8th intended file (`000_concept_index.md`) had already gone to HEAD inside
`e42000bb0` — verified at HEAD, not at the tree. Declared the unavoidable same-file
passenger (the loancalculator lane's DOC-076 row) in the commit message.

Pre-commit hook flagged *"migration + platform code in one commit — needs a staged
rollout order"* as an architecture signal. Read and answered rather than ignored: the
staged order **is** stated (migration first, and why it must be), and the scope
argument is in the submission's `risks` for reviewers to reject if they disagree.
