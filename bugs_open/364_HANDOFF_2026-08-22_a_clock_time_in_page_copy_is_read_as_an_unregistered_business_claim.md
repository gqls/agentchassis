# 364 — a CLOCK TIME in page copy is read as an unregistered business claim, and it blocks the whole page build

**Filed** 2026-08-22 by the `site_ai_agent_orchestration` lane, found while unblocking
`ai-agent-orchestration.com`'s `pricing` rebuild (see `bugs_open/` note below — this is the
*second* claims-layer refusal that page hit, and it is unrelated to the first).
**Severity** medium. Nothing false is published; the cost is that a page asserting **nothing
about the business** is refused, and the refusal reads like a real honesty finding.
**Class** false positive in the claims layer / allow-list exclusion.
**Status** FIX WRITTEN AND TESTED IN THIS COMMIT — **inert until the next fleet roll** (Go).

## 1. The symptom

`page-build-handler` fails at `validate_content`, `0 blockers, 1 errors`, and the page is routed
to `needs_human_review` having written nothing:

```
type: unregistered_number   severity: error   category: claims
value: "2"
location: "and the monitoring needed to debug a failing agent chain at 2am are part of what
           gets deployed from day one, because retr…"
description: Unregistered number "2" — number asserted as a fact about the business matches
             no evidence_base fact value (1 occurrence(s))
```

The copy makes no numeric claim. `2am` is a time of day.

## 2. Root cause

`datahelpers.ScanUnregisteredNumbers` (`claims.go:796`) examines every number token unless
`isExcludedNumber` (`:849`) removes it. That function strips composite tokens (`2026-07-16`,
`24/7`, `14:30`, `v1.0.1124`), currency, range dashes, label prefixes (`Tier 2`), and a list of
**unit suffixes** in `unitSuffixRe` (`:681`):

```
px|rem|em|vh|vw|ms|sec|seconds?|min(ute)?s?\s+read|kb|mb|gb|tb|fps|st\b|nd\b|rd\b|th\b|
[-–]\s*(hour|day|week|month|year|minute|second|token|character|step|person|page)\b
```

**`am` and `pm` are not in it.** So the `2` of `2am` survives to the lexical gate,
`businessClaimContextRe` (`:660`), which matches on `agents?` — present in the same sentence
("a failing **agent** chain") and again in `deployed`. The number is then tested against the
evidence base, matches no fact, and is reported.

⚠ **Note the interaction that makes this specific to real sites.** With a populated evidence base
carrying a `gte` fact, a value as small as 2 is often *supported* by accident — any `gte` fact
whose `context_terms` appear in the window will vouch for it. On this site the agent facts'
context terms are narrow (`"agent definition"`, `"agents in the registry"`, …) and none appear in
that sentence, so nothing vouched for it and the page was refused. **The bug therefore fires more
readily on sites whose evidence base is well-scoped** — the better the fact hygiene, the more
exposed the page.

## 3. Why it matters more than one page

The exclusion list is an **allow-list**, so every unit it has never met costs one false refusal.
This is the **3rd** instance of the same shape as of 2026-08-22:

- `bugs_closed/073` — `Read time: 8–12 minutes` read as a business figure; fixed by adding
  `min read` and the en-dash alternative to this same regex.
- `bugs_closed/102` — guides/explainers read as sales claims; fixed structurally, by page type.
- **this one** — clock times.

Clock times are common in exactly the copy this platform writes: monitoring, on-call, SLA and
operations prose ("paged at 3am", "runs at 11pm", "9am to 5pm"). A refusal writes nothing and
routes to human review, so the failure mode is a **stalled page**, not a bad one.

## 4. The fix (in this commit)

One alternation added to `unitSuffixRe`:

```go
… |fps|am\b|pm\b|st\b|nd\b| …
```

**The `\b` is load-bearing** — without it, "450 amazing clients" would be excluded too, silently
blinding the scanner to a real claim.

**Tests** (`claims_series_test.go`, beside the written-out-dates sibling):

- `TestClockTimesAreNotBusinessNumbers` — `2am`, `3 am`, `11pm`, `4AM`.
- `TestFigureBeforeAnAmWordIsStillScanned` — the control: "450 **amazing** clients",
  "120 **ambitious** agents" must still be scanned.

⚠ **Both tests use an EMPTY `&EvidenceBase{}` deliberately, and that is the point.** The first
version of this test used the shared populated fixture and **passed with the fix reverted** — a
`gte` fact was already supporting the value 2, so the test asserted nothing. It was caught by
mutating the regex out and re-running; with an empty base nothing is supported, so any number
reaching the scan is reported and the test can actually fail. Verified: all four cases FAIL
without the fix, whole package passes with it.

## 5. What is NOT fixed

- The allow-list is still an allow-list. The structural fix — deciding by *what the number
  measures* rather than by suffix — is not attempted here and is the standing residual from 073.
  Candidates not yet met: `%` in non-business prose, `k`/`m` magnitude suffixes, ISO durations.
- **This is Go, so it is inert until the next fleet roll.** Until then the workaround is that the
  copy is regenerated on each attempt and will not always contain a clock time — the same
  non-determinism recorded in the `pricing` shrink investigation. Do not read a later success as
  evidence this was fixed; check the build stamp.

## 6. How to verify after the roll

```bash
# 1. the unit test
go test ./platform/orchestration/datahelpers/ -run TestClockTimes -v

# 2. at the artefact: re-fire the page build and confirm no unregistered_number for a clock time
SELECT context FROM agent_error_log
WHERE error_code='CONTENT_VALIDATION_BLOCKER_DETAIL'
  AND site_id='2a8ebf9c-20a2-4c39-b191-840b012371da'
ORDER BY occurred_at DESC LIMIT 1;
```

## 7. Relations

- `bugs_closed/073` (same regex, previous unit), `bugs_closed/102` (page-type gate),
  `bugs_closed/043` (the original invented-claims work).
- The `pricing` page's *other* claims refusal — the evidence base mandating a wording its own
  facts could not validate — is fixed at source by migration
  `557_aiao_evidence_base_stops_mandating_a_phrase_its_own_facts_cannot_validate.sql`.
  **Different cause, same page, same day; do not conflate them.**
- Lane notes: `docs/agent_docs/docs024_key_docs_latest/site_ai_agent_orchestration/NOTES_site_improvement.md`.
