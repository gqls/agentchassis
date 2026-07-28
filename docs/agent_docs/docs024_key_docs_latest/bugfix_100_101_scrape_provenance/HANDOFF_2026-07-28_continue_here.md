# HANDOFF 2026-07-28 (evening) — `bugs_open/100` + `bugs_closed/101`: continue here

Written because token load ended the session, not because the work stalled. **The
code is live and verified.** What is left is one council round and one owed run.

---

## 1. State in one paragraph

`101` is **CLOSED** (`/bugs_closed/`) — fixed and live on `agent-chassis:v1.0.1192`
and `web-scrape-adapter:v1.0.1192`, pod-grep verified on symbols the fix *created*.
`100` is **still OPEN by choice**: its fix is live and SQL 257 is applied and proven
to enforce, but the file's own two-column acceptance test has never run, because vet
collection has been off since 2026-03-18. Council round 1 returned **REVISE** with
four owed items, none of them a correctness objection to the shipped code.

## 2. What is live, and how it was proved

Verified against **running pods**, never the tag — the fleet index warns *a retag is
not a rebuild* (1188/1189 shared one image id), so every marker below is a symbol
this change **created**, with a pre-existing string as positive control.

```
chassis  agent-chassis-f757fcf65-bg9t7          (v1.0.1192)
  unrecognised_keys               1   (0 pre-deploy)
  "does not read"                 1   (0)
  add_protocol                    3   (0)
  "no fetch provenance available" 1   (0)
  POSITIVE CONTROL scrape_web     1   (1 — proves the grep itself works)

adapter  web-scrape-adapter-c576d96b-bzzz4      (v1.0.1192)
  buildScrapePayload              2   <- created by this fix; 0 before
  POSITIVE CONTROL onlyMainContent 1
```

**SQL 257 applied 18:30Z, after that pod-grep** (the ordering the bug called
load-bearing). Enforcement proven by negative control, not assumed:
`convalidated=f`; an insert with empty provenance **errors**; 2,970 historical rows
untouched.

**`bugs_closed/062` payload watch, post-roll:** 0 `Message Size Too Large`,
0 `Failed to produce`. Re-run it — three steps now receive full pages for the first
time and those steps may not have run yet:

```bash
kubectl -n ai-persona-system logs -l app=web-scrape-adapter --tail=-1 --since=3h \
  | grep -i "Message Size Too Large\|Failed to produce"
```

> **CORRECTED 2026-07-28 ~20:30 — this command said `deploy/web-scrape-adapter`,
> which reads ONE POD OF THREE.** 3 replicas, 1 consumer group, 1 partition ⇒ two
> pods are idle for life and `logs deploy/…` may pick one, giving a permanently
> clean log regardless of what the working pod does. `bugs_open/133`.
>
> **DONE 2026-07-28 ~19:35 — the watch has been exercised**, so §8 item 2 is closed.
> One probe scrape (corr `1e97bd22`): **1 attempt, 0 / 0.** The 062 risk did not
> materialise. But the reply was only deliverable because the adapter had truncated
> `raw_html` 53,805 → 50,000 and appended *"full version in S3"* with
> `upload_results:false` and **no upload performed** — a data loss that completes
> successfully and is therefore invisible to this watch. Filed as `bugs_open/133`
> (also: the single-URL reply path never sends a deliverable error, the gap
> `bugs_closed/062` left when it fixed only the batch sibling).

The 062 failure is **silent to the caller** (~12 min of timeout retries), so absence
of a workflow error proves nothing. Worst exposure: `site-scraper/scrape_site` (no
`formats` override). Mitigation is config-only, no roll: set
`scrape_config.formats` on the offending step.

## 3. NEXT — in priority order

### 3a. Council resubmission (round 2) — the main outstanding task

Round 1: **REVISE**, corr `f4cf0aab-5a08-4475-91ea-fa831cff323c`, 11 reviewers,
**7 approve / 4 object**, `decided_by = "gating objection from tooling_provenance"`.
`unreadable: 1` exists but is **not** the decider, so this is a real REVISE and not
the `bugs_open/119` harness artefact — that was checked first.

**Resubmit on the SAME correlation** so the trail accumulates:
```bash
RESUBMIT_CORR=f4cf0aab-5a08-4475-91ea-fa831cff323c \
  ./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <new.json>
```
The round-1 submission JSON is in this session's scratchpad; rebuild it from the
PLAN if lost. Four things to change:

1. **`doc_notes` against the subjects touched** *(the gating objection)*. The platform
   has `doc_plans`/`doc_notes` keyed by `subject_type`+`subject_key`, and an
   `append_doc_note` action already in the registry. I used a self-built trail instead
   and wrote none. Leave notes against `scrape_web` and `firecrawl_scrape` carrying
   the two non-obvious findings: the `add_protocol`/`add_protocol_if_missing` typo,
   and the `onlyMainContent` inversion. **Do this one first — it is cheap and it is
   what gated the round.**
2. **A `ConditionalKeys` state** *(see §4, residual 1)*. Code change.
3. **The absence claim into `grounded_in`, pinned to a fixed ref:**
   `git grep -n "GetActionInputSpec" 2ebabf2ca^ -- '*.go'` returns only its own
   definition and doc comment. Do **not** cite `HEAD~n` — the claim is now
   self-referentially false at HEAD because its only callers are mine.
4. **Name the three non-vet pipelines** the firecrawl change touches
   (`site-scraper`, `site-adoption-agent`, `website-capture-firecrawl`) and state the
   062 result above. `guardian` asked for owners named or notified.

**No `Council-Reviewed:` trailer is claimed on any commit** — it is earned by
APPROVED only. Re-read `decided_by` per round: a later approval can attach to a
different plan.

### 3b. Close `bugs_open/100` when the first verification runs

Blocked on `vetcomparison` restarting collection — **their P1 blocker is now lifted**,
which they do not yet know. The closing test is the bug's own, and both columns matter:

```sql
SELECT source_url, source_type, raw_data ? 'source_url' AS llm_claimed, collected_at
FROM business_intel.data_observations ORDER BY collected_at DESC LIMIT 5;
```
`source_url` non-empty **AND** `llm_claimed` still **false**. A populated column alone
proves nothing about *where it came from* — if `llm_claimed` is true the fix is the
rejected candidate 4 and must be reverted.

If provenance comes back empty, grep the chassis log for
`no fetch provenance available`: that warning exists so a shape mismatch is
distinguishable from a genuine absence, and it names the field it looked in. The
`data.url` shape is `[UNVERIFIED]` — traced through code, never observed, because no
run carrying `scraped_data` survives the retention clock.

### 3c. Drive the coverage number down

`./scripts/audit-config-keys.sh` — **208 actions / 726 (action,key) pairs** are still
undeclared. Each declaration makes that action's dead keys detectable. This is the
adoption ratchet the opt-in design depends on for its justification.

## 4. Residuals — read these before claiming anything is finished

1. **The audit reads clean while two live steps still misdescribe themselves.**
   Declaring `max_pages`/`follow_links` made them *recognised*, so
   `UNKNOWN KEYS: none` is printed although `vet-practice-verifier/scrape_website`
   and `domain-research-classifier/scrape_site` still advertise a three-page crawl a
   single-page scrape cannot perform. **The design has three states and I built two.**
   Fix: a `ConditionalKeys` notion (key → the condition under which it takes effect)
   reported in its own section. Until then that "none" must not be read as "no step
   misdescribes itself". `WRONG_CALLS.md` 2026-07-28.
2. **Switching those two agents to `action: "crawl"` was deliberately not done** — a
   behaviour change to two other owners' agents (one unowned) under a lane off since
   March. They warn now instead of silently fetching one page. Somebody's deliberate
   call, not a side effect of a bug fix.
3. **`domain-research-classifier` still has no owner** and carries two of the affected
   keys plus the `add_protocol` typo (now implemented, so it works — but nobody owns
   the agent).
4. **`scrape_web` is not `StrictConfig`** and should not be until both definitions are
   clean. Flipping it makes an unknown key a hard validation failure, which would break
   running agents to make a point about their config.

## 5. Landmines this session paid for

- **A retag is not a rebuild, and `IMAGE_TAG` sat at the already-deployed value.**
  Verify by pod-grep on a symbol the change **created**, with a positive control in the
  same command. `unrecognised_keys` 0→1 was discriminating; a typed const would have
  been vacuous.
- **SQL 257 must never be applied on a tag or merge signal.** Against a stale binary
  the CHECK refuses writes the running code cannot satisfy — a silent data defect
  becomes a hard outage of vet verification.
- **`gofmt -l` the files you are committing, not the package.** The build gate rejects
  un-gofmt'd code; package-level noise is what makes drift read as "not mine".
- **An absence claim needs a pinned ref.** "No callers" is true only when you looked,
  and only at the commit you looked at.
- **After adding any declaration/allow-list, re-run the detector on the case that
  motivated it.** Exempting is not fixing, and it prints as clean.

## 6. Where everything is

| what | where |
|---|---|
| the five standing docs | this directory (`PLAN` has dated Corrections; `NOTES §7` has the four missteps) |
| the closed bug | `bugs_closed/101_HANDOFF_2026-07-26_scrape_web_silently_ignores_four_config_keys.md` |
| the open bug | `bugs_open/100_HANDOFF_2026-07-26_verification_write_path_cannot_record_provenance.md` |
| commits | `2ebabf2ca` (fix), `70885daf0` (gofmt sweep + doc correction), `b2a12ae99`, `4a5eeb5b1`, `01d1bd32e` (docs) |
| transferable patterns | `016b §9` — *"Omitting a key is not neutral"* and *"A registry that everything registers with and nothing reads"* |
| new callable mechanisms | concept register `adopting-and-scraping.md` SCR-002/003/004 |
| wrong calls | `WRONG_CALLS.md` — two entries dated 2026-07-28 |

---

## 7. ROUND 2 — submitted 2026-07-28 ~18:50, supersedes §3a

**All four owed items are done and committed** (`9cdc08838`, `9ed2d6a87`). Resubmitted
on the same correlation (`RESUBMIT_CORR=f4cf0aab-5a08-4475-91ea-fa831cff323c`), so
§3a's list is now history — **do not redo it**. Round 2 is 5 edits / 11 grounded_in.

| round-1 objection | seat | discharged by |
|---|---|---|
| parallel doc trail, no `doc_notes` *(GATING)* | `tooling_provenance` | SQL `258` — three notes APPLIED and verified |
| declaring the keys silenced the audit | `editquality` | `ConditionalKeys` + a third report section + regression test |
| "no callers" asserted, not shown | `prior_art_librarian` | `git grep … 2ebabf2ca^` in `grounded_in`, pinned ref |
| name the non-vet pipelines; 062 risk | `guardian` | all three named with their `formats`; post-roll 0/0 measured |
| how is "chassis is live" confirmed? | `debug_historian` | pod-grep of created symbols, in the RUNBOOK and the PLAN |

**Round 2 needs no roll** — `ConditionalKeys` feeds only the audit path, which runs
from source via `go run`. Chassis behaviour is unchanged by it.

### If round 2 comes back APPROVED
Commit the trailer `Council-Reviewed: f4cf0aab-5a08-4475-91ea-fa831cff323c` on a
follow-up (the code is already committed — the trailer is what makes the `098`
report's commit↔verdict join exact). **Re-read `decided_by` for THAT round first:**
a later approval can attach to a different plan.

### If round 2 comes back REVISE again
Read `decided_by` and `unreadable` **before** reading the objections — round 1 had
`unreadable: 1` that was *not* the decider, and ~11% of rounds are decided by one
seat's unparseable JSON (`bugs_open/119`). Then check whether the reviewers'
read-only queries returned zero again: in round 1 they returned **0 for facts
measured true twice** (228 actions/1,155 steps; 3 live steps at
`only_main_content:false`; 1 `add_protocol` carrier). That is the harness, not
evidence — but it was raised in round 2 as a caveat, never as a defence, because no
round-1 objection depended on those queries.

## 8. What is genuinely left after round 2

1. **`bugs_open/100`'s two-column verification run** — blocked on `vetcomparison`
   restarting collection. Unchanged, and it is the only thing standing between 100
   and closure. §3b has the query.
2. **The coverage ratchet** — 208 undeclared actions. §3c.
3. **Nobody has decided** whether `vet-practice-verifier` and
   `domain-research-classifier` should switch to `action: "crawl"`. They now warn at
   runtime *and* appear in the audit's CONDITIONALLY HONOURED section, so the state
   is visible from two directions instead of none. Still an owner's call.
4. **`domain-research-classifier` has no owner** — worth raising separately from
   this bug.

---

## 9. ROUND 3 — **APPROVED** 2026-07-28 ~20:09. §7's verdict branches are now spent.

```
round 1: revise   | decided_by "gating objection from tooling_provenance"
round 2: revise   | decided_by "unreadable reviewer(s): review_editquality.result"   <- harness (bugs_open/119)
round 3: APPROVED | decided_by "approved with 1 advisory objection(s) — none high-severity" | unreadable 0
```

7 approve / 1 advisory object. `editquality` — unreadable in round 2 — approves with
**zero** objections, which is the verdict `ConditionalKeys` was built to earn.
Trailer `Council-Reviewed: f4cf0aab-5a08-4475-91ea-fa831cff323c` is claimed.

**Nothing is owed to the council.** The one medium advisory (no reuse-coverage search
before adding `ConditionalKeys`) was discharged after the fact: the search was run,
returned only comment prose, and `ConditionalKeys` is genuinely new. NOTES §12.

### The whole of what remains

1. **`bugs_open/100`'s two-column run** — blocked on `vetcomparison` restarting
   collection. §3b. This is the only thing between 100 and closure.
2. **The 062 watch is still unexercised** — `grep -c "Starting scrape"` was **0**, so
   the clean error log has a zero denominator. §2, and the RUNBOOK puts the attempt
   count above the watch.
3. **208 undeclared actions** in the coverage ratchet. §3c.
4. ~~**Two owner calls, neither this thread's:**~~ **ONE of the two is now RULED
   (2026-07-28, owner): `vet-practice-verifier` and `domain-research-classifier`
   STAY on scrape, warning.** Not a deferral — a decision. The mismatch is visible
   from two directions (runtime warning + the audit's CONDITIONALLY HONOURED
   section), which is what 101 was about; changing what the agents *do* is a
   separate change for whoever owns them. **Do not flip the action "to finish the
   job".** Recorded in `bugs_closed/101` residual 2.
   **Still open:** who owns `domain-research-classifier` at all.
