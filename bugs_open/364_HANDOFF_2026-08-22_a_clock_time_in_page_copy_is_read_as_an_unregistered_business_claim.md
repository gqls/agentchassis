# 364 — a CLOCK TIME in page copy is read as an unregistered business claim, and it blocks the whole page build

**Filed** 2026-08-22 by the `site_ai_agent_orchestration` lane, found while unblocking
`ai-agent-orchestration.com`'s `pricing` rebuild (see `bugs_open/` note below — this is the
*second* claims-layer refusal that page hit, and it is unrelated to the first).
**Severity** medium. Nothing false is published; the cost is that a page asserting **nothing
about the business** is refused, and the refusal reads like a real honesty finding.
**Class** false positive in the claims layer / allow-list exclusion.
**Status** ⚠ **CORRECTED 2026-08-24 — the clock-time fix is LIVE, and the line below was stale
for two days.** Original text, left as written: *"FIX WRITTEN AND TESTED IN THIS COMMIT — inert
until the next fleet roll (Go)."* It shipped. Proven two independent ways, neither of which has a
shelf life:

- the running chassis binary (v1.0.1334) carries the post-fix regex form `fps|am\b|pm\b` (1 hit),
  does **not** carry the pre-fix form `fps|st\b` (0 hits), and the positive control
  `px|rem|em|vh|vw` passes (1) — the absent negative is what makes this a real probe rather than a
  promiscuous grep;
- `SELECT git_commit FROM service_binary_capabilities WHERE service='agent-chassis'` gives
  `48f55f21`, and `git merge-base --is-ancestor ebe8f4323 48f55f21` returns true.

**Why this mattered enough to write out:** an "inert until the roll" line makes the correct next
action look premature. A detector elsewhere on this estate sat switched off for nine days for
exactly this reason. Do not leave a status like this to be discovered.

**Status of the RESIDUAL (§5), which is now the live half of this bug:** interim fix committed
`a9002793b` (page-type gate extended to the three tracker/directory types; council submission
`b8df25dc-7d19-48b9-9b52-b93b25523d4a`, **APPROVED round 1**), plus the plural fast-follow
`0f9f7f3ff`. ~~Inert until the next chassis roll~~ — **BOTH VERIFIED LIVE 2026-08-25 on v1.0.1337,
see §6f.** Phase 2 (component-grain `ClaimSurface`) is filed as `RFC_053`, not built.

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

> **CORRECTED AND WIDENED 2026-08-24** (`bugs_open/364` lane, resuming this bug). The paragraph
> below framed the residual as *units the allow-list has not met*. That framing is too narrow and
> it sent the next reader looking in the wrong place. Measured, the residual is that **the scan
> cannot tell WHOSE number it is** — the unit suffix is one symptom of that, and not the biggest.
> Original text kept verbatim:
>
> - *"The allow-list is still an allow-list. The structural fix — deciding by what the number
>   measures rather than by suffix — is not attempted here and is the standing residual from 073.
>   Candidates not yet met: `%` in non-business prose, `k`/`m` magnitude suffixes, ISO durations."*

### 5a. What the residual actually is `[MEASURED 2026-08-24]`

`cmd/claimscan` over live `rendered_html`, each of the 19 opted-in sites against its **own current**
register, export **asserted row-for-row against the DB** (115/115 on ai-agent-orchestration.com —
the export truncates silently, see WRONG_CALLS): **44 findings fleet-wide.** Read in ≥300 chars of
context, not from the 100-char snippet (CLM-019's landmine):

| class | n | example | verdict |
|---|---|---|---|
| third-party figures in aggregated listings | **20** | `rollout_scope Over 80% of Fortune 500 deploying active agents` · `200,000 onboarded users` (someone else's) · `JSON-RPC 2.0` (a version) · the `2` inside **A2A** | false, **0% precision** |
| genuine first-person claims on content pages | 16 | "170+ Agents Running in Production" · the Kafka case study's 40-agent figures | **true — the layer working** |
| stale counter facts | 5 | page renders `11513`, register now holds `11646` | true-but-drift → `bugs_open/386` |
| formula / legal threshold / hypothetical | 3 | `throughput = 3600 / cycle time` · "anyone under the age of 16" (privacy policy) · "assuming 100% uptime" | false |

⚠ **This census can only see copy that PASSED the gate** — a refusal writes nothing, so the refused
copy was never stored. It systematically **under-counts**. The damage is better read from
`agent_error_log`: **70 refused build attempts across 23 distinct pages on 8 sites in 60 days**,
still occurring daily; `model-directory` alone refused 20 times since 2026-07-29. The refusals are
*intermittent* (the writer regenerates copy each attempt), so **a later success is not evidence the
bug is fixed** — check the build stamp.

### 5a-note. No `090` run — the substitution, stated plainly (owner ruling 2026-07-31)

§5b below asserts a **structural** root cause, so CLAUDE.md requires either a diagnosis-loop run or
a stated reason for substituting equivalent first-hand verification. **No `090` was run.** What was
done instead, and why it is equivalent for this specific claim:

- the claim is about a **regex and a control-flow gate in one package**, not about a cause living
  somewhere unexpected — the failure mode `090` exists to catch. The scan, its gate and its
  exclusions were read end to end (`claims.go:650-953`);
- the mechanism was **executed**, not reasoned about: `cmd/claimscan` runs the *same shared scan
  engine* as the deploy gate, over live `rendered_html`, per site against its own register. 44
  findings were read individually in ≥300 chars of context;
- both directions were **mutation-proven** — every new fixture fails with the fix reverted and
  passes with it restored;
- consumers were **enumerated** rather than assumed (`editorialPageTypes` → one reader → one caller);
- and it was **independently reviewed**: 13 council seats, round `b8df25dc`, of which two found real
  defects that changed the code (§6d).

⚠ **Where that substitution did NOT hold, it is marked as such.** `bugs_open/387`'s root cause is
explicitly **not established** — that file names three candidate causes and the cheap discriminator
instead of picking one. And one structural claim in this lane *was* wrong and was caught by a test
rather than by reasoning (the `gte`-vouching explanation — `WRONG_CALLS.md` 2026-08-24 #2), which is
the honest argument for running `090` next time the claim is about a cause I cannot execute.

### 5b. The mechanism, restated

`businessClaimContextRe` is a **lexical** gate: it asks whether business-ish words sit near a number.
On a site *about* agents, `agents`, `uptime`, `verified`, `collected` and `items` are in every
sentence, so it cannot separate a claim about **us** from a statistic about **Microsoft**. Each fix
so far has bolted another exclusion underneath it — 073 → 102 → 364 are one defect three times.

**And the lexical gate is itself an allow-list, of NOUNS, with the same unbounded-miss property.**
Found while writing the test for the interim: the gate carries `orchestration` **singular, no `s?`**,
while the live copy says "orchestration**s**" — so *"We run over 1,600 orchestrations a day across 13
live production systems"*, a first-person quantified claim on a live CTA, **has never been scanned at
all**. A plural, synonym or hyphenation the gate has not met is a silent miss anywhere on the fleet.
Not fixed here; widening it is a precision change that needs its own measurement in both directions.

### 5c. Two designs measured and REJECTED — recorded so they are not re-attempted

- **Slot-name rule** (`*-listing` ⇒ third-party). **Refuted:** `case-studies-list` on the same site
  is a list of *our own* work ("orchestrates 30+ specialised agents", "under 4 hours"). The suffix
  does not discriminate, and keying on it would have blinded three real claims.
- **Attribution-cue subtraction** reusing `namedSourceRe`. **Refuted on evidence already in the
  tree:** `claims_attributed.go:49-57` records a near-identical arm that looked like nine true
  positives and evaporated when read in full paragraphs; and `namedSourceRe` is deliberately tuned
  generous for *exoneration*, so reusing it as a *classifier* inherits a one-way false-positive
  profile.
- **A marker in the rendered HTML** (`data-claims-scope="third-party"`). Rejected without measuring:
  the HTML is LLM-generated, so the thing being policed could emit its own exemption. The
  declaration has to come from trusted DB-side structure.

### 5d. What was done about it, and what was not

**Interim, committed `a9002793b`** (council `b8df25dc-7d19-48b9-9b52-b93b25523d4a`, pending): the
three tracker/directory page types added to `editorialPageTypes`. Result: **36 → 16** findings on
ai-agent-orchestration.com, zero survivors on any tracker page, all 16 genuine claims retained.
Mutation-proven both ways.

**It knowingly fails** the second half of that map's own membership bar ("a body that is never
marketing") — the ground `section-index` was refused on. Those pages carry a marketing `hero` and
`call-to-action`, and page-grain gating blinds them too. Measured loss today is zero **only because
of the plural blindness in 5b**, which is a second defect, not a safety property.
`TestTrackerPagesGiveUpTheirFirstPersonClaims` pins the cost; `LANDMINES.md` carries the blind spot.

**Phase 2, filed not built:** widen `ClaimSurface` from page-grain to **component-grain** so `hero`
and `call-to-action` stay scanned while the listing does not — the mechanism `claims.go`'s own
`report` note asks for when it says excluding a page type *"would fix those by coincidence, not by
mechanism"*. Traced: 4 of the 5 `ClaimSurface` construction points can pass component identity in
**one line**; only `validate_page_content.go` — the gate that actually refuses — needs real work, and
`sections_metadata` in `CollectedData` is the reachable signal there. Likely **not** architecture-scope
under RFC_022's 2026-08-11 narrowing (opt-in, unsafe side default-OFF, no live consumer names it) —
but that must be asserted **with the consumer enumeration**, not without it.
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

## 6b. Council verdict — APPROVED (read 2026-08-22)

`Council-Reviewed: 39d04868-6ce3-472f-a976-49cd387a7860` — *"COUNCIL GATE — APPROVED — all
reviewers approve (round 1)"*.

⚠ **Coverage gap, named rather than hidden.** The fix was committed BEFORE the submission was
fired, so that commit carries neither `Council-Reviewed:` nor `Council-Submitted:` and will list as
un-reviewed in the `098` report. Forward-only forbids an amend. The trailer is on the commit that
adds this section instead, which records the verdict but does not retroactively credit the commit
that holds the code. The lesson is the ordering, not the trailer: **submit before or alongside the
commit**, which is exactly what CLAUDE.md's `Council-Submitted:` trailer exists for.

## 6d. Council verdict on the interim — APPROVED round 1, and what each objection cost

`Council-Reviewed: b8df25dc-7d19-48b9-9b52-b93b25523d4a` — *"APPROVED — approved with 3 advisory
objection(s) — none high-severity (round 1)"*. 13 seats ran, 5 abstained. **Two of the objections
found real things and changed the code**; recording all of them, including the one that was wrong,
because which seats were wrong is as useful as which were right.

| seat | objection | what it cost |
|---|---|---|
| `bug_historian` (med) · `compliance` | the interim's "measured loss is zero" rests on a **second, unfixed bug** (the singular `orchestration`) — that is luck, not a safety property, and a later plural fix would make the blindness real with no signal. Asked for the one-line companion fix now. | **ACCEPTED AND FIXED** — `0f9f7f3ff`. Measured both directions first: fleet-wide only ONE site's components contain "orchestrations" (11, all aiao) and the change adds **zero** findings. Mutation-proven, and the singular direction pinned too. |
| `architecture` (low) | third occurrence of the same interim; **file the Phase 2 RFC in `architecture_review/`, not as a commit-message promise.** | **ACCEPTED** — `RFC_053` filed. |
| `guardian` (med) | `editorialPageTypes` also gates `ScanBannedClaims`, so this silently widens the banned-claim carve-out — unmeasured. | **REFUTED by enumeration, and the enumeration is the point.** `editorialPageTypes` has exactly ONE reader (`ProseNumbersAreClaims`) with exactly ONE caller (the guard at `claims.go:868`). `ScanBannedClaims`, `ScanAllBannedClaims(WithSuppressed)` and `ScanStatClaims` take no `ClaimSurface` and cannot consult it; `TestBannedClaimsAreStillCaughtOnEditorialPages` pins that. I checked all consumers rather than the one I remembered — an objection naming one file names a category. |
| `guardian` (low) | other consumers (`check_unverified_claims.go`, `revalidate_unverified_claims.go`) not named. | Partly fair. They call `ScanUnregisteredNumbers` **with** a surface, so they are affected — in the intended direction: they stop raising the same 20 false positives. That is the fix working at a second seam, not an unmeasured side effect. |
| `bug_historian` (low) | nothing forces Phase 2 to land; "interim" quietly becomes permanent. | Partly answered: `TestTrackerPagesGiveUpTheirFirstPersonClaims` must be INVERTED not deleted, and `RFC_053` now exists. **Still unanswered: there is no expiry that fails loudly.** Recorded as an open residual rather than claimed as solved. |
| `editquality` (low ×2) | the sketch showed one test, not the six fixtures; and the rationale named doc files absent from the edits array. | Bookkeeping, and correct. The commit did contain them (`a9002793b`). The lesson is the runbook's: reviewers judge the **sketch**, so a sketch that under-shows the change draws objections about code that is actually fine. |
| `debug_historian` (low) | no post-deploy verification named, unlike the artefact-level proof given for the am/pm fix. | **ACCEPTED** — the recipe is now §6e below. |

## 6e. How to verify the interim AFTER the next chassis roll

Do **not** read a later successful build as evidence — the writer regenerates copy each attempt, so
a page can pass for reasons unrelated to this change. Probe the artefact:

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
# under test: the three new page types must be IN the binary
kubectl -n ai-persona-system exec "$POD" -- grep -a -c -F 'adoption-tracker' /proc/1/exe
# positive control: a page type that was always there (must be >=1 either way)
kubectl -n ai-persona-system exec "$POD" -- grep -a -c -F 'news-index' /proc/1/exe
# the plural fast-follow, with its own negative control (pre-fix form must be 0)
kubectl -n ai-persona-system exec "$POD" -- grep -a -c -F 'orchestrations?|integrations?' /proc/1/exe
kubectl -n ai-persona-system exec "$POD" -- grep -a -c -F 'orchestration|integrations?' /proc/1/exe
```
**Run the controls in the same breath.** A bare grep for a short literal returns a nonzero count on
binaries that do not support it, which is why the pre-fix form is checked as a must-be-ZERO.
The no-shelf-life alternative, and the better one:
`SELECT git_commit FROM service_binary_capabilities WHERE service='agent-chassis'` then
`git merge-base --is-ancestor a9002793b <that sha>`.

Then, at the data: re-fire a `model-directory` build and confirm no `unregistered_number`
in `agent_error_log` (§6's query). ⚠ **CORRECTED 2026-08-25: the pages DO serve** (at `.html`; see §6c) — so the effect of this
fix *can* be read off the live site, contrary to what this line originally said.

## 6f. VERIFIED LIVE 2026-08-25 — both fixes shipped on v1.0.1337

Chassis rolled 2026-08-25 09:27:48Z, image `v1.0.1337`, build commit `635f2d32`. Both fixes are in
it and both were proven at the artefact, with controls in both directions:

| probe | result | reading |
|---|---|---|
| `adoption-tracker` in `/proc/1/exe` | 4 | interim present |
| `orchestrations?\|integrations?` | 1 | plural fast-follow present |
| `orchestration\|integrations?` (pre-fix form) | **0** | **negative control passes** — the probe discriminates |
| `news-index` (always present) | 11 | positive control passes |

`git merge-base --is-ancestor` puts both `a9002793b` and `0f9f7f3ff` inside `635f2d32`. And
`git diff 635f2d32 HEAD -- claims.go` is **empty**, so a locally built `cmd/claimscan` runs the
fleet's exact scan code — which is what licenses the figure below.

**At the data, post-roll, export asserted 118/118:** ai-agent-orchestration.com is at **16 findings,
ZERO on any tracker/directory page** — down from 36. The 16 that remain are the genuine first-person
claims the layer exists to catch.

⚠ **`agent_error_log` shows 0 refusals since the roll, and that number is NOT yet evidence.** The
demand control says why: 6 pages have been built since the roll and **all six are `guide` / `tool` /
`blog-post`** — page types that were already excluded before this change. **No page of an affected
type has been rebuilt yet**, so the zero measures absence of demand, not the fix working. The
end-to-end confirmation is still outstanding; §6e has the recipe.

> **A near-miss worth recording, 2026-08-25.** The `bugs_open/387` lane offered a clean
> `model-directory` rebuild (orchestration `3384ae13`, COMPLETED, page stamped deployed 06:31:06Z,
> **0** `agent_error_log` rows) as the predicted clean pass, observed. **It is not.** That build ran
> at **06:26Z** and the roll carrying this fix landed at **09:27:24Z** — the rebuild PREDATES the
> binary by three hours, so it ran on the old code and is evidence only that the page can pass
> anyway, which is the non-determinism §5 already warns about. The same arithmetic answers their
> mirror question: the two tracker failures of 08-24 18:41Z / 18:47Z also predate the fix and say
> nothing against it. **A favourable result from the wrong binary is still the wrong binary** — check
> the roll time before accepting a confirmation, including one a peer hands you.

## 6g. The fix does NOT dispose of the two false findings already in the human queue — by design

Two `claims_unverified` items have sat in `needs_human_review` since **2026-08-09** and are pure
false positives of this bug:

| item id | page | findings |
|---|---|---|
| `4405fb38-0201-463a-bd2a-40698bed9db7` | adoption-tracker | 8 |
| `2f8f67dd-07b1-4907-a6a1-d7b2bd86fcf4` | protocol-tracker | 3 |

**They will not self-close, and I was about to predict that they would.** `revalidate_unverified_claims.go`'s
ladder re-scans first (`armScanStillTrips`, which now returns nothing on these pages), then hits the
**claim-granular gate** — and that gate compares the *cited token* against the slot's current text.
The tokens are still there, because the copy never changed; only the standard did. So the verdict is
`armGateClaimsStillPresent`: *"page X no longer trips the check, but N of the M texts this finding
cited are STILL in the component they were cited from — so the standard moved, not the copy; a claim
that stopped being flagged while its words are untouched has not been addressed."*

**That is correct behaviour and must not be "fixed".** It is a deliberate integrity control (council
round 5, `compliance` HIGH, 2026-08-11): a detector change is not allowed to silently dispose of
items on the platform's highest-stakes, deliberately HITL-terminal type. The same guard is what stops
a CSS tweak closing a real fabrication.

**So this is a human decision, not an engineering one.** Someone has to rule that these two items were
always false and cancel them. Note the sibling residual: `ddc90e58` (about, 6 findings) and
`962da5c9` (case-studies, 1) are on content pages and are **genuine** — they stay open on merit.

⚠ Whoever closes them: record the item ids first. Closing ARCHIVES the row out of `site_work_items`,
so a later census cannot see what it succeeded at.

## 6h. OWNER RULING 2026-08-25 — the two queue items were ALWAYS FALSE, and are cancelled

§6g said these could only be disposed of by a human. The owner ruled they were false and they are
now `cancelled` (both rows, verified; the four genuine siblings on content pages were left open and
verified untouched as the control).

- `4405fb38-0201-463a-bd2a-40698bed9db7` — adoption-tracker, 8 findings
- `2f8f67dd-07b1-4907-a6a1-d7b2bd86fcf4` — protocol-tracker, 3 findings

The `resolution_path` on each carries the full reasoning, so the ruling travels with the row rather
than living only here. **The integrity control was NOT bypassed** — `armGateClaimsStillPresent` did
its job, refused to auto-close, and escalated to a human, which is exactly the designed path.

**Pre-cancel snapshot, preserved because closing ARCHIVES the row out of `site_work_items` and a
later census cannot see what it succeeded at.** Every token is a third party's figure:

| page | matched | slot | snippet |
|---|---|---|---|
| `adoption-tracker` | 70 | adoption-tracker-listing | agent_framework 70% of regulated enterprises rebuild their AI agent stack ever |
| `adoption-tracker` | 1,837 | adoption-tracker-listing | rollout_scope Out of 1,837 surveyed engineering and AI leaders, only 95 reported havin |
| `adoption-tracker` | 30 | adoption-tracker-listing | laimed 50% of service desk inquiries resolved autonomously, 30% reduction in agent workload perc |
| `adoption-tracker` | 3 | adoption-tracker-listing | uced the Fuse EDA AI Agent in March 2026 for semiconductor, 3D IC, and PCB design workflows. |
| `adoption-tracker` | 1.65 | adoption-tracker-listing | ntainer Terminal (PNCT) is a US container terminal handling 1.65 million TEU per year, deploying |
| `adoption-tracker` | 125 | adoption-tracker-listing | rollout_scope 125+ live use cases, 20,000 employees actively building agents |
| `adoption-tracker` | 200,000 | adoption-tracker-listing | rollout_scope 200,000 onboarded users within eight months users source |
| `adoption-tracker` | 700 | adoption-tracker-listing | roi_claimed Equivalent work of 700 full-time agents FTE-equivalent source |
| `protocol-tracker` | 2.0 | protocol-tracker-listing | agent_framework JSON-RPC 2.0 client-server with Tools, Resources, Prompts, and Sampling |
| `protocol-tracker` | 2 | protocol-tracker-listing | Agent-to-Agent Protocol (A2A) Google |
| `protocol-tracker` | 3 | protocol-tracker-listing | protocol_governance W3C AI Agent Protocol Community Group working on standardisati |

Note `2.0` (a JSON-RPC version string) and `2` (the digit inside the acronym **A2A**) — two of the
eleven are not statistics at all, which is the sharpest evidence that the lexical gate was never
answering the question it was asked.

## 6i. Phase 2 shipped 2026-08-25 — council APPROVED r1, and three objections changed the code

Commit `52958897f` (the mechanism) + `fa0b513f1` (the round-1 rework).
`Council-Reviewed: 3ed2b792-dbe6-4301-83ab-5df22f188c1e` — *"approved with 5 advisory objection(s) —
none high-severity"*. **Approved is not the same as finished**: three of the five were right and are
fixed in the code rather than noted.

| seat | objection | what it cost |
|---|---|---|
| `bug_historian` (med) | `collectPageSections` collapsed "section is empty" with "section HAS html under a key I do not know" into one silent `continue` — dropping a component from the scan while the page still read as scanned. | **FIXED.** The guard is now a COUNT (fewer sections back than the metadata held ⇒ refuse component grain for the whole page and fall back). Mutation-proven on three shapes. **This was a real silent coverage hole on the only gate that refuses a page**, and my `len(out)==0` check could not see it because it only fired when *every* section failed. |
| `reuse_agent`, `prior_art`, `constitution` (3 seats, independently) | I hand-parsed `sections_metadata` when `extractSectionsFromMetadata` is already the canonical reader. | **FIXED.** Rewritten to reuse it — which also inherits its slot-before-component identity resolution (`bugs_open/189`'s fix) and deletes my near-duplicate of an existing function name. One reader to update when the shape drifts; a second one is what `bugs_open/357` cost a day of production for. |
| `guardian` (med) | The block-equivalence safety case rested on **4 pages / 21 components** — thin for a fleet-wide refusal gate. | **RE-MEASURED over the whole live corpus: 775 pages / 2,042 components, export asserted 2042/2042, ZERO pages where whole-page and per-component extraction differ.** |
| `editquality` (med) | The map keys were *asserted* to be live values, not shown. | **VERIFIED, and it surfaced a real correction** — see below. The check is now in the comment. |
| `prior_art` (med) | The consumer enumeration was asserted, not shown checked. | The grep that produces it is now in the comment; re-run 2026-08-25 → 5 literals, all updated. |

### ⚠ WHAT THESE TWO APPROVALS DO NOT MEAN (added 2026-08-25 — read before citing either)

Both rounds are `APPROVED`. **Neither validated a single measured number in this file**, and the
verdicts do not say so, which is why it is written here instead.

**A seat can read a row in twelve tables and can RUN nothing.** Allowlist measured live 2026-08-25
from `council-gate`'s `load_schema_hint`: `agent_definitions, agent_error_log, content_components,
diagnosis_artifacts, doc_notes, landmine, page_components, pages, site_plan_pages, site_plans,
site_work_items, sites`. **`site_specs` and `orchestration_states` are absent** — the evidence
register and everything in `collected_data`.

So, for this lane's evidence specifically:

| claim | seat could verify? | why |
|---|---|---|
| 44 findings / 36→16 / 775-page block equivalence | **no** | requires running `cmd/claimscan` and Go; no seat can execute anything |
| `sections_metadata` shape and its single supplier | **no** | `orchestration_states.collected_data` not in the allowlist — `editquality` said exactly this |
| the register/`gte` facts, the 1,494–7,281 history | **no** | `site_specs` not in the allowlist |
| `slot_name` vs `content_components.function` (106 of 2,033) | **no — see the correction below** | both tables are in the schema hint, but that is TEXT in the prompt, not a query the seat can run |

> **CORRECTED 2026-08-25 by the `bugs_open/386` lane — the table's last row was wrong and so was my
> framing of it.** I wrote that a seat "reads rows but runs nothing", so a visible table made that
> claim checkable in principle. **A review seat executes nothing at all — not even read-only SQL.**
> Measured live: every `review_*` step's config has exactly seven keys (`ai_service, error_step,
> input_fields, output_format, prompt_template, temperature, tolerate_truncation`) and **no
> sql/query/tool/db key on any seat**; the only `query_database` steps are `load_schema_hint`, which
> injects a schema *description as text* before the reviews, and the two `compose_verdict` steps
> after them. **The allowlist is what the seats are SHOWN, not what they can read** — which is why
> `editquality` says *"SQL checks against the given schema can't reach…"*: it is reasoning about what
> a query would show, never reporting one it ran.
>
> **So every figure in this file was unverifiable by every seat**, allowlisted table or not.

**The un-shown/unseeable distinction still holds, with its first half weakened.** It correctly
separates *"I could have made this auditable and didn't"* from *"no submission could"* — but
including the query makes the **inference** auditable (a reader can judge whether that query would
answer that question), **not the number verified**. The `386` lane's formulation, which is better
than mine:

> **A submission can make its reasoning checkable. Nothing a submission can do makes its evidence
> checked. Only someone re-running it does that.**

I answered those objections by re-running the queries and putting the commands in the code comments.
That makes the figures re-derivable **by a human**, which is the only route there is.

**When citing these approvals, say what they approved: the plan and the reasoning.** They did not
re-derive the numbers. `bugs_open/386` (approved `18dba069`) hit the harder version of this — three
seats flagged that `site_specs` was invisible to them, and its `prior_art_librarian` asked outright
that a human re-run the queries before arming. Landmined 2026-08-25.

### ⚠ The correction that came out of the round: `slot_name` is NOT always `content_components.function`

I wrote in the interim's notes that "slot_name IS the component's registered function on this
platform", citing a code comment. **Measured 2026-08-25: 106 of 2,033 live rows differ** — `prose-0`
vs `ported-prose`, `call_to_action` vs `call-to-action`, `FAQ Section` vs `faq`. The surface is keyed
on the **slot**, which is correct because that is what every call site actually holds, and a mismatch
**fails safe** (the component is scanned). But it is also a **silent no-op**, so a new member of
`thirdPartyDataComponents` must be checked against live `slot_name` values — the query is in the
comment beside the map.

**Outcome unchanged by the rework:** 16 findings, zero on any tracker page.

## 6j. ⚠ A LIVE refusal risk this lane's own "no measured cost" claim was hiding

Found 2026-08-25 by re-checking one of my own claims after the `bugs_open/386` lane published a
falling-facts census. **It is live, it is not caused by Phase 2, and Phase 2 stops hiding part of it.**

`aao-orchestrations` — the fact I cited as making the plural fix free — **is a rolling window, not a
total**: 35 recorded snapshots ranging **1,494–7,281**, fallen **17 times across 34 transitions**, and
**below 1,600 on 3 of the 35**. The site publishes *"We run over 1,600 orchestrations a day"*.

Replaying the site's own historic low through `cmd/claimscan` at HEAD — same corpus, same code, only
the fact's value swapped:

| register value | findings on ai-agent-orchestration.com |
|---|---|
| 7,281 (today) | 16 |
| 1,494 (its own historic low) | **20** |

The four that appear only at the low, all at `error` severity, which **refuses the page build**:

- `adoption-tracker / hero` — `"1,600"`
- `protocol-tracker / call-to-action` — `"1,600"`
- `protocol-tracker / hero` — `"1,699"`
- `enterprise-reference-deployment / case-studies-grid` — `"1,834"` ← **exposed on the CURRENT binary**

**Read the last row carefully: this is not a Phase 2 regression.** One of the four is already live on
a content page with no change from this lane. Phase 2 makes the other three visible by scanning the
tracker pages' own surfaces again — which is the fix working, not breaking: those claims were always
unsupported on a low day and were merely un-scanned.

**The claim itself is never false.** *"Over 1,600"* is a floor, and the copy does not change. Only the
evidence for it moves.

> ### ⚠ CORRECTED 2026-08-25 — I named the wrong fix, and the right one needs no code
>
> This section originally said the remedy was `bugs_open/386`'s work, because
> "exact-match-against-retained-former-values keeps vouching for 1,600, since 7,281 IS a value the
> register held". **That is `gte` applied to history, and it is not what 386 built.** Their
> `historySupports` is **EXACT-match only**: with the fact at 1,494 and the page saying 1,600, the
> `gte` arm fails (1600 ≤ 1494 is false) and the history arm asks whether 1,600 exactly equals a
> retained reading — it does not. **The page stays convicted.** Their lane corrected me and is right.
>
> **It should not be built the way I described, either.** "Supported if the value is at or below any
> reading the register ever held" means support is pinned at the **all-time maximum, for ever**: one
> busy day would vouch for *"over 7,000 orchestrations a day"* on every quiet day thereafter. That is
> this bug's own §2 accidental-support gradient with **time** as the amplifier — strictly worse than
> today's behaviour, which at least tracks the current value. Exact-only is the property that makes
> history safe: it vouches for numbers the register actually HELD and can never invent one between.
>
> **The actual remedy is already sitting in the register, unused.** `aao-orchestrations`' own
> `writer_line` reads *"over a thousand orchestrations a day ({value} in the 24 hours to
> 2026-07-26)"*. **A floor of 1,000 sits BELOW the historic low of 1,494 — so that instruction is
> safe on every day in the record.** The convicted copy says 1,600 / 1,699 / 1,834, all **above** the
> low, so all three deviate from the register's own instruction or predate it.
>
> **So this is not a `386` case at all.** It is the owner's "state a floor, never the exact number"
> ruling applied with **the floor set too high**, and the fix is to bring the copy back to the floor
> the register already specifies. The rule that generalises — and this replay is the best evidence
> for it — is **the floor must sit below `lowest`, not below today's value** (`bugs_open/386`
> RUNBOOK §7). What 386's fix DOES cover is the dated-chart class (§6c, F9–F13: a page stating a
> value that was exactly right on the day it rendered); it does not cover falling lower-bound claims,
> and their council submission says so in its limitations.

**Not fixed here, deliberately.** Narrowing that fact's single broad `context_terms ["orchestration"]`
would make the gate stricter and could newly refuse live pages on a customer site — not a unilateral
change, and the site is `bugs_open/387`'s. Flagged to that lane before they rebuild (a dip would make
their rebuild fail for a reason unrelated to their fix, and look like mine failing), to the `386` lane
as a worked case for the durable fix, and to the owner. `LANDMINES.md` carries it, armed, with the
history query.

**The lesson for this file: a `[MEASURED]` figure about a MOVING value expires, and I wrote one into
shipped code.** "At no measured cost" was true of the day it was measured and of no other. The
correction is in the `claims.go` comment where the claim was made.

## 6k. ⚠ Commit `6548e8d79` carries `bugs_open/386`'s mechanism, and its message does not say so

Recorded because a bisect or a review of that commit will otherwise be misled, and forward-only
means it cannot be amended.

**What happened.** `6548e8d79`'s message is about a rolling-window evidence fact and a landmine. It
also contains the whole of the 386 lane's fact-history mechanism — `RetainHistory`,
`History []FactHistoryEntry`, the `FactHistoryEntry` struct, `FactHistoryMaxEntries`, and the
`historySupports` arm in `numberSupported`. **98 lines went into `claims.go` where I believed I was
adding ~20 lines of comment.** Their tree was dirty in that file while they built and mutation-tested;
my pathspec named `claims.go`, and a pathspec commit takes the named file from the working tree
entire. That is the same-file passenger CLAUDE.md documents and no hook can prevent it.

### ⚠ TWO lanes named `001211abf` as the sweeper, and it swept neither of them

Added 2026-08-25 after checking the second claim as well as the first. **The same wrong sha, from two
independent lanes, on one day** — and the root cause is one command:

| lane's claim | what the pickaxe says | what `001211abf` actually did |
|---|---|---|
| `bugs_open/386`: its fact-history mechanism was swept into `001211abf` | `git log -S 'FactHistoryEntry' -- claims.go` → **`6548e8d79`** (mine) | added **only comment lines** to `claims.go` — 25/6, zero non-comment additions |
| `site_ai_agent_orchestration`: its `WRONG_CALLS` entry was swept into `001211abf` | pickaxe on that entry's heading → **`3d31b86a9`**, the `bugs_open/381` lane's commit | its `WRONG_CALLS` diff holds only my own entries 12 and 13 |

**The cause, which the 386 lane diagnosed for itself and I then confirmed on the second case:**
`git log -N -- <path>` answers *"what last touched this file"*, not *"what introduced this code"*. On
a shared tree those are different commits, and a path-filtered log always returns something recent
and plausible — there is no tell. `-S` has its own trap on top: it matches a commit that merely
*mentions* the symbol in prose, which is exactly how `001211abf` surfaced in a `-S historySupports`
search while containing none of the code. The discriminator is two seconds:
`git show <sha> -- <path> | grep '^+' | grep -v '^+\\s*//'` is empty for a comment-only commit.

**Why it matters more than the attribution**: both shas were written into lane NOTES *specifically so
a `098` coverage flag could be hand-traced*, which is the one use where a wrong sha costs a real
search. Landmined 2026-08-25; the `386` lane has corrected its NOTES, and the site lane's entry is
still uncorrected because that session is not live — **if you are reading this from that lane, your
sweeper was `3d31b86a9`, not `001211abf`.**

⚠ **The 386 lane reported this against the WRONG COMMIT — `001211abf` — and I have told them.**
Verified both ways: `001211abf` added **only comment lines** to `claims.go`
(`git show 001211abf -- <file> | grep '^+' | grep -v '^+\s*//'` is empty), while
`git log -S 'FactHistoryEntry' -- <file>` names `6548e8d79`. **Their NOTES and CLM-028 carry the
wrong sha for hand-tracing**, which matters more than the attribution does.

**Nothing is lost.** Verified independently rather than assumed: each symbol appears exactly once in
HEAD, `historySupports` is defined at `claims.go:1299`, and both packages pass fresh with `-count=1`.

**Consequence for the council join, which is the part worth acting on.** The 386 lane's
`Council-Submitted: 18dba069-c142-4828-8e59-03453d04f72b` sits on their commit `63d95be1f`, which now
carries their tests but not the in-scope platform file. **The platform file landed in `6548e8d79`,
which carries no trailer.** So `098` will list `6548e8d79` as an in-scope commit with no review, while
the review that actually covers that code is credited to a commit holding no in-scope platform file.
Nothing dishonest has been written — the trailer asserts a submission, and the submission is real and
does cover the code — but **if `098` flags `6548e8d79`, the answer is correlation
`18dba069-c142-4828-8e59-03453d04f72b` for the fact-history half and
`3ed2b792-dbe6-4301-83ab-5df22f188c1e` for this lane's Phase 2 half.**

**The check that would have caught it, which I had at session start and did not run:**
`git diff --numstat <file>` before committing — it prints `added deleted path`, and `98 2` against an
expectation of twenty is not a number anyone scrolls past. ⚠ **The commit-scope block does not catch
this case**: it lists files, not lines, so it finds a path that went *missing* (someone committed it
first) and not a path that arrived *fatter than you thought*. Two failures, two checks.
`WRONG_CALLS.md` 2026-08-25 #14.

## 6l. ✅ END-TO-END CONFIRMED 2026-08-25 — the interim works on a real build, measured on the copy that build produced

`adoption-tracker` rebuilt through the scheduled refresh (`needs_page f2673ca7`; chain `bebaf2df` →
`f13006dd` → `f3f9b531` → `9c2d016e`, all COMPLETED; deployed **18:31:31Z**), delivered by
`bugs_open/387`. Build start **18:28Z**, well after the **09:27:24Z** roll; pods still `v1.0.1337`,
so this tested **Phase 1 (the page-type gate)** — Phase 2 has not rolled.

**`agent_error_log` for the site since 18:20Z, any error code: 0.**

**The check that makes it discriminating** — and it had to be re-run, because the figure it rests on
was stale. `387` cited my earlier "17 ungated findings", which was measured on the copy that existed
**before** this rebuild. The build regenerated the content, and §6f's `model-directory` case is
exactly where a regenerated page came back with nothing to suppress. Re-measured on the copy this
build actually wrote:

- **gated** (as the live binary judges it) → **0**
- **ungated** (page_type rewritten to `content`) → **19**, all in `adoption-tracker-listing`:
  `1,837` respondents, `700` agents replaced, `200,000` onboarded users, Fortune `500`'s `80%`,
  Salesforce `360`, `57.3%`, `1.65` million TEU …

**19, not 17** — the copy did change and my number was stale, favourably. So: the page carried 19
figures that would have refused it before this fix, the gate suppressed all 19, and it deployed clean.

**This closes the claim that the interim works. It does NOT close the bug** — see the status block at
the top: the interim's own blind spot (the tracker pages' `hero` and `call-to-action` silenced along
with the listing) is live in production until Phase 2 rolls.

## 6m. ⚠ Phase 2 would have REFUSED protocol-tracker on the roll — caught in the inert window

**A regression this lane introduced, found and fixed before it shipped.** Commit `35f452a0e`,
council `Council-Submitted: 64d852b4-bd70-43a6-9c7a-82202ea5688f`.

Phase 2 restores the prose scan to a tracker page's `hero` and `call-to-action` — correct, and the
entire point of component grain. But protocol-tracker's **live** hero reads *"MCP, **A2A** and half a
dozen other proposals are competing to become how agents pass tasks"*, and the `2` inside that
acronym is then raised as `unregistered_number` at **error** severity, **which refuses the page
build**.

**Measured on the copy the page actually carries** — and that distinction is the whole reason this
was caught, because a rebuild regenerates content and §6l records a case where the regenerated copy
had nothing in it at all. protocol-tracker rebuilt 18:32Z, deployed 18:36:35Z; scanning what that
build wrote:

| binary | findings on protocol-tracker |
|---|---|
| Phase 1 (**live now**) | **0** — it gates the whole page, so the regression is invisible today |
| Phase 2 as committed | **1** — the A2A digit, at error severity ⇒ **build refused** |
| Phase 2 + this fix | **0** |

**The rule: a digit with a LETTER on BOTH sides is inside an identifier, not a quantity** (A2A, W3C,
B2B). Both-sidedness is what bounds it — a real quantity always has a word boundary in front of it,
and one-sided forms (`3D`, `MP3`) stay scanned so the rule cannot grow into a general alphanumeric
exemption.

**Blast radius measured before keeping it**, over **all 2,042 live components** with a register that
supports nothing so every number is reported: **347 → 346**. One finding removed fleet-wide, and it
is the right one — `EvilHack0r`, a username in an XSS tutorial. aiao's 16 genuine claims unchanged.
Mutation-proven on all three live shapes.

⚠ **One test fixture moved, and the failure that forced it is the interesting part.** The
component-grain table held *"Agent-to-Agent Protocol (A2A) Linux Foundation"* — whose **only** number
is the A2A digit. Once excluded upstream it can no longer probe component-grain leakage, and my own
negative control `TestSameTextInAnUndeclaredComponentIsStillScanned` **correctly failed** when it
tried. It moved to the acronym test with a comment saying why re-adding it would make the surface
test pass for the wrong reason. **A fixture whose number is excluded upstream is a test that has
quietly stopped testing.**

**The general point, which is why the inert window is worth having:** Phase 2 was council-approved
and correct, and still carried a live build-refusal for one page. An approval does not find this;
only running the new code against the copy the fleet actually holds does.

## 6c. Found by this bug's census, filed separately (2026-08-24)

Both were turned up by the fleet claims run for §5a and are **not** this bug's mechanism.
Named here so the next reader does not re-find them, and so they are not conflated with 364.

- **`bugs_open/386`** — refreshing a **counting fact** turns every page that already rendered the
  previous value into an "unregistered claim". fundamentallyai.com renders `11513/10194/428/483`
  while the register now holds `11646/10416/437/503`. Framework-wide, periodic, and it convicts
  honest pages at `error` severity.
- **`bugs_open/387`** — the three tracker pages report `build_status='deployed'` with timestamps
  from earlier today and **404 on the live site** (with an invented-URL control proving the
  domain discriminates). ⚠ **CORRECTED 2026-08-25 by the `bugs_open/387` lane — my filing was WRONG and so was
  the caveat I put here.** I wrote that the three tracker pages were "not currently served".
  **They are served**, at `/adoption-tracker.html` etc. (200, curled 2026-08-25 ~11:0xZ). The
  extensionless form 404s for **every** page on that site — `/about` 404s too — because the
  worker does not resolve slashless paths. So the 20 false positives WERE on live, public pages,
  and this bug's damage was larger than I described, not smaller.
  **Why my control did not catch it, which is the transferable part: my invented-URL control
  shared the defect with my claim.** Both were extensionless, so it proved the domain
  discriminates and could say nothing about the URL *form*. The missing control was a page I had
  **not** touched at the same form (`/about` → 404 answers it in one request). `WRONG_CALLS.md`
  2026-08-25.

## 7. Relations

- `bugs_closed/073` (same regex, previous unit), `bugs_closed/102` (page-type gate),
  `bugs_closed/043` (the original invented-claims work).
- The `pricing` page's *other* claims refusal — the evidence base mandating a wording its own
  facts could not validate — is fixed at source by migration
  `557_aiao_evidence_base_stops_mandating_a_phrase_its_own_facts_cannot_validate.sql`.
  **Different cause, same page, same day; do not conflate them.**
- Lane notes: `docs/agent_docs/docs024_key_docs_latest/site_ai_agent_orchestration/NOTES_site_improvement.md`.
- The 2026-08-24 residual work: interim commit `a9002793b`, council `b8df25dc-7d19-48b9-9b52-b93b25523d4a`,
  concept register **CLM-016** (extended, and its index status corrected from a month-stale "INERT until roll"),
  `LANDMINES.md` entry "The claims number scan is now SILENT on three whole page types",
  and the two spin-offs `bugs_open/386` / `bugs_open/387`.
