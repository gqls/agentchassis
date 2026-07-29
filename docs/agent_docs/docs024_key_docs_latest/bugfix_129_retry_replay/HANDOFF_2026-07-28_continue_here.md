# HANDOFF — bugfix 129 retry replay, 2026-07-28 ~22:20 BST

> ## ⚠ SUPERSEDED 2026-07-28 ~22:30 by thread "bugsearch 5" — **§2's owed item IS DISCHARGED and the bug is CLOSED.**
>
> A replay was witnessed on a natural live timeout at 21:25:18Z (`call_scraper`,
> request `30585e6d…`): child orchestration id `ef7e2ddb…` ≠ awaiting `b89b6e5e…`,
> action `process`, reached `processed`, both orchestrations COMPLETED.
> **Do not go looking for it — it is done.** The bug file moved to
> `bugs_closed/129_HANDOFF_…`; the full evidence is in its new top block, and the
> read-out is `SUMMARY_2026-07-28_retry_replay.md`.
>
> **Two defects in this document, corrected there — do not follow §2 or §5 as written:**
> - **§2's grep names the wrong pod.** `-l app=agent-chassis` returns nothing even
>   when a replay has happened: the retry runs in whichever service hosts the
>   **awaiting orchestration** (here `business-intel`). All services run one image,
>   so loop over the labels. This nearly produced a false "no replay yet".
> - **§5's "anything else is a genuine gap" is too strict.** Adapter actions
>   (`target_agent_id = ''` + `system.adapter%` topic) branch at
>   `coordinator.go:2809` into step **re-execution** and never read a payload, so
>   their NULL payloads are harmless. Exclude them before counting.
>
> **What genuinely remains** is §3 and §4 — both owner decisions about *how* the
> change shipped, neither blocking the bug. Those stand exactly as written below.

**Read this first, then `REVIEW_2026-07-28_council_scope_veto.md`.** Everything
else in this directory is background you do not need to start.

**Thread:** "bugsearch 4". ~~**State: the fix is LIVE and half-verified.** One thing
is owed before the bug can close~~ **— now fully verified and closed, see above** —
and one decision is owed by the owner.

---

## 1. What happened, in five lines

- `bugs_open/129` blamed the spawned child. **The child was never the defect** — the
  coordinator's retry path *reconstructed* the timed-out request from the **awaiting**
  orchestration's own state, so every retry carried the PARENT's `orchestration_id`,
  an empty body and `Action:"execute"`.
- Measured: **430 of 430** retried `awaited_requests` in 14 days took that path;
  **294 (68%)** exhausted the budget.
- Fixed with one invariant: **a retry is a REPLAY** — only `retry_version`,
  `message_id`, `timestamp` may differ.
- Council **REJECTED on SCOPE** (guardian veto). Six of ten seats approved; **no seat
  disputed the diagnosis**.
- It went **live anyway** on v1.0.1194 (another session's roll carried my committed
  code). Capture half proven on live traffic; **replay half not yet witnessed**.

## 2. The ONE thing owed to close the bug

**Witness a replay.** The capture half is proven; the replay half has had no chance
to run, because no awaited request has timed out since the roll.

```sql
-- has anything retried since the roll yet?
SELECT step_name, retry_version, status, (request_payload IS NOT NULL) AS has_payload
FROM awaited_requests
WHERE retry_version > 0 AND sent_at > '2026-07-28 20:48:11'
ORDER BY sent_at DESC;
```

```bash
kubectl logs -n ai-persona-system -l app=agent-chassis --since=6h \
  | grep -E 'Replaying original request|RETRY_PAYLOAD_UNAVAILABLE|RETRY_SELF_ADDRESSED|MISROUTED_REQUEST'
```

**What closes it:** a `Replaying original request` line whose logged
`child_orchestration_id` is **different from** the awaiting orchestration id, and
that request subsequently reaching `status='processed'`. Then move the file to
`bugs_closed/`.

**What does NOT close it:** a green run of `TRIGGER_code_indexer_v2.sh`. I already
have one (`c54b3fdf-b556-45f1-8e59-8237bec64d2a` — `spawn_indexer` processed at
`retry_version 0`, child reached `index_symbols`). It proves the lane advances and
the capture path works; **it succeeded on the first attempt, so it never touched the
replay path.** The defect was bursty (2 of 3), so one green happy path is weak
evidence. Induce the fault or wait for a natural timeout — the fleet produced ~430
in 14 days (~30/day), bursty, mostly `spawn_dispatch`/`call_dispatch` from
`build-pipeline-trigger`.

**If `RETRY_PAYLOAD_UNAVAILABLE` appears:** that is the designed loud failure for a
sender that records no payload. Check which step:
```sql
SELECT step_name, target_agent_type, count(*) FROM awaited_requests
WHERE request_payload IS NULL AND sent_at > now() - interval '2 hours' GROUP BY 1,2;
```
Expect only `scrape_pages`/`search_web` (see §5). Anything else is a genuine gap.

## 3. The decision owed by the owner

The guardian vetoed on **venue**, not on the fix: a shared contract plus a schema
column arriving inside a bug patch — the same finding as `bugs_closed/124`'s `$ctx.`
veto, two days after the ruling written for it.

**Do not resubmit to the council.** The 2026-07-28 ruling is explicit that a SCOPE
veto is a judgement about *how* a capability reached production, not a measurement
to be improved — and the seats **contradict each other** on the remedy here (the
guardian's "ship the child-side guard alone" is what `constitution` and
`editquality` approved the plan for *not* treating as primary, and it does not fix
the bug — it removes the silence, not the failure).

The options and their costs are in `REVIEW_2026-07-28_council_scope_veto.md`. Since
the code is now live, option 2 (the contained alternative) has become *a revert of
live code that would restore a 430/430 defect*, and option 3 (hold) turned out never
to have existed — see §4.

## 4. The finding that outlives this bug

**On this tree, committing IS shipping.** I chose "hold the deploy pending an owner
call". That option does not exist: `make build-<service>` builds from committed
`HEAD` (deliberately, so builds cannot bundle WIP), CLAUDE.md forbids holding work
uncommitted, and the fleet rolled ~8 times on 2026-07-28. My commits went out inside
someone else's roll, and that session had no way to know.

**The platform-seam ruling assumes the committing thread controls when its seam
ships. It does not.** The only mechanism that actually holds a seam back on a shared
HEAD is to **commit it dark** — behind a flag or config switch defaulting off. The
ruling does not currently ask for that. Written up as an addendum in the REVIEW doc
and as a 016b §9 entry; **it is the owner's call whether the ruling gains that
clause.**

## 5. Known, deliberate, NOT a gap to "fix"

> **⚠ CORRECTED 2026-07-29 — THIS SECTION ENUMERATES STEP NAMES, BUT THE DEFECT IS
> PER-ACTION, so the list below is incomplete and will keep "discovering" itself.**
> On 07-29 a post-roll census turned up `search_news` as unrecorded, which by this
> section's wording reads as a new gap. It is not: `FetchNewsSearchAction`
> **delegates to `WebSearchAction`** — the same sender named below. Enumerate by the
> ACTION, and the true scope is **10 step instances across 9 agents**:
>
> | action | instances | step names |
> |---|---|---|
> | `web_search` | 8 | `search_web` ×5 (adoption-researcher, directory-researcher, evidence-researcher, grounded-explainer, research-agent), `search_area`, `search_domain`, `search_practice` |
> | `fetch_news_search` → `WebSearchAction` | 1 | `feed-ingester.search_news` |
> | `fetch_scrape` → `WebscrapeAction` | 1 | `feed-ingester.scrape_source` |
>
> One query gets it, and it is the one to re-run rather than trusting this table:
> ```sql
> SELECT v->>'action' AS action, type||'.'||k AS step
> FROM agent_definitions, LATERAL jsonb_each(default_config->'workflow'->'steps') e(k,v)
> WHERE deleted_at IS NULL AND is_active AND COALESCE(is_snapshot,false)=false
>   AND v->>'action' IN ('web_search','webscrape','fetch_news_search','fetch_scrape');
> ```
> **The `≤2 requests a fortnight` cost below was computed over the RETRIED
> population and still holds** (`search_news` has 193 rows in 14 days and has been
> retried **0** times) — but the *exposure* is 5× wider than the two step names
> below suggest. Post-fix capture is **139/145 (95.9%)**, and every one of the 6
> misses is `search_news`. Whoever takes the follow-on diagnosis: scope it to the
> two ACTIONS, not to the step names.

`scrape_web` / `web_search` (6 of 428 retried requests, 14 days) record no payload
and now **fail fast instead of retrying**. They are a *different defect*:
`web_search_action.go:139` puts `params.ExecutionContext.OrchestrationID` — the
**caller's own** id — on the **original** outbound message, so there is no child
identity to replay. Wiring them would only make `RETRY_SELF_ADDRESSED` fire on their
originals. **Needs its own diagnosis; bundling it here is exactly what was vetoed.**
Cost of leaving it: 4 of those 6 exhausted anyway ⇒ **≤2 requests a fortnight** lose
a retry that might have worked, in exchange for a named error instead of silence.

## 6. State of the world (verified, not inferred)

| thing | state | how it was checked |
|---|---|---|
| chassis | **v1.0.1194**, rolled 20:48:11Z | `kubectl get deploy` + pod list |
| the fix in the binary | **LIVE on both pods** | pod-grep: `is_retry` **0**, `RETRY_PAYLOAD_UNAVAILABLE` / `RETRY_SELF_ADDRESSED` / `MISROUTED_REQUEST` / `RETRY_PAYLOAD_BACKFILL_MISSED` all **1** |
| migration 263 | **applied + recorded** | `information_schema`, and `--record-only` |
| capture on live traffic | **4 of 4** awaited requests since the roll carry a payload | SQL, §2 |
| the invariant | **0** self-addressed payloads | SQL, §2 |
| payload size | **~1.1 KB** (`pg_column_size`) | retires the council's size risk |
| replay | **0 occurrences — untested, not failing** | log grep; no timeouts have happened yet |

**Landmine — the discriminating pod-grep is a string this fix DELETED (`is_retry`).**
v1.0.1192 has it and none of the new markers; 1193/1194 are the reverse. The positive
markers alone are vacuous. **Re-grep after any roll you did not do.**

**Landmine — the live retry path is `coordinator.go handleRecoverableError`, NOT
`helpers.go retryTimedOutRequest`.** The latter has the identical defect and is fixed
too, but is dormant. Grepping the *mechanism* lands you there and you will fix
nothing; grep the error string from the evidence (`timed out after 3 retries`).

## 7. Commits

| commit | what |
|---|---|
| `eb70c3dd3` | the fix — replay contract, migration 263, both retry paths, child-side guard, tests |
| `17bd675dd` | council record; back-fill rows-affected fix; the coverage correction |
| `57e7a15c7` | `IMAGE_TAG` v1.0.1193 |
| `b9fc30615` | migration 263 house style (BEGIN/COMMIT + type guard) |
| *(this one)* | live verification, the "committing is shipping" finding, this handoff |

No `Council-Reviewed:` trailer on any of them — correctly, since the verdict was
REJECTED. Do not add one.

## 8. If you want more work in this lane

`bugs_open/` candidates I checked as genuinely unowned and did **not** take, with
the reason (so you do not re-derive it):

- **132** — raw JSON 404 on every B2 site. Fix is a Cloudflare Worker; the emitter is
  in **no local repo** (grepped `objectKey` / `"B2 returned error"` across
  `~/projects`, and there is no `wrangler*` anywhere). Needs Cloudflare access.
- **120** — merge commits skip the deploy. Real, unowned, one-line fix — but it lives
  in `~/projects/sites` (`gqls/sites`), a different repo, not a chassis build.
- **091 / 097** — `bugs_sweep_2026_07` names 091 as its next item. `who-owns.py`
  reads commits, so a session mid-fix is invisible; check the tree too.
