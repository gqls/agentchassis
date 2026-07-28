# 083 — the Gauntlet engine returns a bare 503 and throws the reason away, so an intermittent failure cannot be diagnosed

> ## ⚠ THE NUMBER 083 IS AMBIGUOUS — REFER TO THIS CASE BY SLUG
>
> Two unrelated open cases share it, which is the documented trap in
> `bugs_closed/README.md`:
>
> - **THIS file** — `083_…_gauntlet_engine_503_discards_the_error` (vonc.com's
>   debate engine on the island VM). Owned by `gauntlet_dead_cta`.
> - `083_…_detected_findings_never_reach_a_handler` — a *different* bug, about the
>   `detected` work-item queue having no consumer. **Actively worked by another
>   session** (commits `b1b650b00`, `02da9491e`, `75df951c9`, `e2634eeb7`).
>
> `scripts/who-owns.py 083` returns BOTH and warns. Almost every commit message
> saying "083" refers to the other one. Do not read those as activity here.

> ## 🔒 TAKEN ON — 2026-07-27, by the `gauntlet_dead_cta` session
>
> **Status changed: OPEN, unowned → OPEN, IN PROGRESS (owned).**
>
> Ownership checked before claiming it, because `who-owns.py` reads *commits* and
> is blind to a session mid-fix. All seven signals clear as of 2026-07-27:
>
> | check | result |
> |---|---|
> | commits touching THIS file | 1 — `f32f4a003`, my own filing |
> | other docs describing the mechanism | only this workstream's own two files |
> | `git status internal/tools-api/` | clean — nobody mid-edit |
> | commits to `internal/tools-api/` since 07-25 | none |
> | council runs naming `tools-api` (24 h) | none |
> | open `site_work_items` on the 503s | none |
> | memory index | records `gauntlet_dead_cta` as owning `tools-api` + the island |
>
> **If you are another session and want this, say so in this file before starting
> — do not open a competing fix.** Working notes go in
> `docs/agent_docs/docs024_key_docs_latest/gauntlet_dead_cta/NOTES_gauntlet_dead_cta.md`.
>
> **Order of work, per §4:** candidate 1 (log the discarded error) FIRST and alone
> — §4 says nothing below it can be evaluated until it lands, and the truncation
> theory in §2 is explicitly recorded as *not fitting the evidence*. Do not
> speculatively raise `max_tokens`.
>
> **Two things that make this different from a chassis bug**, both easy to get
> wrong: this code runs on the **island VM under docker compose**, not in the
> cluster, so **a chassis image roll does nothing** (re-verified on v1.0.1172 and
> again after v1.0.1174) — it needs an island rebuild per `RUNBOOK_gauntlet_dead_cta.md` §5.
> And it is a change to `internal/`, so it goes through the council gate.

**Filed:** 2026-07-26 · **By:** gauntlet_dead_cta (P4) · **Severity:** MEDIUM —
the flagship journey visibly fails for a visitor, but recovers on retry and the
page degrades honestly · **Status:** OPEN — **IN PROGRESS since 2026-07-27**

## 1. Symptom

`POST /api/v1/tools/gauntlet/position` and `.../defend` intermittently return
`503 {"error":"gauntlet opponent unavailable"}` /
`{"error":"gauntlet judge unavailable"}`. The same request retried usually
succeeds. Failures are **bursty rather than steady** — clustered in time, with
long clean runs either side.

Measured today, all against the live island through Cloudflare:

| window | calls | failures |
|---|---|---|
| first probe | 1 `/defend` | 1 failed at 24.8 s, then again at 25.0 s on the same round, then **succeeded** on the third attempt at 14.3 s |
| 5-round curl sample | 5 `/position`, 5 `/defend` | 0 |
| live browser verification, ~15:2x | 2 `/position` (one per profile) | **2 failed** — both recovered on the automatic retry |
| 8-round curl sample, immediately after | 12 `/position`, 11 `/defend` | 0 |

So: **not reproducible on demand, and not rare enough to ignore** — it hit the
first attempt of both profiles during one verification window and nothing either
side of it. [UNMEASURED] no honest overall rate can be quoted from this; the
clustering is the finding, not a percentage.

Latency when it succeeds (23 samples): `/position` 8.2–12.9 s, `/defend`
9.9–23.4 s. Failures land at ~25 s, i.e. **longer than a typical success**.

## 2. Why it cannot be diagnosed from outside — this is the actionable part

`internal/tools-api/handlers/defend.go:89-93` (and the equivalent in
`position.go:83`):

```go
text, err := client.GenerateText(ctx, prompt, map[string]interface{}{})
if err != nil {
    httperr.JSONError(c, http.StatusServiceUnavailable, "gauntlet judge unavailable")
    return
}
```

`err` is **discarded** — not logged, not wrapped, not counted. There is no
request logging on the island either (`docker compose logs tools-api` shows only
gin's startup banner), so nothing anywhere records *why* the call failed. Every
distinct upstream condition arrives at the client as the same opaque 503:

- an HTTP error from `api.anthropic.com` (429, 529 overloaded, 500) —
  `aiservice/anthropic.go:150` returns `API request failed with status %d`
- a network/timeout failure — note `anthropic.go:63` builds
  `&http.Client{}` with **no timeout at all**, so nothing here bounds the call
- **a truncated completion** — `anthropic.go:209` returns a non-nil
  `TruncatedError` when `stop_reason == "max_tokens"`, and the handler cannot
  tell that apart from a real failure. Default `max_tokens` is 2048
  (`anthropic.go:72`) and both handlers pass empty options, so nothing raises it.

> **A note on a theory that does NOT fit, recorded so nobody re-walks it.**
> Truncation is the tempting explanation because the failures are *slower* than
> the successes and CLAUDE.md warns about exactly this class. But the successful
> responses measure only ~373 output tokens — nowhere near 2048 — so a 5×
> overrun would be needed, and the ~25 s failure time is too fast for 2048
> tokens at the observed generation rate. **NOT asserted, and not to be treated
> as the cause without evidence.** The point of this bug is that the evidence
> does not currently exist.

## 3. Blast radius

`vonc.com/tools/gauntlet/index.html` — the Position and Defence steps. The
front-end handles it correctly (honest "the AI opponent is offline" message, no
clock started, no objective marked, the visitor's text preserved), so this
degrades rather than breaks. But it is the flagship journey, and a visitor who
gets it on their first attempt may not try again.

It also produced the only failing check in the P4 live acceptance run —
`no_console_errors`, because the browser logs each 503. That check is behaving
correctly; the upstream is what is wrong.

## 4. Fix candidates, in order

1. **Log the error before returning the 503**, in both handlers, including the
   upstream status and whether it was a `TruncatedError` (`aiservice.IsTruncated`
   already exists for exactly this). One line each. **Nothing else on this list
   can be evaluated until this lands** — everything below is currently a guess.
2. **Give the HTTP client an explicit timeout.** `&http.Client{}` with no
   timeout means a hung upstream is bounded only by whatever the platform
   happens to impose.
3. **Retry once on a transient upstream failure** (429/529/timeout), with a
   short backoff. The evidence that a retry usually works is strong: every
   observed failure succeeded on a subsequent attempt.
4. **Raise `max_tokens` for the judge** only if (1) shows truncation actually
   occurring — see the note above; do not do this speculatively.

Any of these is a change to `internal/`, so it goes through the council gate per
CLAUDE.md, and the island must be rebuilt and `compose up -d`'d for it to take
effect (`RUNBOOK_gauntlet_dead_cta.md` §5) — a fix committed in-repo is inert.

## 5. How to verify a fix

After (1), reproduce by sampling until a burst appears, then read the island log
for the recorded reason:

```
ssh root@toolsapisuk.vs.mythic-beasts.com 'cd /opt/island && docker compose logs --since 30m tools-api | grep -i judge'
```

A fix for (3) is verified when a sampling loop of ~30 rounds returns no 503 to
the caller **while the underlying transient is still occurring** — i.e. the log
from (1) must still show retried failures. A quiet period proves nothing here,
because the fault is bursty: 23 consecutive clean calls were recorded today
within minutes of two live failures.

---

## 6. WORK LOG — candidate 1 written and committed, 2026-07-27

**Commit `a37a2037c`** · **Council `SUBMISSION_CORR = e004fd81-5126-45c0-b580-635a28187995`**
(submitted ~18:24Z; verdict pending — the dispatch lane was 8 deep, one item at 373 s).
**No `Council-Reviewed:` trailer on `a37a2037c`**, deliberately: a trailer is earned
by an APPROVED verdict, and this verdict post-dates its commit, so it can never
carry one honestly. If APPROVED, the trailer goes on a follow-up commit.

### It was SEVEN discard sites, not the two §2 names

§2 quotes `defend.go:89-93` and `position.go:83`. Reading the handlers found that
**both** LLM endpoints discard at `client_init` **and** at `generate`, and that
**both additionally have two unlogged branches returning a *different* 503
message** — `"gauntlet judge/opponent response was invalid"` — which §1 never
separated from `"unavailable"`. That distinction is load-bearing: if the live
failures were the *invalid-response* kind, the cause is a malformed completion
rather than an upstream outage, and §4's candidate order points the wrong way.
The log now tells them apart.

### A THIRD endpoint was affected, and its status code was destroying the evidence

`round.go` discarded its error too, **and returned a literal `502`**:

```go
prov, err := FetchProvocation(domain)
if err != nil {
    httperr.JSONError(c, 502, "provocation unavailable")
    return
}
```

Commit `b498df16b` moved `/position` and `/defend` from 502 to 503 precisely
because **Cloudflare replaces an origin 502's body with its own HTML** — so this
endpoint's JSON error shape never reached the browser at all. `/round` was missed
by that fix. It is the endpoint the other two depend on, so its failures were the
most consequential to misreport. Now 503, and logged.

### The service had no request logging whatsoever

`api/server.go` ran `gin.New()` with only `Recovery`. That is the real reason §5's
`docker compose logs tools-api | grep -i judge` was never going to find anything,
and why §1 carries `[UNMEASURED]` on the rate — **there was no denominator.**
`gin.Logger()` is now attached, ahead of `Recovery` so a request that panics is
still recorded.

### What was deliberately NOT done

`max_tokens` is untouched. §2 records the truncation theory as *not fitting the
evidence*, so this **instruments** the question rather than guessing at it: a
truncation now logs as `TRUNCATED` carrying the provider's own `reason` and
`output_tokens`, distinctly from a generic `FAILED`. **If that branch never fires,
that is itself the finding** — and §4's candidate 4 can be closed rather than
attempted. Candidates 2 and 3 are untouched, per §4's own ordering.

### The tests discriminate — verified by inducing the fault, not by going green

A passing test proves nothing by itself, so the naive implementation was induced:
a version that still *detects* truncation but reports it generically **compiles,
still logs a line**, and fails exactly `TestLogAIFailure_NamesTruncationDistinctly`
and `TestLogAIFailure_FindsTruncationThroughWrapping`. Restored, re-run green.
(A first attempt at inducing it left `aiservice` imported-but-unused, so the
package failed to BUILD — a build failure is not a discriminating test result,
and it was redone.)

### STILL OPEN — this fix is INERT

- **The island has not been rebuilt.** Committing changes nothing on
  `tools.apis.uk`; it needs `RUNBOOK_gauntlet_dead_cta.md` §5. **A chassis image
  roll does nothing for this service** — re-confirmed today across v1.0.1172,
  v1.0.1174 and v1.0.1175.
- §5's verification is unrun for that reason: it needs the deployed build, and
  then a burst to be sampled.
- **So this bug stays OPEN.** Per `/bugs_closed/README.md` the bar is fixed AND
  live; a fix committed but inert until a rebuild does not meet it.

---

## 7. COUNCIL: round 1 REVISE → round 2 APPROVED (corr `e004fd81-5126-45c0-b580-635a28187995`)

**Round 1 — REVISE**, `decided_by: gating objection from editquality` (8 reviews:
6 approve, 2 object). Note the metadata read `"abstained": 8` of `"reviewers": 8`,
which is **not** what it looks like — it is the filtered-seat counting artefact,
not eight abstentions. Read the `reviews` array in `body`, never the counters.

**The gating objection was worth more than its literal claim.** editquality said
only that it could not tell an abridged sketch from missing work. The code *was*
complete for the AI paths — but checking properly showed **my stated count of
"seven" was wrong (it was nine)**, and auditing every error return then exposed
**seven further discarded errors on the 500 paths** that nobody had flagged. A
reviewer who could not see the diff still roughly doubled the fix's coverage.
Final: **16 of 16 5xx fault paths logged; 6 of 6 4xx caller paths deliberately not.**

**Round 2 — APPROVED**, "with 2 advisory objection(s) — none high-severity"
(9 reviewers: 7 approve, 2 object, advisory only). Commit
`9474e6b68` carries the `Council-Reviewed:` trailer; the three earlier commits
(`a37a2037c`, `7f281cea9`, `74795f6ef`) cannot, because the verdict post-dates
them — a trailer asserted ahead of its verdict is a permanent false claim.

### The four advisories, and what was done about each

1. **guardian [medium] — the logged snippet needs a HUMAN, not this council.**
   `logAIBadResponse` writes a 300-char-capped extract of live model output, which
   on some failure shapes echoes the visitor's own argument text. Guardian
   explicitly says *"the plan itself flags this as needing human sign-off rather
   than code review and does not resolve it — that is the right call, but it means
   this item cannot be closed by this council alone."*
   **STATUS: OPEN, awaiting an owner ruling.** If the answer is no, the snippet
   reduces to a shape summary (length + first byte), at the cost of no longer
   distinguishing a prose wrapper from a double-JSON emission — which is the one
   thing it exists to do.

2. **guardian [low] — unbounded log driver.** Correct, and it was a risk *this
   change created*: attaching `gin.Logger()` took the service from near-silent to
   a line per request, into compose's unbounded `json-file` default, on a VM where
   a full disk takes Postgres down with it. **FIXED, not deferred** —
   `infra/island/docker-compose.yml` now pins `max-size: 10m`, `max-file: 3`.

3. **debug_historian [medium] — I overstated the blast-radius answer.**
   My claim that "no un-allowlisted origin can reach /round" rests on
   `store/sites.go:34`: `SELECT id, domain FROM sites WHERE status = 'deployed'
   AND domain = $1`. That is **exactly the documented `sites.status` trap** —
   status is informational elsewhere in this platform, and blast radius should not
   be scoped by it without enumerating the real values first.
   > **CORRECTED:** the claim stands only as far as *"CORS filters by
   > `status='deployed'` on the island's own `sites` table"*. Whether that table
   > contains anything besides vonc is **[UNVERIFIED]** — the island DB is not
   > reachable from the cluster. **Owed at rebuild time, when SSH is open anyway:**
   > `docker compose exec -T postgres psql -U tools_api -d tools_api -c "SELECT domain, status FROM sites;"`

4. **debug_historian [low] — no precedent check.** Fair; run now.
   Nine prior `council_report` rows touch this surface, all this workstream's own
   build rounds (`70c8893b`, `64e6112c`, `cff7ff61`, `c379f7b7` …). **Concordant,
   not contradictory** — none litigated the error-discard pattern or reached a
   different conclusion on it; they reviewed the service being built, not how it
   reports failure. Query for the next person:
   ```sql
   SELECT created_at::date, correlation_id, metadata->>'decision'
   FROM diagnosis_artifacts WHERE kind='council_report'
     AND (body::text ILIKE '%tools-api%' OR body::text ILIKE '%defend.go%')
   ORDER BY created_at DESC;
   ```

**083 REMAINS OPEN.** Approval is not deployment: the island has not been rebuilt,
so every line of this is still inert, and `/bugs_closed/`'s bar is fixed AND live.

---

## 8. OWNER RULING 2026-07-27 — advisory 1 CLOSED: fingerprint, not text

The council's one item it said it could not close itself (guardian/medium) has
been ruled on: **the failure log records the SHAPE of an unusable response, never
its text.** Commit `1e2762809`.

**What decided it.** The excerpt was capped at 300 characters, and
`bugs_closed/088`'s shape is *"a complete JSON object, then commentary, then a
second complete JSON object"* — a complete judge verdict with real reasons runs
past 300 chars on its own, so **the second object starts ~1,500 chars in and the
excerpt could never see it.** The excerpt could not detect the case that
justified it. That turned a trade-off into no trade-off at all.

`aiservice.Fingerprint` (new, `platform/aiservice/fingerprint.go`) reports:

```
chars=1834 first=L fence=yes objects=2 parses=false keys=[]
```

which answers every question the excerpt was there for — prose wrapper
(`first` ≠ `{`), markdown fence, **double JSON (`objects=2`)**, empty completion,
parsed-but-empty-fields — while emitting no model or visitor text at all. Its
object scanner is string- and escape-aware; a brace inside a quoted reason would
otherwise miscount, and the count is the whole point.

**Automated, so it stops depending on a reviewer noticing.**
`scripts/pattern-check.py` gains `check_logged_model_output` (8th check;
advisory, runs in `.githooks/pre-commit`). It flags a log sink passing an
unwrapped payload identifier inside any package that calls `GenerateText`.

> **MISSTEP, recorded because the check itself nearly shipped useless.** The
> first version gated on *the file* containing `GenerateText` and was
> **VACUOUS**: the LLM call lives in `defend.go`, the log sink in its sibling
> `ailog.go`, so it scanned nothing and reported clean. **A check that reports
> clean and a check that is not running look identical.** Caught only by a
> positive control — auditing `a37a2037c`, the commit that introduced the raw
> excerpt, and demanding a finding. Now gated on the package and verified three
> ways: **flags `a37a2037c` `ailog.go:72`; silent on the fingerprint version;
> 0 findings fleet-wide.**

### Remaining advisories

- guardian/low (log rotation) — **CLOSED**, `max-size: 10m`/`max-file: 3`.
- debug_historian/medium (`sites.status` blast-radius claim) — **corrected in
  place and marked [UNVERIFIED]**; enumeration owed at rebuild.
- debug_historian/low (precedent check) — **CLOSED**, run and concordant.

**083 still OPEN**: the island has not been rebuilt, so all of this is inert.

---

## 9. DEPLOYED 2026-07-27 — candidate 1 is FIXED AND LIVE; 2–4 are now unblocked

Island rebuilt on **`aqls/tools-api:v1.0.1178`**, built from committed HEAD via
`make build-tools-api-ref`, shipped `docker save | gzip | ssh docker load`,
container recreated. Commit `a0d275916`.

**Verified against the RUNNING container, never the tag** (RUNBOOK_island.md):

| check | result |
|---|---|
| `strings /tools-api \| grep -c logInternalFailure` | **4** (0 on v1.0.1163 — a real before/after pair) |
| `logAIBadResponse` / `TopLevelJSONObjects` | 4 / 1 |
| negative control `logNeverExisted` | **0** — the grep discriminates |
| `HostConfig.LogConfig` | `max-size 10m, max-file 3` — rotation live |
| container image | `v1.0.1178`, recreated (not a kept container) |

### The failing branch was INDUCED — a green path proves deployment, not correctness

Run off production against a throwaway Postgres and the **identical image**, with
a deliberately invalid API key:

```
v1.0.1178 →  gauntlet/position: generate FAILED round_id=cbf64469-4d06-46da-bac2-f642e6822d4b
             err=API request failed with status 401: {"type":"error","error":
             {"type":"authentication_error","message":"invalid x-api-key"},
             "request_id":"req_011CdT8yka1p5Ahy3Ye5gBU2"}
             [GIN] … | 503 | 164.116709ms | POST "/api/v1/tools/gauntlet/position"

v1.0.1163 →  0 diagnostic lines, 0 request lines
```

**That silence is the whole of this bug.** The upstream status, the provider's
own error type and its `request_id` are all now on the record, and the request
log gives the failure a denominator for the first time.

Live round-trip after the roll: `/round` 200 · `/position` 200 (10.0 s, real
counter+challenge) · `/defend` 200 (14.9 s, real verdict). Smoke: preflights 204,
denied origin 403, missing round 404, `vonc.com/tools/gauntlet` 200.

### Why this bug stays OPEN

The **diagnosability** defect in the title is fixed and live. The **intermittent
503s themselves are not** — nothing has been done to stop them, and by design:
§4 orders candidate 1 first precisely so 2–4 stop being guesses.

**Next step is to wait, not to code.** The fault is bursty (23 clean calls minutes
after two live failures), so it must be *caught*, not reproduced:

```bash
ssh root@toolsapisuk.vs.mythic-beasts.com \
  'cd /opt/island && docker compose logs --since 24h tools-api | grep -E "gauntlet/(round|position|defend): "'
```

Then read what it says and act on THAT:
- `err=…status 429/529` → candidate 3 (retry once on a transient).
- a timeout / context deadline → candidate 2 (the client is still `&http.Client{}`
  with no timeout, `anthropic.go:63`).
- `TRUNCATED` → candidate 4 becomes justified. **§2 predicts it will not appear;
  if it never does, close candidate 4 as refuted rather than leaving it open.**
- `response UNUSABLE … objects=2` → a bugs_closed/088-class double emission, which
  is a prompt fix, not a transport fix.

[UNMEASURED] no failure has been captured in the wild yet — the build is 5 minutes
old at time of writing. **Recheck after 24–48h of real traffic.**

---

## 10. 2026-07-28 — the logging is armed and has caught NOTHING, and §9's own next step was wrong

> **CORRECTED — §9 said "wait 24–48h of real traffic, then read the log". That
> premise is false: THERE IS NO REAL TRAFFIC.** The island's request log for the
> 24h after the roll holds **8 lines, all of them mine** from the 19:43–19:47
> verification. Waiting cannot produce evidence here, and the guidance would have
> had the next session wait indefinitely for a burst that no visitor is generating.
> This only became visible *because* the fix added a request log — the first
> measurement it produced was that the denominator is zero.

### What was done instead — deliberate sampling, per §5

§5 says "reproduce by sampling until a burst appears", which is the correct
instruction; §9 contradicted it. 12 rounds fired at the live engine
(`p4_sources`-adjacent script, 24 LLM calls):

| | result |
|---|---|
| client-side | **24 of 24 LLM calls 200** |
| server-side | **36 of 36 requests 200** (12 × round+position+defend) |
| latency | 5.9 – 22.5 s; slowest server-side 22.1 s — inside the known band |
| **fault lines logged** | **0** |
| `TRUNCATED` lines | **0** |

### Cumulative: no failure since 2026-07-26

Clean LLM calls recorded since the last observed failure, by window:

- 07-26, the 8-round sample taken immediately after the two live browser
  failures — **23**
- 07-27, post-deploy round-trip — **2**
- 07-28, this sample — **24**

**≈49 consecutive clean LLM calls across three days.** The fault has not recurred.

### What that changes

- **Candidate 3 (retry on a transient) is NOT justified by current evidence.** The
  thing it would paper over has not happened in 49 calls.
- **Candidate 4 (raise `max_tokens`) is refuted so far**, exactly as §2 predicted:
  the `TRUNCATED` branch exists, is live, and **has never fired**. Do not raise the
  cap. If it still has not fired after the next real burst, close candidate 4.
- **Candidate 2 (explicit HTTP client timeout) stands on its own merits, and is
  MORE expensive than §4 implies.** `&http.Client{}` (`anthropic.go:63`) is
  **fleet-wide chassis code — 17 Go files reference `aiservice`** — so this is not
  an island-only edit. An unbounded client is a genuine latent defect for every
  agent, not just this one, and it should be argued on that basis with its own
  council round, not slipped in as a fix for a burst that has stopped.

### Correct posture now: leave it armed, do not speculatively fix

The engine's failures are diagnosable for the first time and nothing is failing.
The evidence to act on does not exist yet, and manufacturing a fix without it is
what §4's ordering was written to prevent.

**Re-check trigger, not a date:** whenever the Gauntlet next gets real use, or
after any Anthropic incident, run:

```bash
ssh root@toolsapisuk.vs.mythic-beasts.com \
  'cd /opt/island && docker compose logs tools-api | grep -E "gauntlet/(round|position|defend): "'
```

[UNMEASURED] no failure has been observed under the new build, so the fix's
ability to capture a REAL burst is proven only by the induced fault in §9, not by
a wild one.
