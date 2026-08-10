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
right one.

> **⚠ CORRECTED later the same day — see MISSTEP 4 at the foot of this file before
> quoting the sentence above.** It holds for `rebuild_policy='generic'` pages and **not**
> for this one: `AssemblePageAction` refuses an `owned` page before assembly (208's
> guard), and this page is `owned`. I measured an absence in that function's body and
> reported it as a behaviour, without reading its entry conditions.

Landmined + registered as SEO-003's open question; deliberately **not** widened into this
bug (architecture scope, pre-existing, and the guardian seat vetoes exactly that kind of
ride-along).

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

---

## 2026-08-09, later — council verdict REVISE, and the objection that was worth the round

`1139cbbe-…`, gated by `editquality` (high). 13 seats: 6 approve, 5 abstain, 7 objections.
Verdict read from `diagnosis_artifacts.body` keyed on **my** correlation, not from the
newest `doc_note` — CLAUDE.md's documented recipe returns whichever lane's verdict is
newest, which the 168 lane found the same day.

**The gating objection is the one I should have anticipated.** I proved the *general*
two-producer divergence, exhaustively, and never proved *which producer builds this page*.
If it were the other one, my whole fix would be inert for the single page it targets — the
same trap I was busy warning everyone else about, pointed at me. Answered with three
measurements: 3/3 work items ever filed against the page are `page_rerender` →
`page-rerender`; `rerender_single_page` is dispatched by `page-rerender` + `report-builder`;
and the page is `rebuild_policy='owned'`.

## MISSTEP 4 (real, and published in three places before I caught it)

Reading `rebuild_policy='owned'` to answer that question **refuted a claim of my own**.
`AssemblePageAction` **refuses** an owned page before assembly (208's guard,
`multipage_actions.go:42-86`) — so for this page the other path does not silently drop the
tag, it declines to touch the page. My landmine, my register entry and my bug-file section
all said otherwise *about this page*.

**Why I got it wrong:** I greped the other producer for the four inject helpers, found
none, and stopped. **I established what that function does not DO and never read what it
refuses to ENTER.** Its first 45 lines are an ownership guard, sitting above every line I
looked at. One `sed -n '1,90p'` on the function I was making claims about would have shown
it. Generalised: *an absence in a function's body says nothing about whether control
reaches the body* — I measured an absence and reported it as a behaviour.

Corrected in place, visibly, at all three sites, and the correction keeps the landmine
because the `generic` case is untouched by it. Recorded in WRONG_CALLS. Note the shape of
the catch: **the question that found it was about something else entirely**, which is the
argument for the gate rather than for being more careful.

The residual is now countable rather than hypothetical — 208's guard **fails open** when
`rebuild_policy` cannot be read (`checked=false`, logs `OWNED_PAGE_GUARD_UNCHECKED` at
high), so:
```sql
SELECT count(*) FROM agent_error_log WHERE error_code='OWNED_PAGE_GUARD_UNCHECKED';
```

## The false positive, and it is self-inflicted

`constitution` objected at **high**: the migration adds a column that already exists. True
of the schema it read — because I applied the migration ~8 minutes before submitting. Under
this estate's DB-leads-commit ordering on a shared tree, that objection is **structurally
unavoidable** for any additive migration submitted after apply. Same family as "your own
action silences your own detector", arriving at a reviewer instead of a check. **Fix for
next time: say so in the rationale.** Mine did not, and it bought a high-severity objection
on a non-issue.

## Checks the seats asked for, all run, all clean

- **`reuse_agent`** — was any robots/indexing control already available? Six stores, all
  **0** (`site_specs`, `sites.settings`, `sites.deploy_config`, `content_components`,
  `site_components`, `page_components`), plus zero pre-existing `noindex`/`X-Robots` in Go.
  Nothing duplicated.
- **`guardian`** — the symmetric blast-radius census I had only run on the *other* path:
  `rerender_single_page` has two consumers, `page-rerender` and `report-builder`.
- **`debug_historian`** — fair procedural hit: the *submission* never committed to a
  pod-grep, even though the bug file and RUNBOOK always did. A reviewer reads the
  submission. Its second point is a genuine improvement to adopt: the DO guard asserts an
  aggregate count, where a `RETURNING`-tied post-condition would bind to the specific row.
  Mitigated here only because row identity was verified separately.
- **`architecture`** — "third feature to land only in assemblePage; worth a tracking item
  (not this bug)". Accepted; recorded as the lane's next action. Deliberately not filed in
  this round because misstep 4 changes its severity story, and a structural root cause
  needs `090` or a declared substitute.

**Not resubmitted.** The code did not change; the verdict asked for evidence and a
correction, and re-reviewing an identical plan would spend a round to be told the same
thing. `RESUBMIT_CORR=1139cbbe-…` if a later thread disagrees.

---

## 2026-08-10 — the roll landed; fix is LIVE and proven in both directions

Chassis **v1.0.1277**, pods up since 2026-08-09 21:35Z (so the ≥300s dispatch window was
long past — no wait needed).

**Pod census over ALL containers running the image, not the label.** `-l app=agent-chassis`
returns **2**; five containers actually run the image (`business-intel`, `vet-intel` and a
completed `agent-med-price-collector` job pod also carry it). On each of the 4 live ones:
`injectRobotsNoindex` **2**, `injectCanonicalLink` **4** (pipeline positive control),
fabricated `zzzNotARealSymbol232` **0** (negative control — nothing was removed by this
change, so a natural negative did not exist and had to be manufactured).

**Both directions, same binary, minutes apart** — this pairing is the actual proof:

| page | `noindex` | tag in rendered_html | served | deploy |
|---|---|---|---|---|
| `/tools/gauntlet/round.html` | true | present | 1 | ok |
| `/about.html` | false | absent | 0 | ok, 51,424 B |

Tag sits inside `<head>` (idx 9,779 vs `</head>` 9,822); page intact at 30,327 B.
Deploy leg read from `deploy_result.response.data` — `success:true`, one file, `gqls/sites`.

## MISSTEP 5 (small, and the standing rule caught it) — I guessed the result paths

First read of the completed run asked for `collected_data->'rerender_result'->'skipped'`,
`->'deploy_result'->'success'`, `->'commit_sha'`. **All four came back `null`** — and a
jsonb path read cannot distinguish "absent" from "null", so that output was
indistinguishable from a run that had done nothing. Enumerated the keys instead
(`jsonb_object_keys`) and the real shape is `render_page` (not `rerender_result`) and
`deploy_result.response.data.success` (nested two deeper than I guessed). The standing
landmine — *a jsonb path read cannot see the shape change underneath it; enumerate keys* —
is exactly this, and it cost one query because I followed it on the second attempt rather
than believing the nulls.

## ⚠ TWO TRAPS WORTH MORE THAN THIS BUG

**(a) The documented RUNBOOK route hung; the bypass worked.** `rerender_page_vonc.sh` (the
`spawn_agent`→`call_agent` wrapper the gauntlet RUNBOOK §18 prescribes) parked at
`spawn_rerender/AWAITING_RESPONSES` and **FAILED at 634s** with **no child orchestration
ever created** (corr `04b6176c` — one row, no parent/child pair). Checked before blaming
it: `dispatch-queue-depth.sh` reported **LAG 0, lane clear**, my run the only thing in
flight — so not queue latency, which is the usual and usually-correct explanation.
`049b_deploy_single_page.sh` dispatches `page-rerender` **directly with no spawn step** and
completed in ~25s (corr `1f3d125c`). Did **not** cancel the hung row: it is the evidence,
and destroying it pre-diagnosis is a documented compounding error.

**(b) The failure is invisible in the table the standing account says to measure it in.**
`spawn-call-handshake-races` says count these in `agent_error_log` because
`orchestration_states` under-reports (166 COMPLETED / 0 FAILED vs 79 logged timeouts).
**Here it is exactly inverted:** `orchestration_states` = `FAILED`, `agent_error_log` =
**0** `%timed out after%` rows in the surrounding 2 hours. So the under-counting runs
**both ways**, and a clean `agent_error_log` is *not* evidence the handshake is healthy.
That also makes this run a counter-example to the "`page-rerender` 271/0, clean" figure —
it is clean *in that table*. Not chased further: it belongs to `bugs_open/029`, which is
owned, and contributing a measurement is the right move rather than forking a diagnosis.

## Verifier still held, re-checked rather than assumed

`code_symbols` has grown **5,755 → 5,837** since yesterday and still holds
`injectRobotsNoindex` **0** against `injectCanonicalLink` **1**. The index has not reached
`c3d7841f9`, so dispatching the landmine verifier would still produce `bugs_open/223`'s
false STALE against a correct entry. Held again, with the measurement rather than the
memory of yesterday's measurement.
