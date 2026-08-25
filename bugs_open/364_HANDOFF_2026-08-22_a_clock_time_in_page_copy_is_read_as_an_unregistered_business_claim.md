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
in `agent_error_log` (§6's query). ⚠ **And note `bugs_open/387`** — those three pages currently 404,
so "the page is fine now" cannot be read off the live site either.

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

## 6c. Found by this bug's census, filed separately (2026-08-24)

Both were turned up by the fleet claims run for §5a and are **not** this bug's mechanism.
Named here so the next reader does not re-find them, and so they are not conflated with 364.

- **`bugs_open/386`** — refreshing a **counting fact** turns every page that already rendered the
  previous value into an "unregistered claim". fundamentallyai.com renders `11513/10194/428/483`
  while the register now holds `11646/10416/437/503`. Framework-wide, periodic, and it convicts
  honest pages at `error` severity.
- **`bugs_open/387`** — the three tracker pages report `build_status='deployed'` with timestamps
  from earlier today and **404 on the live site** (with an invented-URL control proving the
  domain discriminates). ⚠ **This bears directly on how you read §5a**: the 20 false positives
  this bug's interim removes are on pages that are *not currently served*. The **refusals** in
  `agent_error_log` are real regardless — a refusal is recorded at build time — but do not
  describe the tracker findings as live public damage. I did, before I curled, and was wrong.

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
