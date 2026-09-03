# HANDOFF — bug sweep lane, 2026-09-03c (post-roll, mechanism LIVE)

> ⚠ **THERE ARE TWO `09-03c` HANDOFFS IN THIS DIRECTORY AND THEY ARE DIFFERENT SESSIONS.**
> `HANDOFF_2026-09-03c_bugsweep4_continue_here.md` is a concurrent sweep covering bugs **361,
> 366, 400** — disjoint from this one, and neither supersedes the other. **This file is the
> 442/464 thread** (and carries forward 338/320/404/407). Read whichever thread you are resuming.

**Read this instead of `HANDOFF_2026-09-03b_continue_here.md` and
`HANDOFF_2026-09-03_continue_here.md`.** Both are still correct about 338/320/404/407 and the four
owner rulings; what changed is that **`bugs_open/442` is BUILT AND LIVE**, a second bug
(`bugs_open/464`) was filed out of it, and the council is mid-round-3. Where the earlier files are
now wrong I say so here rather than editing them.

---

## 0. STATE IN ONE TABLE

| item | state 2026-09-03c |
|---|---|
| **442** silent refusal | ⚖ **BOTH HALVES LIVE.** Config `728`+`734`; Go on **`v1.0.1359`** (pods 13:28Z), verified at the binary on both pods. **Never yet exercised** — see §3 |
| **442** council | ⚖ **APPROVED at round 3** (`76288ff9-…`, 1 advisory, none high). r1 REVISE, r2 REVISE, r3 approved — **each round found something real**. Verdict READ, advisories answered (§10j) |
| **464** unread call sites | ⚖ **FILED AND CLOSED the same day** → `bugs_closed/464`. Every copy-gate caller READ; **442's shape has no second instance**. ⚠ Its own population was wrong in both directions — §7.1 |
| **SEO-008** register | LIVE, with its index row. SEO-004 corrected (it held a 4th copy of the stale reason list) |
| 338 / 320 / 407 / 359 | unchanged — see the 09-03 handoff §0 |
| **404** r4 objections | still owed a read. Not this lane's |

---

## 1. WHAT IS LIVE, AND HOW TO CHECK IT WITHOUT TRUSTING ME

**The mechanism.** When a site's copy gates refuse a meta description,
`save_page_meta_description` no longer just logs it. It files a `meta_description_refused` work
item **at `meta-description-repair`**, which re-asks for the sentence **with the refusal quoted
back** and saves through the **same gated action**, so the same rules judge the retry. A second
refusal parks at `needs_human_review` carrying both attempts.

```bash
# 1. Is the Go actually in the running binary? BOTH CONTROLS ARE THE POINT.
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
for s in meta_description_refused meta-description-repair candidate_looks_internal ZZZ_cannot_exist_9f3a; do
  kubectl -n ai-persona-system exec $POD -- grep -aq "$s" /proc/1/exe && echo "PRESENT $s" || echo "absent  $s"
done
# expect: PRESENT, PRESENT, PRESENT (positive control), absent (negative control)
# ⚠ DO NOT use `kubectl logs | grep 'build provenance'` on the CHASSIS — it matches the
#   landmine corpus ABOUT build provenance and returns megabytes. Confirmed again today.

# 2. Is the agent live and still fill-blanks-only? (see §4.1 — this one matters)
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT type, is_active,
       default_config#>'{workflow,steps,save_description,config}' ? 'overwrite_existing' AS declares_overwrite
FROM agent_definitions WHERE type='meta-description-repair' AND deleted_at IS NULL;"
# expect: t | f      (f, or the key absent, is the OWNER RULING — see §4.1)
```

---

## 2. THE ONE MEASUREMENT ANY LANE FILING FINDINGS SHOULD STEAL

This decided the whole design and took one query.

`[MEASURED 2026-09-03, `site_work_items` UNION `site_work_items_archive`]`

| shape | items | complete | **%** | parked |
|---|---|---|---|---|
| **WITH a `handler_agent`** | 56,315 | 46,465 | **83%** | 407 |
| **NO handler (flag-only)** | 6,699 | 1,142 | **17%** | 989 |

And `voice_tells`, the queue a copy-gate refusal would naturally have joined: **69 rows, every one
`handler_agent = ''`** — 3 complete, 66 parked, nothing filed since 2026-08-27.

**A review queue is not a graveyard because people are busy. It is a graveyard because nothing is
pointed at the row.** A flag-only `needs_human_review` item looks like a fix and IS the 17%. File
at an actor, or state why no actor can act.

---

## 3. ⚠ NOTHING HAS EXERCISED IT, AND "NO ROWS" IS NOT "IT DOES NOT WORK"

`[MEASURED 2026-09-03]` **zero** active pages are both blank and past the backfiller's gate, so
**no page can be refused and no item can exist.** Run the demand control before concluding
anything:

```sql
SELECT count(*) FROM pages WHERE status='active'
  AND COALESCE(meta_description,'')='' AND page_visible_text_len(id) > 200;   -- 0 today
-- and the item table, once that is non-zero:
SELECT status, count(*) FROM site_work_items WHERE item_type='meta_description_refused' GROUP BY 1;
```

The first real `meta_description_refused` row is the acceptance evidence. Until one exists, the
mechanism is **live and unproven**, and saying otherwise would be the `bugs_open/338` mistake
(an acceptance test that could never fire) one lane along.

---

## 4. ⚠ FOUR THINGS THAT MUST NOT BE LOST

### 4.1 `overwrite_existing` must stay UNDECLARED — it is an owner ruling, not a default
`102_coverage_ratchet.txt` line 105 and `bugs_open/320` §15: rewriting **published** copy on an
automated finding is authority the owner **explicitly withheld on 2026-08-21**, granted once for a
one-off dispatch and never armed on any agent. `meta-description-repair` is exactly the shape that
note warns about and does **not** carry it: the key is undeclared, so it defaults `false` and can
fill a blank only. **Adding that one key is a one-line, entirely innocuous-looking config edit that
crosses an owner ruling.** Also a LANDMINE (footprinted on both agents), which carries the trap
inside the trap: the *backfiller's* save step is NESTED in `backfill_loop.sub_workflow`, so the
top-level jsonb path returns **NULL** for it — and NULL is not `f`.

### 4.2 A commit in HEAD + a roll that happened ≠ your code is live
**`v1.0.1358` rolled at 12:06Z with both my commits already ancestors of HEAD, and did NOT carry
them** — the build was cut from an earlier HEAD. Measured at the binary, both controls behaving.
It shipped in `v1.0.1359` at 13:28Z. `git merge-base --is-ancestor` answers *"is my commit in the
source"*, which is a different question from *"is my code in the binary"*, and this is the case
that separates them. `bugs_open/442` §10h.

### 4.3 A roll KILLS an in-flight council round
Round 2 was mid-review at 12:05; pods were replaced at 12:06; **7 of 11 in-flight orchestrations
fleet-wide froze in the same minute**, which is how you tell a roll from a sick submission. The
correlation survives, the run does not — resubmit under `RESUBMIT_CORR`. Second recorded instance.

### 4.4 ⚠ Another session is mid-edit in THIS lane's test file — do not sweep it, and note the corruption
`platform/orchestration/actions/save_page_meta_description_test.go` has **uncommitted** changes
that are not this lane's: a `gofmt` comment reflow (spaces → tab), plus one real corruption. In
`TestSavePageMetaDescription_EmptyCandidateWritesNothing`'s **MUTATION** instruction, the two
ASCII quotes in *"an UPDATE writing `''`"* have become a single curly `”` (U+201D). It compiles —
it is a comment — but **the mutation instruction now names the wrong literal**, so anyone
following it writes something else and draws the wrong conclusion about whether the guard holds.
Not ours to commit (it is their working state); worth telling them, and worth not sweeping into
a pathspec commit of that file.

### 4.5 The working tree may not compile, and it may not be yours
Another session has an **untracked** `criteria_value_assertions.go` whose `itoa` collides with a
test helper at HEAD. `go test ./platform/orchestration/actions/` fails in the tree and proves
**nothing** about HEAD. Everything this lane ran went through
`scripts/verify-head-builds.sh --test [--with <path>]`, which was green throughout. ⚠ `--with`
takes a **repo-relative** path; an absolute one exits **2** ("could not run"), and a grep for
`FAIL` cannot tell that from a passing run.

---

## 5. WHAT IS LEFT

**Owed on 442: nothing.** Round 3 **APPROVED** (1 advisory, none high-severity; 7 seats clean).
All three advisories answered from the code in §10j — including `recurrenceExpected`, where the
brake's own counting query (`status IN ('complete','failed')`) shows a parked repair is **not** a
strike, so it holds the dedup slot and the next refusal coalesces onto it. `098` credits the
`Council-Submitted:` commits automatically now the correlation is approved; no amend.

**Open on 442, so the file stays OPEN:**
- **No verifier** for `meta_description_refused`. Registering one fails **five** build guards and
  needs a live migration amending the claimed-item-timeout sweep's `pre_query`, **merged with any
  other lane's pending amendment** — `discovery_checks/verify_required_fields_missing.go`
  documents the exact sequence. Do not start it casually.
- **`voice_gate_unreadable` is still silent**, correctly (a rewrite handler is the wrong actor for
  an infrastructure fault) — but it still strands a page for ever. Needs an operational surface.
- **§9d's second silent path is unbuilt.** The writer's own prompt says *"omit that page entirely…
  returning fewer entries than you were given is a correct answer"*, and nothing compares
  `pages_missing_meta.count` against `jsonb_array_length(written.result.descriptions)`. **The fix
  filed here cannot see that path** — it fires from inside the action, and a dropped page never
  reaches it. Fifth candidate: compare the two integers already in `collected_data`.

**`bugs_closed/464` — DONE, and the result is worth knowing before you touch this area:**
Every copy-gate caller on the tree was read (ten call sites across `platform/ internal/ pkg/ cmd/`,
with a control). **`save_page_meta_description` was the ONLY one that returned a refusal as
`(map, nil)` with nothing asserting on it.** The others error, block to human review, record into a
structured `rejected`/verdict list, file a work item, or write a durable record by design. So
442's blast-radius claim is **forward-looking, not realised** — the shape travels with the gate and
will bite the next single-value caller.
⚠ **And 464's own population was wrong in BOTH directions**, which is now a landmine: a false
positive from `grep -l` matching a **comment** that said the file does *not* use the symbol, and
false negatives from scoping the census to **one directory's top level** (5 files that way, 10
whole-tree, including a whole second package under `internal/`).

**Adjacent, other lanes' (carried forward, not re-verified today):**
`platform/livespec` RED at HEAD (405 lane) · `TestNoHandSpelledTombstonePredicate` RED at HEAD
(114/IMG-077) · `_RELOCK` unclassified migration suffix, still WARNed by 097 · `WII-035` duplicate
index row · ⚠ **two different migrations are numbered `728`** (mine and a boxingonline one), both
applied and both recorded — no damage, the ledger keys on filename, but the numbering collided.

---

## 6. THE MISSTEPS, BECAUSE THEY ARE THE PART THAT DOES NOT RE-DERIVE

Full detail in `NOTES_bugsweep.md` and `WRONG_CALLS.md` (six rows added today).

1. **A zero with two sufficient causes, attributed to one.** `442` §2 blamed retention for finding
   no refusal records; the records were there and there had simply been no refusals. That wrong
   half became an *instruction* in §6 telling the next session not to look where the evidence was.
2. **A negative test that asserted nothing, whose comment said it couldn't.**
   `ExpectationsWereMet()` with no expectations registered is nil unconditionally. Only the
   mutation found it. **The arm I got wrong was the negative one** — always the easier one to
   write vacuously.
3. **A mutation run that never executed** (exit 2, absolute overlay path) and would have read as
   "not killed" to a grep for `FAIL`. **A mutation harness needs a three-way exit check: 0
   survived, 1 killed, 2 you learned nothing.**
4. **Citing instead of showing — three consecutive council rounds, four `WRONG_CALLS` rows, and
   the third one cost a gating HIGH.** Every claim was TRUE. That is the point: a reviewer cannot
   tell a verified claim from an unverified one, and I had logged the remedy myself that morning
   and then not applied it to the next submission I wrote. **Logging a lesson is not adopting it.**
5. **Backticks in `git commit -m` executed** and ate the subject of a sentence in `356196fe9`.
   Documented landmine; walked into it while writing a commit message about being careful.

---

## 7. WHAT SHIPPED THIS LANE TODAY

`5a8728db9` mig 728 (the operator message, council **APPROVED** `2ed33c57`, 10/10 seats) ·
`a3e092368` 442 §9 · `b811964f7` + `6815e3cb0` + `78cfdd170` ledgers · `652f32f74` 016b §9 ·
`699b6e439` NOTES + the lane's first `README_where_we_are.md` · `75fa195a2` §9g ·
`776511e70` **the Go + mig 734** · `356196fe9` **r2 revision** (the filing reports its own outcome)
· `789678c49` + `87b37779f` + `5a12c163e` register **SEO-008** · `a3fc809eb` LANDMINE ·
`9e72f75b0` **`bugs_open/464`** · `d1693c728` §10i · `185de4cd6` §10g · `99e7be3ad` §10h.

Migrations `728` and `734` applied and ledger-recorded. Council `2ed33c57` **APPROVED**;
`76288ff9` at **round 3**, pending.
