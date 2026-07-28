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
kubectl -n ai-persona-system logs deploy/web-scrape-adapter --since=3h \
  | grep -i "Message Size Too Large\|Failed to produce"
```

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
