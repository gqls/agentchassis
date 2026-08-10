# HANDOFF 2026-08-10 — `bugs_open/232` cold-start for a fresh chat

**The bug is FIXED, LIVE and BEHAVIOURALLY PROVEN.** This handoff exists for the two
things still owed and the three findings that outlived the fix. Read in this order:
this file → `bugs_open/232` (the last two sections) → `NOTES` tail → `RUNBOOK` §8.

Constants you will need:
- page id **`4629451e-e4f2-4fe2-b258-35107b5cb51e`** (`/tools/gauntlet/round.html`)
- site id **`9ec3b9ee-5b08-461b-b4f8-9e1e03579c74`** (vonc.com)
- council corr **`1139cbbe-3173-4886-846b-c25daeeda93c`** (REVISE, read + answered)
- commits: fix `c3d7841f9` · lane docs `2c0d1b51b` · ratchet `afa74c9b8` ·
  council response `7e685b250` · live proof `63824441e`

## 0. State in one paragraph

`pages.noindex` (opt-in boolean, `NOT NULL DEFAULT false`, migration **352 applied**) is
read by `getPageInfo` and gates `injectRobotsNoindex` at `assemblePage` step 5a3
(`rerender_single_page_action.go`). Live on **chassis v1.0.1277**, pod-verified on all 4
live replicas, and proven **in both directions** at the served artefact. Registered as
**SEO-003**. Council returned **REVISE**; every objection is answered with a measurement in
`bugs_open/232` §"COUNCIL VERDICT READ" — the code did not change, so it was deliberately
**not** resubmitted.

## 1. STILL OWED — tools-api `X-Robots-Tag` (the only outstanding work on this bug)

`internal/tools-api/handlers/publish.go` → `PublicRoundHandler` sets
`X-Robots-Tag: noindex, nofollow`. **Committed but NOT live.** It is not in this cluster:

```bash
kubectl get deploy -n ai-persona-system | grep -i tools    # no rows
```

It runs on the **island VM** under docker compose — rebuild + `docker save|load` +
`compose up -d`, per RUNBOOK `gauntlet_dead_cta` §5. That is SSH to another host, so it is
**owner-adjacent and was deliberately not done unasked**.

⚠ **Coordinate before deploying.** Another lane has since added RFC_020 §5.2's `namecheck`
publish refusal to the *same file*, so that deploy now ships **both** changes.

Verify with a **real published slug** — a 404 returns no header either, which reads
identically to a fix that is not there:
```bash
curl -sI 'https://tools.apis.uk/api/v1/tools/gauntlet/round/<slug>' \
  -H 'Origin: https://vonc.com' | grep -i x-robots
```

## 2. STILL OWED — the two-head-producer tracking item (the architecture seat's ask)

Three features have now landed in `assemblePage` while `AssemblePageAction`
(`multipage_actions.go`) stays a silent non-consumer: JSON-LD (07-28), canonical (08-02),
noindex (08-09). The council's `architecture` seat asked for an explicit tracking item,
**not in this bug**. Not filed yet, for two stated reasons: the correction in §3 changes its
severity story, and a file asserting a structural root cause needs `090` or a declared
substitute (owner ruling 2026-07-31).

**Grep `bugs_open/` and `bugs_closed/` first — it may already exist.** The measurements are
in `bugs_open/232` and the `LANDMINES.md` entry *"There are TWO page-`<head>` producers…"*.

## 3. ⚠ READ THIS BEFORE QUOTING THE LANDMINE — I overstated it, and corrected it

The landmine says a rebuild through the other path drops the tag while `pages.noindex` reads
true. **True for `rebuild_policy='generic'` pages; FALSE for `owned` ones.**
`AssemblePageAction` carries `bugs_open/208`'s ownership guard
(`multipage_actions.go:42-86`) and **refuses an owned page before assembly** — it does not
strip the tag, it declines to touch the page. The flagged page is `owned`.

Residual that keeps the landmine alive: the guard **fails open** when `rebuild_policy`
cannot be read, logging `OWNED_PAGE_GUARD_UNCHECKED` at high severity. So the exposure is a
**count, not a hypothesis**:
```sql
SELECT count(*) FROM agent_error_log WHERE error_code='OWNED_PAGE_GUARD_UNCHECKED';
```
Corrected in place at all three sites (bug file, `LANDMINES.md`, SEO-003) and logged in
`WRONG_CALLS.md`. **What caught it was a question about something else** — the council's
gating objection asked which path builds this page; the answer refuted a neighbouring claim.

## 4. Findings that outlived the bug — worth more than the fix

**(a) Our own runbook's rerender route hangs.** `rerender_page_vonc.sh` (the
`spawn_agent`→`call_agent` wrapper, gauntlet RUNBOOK §18) parked at
`spawn_rerender/AWAITING_RESPONSES` and **FAILED at 634s with no child orchestration ever
created**, on a **clear lane (LAG 0 — checked first)**. `049b_deploy_single_page.sh`
(direct, no spawn step) did the same work in ~25s. **Use the direct route for a single
page.** This lane's RUNBOOK §8 is corrected; gauntlet RUNBOOK §18 is **not** (not mine).

**(b) That failure is invisible in the table we tell each other to use.** The standing
`spawn-call-handshake-races` account says count these in `agent_error_log` because
`orchestration_states` under-reports. **This run was inverted:** `orchestration_states`
FAILED, `agent_error_log` **0** rows in two hours. The under-counting runs **both ways**;
a clean `agent_error_log` is *not* evidence the handshake is healthy, and this is a dated
counter-example to the "`page-rerender` 271/0, clean" figure. Contributed as a measurement
to `bugs_open/029` (owned — do not fork a diagnosis). The hung row
`04b6176c-2c63-4245-9167-03e056f8aa62` was **deliberately not cancelled**; ~24h reap.

**(c) Four of my doc edits were swept into other sessions' commits** within minutes
(`9a9fef332`, `e42000bb0`, `bc6b03ec4`, `8d065fc09`). Nothing lost. The trap: the file then
reads **clean**, which looks exactly like "my edit never saved" and invites a duplicate
re-append. **Ask HEAD for your CONTENT, not the tree for your file:**
`git show HEAD:<file> | grep -c "<your distinctive phrase>"`. In `WRONG_CALLS.md`.

## 5. Landmine verification — deliberately HELD, re-measure before dispatching

The entry's own `injectRobotsNoindex` footprint is not yet in the index. Re-measured
2026-08-10 with positive controls in one query (5,755 → **5,837** total, still 0):

```sql
SELECT count(*) FILTER (WHERE symbol='injectRobotsNoindex') AS new_symbol,   -- want >=1
       count(*) FILTER (WHERE symbol='injectCanonicalLink') AS control,      -- 1
       count(*) AS total FROM code_symbols;
```
Dispatch **only when `new_symbol >= 1`** — before that, a verdict is `bugs_open/223`'s
false STALE against a correct entry. Note the entry's NEEDS_VERIFICATION signal was
consumed once already by a concurrent session's `landmines-sync.py --apply` (223's second
trap), so it will not surface on its own; the correction re-armed it.

## 6. How to re-verify the whole thing in four commands

```bash
# 1. binary, every replica (the label matches 2 of 5 — enumerate by image instead)
kubectl exec -n ai-persona-system <pod> -- sh -c \
  'strings /app/agent-chassis | grep -c injectRobotsNoindex; strings /app/agent-chassis | grep -c injectCanonicalLink'
# 2. flagged page carries it
curl -s "https://vonc.com/tools/gauntlet/round.html?cb=$(date +%s)" | grep -io '<meta name="robots"[^>]*>'
# 3. THE CONTROL — unflagged page does not (must be re-rendered on the NEW binary)
curl -s "https://vonc.com/about.html?cb=$(date +%s)" | grep -ic 'name="robots"'    # 0
# 4. the flag itself, by row identity
```
```sql
SELECT p.id, s.domain, p.url, p.noindex FROM pages p JOIN sites s ON s.id=p.site_id WHERE p.noindex;
```

**Do not skip 3.** A page rendered *before* the roll lacks the tag for a trivial reason and
proves nothing about the gate. That pairing is the whole proof.
