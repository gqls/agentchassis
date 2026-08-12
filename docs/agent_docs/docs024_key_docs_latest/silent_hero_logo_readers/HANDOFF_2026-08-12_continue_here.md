# HANDOFF — 2026-08-12 — continue here

**Lane:** `silent_hero_logo_readers` — commission item 2, owner ruling 2026-08-10 (*"2. yes."*).
**Read first:** `PLAN_2026-08-11_…` (design + the two decisions that depart from the commission's
letter), `NOTES_…` (the missteps — three of them are mine and one is expensive), `RUNBOOK_…` (every
command with its gotcha).

---

## 1. State in one paragraph

**Item 2 is DONE, council-APPROVED, COMMITTED and LIVE on `agent-chassis:v1.0.1290`** (both
replicas, verified at the binary with a negative control). Its **behavioural proof has not
happened**: zero rows, but the demand control says zero hero/logo deploys since the roll, so the
path has had no opportunity to fire. **Nothing in this lane is blocked on me.** The open items are
(a) a real site build that deploys an image, which this lane cannot force, and (b) a **newly
measured defect in the diagnosis harness's code tier**, which is now the thing standing between
`bugs_open/236` and its root cause. Section 4 is the recommended next move; section 6 is the list of
owner decisions, unchanged from yesterday except that one of them is now better informed.

## 2. What shipped

| commit | what |
|---|---|
| `038211dd8` | **the code — NOT this lane's commit.** Another session's `git add -A` swept all four files into HEAD mid-verification (see §7) |
| `d553f1d73` | the standing five + the submission |
| `258b8beb1` | ratchet line (not a registrable mechanism) |
| `c98c20c9d` | council response: both mediums answered with measurements |
| `b4f005fbb` | milestone summary + four corrections written back into the commission |
| *(this handoff's commit)* | 090 verdict read, the code-tier finding, `WRONG_CALLS`, roll verification |

**The mechanism**, all in `platform/orchestration/actions/deployed_image_read_audit.go`:

1. `deployedImageURL` — the single door all three readers now use
   (`v3_site_actions.go:1125`/`:1136`, `assemble_from_library.go:452`). On a container that is
   **present but carries no usable `image_url`**: a `Warn` **plus** an `agent_error_log` row
   (`error_code DEPLOYED_IMAGE_RESULT_MISSING_URL`, severity `warning`) via the existing
   `LogActionError` door.
2. **The demand gate** — an ABSENT container records nothing. This is the load-bearing part: the
   action runs for every page build and most pages deploy no image. Proven by mutation.
3. `context.fallback_sibling_present` — the discriminator between "the container lost the key" and
   "the URL is lost everywhere". Aimed at `bugs_open/236` §5.

## 3. THE 090 CAME BACK — and its failure is the most valuable thing in this handoff

**Run `dbcc4259-ab84-494b-a48b-1df647209a40`, COMPLETED 2026-08-11 18:42Z, 4 iterations, verdict
`UNVERIFIABLE`.** My candidate for §5's root cause is **neither confirmed nor refuted** — do not
quote it either way.

**Why it could not decide**, its own `needed_evidence` verbatim:

> *"The bundle never renders the bodies of `persistAwaitingStateWithRetry`, `processAwaitResponse`,
> or `applyResponseToState` — only `storeActionResult`'s body and a bare signature line for
> `applyResponseToState` are present."*

**The index is not the problem. Measured 2026-08-12:**

| symbol | kind | `length(body)` | lines |
|---|---|---|---|
| `persistAwaitingStateWithRetry` | func | 2,058 | 2067–2132 |
| `processAwaitResponse` | func | 5,619 | 1914–2063 |
| `(*SagaCoordinator).applyResponseToState` | method | 4,746 | 2650–2779 |
| `storeActionResult` | func | 970 | 1863–1892 |

**The index held four bodies; the bundle rendered one.** So this is a selection/rendering defect in
the **code tier** of the diagnosis bundle — and the verdict prompt's cite-or-abstain rule then acts
on that absence, which is *exactly* the mechanism commission item 5 fixed one tier over (a
filtered-out table and a non-existent table rendered identically).

**This closes a question item 5 explicitly left open.** That lane's `PLAN` §3: *"whether the code
tier has an analogous blind spot is unexamined `[UNMEASURED]`"*. Measured now; the answer is yes.

## 4. RECOMMENDED NEXT MOVE — fix the code tier, and why it beats the alternatives

**Not** another `090` on 236: a third run on the code-path question will fail identically until this
is fixed, and that is two runs' credits already spent learning the same thing.

The case for it, briefly: it is the **same shape as a fix that already passed council and worked**
(item 5, `df9dae6c`, approved round 1); it **unblocks `bugs_open/236`'s root cause**, which is
otherwise stuck; and it unblocks every future question whose evidence is a function body, which is
most of them. The likely target is the code-tier gather/assemble path beside
`diagnose_load_runtime_action.go`'s `gatherSchema` — but **find the producer first and do not assume
it is symmetrical with the schema tier**; item 5's commission guessed the producer wrong and the
guess cost a search.

⚠ **Before starting, read the bundle artefact itself**, not the index — that is the whole lesson of
§7's wrong call:

```sql
SELECT left(body, 4000) FROM diagnosis_artifacts
WHERE correlation_id='dbcc4259-ab84-494b-a48b-1df647209a40' AND kind='bundle'
ORDER BY created_at DESC LIMIT 1;
```

⚠ **Scope routing is NOT free here.** Item 5's own approval carries a recorded forward concern from
the `architecture` seat: *"if this phrasing pattern gets reused across other diagnosis actions it
could accumulate into a de facto shared vocabulary without ever passing through architecture
review."* The commission records the ruling that follows from it — **the SECOND diagnosis action to
teach its prompt a self-service capability in this style is the one that needs an RFC.** If your fix
adds instructive prompt text in item 5's style, you are the second. Read
`COMMISSION_2026-08-10_owner_rulings_five_pieces.md` §5's status block before choosing the route.

## 5. Item 2's own outstanding verification — and how not to misread it

| check | result | date |
|---|---|---|
| live on the chassis | **YES** — `v1.0.1290`, literal present in `/proc/1/exe` on both replicas, control absent | 08-12 |
| rows with the new `error_code` | **0** | 08-12 |
| **DEMAND CONTROL** — `hero_deployed` / `logo_deployed` present anywhere | **0 / 0** of 6,364 retained | 08-12 |

**The zero is unfalsifiable, not reassuring.** No demand has reached the path since the roll. Per
`bugs_open/236` §3, do not read the count as an incidence rate in either direction. The proof needs
a site build that deploys a hero or logo, and then the RUNBOOK's paired query **within four hours**
(`AWAITING_RESPONSES` is pruned on that clock — measured first-hand off the live
`scheduled_tasks.pre_query`, not inherited).

## 6. Decisions waiting on the owner

**None are blocked on me.** Unchanged from the item 5 handoff except where noted:

1. **Which commission item next.** Item 2 was the last one needing no ruling. Remaining: item 1
   (large, mostly investigation, design decision reserved to the owner) and item 3 (medium, spans
   three layers, needs a routing call). **NEW OPTION, and my recommendation: the code-tier fix in
   §4**, which is not a commission item but now gates item 1.
2. **Item 3's routing:** council gate or architecture RFC? It changes a client return signature and
   adds a field to a shared adapter response payload. The commission: *"if in doubt write the RFC —
   the cost is one document."*
3. **Item 3's modelling question:** `deploy_commit` is per-component, but a page is many components
   across possibly several commits. The column's original author never answered whether the page
   level wants it too.
4. **Item 1's design decision** (Design 1 vs Design 2) is explicitly reserved. The census's *"0
   breaks"* premise for Design 2 is contradicted by production — **re-measure the baseline before
   scoring either**. My §5 candidate, if it survives, bears directly on this.

## 7. Traps this lane paid for

- **A wrong call of mine, and the shape is general.** I declared 236 §5a's "loop cannot read the
  bodies" blocker CLEAR because *the index was fresh*. The loop had complained about the **bundle**.
  A fresh index returns "present" whether or not the bundle renders it, so **the check could not
  have come out false.** Logged in `WRONG_CALLS.md`. General form: *"the data exists" and "the
  consumer received it" are independent facts, and the second one is always the claim.*
- **A 090 verdict is NOT in `diagnosis_artifacts`.** That table holds only `kind='bundle'` rows for
  a 090 correlation. The verdict is on the run's own row, at
  `orchestration_states.collected_data->'verdict'`. Two of my polls said "NOT YET" about a run that
  had finished five hours earlier. Always print a connectivity control beside a poll.
- **The 090 row is found by `correlation_id`**, not by
  `collected_data->'input_data'->>'fix_correlation_id'` — that is the *council's* key shape, and
  using it on a 090 looks like a dropped dispatch.
- **`status='COMPLETED'` is not success** (`bugs_open/099`): a failed step shows COMPLETED with
  `error` NULL. Read the payload.
- **Guess a column name and you get a confident error, or worse.** Two of my queries died on
  `created_at` (`agent_error_log` uses `occurred_at`) and `name` (`code_symbols` uses `symbol`).
  `\d <table>` first, as CLAUDE.md says.
- **Never leave a mutation in the tree across a tool call you do not control.** `038211dd8` swept
  this lane's files into HEAD minutes after a mutation window closed; had it landed inside one, a
  deliberately disabled guard would now be in HEAD under someone else's commit message. Back up,
  mutate, restore, and `diff` against the backup **in the same breath**.
- **The commission's line numbers were stale on day one.** Grep the symbol.

## 8. If you change this code

The demand gate is what keeps this quiet on the fleet, and
`TestDeployedImageURL_AbsentContainerRecordsNothing` is what protects it. **If you widen what the
readers accept, you are in item 1's territory** — that is `bugs_open/236` §4 candidate 2, refused
here on the commission's explicit instruction and reserved to the owner. The four dynamic-key
readers listed in the census (NOTES, 08-11 evening) are the honest residual: they resolve a key from
step config, so they cannot be classified by reading the Go, and they are `[UNVERIFIED]` rather than
clean.
