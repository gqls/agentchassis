# 376 — one unscrapable exemplar discards the WHOLE vertical-research step, and retrying is structurally incapable of routing around it

**Filed:** 2026-08-23 by the `loanzy_uk_example_site` lane (one-shot build route), from the live
`garden-tools.uk` greenfield build. **Status: OPEN, unowned.**
~~**Severity: medium**~~ **Severity: HIGH — this does not degrade a greenfield build, it KILLS it.**
100% reproducible per affected vertical, silent for 30-60 minutes at a time, and the step's own
stored record reads `success: true` while it is failing.

> **UPGRADED 2026-08-23 18:30Z, ~20 minutes after filing, by the filer.** I first wrote that the
> build "then continues without vertical research, or stalls, depending on whether the next stage
> gates on it" — I had not checked which. It stalls, terminally, and §2a is now the most important
> section of this file. Filing a severity before checking the chaining was the wrong order.

> **On the 2026-07-31 owner ruling** (a `bugs_open/` file asserting a cross-cutting root cause is
> not filed until it has been through `090`, or the filing session says plainly why it substituted
> equivalent first-hand verification): **substituted, and here is the why.** The cause is not
> inferred from symptoms — it is read directly from the live `agent_definitions` config (the crawl
> step has no error tolerance; the selector pins no temperature) and confirmed by **two controlled
> observations of the same build** in which the candidate set was identical and only the
> permutation moved, with the failing request_id matching the error on the item both times. There
> is no hypothesis here for a diagnosis loop to narrow: the config is the mechanism and the two
> attempts are the control. What `090` could still add is estate-wide blast radius, which §6 leaves
> as the open question.

## 1. Symptom

A greenfield build's `needs_vertical_research` item fails, retries, and (on this evidence) parks
`failed` after three attempts spanning ~1h37m, having produced no vertical research at all. The
build's own record of the crawl steps says they succeeded.

## 2. The mechanism, in four lines

1. `vertical-exemplar-researcher` asks an LLM (`select_exemplars`) to pick **three** exemplar sites
   for the vertical, then crawls each with `firecrawl_crawl` in three sequential steps.
2. Firecrawl **refuses some domains outright** — `"We apologize for the inconvenience but we do not
   support this site … (code: WEBSCRAPE_ERROR)"`. Large publisher properties are exactly the sites
   an LLM nominates as "the best in this vertical".
3. The crawl steps have **no `on_error`, no `continue_on_failure`, no fallback** — so ONE refusal
   fails the child orchestration (`CHILD_ORCHESTRATION_FAILED`) and **discards the crawls that
   already succeeded**.
4. The refusal is **never persisted anywhere the selector reads**. Each retry re-asks the same LLM
   the same question, so it re-nominates the same sites.

## 2a. Why this is terminal: the step that chains the build is the LAST step, and it is the ONLY producer

`vertical-exemplar-researcher`'s step order is
`select_exemplars → crawl_exemplar_1 → format_1 → crawl_exemplar_2 → format_2 → crawl_exemplar_3 →
format_3 → synthesise → write_landscape_spec → create_next_item → complete`.

**`create_next_item` is what queues `needs_strategy`** (`handler_agent: domain-strategist`,
`item_key_prefix: strategy`, `recurrence_expected: true`). It is reachable only after **all three**
crawls succeed. A refusal at crawl 2 or 3 kills the orchestration, so it is never reached.

**And it is the only thing in the estate that produces that item type** `[MEASURED 2026-08-23 18:29Z]` —
one row, from a sweep of every live agent's steps:

```sql
SELECT type, s.key FROM agent_definitions ad, jsonb_each(ad.default_config->'workflow'->'steps') s
WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
  AND s.value->'config'->>'item_type' = 'needs_strategy';
-- vertical-exemplar-researcher | create_next_item      (1 row)
```

So the documented cascade — *research → strategy → briefing → site plan → composition → design →
pages* — **has a single point of failure at its second hop, with no alternative producer and no
bypass.** `garden-tools.uk` has **zero** `needs_strategy` rows and will never get one from this
attempt. No strategy, no briefing, no plan, no pages, no site. The site row and its four classifier
specs simply sit there.

⚠ ~~**Compound with `bugs_open/326`.** When attempt 3 parks the item `failed`, re-submitting is
suppressed by `writeWorkItem`'s two-strike block for 3h from the sibling's `created_at`, so the
front door stays shut until 20:17:15Z with nothing saying so.~~

> **RETRACTED 2026-08-23 18:35Z, within an hour of writing it — the ground moved while I was
> typing.** Migration **572** is applied and live, and it declares `recurrence_expected: true` on
> the build-chain handoffs, which **skips the two-strike block entirely**. Verified here at the
> live config rather than taken from the report `[MEASURED 18:34Z]` — all five hops now carry it:
>
> ```sql
> SELECT ad.type, s.key, s.value->'config'->>'item_type',
>        COALESCE(s.value->'config'->>'recurrence_expected','(absent)')
> FROM agent_definitions ad, jsonb_each(ad.default_config->'workflow'->'steps') s
> WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
>   AND s.value->'config'->>'item_type' IN ('needs_domain_research','needs_vertical_research',
>        'needs_strategy','needs_briefing','needs_site_plan');
> -- all five: recurrence_expected = true, including domain-submitter.create_research_item
> ```
>
> **So re-submitting after this build dies WILL queue a fresh `research_<domain>`, at any offset.**
> `idx_swi_dedup` is untouched, so a genuinely concurrent duplicate is still refused — that
> protection lives in the database, not in config.
>
> **The lesson is the retraction itself, not the fact.** I filed a compound claim about another
> lane's live mechanism without re-reading that mechanism's config *at the moment of writing* — and
> it had changed within the hour, because that lane was actively fixing it. A cross-bug interaction
> is the least stable claim you can make: it depends on two mechanisms, either of which may be
> under repair by whoever you are writing it for. **Re-read the other bug's config before asserting
> a compound, and date the assertion.**

> **Still true elsewhere, scoped and dated so this line cannot go stale in turn:** 572 declared the
> five build-chain hops only. **As of 2026-08-23 there are 14 keyed steps still undeclared**
> (`scripts/audit-undeclared-recurrence.sh` names the current set — run it rather than quoting this
> number) plus 36 non-test `insertWorkItem` call sites in Go. For any of those, "the work died and
> the front door silently reports `COMPLETED` for the rest of the window" remains exactly right.
> That is `bugs_open/326`'s territory, not this file's — pointer, not a second account.

**Not a motivating case for `RFC_048`** (*the anti-churn brake may delay work but may not destroy
it*), and it should not be cited as one. **The brake never runs in this bug.** `needs_strategy` is
not destroyed by it — the producing step dies upstream, so the item is **never created**. Absence of
work that should have existed, not destruction of work that did; a fix for either leaves the other
untouched. What this bug *does* support is the premise beneath that RFC — that the cascade has hops
with **no producer of last resort** (§2a), so anything stopping one hop stops the build for good.

**What survives the retraction, and it is the whole severity case:** the build is still terminally
dead at hop two. Recovery now requires a human to notice and re-submit; nothing retries it, and the
re-submission will hit the same refused exemplar (§4).

## 3. Evidence — two attempts, same build, `garden-tools.uk` `[MEASURED 2026-08-23]`

Site `16784842-f7d8-4467-bb5b-eb1fb5c1caba`, item `needs_vertical_research` created 17:44:56Z.

| slot | attempt 1 (17:48:27Z) | attempt 2 (18:19:48Z) |
|---|---|---|
| 1 | `gardenersworld.com` — dispatched OK | `gardenersworld.com` — dispatched OK |
| 2 | **`thespruce.com` — REFUSED** | `which.co.uk` — dispatched OK |
| 3 | `which.co.uk` — never reached | **`thespruce.com` — REFUSED** |
| died at | `crawl_exemplar_2` | `crawl_exemplar_3` |
| request_id | `4ac4c952-55c0-4a94-b66d-09bc9cfd3a02` | `1607dc02-cc7f-4a94-b0e2-b165dd58f90d` |

Both request_ids appear verbatim in `site_work_items.error`. Back-off doubled 30min → 60min;
`attempt_count=2/3`, `retry_after=2026-08-23 19:20:32Z`.

**The load-bearing observation: the SET is identical and only the ORDER changed.** Attempt 2 got
one step further and therefore threw away **two** successful crawls instead of one.

## 4. Why "just retry" cannot work, and why that is the interesting part

`select_exemplars` is an `ai_service` step (`claude-sonnet-4-6`, `max_tokens: 1500`) with **no
`temperature` key**, so it samples at the provider default. It is tempting to conclude that a retry
will route around a bad pick. **It does not, and the two attempts above are the disproof.**

Sampling varies the *ordering*; it does not vary the *candidate pool*, because the pool is fixed by
the prompt — *"the THREE best EXISTING websites … the sites a person in this niche would call the
best"* — against a vertical that contains about four such sites. **Retry re-rolls a die whose faces
are all the same.** Any fix built on retrying, widening `max_attempts`, or nudging temperature is
therefore treating the one variable that provably does not move.

## 5. Fix candidates, ordered by what closes the door

1. **Tolerate partial results (smallest, closes the consequence).** N-of-3 is research, not a
   transaction. Give each `crawl_exemplar_*` step an error-tolerant path so a refusal yields an
   empty crawl rather than a dead orchestration, and let `format_*`/downstream proceed on what
   returned. On attempt 2 this alone would have delivered two good exemplars. **Requires a stated
   floor** — decide and record whether 1-of-3 is acceptable research or whether the step should
   fail loudly below 2, because "silently proceeds on one exemplar" is its own defect.
2. **Persist refusals and exclude them at selection (the only one that gets cheaper over time).**
   On `WEBSCRAPE_ERROR`, record the host in a small `firecrawl_unsupported` store and inject it
   into the `select_exemplars` prompt as an exclusion list. This is the only candidate that makes
   the *estate* smarter rather than this *step* luckier, and it converts a recurring cross-vertical
   tax into a one-off per host.
3. **Ask the provider first.** If Firecrawl exposes a support/blocklist probe, check the three URLs
   before crawling any and re-select once, cheaply, rather than discovering it a step at a time.
4. **Not a fix: raising `max_attempts` or setting a temperature.** See §4 — pool, not sampling.

## 6. Blast radius — MEASURED for occurrence, OPEN for exposure

`[MEASURED 2026-08-23 17:50Z]` **one** `site_work_items` row carries `WEBSCRAPE_ERROR` in 30 days —
this one. Nothing in `/bugs_open/`, `/bugs_closed/` or `LANDMINES.md` mentions `WEBSCRAPE_ERROR` or
the refusal string.

**Do not read that as "rare".** It is a count of *occurrences*, not of *exposure*, and the
exposure is one greenfield build per new domain — which this estate has done **twice in a month**.
`needs_vertical_research` has run few times; it has failed on a meaningful share of them. The open
question, which a sweep or `090` could answer: **how many verticals nominate a Firecrawl-refused
exemplar?** Recipes, health, consumer reviews and personal finance all point at exactly the
publisher properties most likely to be on a blocklist.

⚠ **A COUNT OF THINGS CARRIES ITS DATE (owner ruling 2026-08-22):** "one in 30 days" is **as of
2026-08-23**, and this class grows by ADDITION every time a new domain is submitted.

## 7. The trap this bug leaves for whoever verifies the fix

**`collected_data.crawl_exemplar_N` says `success: true` on a crawl that was REFUSED.** That flag
records that the request was published to `system.adapter.webscrape.requests` — it is a dispatch
receipt, not an outcome. The refusal arrives asynchronously and is written **only** to
`site_work_items.error`. So:

- **Do not verify this fix by reading the step outputs.** They were already green.
- Join on `request_id`, which appears in both the optimistic step record and the true error.
- The house rule applies unchanged: *trust the artefact, not the status* — here the artefact is
  whether a `vertical_research` spec row actually exists afterwards.

## 8. How to verify a fix

```sql
-- the artefact, not the item status: did vertical research actually land?
SELECT ss.aspect, ss.source_agent, ss.created_at FROM site_specs ss JOIN sites s ON s.id=ss.site_id
 WHERE s.domain='<domain>' AND ss.aspect LIKE '%vertical%' AND ss.is_current;

-- and that a refusal no longer kills the step:
SELECT attempt_count, status, left(error,120) FROM site_work_items w JOIN sites s ON s.id=w.site_id
 WHERE s.domain='<domain>' AND w.item_type='needs_vertical_research';
```
A fix is proven only against a vertical whose exemplar set **contains a refused host** — otherwise
the test passes for the wrong reason. `thespruce.com` is a known-refused host as of 2026-08-23 and
makes a ready positive control.

## 9. Provenance

Live run: `garden-tools.uk`, submitted 17:17:18Z 2026-08-23 with no prompt, no mission, no seed
(the lane's one-shot route test). Full running record, including the two attempts and a correction
to this filer's own reasoning about them:
`docs/agent_docs/docs024_key_docs_latest/loanzy_uk_example_site/NOTES_loanzy_uk_example_site.md`
(entries 17:49Z, 17:52Z, 18:20Z). Route defects index:
`docs/agent_docs/docs024_key_docs_latest/loanzy_uk_example_site/HANDOFF_2026-08-19_fixing_the_one_shot_route.md`.
