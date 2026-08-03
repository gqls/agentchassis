# NOTES — bugs_open/123, content-creator's copy cannot reach the claims assessor

Append-only, newest at the bottom. Technical log: evidence, commands, what the
system actually said, and every misstep.

---

## 2026-08-03 — session start: bug selection, and why 123

Ran the sweep the owner asked for: pick the next `bugs_open/` case no other thread
is working. Two other sessions are running the **identical** prompt right now
(their first user message is byte-similar to mine), so "unowned" had to be
established against live sessions, not only against git.

**How ownership was actually measured**, because `scripts/who-owns.py` alone was
not enough:

1. `who-owns.py` returned **OWNED or recently active** for *every* candidate I
   tried (033, 066, 071, 084, 085, 093, 096, 107, 113–116, 120–123). Its verdict
   fires on any commit touching the bug file in 14 days, and a triage sweep on
   2026-07-27 touched most of them at once. **[MISSTEP, mine]** I ran it 16 times
   before reading its source and noticing the verdict could not discriminate.
   The useful part of its output is the *owning workstream* block, not the verdict
   line.
2. `git log` per bug file + commits whose SUBJECT names the number in the last 4
   days. That separates "was swept" from "is being worked".
3. **Live session transcripts**, because every git-based check is lagging by
   construction (`~/.claude/projects/-home-ant-projects-agentchassis/*.jsonl`,
   modified in the last 6h). Counting `bugs_open/NNN` per transcript names what
   each live session is actually on: 182, 091, 149, 010, 183, 098, 175, 117, 151,
   168, 139, 172, 140, 153, 115, 178, 138, 155, 159, 157, 128, 150.
   **123 appears in none of them as a subject.**

**[MISSTEP, mine — the measurement answered the question I encoded]** My first
validity check on a different candidate (`bugs_open/116`, "the link-integrity
checks have never run") queried `site_work_items` for item types
`phantom_internal_links`, `dead_controls`, `misdirected_cta` and got zero rows,
which looked like confirmation. Those are **check names**. The **item types** are
singular — `phantom_internal_link` (25 rows, latest 2026-08-03), `dead_control`
(2), `cta_names_unknown_destination` (40). 116's central claim is **stale**: the
agent runs and the checks do produce findings. The cheap check I skipped: list
the distinct `item_type` values the agent actually produced before asserting an
absence, instead of filtering on names I had read in prose.

```sql
-- the query that corrected it
SELECT item_type, count(*), max(created_at)::date
FROM site_work_items WHERE created_by='completeness-discovery-agent'
GROUP BY 1 ORDER BY 2 DESC;
```

### 123 re-verified against the live system before taking it on

| claim in the bug file | check run 2026-08-03 | verdict |
|---|---|---|
| content-creator has no validation of any kind | `grep -rniE "validate\|claims\|evidence\|fabricat\|banned" internal/agents/contentcreator/` → **0 matches** (828 lines, one file) | STILL TRUE |
| it writes nothing to `llm_call_log`, so no per-agent usage query sees it | 14-day `llm_call_log` group-by `agent_type ILIKE '%content%'` → `content-quality-auditor`, `page-content-writer`, `content-reviewer`, `content-gap-planner`. **No `content-creator`** | STILL TRUE |
| the claims machinery cannot be pointed at site-less text | `save_sections_claims_guard.go:81` says so in terms: *"bugs_open/123's content-creator path has no site and no page row, so this seam cannot reach it at all"* | STILL TRUE, and now stated by the platform itself |
| "which workflows call content-creator is untraced" | **NOW MEASURED** — see the correction below | CORRECTED |

### CORRECTION to 123's severity, measured not assumed

123 is filed **HIGH** ("HIGH if any content-creator output is published") and asks
for an owner call "before blog or social output is published anywhere". The
untraced half is now traced, and it lowers the severity:

```sql
SELECT type, is_active, COALESCE(is_snapshot,false) snap, deleted_at IS NOT NULL del
FROM agent_definitions WHERE default_config::text ILIKE '%content-creator%';
--  website-builder           | f | f | t
--  multipage-website-builder | f | f | t
--  multipage-website-builder | f | f | t
```

**All three rows that reference it are deleted and inactive.** No live agent
definition dispatches content-creator. It is reachable only by a direct Kafka
publish to `system.agent.content-creator.requests` — which is exactly how the
2026-07-27 fabrication was produced.

So the honest reading is **latent, not biting** — the same shape as
`bugs_open/134`. It is still worth fixing and the fix is still cheap, but the
"owner call before publishing" framing overstates what is live today. The service
itself is deployed and running (`content-creator-agent-8576b699d4-rgmq2`,
1/1 Running), so a guard placed in it is not a mechanism rotting unexercised — it
is on the path any direct dispatch takes.

### What the platform has grown since 123 was filed (2026-07-27 → 2026-08-03)

This is the part that changes the fix, and it is why the case is worth taking now
rather than when it was written:

- `datahelpers.ScanAllBannedClaims(blocks, eb)` (`claims_global.go:248`) is
  **nil-safe by design** — *"eb may be nil, and a nil eb means 'this site has no
  register', not 'do not scan'"*. That is precisely the site-less entry point
  123's fix candidate 1 asks someone to build. It already exists, shipped by
  `bugs_open/104`'s lane on 2026-07-28, the day AFTER 123 was filed.
- The **claims floor** (`save_sections_claims_guard.go`, `bugs_open/149` C1)
  established the estate's severity precedent: a banned claim REFUSES the write; an
  unregistered number is RECORDED and allowed, because "number extraction has false
  positives by design".
- `platform/voicestyle/` is the working precedent for one shared package consumed
  by both the chassis (`*sql.DB`) and content-creator (`pgxpool.Pool`).

So the fix is now mostly *wiring*, not building — which is the outcome CLAUDE.md's
"reuse existing machinery before building new" is aiming at.

**Still an open gap, and it is the half that matches the observed damage:** the
fleet-wide banned set is about **self-accuracy overclaims** ("guaranteed accurate",
"independently verified", "you can rely on us"). Read all ten patterns at
`claims_global.go:111-215`. **None of them matches "Industry data shows … between
3% and 10%"** — an attributed-but-uncited statistic. That is 123's fix candidate 3
and it needs a new detector.

---

## 2026-08-03 — the constraint that shapes the fix: the delivery path is another lane's

Grepped `LANDMINES.md` for the symbols I was about to touch **before** touching
them (the SessionStart hook only matches files already dirty in the tree, and
`internal/agents/contentcreator/` was clean). It caught something that changes the
design:

> `platform/kafka.DeliverReply` returns the outcome that stops the silent starve —
> and a caller that ignores it compiles, passes platform/kafka's tests, and is unfixed
> · **fires when:** you adopt the shared reply-delivery policy at one of the 8
> remaining log-and-return sites (`bugs_open/158` item 1: reasoning, **contentcreator
> ×2**, websearch ×2, thunder ×2)

So `sendSuccessResponse` / `sendErrorResponse` in `agent.go` are **owned by the 158
lane, which is active right now**. Two consequences, and the second is the more
interesting:

1. I must not edit either function, and must not adopt `DeliverReply`.
2. **A refusal design is ruled out on record, not on taste.** The same entry
   records that the council's architecture seat ruled on round `7478233b` that
   widening that adoption "is an RFC moment ... because it changes 4 services'
   caller-observable failure behaviour from timeout to error response". Making
   content-creator return an error where it previously returned text is the same
   class of change to caller-observable behaviour. The claims floor REFUSES a
   banned claim, and the temptation was to mirror it for symmetry; the floor sits
   on a *persistence* seam where refusing means "the old content stays", whereas
   here refusing means "the caller gets an error instead of prose". Different
   seam, different affordance.

So the design space is **record + annotate, never refuse**, leaving the delivery
path byte-identical.

### Where a site-less finding can be recorded — checked, not assumed

`agent_error_log.site_id` is **NULLABLE** (`\d agent_error_log`), and NULL-site
rows are already the majority shape:

```sql
SELECT site_id IS NULL AS site_is_null, count(*), max(occurred_at)::date
FROM agent_error_log WHERE occurred_at > now() - interval '7 days' GROUP BY 1;
--  f | 201 | 2026-08-03
--  t | 283 | 2026-08-02
```

So the claims floor's "a pod log line is not a record" rule can be honoured here
by exactly the mechanism it uses, with no schema change and no invented shape.

**[UNVERIFIED]** Whether the immune-system sweep actually *reads* NULL-site rows.
I have established the row can be written and can be queried by `error_code`; I
have NOT established that anything sweeps it. Do not write "the immune system will
pick this up" anywhere until that is checked.

### The seam in the code

`handleMessage` (`agent.go:192`): generation returns at `:290`, the response is
assembled at `:332-346`, and `sendSuccessResponse` is called at `:348`. The scan
belongs between `:297` and `:332` — after generation has succeeded, before the
payload is built, so the findings can ride in `Metadata` without touching the
delivery call. `a.db` is a `*pgxpool.Pool` and is **optional** (`NewAgent:127-136`
runs the agent without it if the DB is unreachable), so any durable record must
degrade to a log line rather than assume a connection.

---

## 2026-08-03 — blast radius MEASURED before the plan came back, and it changed the design

`bugs_open/124` drew a REJECTED verdict for asking the council to check a
blast-radius claim instead of measuring it. So this was measured first, with the
platform's own engine (`cmd/claimscan`, CLM-014) over the **complete live corpus,
1,130 components** exported today.

**Re-measure, never quote:** the corpus was 908 components on 07-28, 919 on 07-29,
949 on 07-30 and is **1,130 today**. Every figure in the register entries above is
already stale.

Candidate pattern set for the attributed-but-uncited statistic (4 patterns, run
with `-no-global` so only the candidates fire). This is an **UPPER BOUND**: the
real detector subtracts sentences that carry a citation, and this run has no such
subtraction.

```bash
go run ./cmd/claimscan -evidence cand.json -no-global -components corpus.tsv | grep '^BANNED'
```

**14 findings / 1,130 components = 1.24%.** By pattern: 9 × "bare percentage-of-
population", 3 × "industry data near a figure", 2 × "research shows near a figure".

### Reading them one by one is what earned the design — 5 of the 14 are FALSE POSITIVES, in three distinct ways

| finding | verdict |
|---|---|
| `can-you-trust-ai-with-your-data` "66% of people" ×7 | **TRUE** — an unsourced statistic, live |
| `learn-index` / `learn-security-xss-vulnerability` / `tool-csp-builder` / `tool-jwt-inspector` "90% of websites" | **TRUE** ×4 — same unsourced figure repeated across four ported pages |
| `guide-total-cost-of-borrowing` "51% of people" | **TRUE** |
| `ai-data-trust-in-healthcare` "97% of organisations", `…financial-services` "78% of customers", `…hiring-and-hr` "87% of organisations" | **TRUE** ×3 |
| `index` + `news` "research showing its AI orchestration framework reduces token usage by 38%…" | **FALSE** ×2 — the sentence begins `[VentureBeat] Writer released research showing…`. The attribution IS cited; the citation just sits outside my window |
| `noticias-index` "market analysisChrono24" | **FALSE** — a news listing whose source (Chrono24) is fused to the match by the HTML extraction |
| `can-you-trust-ai-…` and `…financial-services` "industry reported more AI-related incidents in the first half of 2026 than in all of 2025" | **FALSE** ×2 — the `\d` my pattern required is **the YEAR**. No statistic in the sentence at all |

**Three lessons, all of which the plan must carry:**

1. **A citation subtraction is not optional, it is the detector.** Without it the
   news-listing class alone produces false positives on every site that carries a
   feed — and those are the pages most likely to quote a sourced figure.
2. **A bare `\d` matches a year.** `datahelpers` already has the guard for exactly
   this shape — `isExcludedNumber` (`claims.go:807`) drops composite tokens: dates,
   times, versions, currency, en-dash ranges, "found by measurement 2026-07-26" on
   a live sweep. **Reuse it; do not write a second one.** This is CLM-017's
   recorded landmine ("narrowing a pattern by reasoning can make it inert") in its
   mirror form: widening by reasoning made mine noisy, and only the corpus said so.
3. **The nine true positives are the argument for the detector and against
   blocking on it.** Eight live components on four sites assert an unsourced
   population statistic today. A blocking detector would strand all of them on
   their next rebuild — and the observed 5/14 false-positive rate is exactly the
   reason the estate's numeric scan is "never a blocker". So: **opt-in field,
   default OFF, record-and-annotate.** That is not deference to a rule, it is what
   this measurement says.

**[UNMEASURED]** How often the shape appears in content-creator's *own* output. It
has produced no corpus to sample — one known generation, one fabrication in it.
The 1.24% above is a page-copy rate and must not be quoted as a generation rate.
