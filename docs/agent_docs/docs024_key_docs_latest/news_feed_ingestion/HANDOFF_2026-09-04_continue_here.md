# HANDOFF — news_feed_ingestion — continue here (2026-09-04)

Written 2026-09-04. **Read this first**, then `PLAN_2026-09-02_news_feed_ingestion.md`
(design + decisions), `NOTES_news_feed_ingestion.md` (chronological, newest at the
bottom), `RUNBOOK_news_feed_ingestion.md` (every command, with status blocks now marking
what is already done), `README_where_we_are.md` (plain prose for the owner).

**Supersedes `HANDOFF_2026-09-03_continue_here.md`.** That file's §1 — "THE ONLY THING
BLOCKING THIS LANE: two migrations need an authorised apply" — is **fully discharged**;
read it only for background. Its §5 (designblog) and §6 (earlier work) still stand as
written.

## 0. What this lane is

Charter: the raw news-ingestion pipeline — source enrollment, fetch scheduling/cadence/
queue fairness, and normalising/extracting structure out of what arrives. Concretely
`content-feed-orchestrator`, `find_news_sites`/`content-feed-trigger`, `feed-triage`,
`content-feed-refresh`, and `content_feed_items` as a data asset. Full charter and
exclusions in PLAN's opening section.

## 1. ⚠ READ FIRST — a FLEET-WIDE outage is live and it is NOT this lane's

Every Anthropic LLM call has failed since **2026-09-04 11:17:05Z**:

```
API request failed with status 400: {"type":"invalid_request_error",
"message":"Your credit balance is too low to access the Anthropic API.
Please go to Plans & Billing to upgrade or purchase credits."}
```

`[MEASURED 2026-09-04 11:37Z]` **9 distinct `owner_agent_type`s** affected
(landmine-verifier, generic, diagnose-agent, diagnose-orchestrator, directory-researcher,
feed-triage, build-briefing-agent, tool-improver, content-feed-orchestrator) — the estate,
not a lane. `llm_call_log` ran **0–1 failures/hour for the preceding 12 hours**, then
**70 failed / 14 ok** after 11:17, with the **last successful call of any kind at
11:20:49Z**. Escalated to the owner at once; only he can act on it.

**If you are picking this lane up and LLM steps still fail, that is why — do not diagnose
it as a feed bug, and do not file it as one.** Re-check before assuming it persists:

```sql
SELECT max(created_at) FILTER (WHERE success) AS last_success,
       count(*) FILTER (WHERE NOT success AND created_at > now() - interval '15 min') AS recent_failures
FROM llm_call_log WHERE provider='anthropic';
```

**This is NOT the 2026-08-23 incident** — that was *"you have reached your specified API
usage limits"* (a monthly **cap**); this is prepaid **credits** exhausted. Different lever,
same trap: there are ≥2 Anthropic orgs and the console's default is not the fleet's, so
credit bought on the wrong one changes nothing. Decisive check is **API keys → `Last
used`** (a failed call is still a use). Memory: `the-fleet-key-is-not-on-the-default-console-org`.

## 2. DONE this session — all three of yesterday's blockers, on the owner's authorisation

| action | outcome |
|---|---|
| `691_uk_news_search_region_default.sql` | **APPLIED + VERIFIED + RECORDED.** 26/26 `.uk` `news_search` rows carry `region=uk`, 0 non-UK touched |
| `746_advertise_news_feed_enablement.sql` | **APPLIED + VERIFIED + RECORDED.** `recommended=true`; 6 sources (1 rss WebProNews + 5 `news_search` `region=uk`); trigger predicate selects the site |
| 746 council review | **DISPATCHED, then KILLED BY THE OUTAGE — the correlation is spent and must be re-fired.** See §3a |

All four preconditions were re-verified first-hand before asking, and none had drifted
(both files absent from `schema_migrations`; 26/0 on the region census; advertise on 0
sources with no `content_features`; 746's pinned spec `ec005136…` still current).

## 3a. ⚠ The 746 council review DID NOT HAPPEN — and it does not look that way

`SUBMISSION_CORR = 70f500ff-fb38-4fef-802a-8f25e8535367` was dispatched 11:29:28Z and the
run **finished**: `status='COMPLETED'`, `error` NULL, `current_step='complete_invalid'`.
All three signals read as "your submission was rejected as invalid". It was not.
`collected_data->'__step_errors'` gives the real cause:

> `council_decide` — *"no reviewer produced a readable opinion (1 abstained, 16 unreadable:
> review_editquality.result, review_bug_historian.result, … review_architecture.result) — a
> council with no opinions cannot decide"*

…with each of those 16 seats carrying the **same** `provider=anthropic … "Your credit
balance is too low"` error from §1. Every seat was down; the gate reported a **billing
outage** as an **invalid submission**. The JSON is fine — it had passed `DRY_RUN=1`
validation *and* scope admission minutes before.

**So: the correlation is SPENT.** It will never write a `council_report`, so it can never
be resolved by `098` and nothing should carry it as a trailer. **Re-fire once §1 clears**
(the submission file needs no rework) and record the NEW correlation:

```bash
DRY_RUN=1 ./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh \
  docs/agent_docs/docs024_key_docs_latest/news_feed_ingestion/COUNCIL_SUBMISSION_746.json
# then without DRY_RUN, and SAVE the printed SUBMISSION_CORR
```

Do **not** read the missing verdict as queue latency. CLAUDE.md's "a missing row is almost
always latency" is about a missing *orchestration row*; here the row exists and is finished.
Full entry in `LANDMINES.md` ("A council-gate run that ends `COMPLETED` at
`complete_invalid`…").

## 3. DONE and PROVEN — the UK-region fix works at the provider

`[MEASURED 2026-09-04 11:34Z]` at `web-search-adapter`: advertise's five `news_search`
sources went to Firecrawl as **`region='uk'`**, and in the **same window, same pod, same
provider**, another site's region-less search went as **`region=''`**. The control could
have come out otherwise and did — the field discriminates rather than being a constant.
This retires HANDOFF 2026-09-03 §2's "what is NOT yet proven".

> **One narrowing, stated rather than rounded up.** The proof came from **746's new**
> sources, not from **691's backfilled** ones. idea.uk's five rows are stamped in the DB
> but had already fetched at 09:15Z (`next_fetch_at` 2026-09-05 09:15Z), so the orchestrator
> correctly skipped them and no idea.uk search fired. Same key, same code path — but they
> have not been watched doing it. **Close this tomorrow** (§5.2).

## 4. ⚠ The instrument that nearly reported the opposite — `kubectl logs -l` truncates

`logs -l <selector>` applies a default **`--tail=10` per pod**, and `--since` does not lift
it. On the **single**-pod adapter, the RUNBOOK's proof command returned **2** lines where a
pod-addressed read of the same window returned **12** — so this is not the "one pod of N"
trap and counting pods will not reveal it. A tail keeps the *newest* lines, so what survived
was an unrelated site's `region=""`: **the truncation preserved exactly the evidence that
reads as "the fix does not work"**, with nothing in the output saying it was truncated.

**Always pass `--tail=-1` with `-l`.** `[MEASURED 2026-09-04]` **120** occurrences of the
untailed shape exist across `docs/` and `scripts/` — this lane's two are fixed, the rest are
not, and each is a potential false negative. Full entry in `LANDMINES.md`
("`kubectl logs -l <selector>` silently applies `--tail=10` PER POD").

## 5. RESUME HERE

1. **Blocked on §1.** advertise.co.uk fetched **19 items, all `status='ingested'`, none
   scored.** Two causes, neither a 746 defect: the hand dispatch raced its own ingestion
   (triage 11:34:16→26 while items landed 11:34:17→35 — the 6-hourly route does not race),
   and the outage stopped `score_relevance`. **Once LLM calls succeed**, re-dispatch
   `scripts/dispatch_content_feed_orchestrator.sh advertise.co.uk` (in this dir; it is
   idempotent at the seeding step) or wait for `content-feed-refresh`, then judge at the
   artefact: `746_..._VERIFY.sql`'s relevant/review/rejected split, then
   `https://advertise.co.uk/data/news-archive.json` (404 before).
   ⚠ **Never at the served `/news/index.html`** — it fills client-side, so `curl | grep`
   reads 0 for ever regardless of success (HANDOFF 2026-09-03 §4's correction; the live DOM
   via `browser-runner-adapter` is the right instrument).
2. **Tomorrow ~09:15Z**, idea.uk's backfilled sources fetch. Confirm `region='uk'` at the
   adapter **with `--tail=-1`** to close §3's narrowing. Only then may anyone write
   "691 proven at the provider".
3. **Re-fire the 746 council review** — §3a. The old correlation is spent; do not query it
   for a verdict and do not put it on a commit. A later commit may carry
   `Council-Reviewed: <NEW corr>` **only after reading an approved verdict**.
4. **designblog.co.uk `/the-design-feed/`** — unchanged: a decision, not a build. HANDOFF
   2026-09-03 §5 is still the current statement, including this lane's **withdrawn**
   preference for route (1) and the third gap
   (`create_blog_posts_action.go:196` never passes `ParentSection`). Track `bugs_open/460`,
   `463`, `468`, `457`.

## 6. New in this lane's toolbox

`scripts/dispatch_content_feed_orchestrator.sh <domain>` (this dir) — domain-generic feed
dispatch with a receipt. Generalised from the idea.uk-hardcoded copy under
`idea_uk_vm_site/scripts/` (left untouched; another lane owns it). It resolves `site_id`
from the domain and **refuses before publishing** if the site's spec lacks
`content_features.news_feed.recommended=true` or it has zero active sources — the silent
failure the 746 work exposed, where an "enabled" site fetches nothing because
`seed_content_sources_action` skips `rss` and returns early on any existing active source.
Both guards proven in the negative direction (`designblog.co.uk` and a non-existent domain
each exit 1), not merely asserted.

## 7. Standing practices this lane follows

Unchanged from HANDOFF 2026-09-03 §7 — `\d` before SQL, pathspec commits, apply your own
migration by hand and `--record-only`, dry-run by hand (the runner's probe declines any
file whose text contains "rollback"), verify a deploy at the binary with both controls,
a single-service cluster deploy is the owner's call. Add two from today:

- **A denial attaches to the ACTION and is lifted only by the owner naming it** — which is
  exactly what happened here. Yesterday's session was right not to re-fire 746 on the
  evidence that a sibling dispatch worked.
- **Validate a trailer against the COMMIT's paths, not your recent memory.**
  `bash scripts/council-scope.sh <paths>`. This session put 746's `Council-Submitted:`
  correlation on a `docs/` shell script the council does not review — disclosed in NOTES
  and `WRONG_CALLS.md` rather than amended.
