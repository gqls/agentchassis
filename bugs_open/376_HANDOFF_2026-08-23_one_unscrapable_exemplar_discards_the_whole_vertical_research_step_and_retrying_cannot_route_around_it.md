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

| slot | attempt 1 (17:48:27Z) | attempt 2 (18:19:48Z) | attempt 3 (19:20:48Z) |
|---|---|---|---|
| 1 | `gardenersworld.com` — OK | `gardenersworld.com` — OK | `gardenersworld.com` — OK |
| 2 | **`thespruce.com` — REFUSED** | `which.co.uk` — OK | **`thespruce.com` — REFUSED** |
| 3 | `which.co.uk` — never reached | **`thespruce.com` — REFUSED** | `which.co.uk/reviews/garden-tools` — never reached |
| died at | `crawl_exemplar_2` | `crawl_exemplar_3` | `crawl_exemplar_2` |
| request_id | `4ac4c952-55c0-4a94-b66d-09bc9cfd3a02` | `1607dc02-cc7f-4a94-b0e2-b165dd58f90d` | `b480af93-41f8-4816-9234-228c18d57f88` |

All three request_ids appear verbatim in `site_work_items.error`. Back-off doubled 30min → 60min.
Item now **`failed`, `attempt_count=3`** at 19:22:13Z — **1h37m from creation to death.**

**The load-bearing observation: THREE attempts, THREE identical organisation sets, and it died at
whichever slot `thespruce.com` occupied.** Attempt 2 got one step further and therefore threw away
**two** successful crawls instead of one. Attempt 3 nominated a *deep path* for Which?
(`/reviews/garden-tools`) rather than the front page the prompt asks for — a variation in the URL,
not in the set, and irrelevant to the refusal.

> **The caveat, and its DISCHARGE.** All three attempts above read the **same** classifier specs, so
> strictly they showed the selection is stable *given fixed input* — not that the vertical always
> yields this set. **The control has now run and it settles it.**
>
> The domain was re-submitted 19:23:06Z, which re-ran `domain-research-classifier` from scratch. It
> produced **materially different specs** — `industry_tags` came back with 10 entries against the
> first run's 8, and reworded (`buying-guide-platform → buying-guides`, `uk-gardening → uk-retail`,
> new `allotment`/`tool-directory`). Off those fresh specs, `select_exemplars` returned
> `[MEASURED 19:30Z]`:
>
> | slot | attempt 4 — FRESH specs |
> |---|---|
> | 1 | `which.co.uk` |
> | 2 | `gardenersworld.com` |
> | 3 | **`thespruce.com`** |
>
> **The same three organisations, a fourth time, in a fourth permutation, from re-derived input.**
> So the candidate pool is a property of the **vertical**, not of the specs and not of the sampling.
> "Sampling permutes, it does not re-draw" is now unconditional, and any fix that relies on a retry
> eventually picking differently is disproved rather than merely doubted.

## 4. Why "just retry" cannot work, and why that is the interesting part

`select_exemplars` is an `ai_service` step (`claude-sonnet-4-6`, `max_tokens: 1500`) with **no
`temperature` key**, so it samples at the provider default. It is tempting to conclude that a retry
will route around a bad pick. **It does not, and the two attempts above are the disproof.**

Sampling varies the *ordering*; it does not vary the *candidate pool*, because the pool is fixed by
the prompt — *"the THREE best EXISTING websites … the sites a person in this niche would call the
best"* — against a vertical that contains about four such sites. **Retry re-rolls a die whose faces
are all the same.** Any fix built on retrying, widening `max_attempts`, or nudging temperature is
therefore treating the one variable that provably does not move.

## 4a. WHY the pool is fixed — the site-specific inputs never reach the decision (found by the `bugs_open/326` session, verified here)

§4 says the candidate pool is a property of the vertical. True, but the mechanism is one level
deeper and it changes which fix is right.

`select_exemplars`'s prompt says: *"Prefer sites named in `identity.competitors_found` when they are
genuinely strong; otherwise use well-known leaders of the vertical."* **The `competitors_found`
branch has never been taken** `[MEASURED 2026-08-23 19:35Z]`:

```sql
-- 0 of 4 runs chose ANY exemplar appearing in that site's competitors_found; 6 were available each time
```

| run | chosen exemplars | drawn from `competitors_found` | available |
|---|---|---|---|
| 17:48 | gardenersworld / thespruce / which | **0** | 6 |
| 18:19 | gardenersworld / which / thespruce | **0** | 6 |
| 19:21 | gardenersworld / thespruce / which | **0** | 6 |
| 19:30 | which / gardenersworld / thespruce | **0** | 6 |

`competitors_found` holds UK garden-tool **retailers** (`burgonandball.com`, `kentandstowe.com`,
`gardentoolcompany.com`, `gardena.com/uk`, `marshallsgarden.com`, `sgs-engineering.com`); the
exemplars are **editorial** properties. Zero overlap, every time.

**And that input demonstrably varied while the output did not** — run 2's identity spec added
`sgs-engineering.com` (5 → 6), and `industry_tags` went 8 → 10 reworded. `select_exemplars` reads
`{{.site_specs}}` whole, so both were in scope. **Two independent site-specific inputs moved; the
chosen set did not.**

**So the pool is not fixed *despite* fresh specs — it is fixed *because the fresh specs never reach
the decision*.** Every selection so far has come from the fallback branch, i.e. the model's own
priors for a well-known vertical, which are near-deterministic. That is why permutation is the only
thing that varies.

**This separates two fixes that would otherwise be confused:**
- If the pool were genuinely vertical-derived, the remedy is §5's exclusion list.
- If the site-specific branch never fires, there is a **second, separable defect upstream**: either
  the prompt's "genuinely strong" test is too strict, or — more likely here — `identity` is finding
  the wrong *kind* of competitor for the classification. A site classified `hub`/`content` is
  compared against retailers; its real comparators are other content sites. That is a different
  owner from the `on_error` gap.

⚠ **BOUND THE EVIDENCE — and the first version of this bound was itself wrong.**

> ~~`vertical-exemplar-researcher` has **4 runs in its entire history**~~ **CORRECTED 2026-08-23
> 19:36Z, caught by the `bugs_open/326` session and verified here.** That count came from
> `orchestration_states`, which is a **~24-hour rolling window** — measured: its oldest row is
> `2026-08-22 19:22:13Z`, exactly 24h before the query. "Four runs" was the retention window, not a
> history. The durable tables say otherwise:
>
> ```sql
> SELECT count(*), min(created_at)::date, count(DISTINCT site_id) FROM (
>   SELECT created_at, site_id FROM site_work_items         WHERE item_type='needs_strategy'
>   UNION ALL
>   SELECT created_at, site_id FROM site_work_items_archive WHERE item_type='needs_strategy') q;
> --  32 | 2026-04-02 | 27
> ```
>
> **32 items across 27 sites since April.** This step has run many times; four is what survived
> reaping.

**So the honest bound is stronger than "small sample".** The historical *selections* lived in
`orchestration_states.collected_data`, which is reaped — so **"has the `competitors_found` branch
ever fired?" cannot be answered from that table at any sample size.** The 0/4 was not
under-powered; it was measured against a table structurally unable to hold the answer, which is the
`could-not-have-come-out-otherwise` shape.

**What the 0/4 does license:** that across four selections on one domain, with two site-specific
inputs demonstrably varying, the branch did not fire — enough to say it is **unexercised here** and
to justify looking. **It licenses nothing about the branch's history or about the fleet.** Do not
quote it without this paragraph.

**The cheap discriminator** remains one greenfield build in a vertical whose competitors genuinely
*are* content properties. Note also that the branch not firing and the branch **working correctly on
unsuitable input** are indistinguishable on this evidence: `identity` found retailers for a domain
classified `hub`/`content`, so "not genuinely strong" is a defensible reading of the prompt, not a
proven defect. The upstream question stays a question.

## 4b. ⚠ CORRECTION 2026-08-23 20:03Z — §4 and §4a OVERCLAIMED. The pool is BIASED, not FIXED, and the `competitors_found` branch DOES fire

**A fifth selection refutes both.** Attempt 5 (20:02Z, off the run-2 specs) chose
`[MEASURED 20:03Z]`:

| slot | attempt 5 |
|---|---|
| 1 | `gardenersworld.com` |
| 2 | `which.co.uk` |
| 3 | **`burgonandball.com`** |

**`thespruce.com` is absent** — the set re-drew. And `burgonandball.com` is **in
`identity.competitors_found`**, so the branch §4a said had never fired **just fired.**

**What is retracted:**
- ~~"Sampling permutes the order; it does not re-draw the pool."~~ It does re-draw, just rarely —
  **4 of 5 draws** contained the refused host, the fifth did not.
- ~~"Any fix premised on a retry eventually picking differently is disproved."~~ **Not disproved.**
  A retry *can* escape; it is a low-probability escape, not an impossible one.
- ~~"The `competitors_found` branch has never fired."~~ It fired on the next observation after I
  wrote it. The 0/4 was a run, not a property.
- ~~"The candidate pool is a property of the vertical."~~ Too strong. It is **heavily biased toward**
  a small editorial set, and it samples outside it.

**What SURVIVES, and it is still the whole severity case** — none of it depended on the pool being
fixed:
- The crawl steps have **no `on_error`**, so one refusal discards the stage including successful
  crawls. Config-verified. Unchanged.
- `create_next_item` is the last step and the **only** producer of `needs_strategy` estate-wide, so
  a refusal is **terminal** for the build. Unchanged.
- On this vertical the refused host appeared in **4 of 5** draws, so the expected cost is several
  full retry cycles (~30-60min each) before a lucky draw — and `max_attempts=3` means **the item
  usually dies first.** That is a strong argument for fix candidates 1 and 2 and a weak one against
  them being urgent; it is *not* the "retry is structurally futile" argument I made.

**Fix ranking changes slightly:** candidate 2 (persist refusals, exclude at selection) is still the
one that gets cheaper over time, but its case is now "removes a 4-in-5 tax" rather than "is the only
thing that can ever work". Candidate 1 (`on_error`) is unaffected and remains the cheapest real fix.

> **The meta-lesson, and it is the SECOND time today in the same shape.** This morning I built a
> dispatch-walk theory on **14** consecutive ordered samples; it broke 20 minutes later. This
> afternoon I wrote "disproved, not doubted" on **4** consecutive identical samples; it broke on the
> fifth. Both times a run of identical observations was read as a mechanism, and both times the
> counter-example arrived within the hour, for free, because the system kept running.
> **A run is not a law, and the number of repetitions is not the evidence — the absence of a
> counter-example you actively looked for is.** In both cases I could not name what would break the
> pattern, which is exactly the check this repo already prescribes and which I skipped twice.
> **Practical form: state the claim at the strength the sample supports ("4 of 5 draws", "0 of 4
> observations") and never in the modal form ("cannot", "structurally incapable") unless a
> mechanism, not a tally, forbids it.**

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

---

## 10. A SECOND live draw, on a SECOND vertical — hop two PASSED, and it exposed a failure mode this file did not know about `[MEASURED 2026-08-25 10:53–10:54Z]`

**Provenance:** `homegarden.uk` (site `5904bd0f-33fd-4212-9c1b-50b28fe72fdb`), the owner-authorised
greenfield canary, dispatched 10:21:49Z. Orchestration `5937f08b-63ad-4de2-a5ea-97b17cacbb04`.
Raw capture, taken because `orchestration_states` reaps inside ~25h:
`docs/agent_docs/docs024_key_docs_latest/loanzy_uk_example_site/EVIDENCE_2026-08-25_homegarden_hop_two_exemplar_draw_and_crawls.json`.
The `bugs_open/381` lane flagged the draw in flight; the reading below is first-hand and differs
from their summary in one material respect (§10b).

### 10a. The draw, and what it does and does not say about §4b's rate

| slot | exemplar | crawl | formatted result |
|---|---|---|---|
| 1 | `https://www.rhs.org.uk` | `success: true` | **6 sources, `content_quality: good`** |
| 2 | `https://www.gardenersworld.com` | `success: true` | **6 sources, `content_quality: good`** |
| 3 | `https://www.which.co.uk/reviews/home-and-garden` | `success: true` | **0 sources, `content_quality: none`** |

**`thespruce.com` is NOT in this draw**, and hop two passed — the orchestration reached `synthesise`
at 10:54:06Z. **This is a genuine observation and a weak one.** It is a *different vertical*
("Home and Garden Publishing") on a *different site*, so it does **not** make §4b's figure "4 of 6".
The honest statement: **first draw on an adjacent vertical, refused host absent, stage survived.**
`gardenersworld.com` appears in this draw and in all five garden-tools draws; `rhs.org.uk` is new.

⚠ **This is not evidence the bug is milder than filed.** §4b already established the pool is biased,
not fixed. One draw without the refused host is the outcome §4b predicts one time in five.

### 10b. THE NEW FINDING — a crawl can report `success: true`, deliver NOTHING, and the chain proceeds with no floor and no record

`crawl_exemplar_3` returned `"success": true`. `format_exemplar_3` returned
`{"sources": [], "source_count": 0, "research_text": "Crawl completed but no usable page content was
found.", "content_quality": "none"}`. The workflow moved to `synthesise` anyway.

**Read from the live workflow the same minute** (`agent_definitions`, `vertical-exemplar-researcher`,
active non-snapshot): the chain is strictly linear —
`select_exemplars → crawl_1 → format_1 → crawl_2 → format_2 → crawl_3 → format_3 → synthesise →
write_landscape_spec → create_next_item → complete` — with **no `on_error` on any crawl step**
(this file's §2 claim, re-confirmed on live config today) and **no `condition`, and no reference to
`content_quality` or `source_count`, anywhere in it.**

**So there are TWO upstream outcomes, not one, and they need different guards:**

- **(a) the crawl ERRORS** → no `on_error` → the stage dies, discarding whatever already succeeded,
  and `create_next_item` never runs. Terminal. **This file, as filed.**
- **(b) the crawl SUCCEEDS AND DELIVERS NOTHING** → passes straight through. The vertical landscape
  is then synthesised from 2 exemplars while every status in the system says 3. **Silent, and it has
  just happened on a live build.**

**Why this changes the remedy, and it is the whole point of recording it.** §5's fix candidate 1 is
*"`on_error` tolerance on each crawl step (N-of-3 is research, not a transaction), with a stated
floor"*. **A floor evaluated on step success is blind to (b)** — `which.co.uk` would count toward it.
Implemented naively, the estate would tolerate "3 of 3 succeeded" while one delivered zero sources,
and in the degenerate case would write a vertical landscape from **no research at all** with every
step green. That is strictly worse than today's failure, because today's at least stops.

> **The floor must be evaluated on CONTENT — `source_count` / `content_quality` — not on step
> success.** Same disease as this file's own §2 warning that a refused crawl's record reads
> `"success": true`: **the receipt is not the result, and it is not the result in TWO directions.**

**What is NOT claimed here.** I have measured that 2-of-3-with-content proceeds. I have **not**
observed a 0-of-3 build, and I am not asserting the degenerate case from the absence of a gate alone
— though the absence of any `content_quality` reference in the workflow is what makes it the
expected behaviour rather than a guess. **A cheap disconfirming test exists**: a vertical whose three
exemplars are all thin, or a deliberate spec with three unscrapable hosts, should be watched at
`format_exemplar_*` rather than at the item status.

### 10c. Consequence for §8's verification recipe

§8 says a fix is proven only against a vertical whose exemplar set **contains a refused host**. That
stands, and **add a second control**: the fix must also be exercised against a host that crawls
successfully and yields nothing, or it will pass while (b) remains live. `which.co.uk/reviews/home-and-garden`
is a dated, working example of that shape as of 2026-08-25 10:52:41Z.
