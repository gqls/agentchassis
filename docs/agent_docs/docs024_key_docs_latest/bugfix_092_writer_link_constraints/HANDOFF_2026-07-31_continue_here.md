# HANDOFF — continue here (bugfix_092 lane, 2026-07-31 evening)

**This lane is FINISHED.** `bugs_closed/092`. Nothing is owed on it. This document exists so
a fresh session can (a) confirm that in thirty seconds without re-reading the workstream, and
(b) pick up the next bug with the traps this lane paid for already in hand.

---

## 1. State: DONE, and how to re-confirm it cheaply

`bugs_open/092` → **`bugs_closed/092`**. Fixed, council-APPROVED at round 1
(`4b8c5e21-011b-40f0-819a-3dfa4b4c7b1d`), LIVE on chassis **v1.0.1219**, and proven by an
induced run at 19:16Z.

Five commits, all with a `Council-Reviewed:` or `Council-Submitted:` trailer:

| commit | what |
|---|---|
| `2e1bfb39e` | the fix + 9 tests; `link_constraints.go` deleted |
| `9a57d2395` | round-1 review answer: the shared `linkablePageStatusPredicate` + its test |
| `b3ceec3d0` | verdict record, the sibling audit, 016b §9 pattern, WRONG_CALLS ×3, LANDMINES ×2, register |
| `95d74e1ea` | milestone summary (now carrying a supersession banner) |
| `902c0579f` | the close: pod proof + induced run, `bugs_open` → `bugs_closed` |

Re-confirm in one query — this should keep returning `database` / non-zero:

```sql
SELECT created_at, collected_data->'link_context'->>'page_count' AS pages,
       collected_data->'link_context'->>'source' AS source,
       collected_data->'link_context'->>'degraded' AS degraded
FROM orchestration_states WHERE collected_data ? 'link_context'
ORDER BY created_at DESC LIMIT 3;
```

Pre-fix rows are `0` with NULL `source` (the field did not exist), so old and new rows are
self-labelling — you do not need the roll time to tell them apart.

**If it regressed**, the two things to check first: `agent_error_log WHERE
error_code='LINK_CONTEXT_UNAVAILABLE'` (the loud arm — if it is firing, the site id stopped
resolving), and whether anyone widened or moved `extractSiteID` (see the landmine below).

## 2. What was deliberately NOT done — do not mistake these for oversights

- **No already-deployed page was repaired.** This is a write-path fix. The live 404s belong
  to `bugs_open/071` and `bugs_open/097`. Do not "finish the job" by rewriting pages; that is
  a different bug with a different owner and its own risk profile.
- **`extractSiteID` was NOT widened** despite two of its five callers being exposed. Five
  callers, several treating `""` as "skip this work" — widening it is a silent behaviour
  change to unrelated actions. The audit is in `bugs_closed/092`.
- **The `link_registry` question was left `[UNDETERMINED]`, on purpose.** The table has **0
  rows all-history** and `ExtractAndSyncLinksAction` returns a success-shaped
  `{"links_extracted": N, "persisted": false}` when the site id does not resolve — but its
  only agent has **0 orchestrations in the retained window**, so "the exposure fires" and
  "the agent never runs" are indistinguishable. Contributed to `bugs_open/165` (which owns
  that table). **If you pick this up, the missing evidence is a single `multipage-website-builder`
  run** — get one and the ambiguity collapses. Do not conclude without it.
- **The dead `link_constraints` config block** on `page-content-writer` was left in place. It
  is now provably unread; removing it is a live config change with its own risk and no gain.

## 3. Traps this lane paid for — these are the reusable part

All are in the fleet files (`LANDMINES.md`, `WRONG_CALLS.md`, 016b §9); repeated here because
a handoff that only points at four other documents does not get read.

1. **`extractSiteID` cannot see `input_data.site_id`** — the only place `page-content-writer`
   keeps it (26/26 runs vs 0 for every location the helper knows). A DB read wired to the
   shared helper resolves `""` and fails *exactly like the bug it is fixing*. I nearly did
   this. Ask the runs where the identity is, not the helper.
2. **Choose an induction target by where your evidence is RECORDED, not by whether the run
   can succeed.** The 079 lane induced successfully and got a null result. I targeted an
   `owned` page on a non-deployed site *because* `save_page_sections` refuses it — guaranteeing
   zero writes — while `prepare_link_context` records long before the refusal.
3. **`kubectl run -i | kcat -P` silently sends nothing ~4 times in 5.** Put the payload in the
   container COMMAND with `--command`, and append `&& echo PUBLISH_OK`. No receipt, no publish.
4. **Pod-grep in BOTH directions.** A string you ADDED proves the binary contains it; only a
   string you REMOVED proves it is the *new* binary. Plus a positive control in the same exec.
5. **A `+1` overflow probe answers a boolean, not a quantity** — `len - limit` from a
   `LIMIT n+1` can only ever be 1. Ask what the maximum possible value of any number you emit
   is, under the query that produced it.
6. **After adding a sentinel, grep every comparison that reads it.** I added `-1` and left two
   `> 0` guards, reintroducing the silent-cap shape inside the fix for the silent cap.
7. **Run the mutation you are about to describe.** I wrote a mutation prediction into a test's
   doc comment; it was the opposite of the truth, and only running it caught that.

## 4. Environment notes that will bite the next session

- **`cmd/reasoningset` does not compile at HEAD** (`main.go:504`, three declared-and-not-used
  vars). Not this lane's, unmodified in the tree — it is committed breakage. `go build ./...`
  therefore fails and will look like *your* fault. It does **not** block any service: each
  dockerfile builds only its own `./cmd/<service>`, and `go build ./cmd/agent-chassis/`
  succeeds. Build `./platform/...` and your own service, not `./...`.
- **`/tmp` is a 16G tmpfs shared by ~30 sessions.** A `git archive HEAD` checkout plus its
  build cache is ~220MB; it hit 100% during this session. When full, Bash reports ENOSPC but
  **the command may have succeeded with only its output lost** — check before re-running
  anything non-idempotent. `rm -rf` your checkout in the same breath as the test that needed
  it. I removed two scratchpads whose session *and* transcript had both been idle >24h
  (freeing 2.2G, with the owner's go); two others with stale scratch but live sessions were
  left alone.
- **The council gate turned round in ~8 minutes**, not the ~30 the runbook budgets. Do not
  assume you must commit-then-wait; you may well have the verdict before you finish the docs.
- **`0 of 185` active agents declare `input_contract`/`output_contract`.** The `guidelines`
  seat raises DECLARED CONTRACTS on submissions touching action outputs; it is inert
  fleet-wide, and one query settles it rather than an edit.

## 5. Picking the next bug — what this lane learned about the choice itself

`who-owns.py <n>` reads **commits**, so a session mid-fix is invisible to it. It named a lane
for `092` that had closed its own bug two days earlier. What actually worked:

```bash
cd ~/.claude/projects/-home-ant-projects-agentchassis/
for f in $(find . -name '*.jsonl' -mmin -180 -size +10k); do
  n=$(grep -c -E '<CODE SYMBOLS of the target>' "$f"); [ "$n" != 0 ] && echo "$f :: $n"
done
```

**Grep the code symbols, not the bug number** — sessions working the same class arrive via a
sibling bug and never type your number, while several of the hits you *do* get are the same
`MEMORY.md` line loaded into three contexts. Then read the context around each hit: a hit is
a lead, not a verdict.

Bugs confirmed **free** at 19:00Z today (no active transcript, no recent commits) and not
taken by this lane — a starting shortlist, but **re-run the check, it goes stale in minutes**:
`080`, `081`, `093`, `096`, `097`, `100`, `107`, `111`, `113`, `114`, `115`, `116`, `118`,
`121`, `122`, `123`, `126`, `128`, `132`, `134`, `150`, `158`, `160`.

Two notes on that list from reading around them: `071` is a *bundle* of one closed mechanism
and three open ones (its own triage recommends splitting it, and says the owning lane should
pick the split — not a sweep); `081` and `126` are both recorded as needing an **owner call**
rather than an implementation.
