# HANDOFF — routing capability guard (`bugs_open/395` root cause) · 2026-08-26

**COLD-START = this file + `bugs_open/395` §9/§10/§11 + `bugs_open/320` §4/§5/§9/§15 + register
`WII-035` (and `WII-033`, `CLM-024`) + `architecture_review/RFC_057`.**

> **Why there is no set of five standing docs here.** The technical record, the missteps and the
> mechanism are already in `bugs_open/395` (§9–§11), `WII-035`, `LANDMINES.md`, `WRONG_CALLS.md` and
> `RFC_057`. CLAUDE.md's working-docs rule says *"point at bugs, don't restate them"* and warns
> against forking a second account that drifts, so this directory holds the handoff and the OPEN
> DECISIONS only. If this becomes an ongoing programme (it will if decision 1 goes a certain way),
> open the five then.

---

## The one-line state

> **The root cause of `bugs_open/395` is FIXED AND LIVE AND PROVEN. Rule 3b (`WII-035`) went live on
> `v1.0.1341` at 2026-08-25 23:11:52Z and fired twice within 33 minutes on two sites nobody had
> measured. What remains is not engineering — it is FOUR DECISIONS, one of which is the owner's and
> blocks everything downstream: nothing in the estate may currently rewrite a published page
> description, so these findings are correctly recorded and permanently unrepairable.**

---

## 1. What is DONE — do not re-take any of this

| | state |
|---|---|
| **routing rule 3b** — a finding is not routed at a handler that cannot WRITE the field its own criterion names | **LIVE AND PROVEN, `v1.0.1341`.** `af3194204` + `a48c5c942` + `f4aa19ae7`. Register `WII-035` |
| council | **APPROVED round 1**, corr `021cb965-c12b-482d-991a-a1a93f52edea`. All objections fixed or explicitly conceded (`a48c5c942`, `f4aa19ae7`) |
| mutation proof | **FOUR mutations RUN and observed failing**, not asserted. Listed in §5 below |
| completion **gate 1c** (`WII-033`, the other lane's) | LIVE on `v1.0.1339`, verified independently by two sessions. Records, does not refuse |
| the diagnosis | `bugs_open/395` §9 (recurrence of `320` §5; the criterion was UNREACHABLE) and §11 (live proof) |
| `bugs_open/320` §4's "no UPDATE path exists" | **corrected as stale-by-addition** in a CONTRIB into that lane — its own fix added one |
| `LANDMINES.md` ×3, `WRONG_CALLS.md` ×3 entries, owner's plain-prose log | all written and committed; landmines dispatched to the verifier |
| `RFC_057` (the shared-contract question) | **filed by the `vigilant_designer_offer_analysis` lane**, awaiting an owner ruling |

### 1a. The proof, so nobody re-derives it

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system exec "$POD" -- grep -aq 'no_writer_for_page_field'   /proc/1/exe  # PRESENT
kubectl -n ai-persona-system exec "$POD" -- grep -aq 'handler_reported_no_change'  /proc/1/exe  # PRESENT (control)
kubectl -n ai-persona-system exec "$POD" -- grep -aq 'zzz_invented_control_395'    /proc/1/exe  # absent  (control)
```
⚠ **Do NOT use CLAUDE.md's `grep -m1 'build provenance'`** — that string is in no Go source in this
repo and its documented "not in range" caveat absorbs the real failure. `LANDMINES.md` has the entry.

```sql
-- rule 3b's own firings
SELECT s.domain, wi.created_at, wi.spec->>'unwritable_field', wi.spec->>'would_have_routed_at'
  FROM site_work_items wi JOIN sites s ON s.id=wi.site_id
 WHERE wi.spec ? 'unwritable_field' ORDER BY wi.created_at DESC;
```

---

## 2. ⚠ THE FOUR DECISIONS. Everything below is blocked on 1; 2–4 are independent of it.

### DECISION 1 — may an automated finding cause a published page description to be REWRITTEN? (OWNER)

**This is the one that matters and it is not a lane's to take.**

*What is true today.* Nothing in the estate can rewrite a non-empty `pages.meta_description`.
`[MEASURED 2026-08-25]` every writer is create-or-fill-blank
(`site_db_actions.go:1235`, `apply_adoption_plan_action.go:84`, both
`COALESCE(NULLIF(EXCLUDED,''), existing)`) except `save_page_meta_description_action.go:211`, which is
reachable from one agent whose scheduled `pre_query` selects `COALESCE(meta_description,'')=''`.

*Why it is yours.* `bugs_open/320` §15 records you granting `overwrite_existing: true` for a **one-off**
681-page regeneration and **explicitly withholding it for the standing mechanism** — verified
afterwards that the seeded agent was left unarmed. A standing path that rewrites published copy on an
automated finding is precisely that withheld authority.

*The cost of leaving it.* The producer keeps generating these — **two in 33 minutes** on the first
night. They are now recorded honestly rather than closed green, but nothing repairs them, and the
capability-gap list grows without bound.

| option | what it buys | what it costs |
|---|---|---|
| **(a) leave it** | nothing to build; the record is honest | descriptions stay wrong for ever; the list grows; the analyser's work on this field is permanently wasted |
| **(b) scoped standing authority** — a work-item-driven meta-description writer, opt-in per site, with the 681-pass's guards (backup table, reversible by one UPDATE) | the findings become repairable; gate 1c gets a reachable negative control | automation edits published copy; needs its own council round and an RFC |
| **(c) authority ONLY over machine-written descriptions** (the backfiller's own products), never human-authored copy | most of the value, much less exposure | needs provenance on the column, which does not exist today — that is `bugs_open/403`'s territory |

**Recommendation: (c) if the provenance exists or is cheap, otherwise (b) scoped to one site as a
canary.** (a) is defensible but means accepting that one whole class of the analyser's output is
decoration.

### DECISION 2 — does the predicate vocabulary stay this narrow? (owner-level scoping; the other lane's seam)

An acceptance predicate may only grade `meta_description` or `title`
(`acceptancePredicateTextFields`). **Both are effectively unwritable on the audit-routed path.** So
*every* predicate the analyser can currently emit is doomed, and **gate 1c's negative control is
unreachable by construction** — which is now written into its `PromotionOwes`.

- **(a) widen the vocabulary** to a field some handler CAN write (section content is the obvious
  candidate) → predicates become satisfiable, gate 1c gets its control, promotion becomes possible.
  ⚠ `bugs_open/395` §8f/§5 notes body-text shapes are excluded today *because the page surface carries
  no content* — so this is not a one-line change.
- **(b) leave it narrow** → gate 1c stays unexercised indefinitely and its refusal arm stays a
  third instance of CLM-023's residual.

### DECISION 3 — build the staleness audit, or accept the residual? (a lane can do this; needs a call on priority)

The council's `constitution` seat's objection, **conceded not answered**: the roster is a NEGATIVE
capability claim, its staleness re-check is **prose**, and *"that is the exact enforcement gap
`320` §9 was filed about, now reproduced one layer down."* It is correct.

⚠ **If you build it, read `RFC_057` §3 FIRST.** An audit that greps config for the column name **runs
GREEN while wrong about two thirds of the writers** — `upsertPage` writes the column and is reached
via the action `sync_pages_to_db`; three live agents carry that step and one names the column.
**Resolve steps through the ACTION REGISTRY.** Building it the obvious way reproduces this
mechanism's blind spot inside the check meant to detect it.

`RFC_057` §4's position — which I think is right — is that this is a **condition of the roster
GROWING**, not a nice-to-have: one entry is a measurement; ten entries with no audit is a stale map
with an enforcement mechanism bolted on.

### DECISION 4 — `RFC_057`, already filed and awaiting your ruling

Two questions in it: whether `HandlerCanWriteField` as a new shared contract needs more than the
written contract note now in the file; and **§6 Q2 — whether the estate wants a seam for "these two
changes must land together" at all.** The second is the general form of a real problem: the council
objected that our cross-lane agreement was *a chat message*, which is the mitigation for the very
failure mode the founding incident describes. A build-time test naming one symbol fixed this instance
and does not generalise.

---

## 3. What the next session should do

1. **Nothing on rule 3b unless a decision above lands.** It is live, proven, and its residual is
   stated. Resist tidying it.
2. **If decision 1 goes (b) or (c):** the build is a work-item-driven route to
   `save_page_meta_description`, and `scripts/regen-meta-descriptions.sh` is the worked precedent for
   how the authority was handled last time (inline on a one-off dispatch, seeded agent left unarmed,
   full backup in `meta_description_pre_regen_20260821`). Council round required; `RFC_057`-shaped
   scope question first.
3. **If decision 3 goes "build":** `RFC_057` §3, action registry not config text, and build it ONCE so
   it covers both this roster and the other lane's emit-side stamp.
4. **Watch the capability-gap list rather than the gate-1c census.** The live query is in §1a. Two
   existing readers already consume it (`diagnose_triage_action.go:361`,
   `fixloop_digest_action.go:358`).
5. **Re-run the roster census before quoting it.** Both entries are dated `2026-08-25`:
   `git log --since=2026-08-25 --diff-filter=AM -- platform/orchestration/actions` — a non-empty
   result means re-measure before trusting the roster.

---

## 4. Watch-outs this thread paid for (newest first)

- **⚠ CLAUDE.md's deploy check does not exist.** `grep -m1 'build provenance'` matches nothing in any
  Go source here; its documented "not in range" note absorbs the real failure. Probe the CAPABILITY
  with a control on BOTH sides. `LANDMINES.md`.
- **⚠ A probe whose every result is "absent" means NOTHING.** Three candidate build shas, three
  absents, no PRESENT control — reads exactly like "did not ship". `WRONG_CALLS.md`.
- **⚠ "Can this agent write X?" is a GO question.** Config names the ACTION, not the effect. A
  column-name census over `agent_definitions` sees 1 of 3 writers. `LANDMINES.md`, `RFC_057` §3.
- **⚠ A demand control cannot tell you the instrument is pointed at the wrong thing.** I answered a
  council objection with a checkable query that could not have answered it. `WRONG_CALLS.md`.
- **⚠ A `workflow.steps` walk is a LOWER BOUND** — it misses steps nested in a loop's `sub_workflow`
  and returns a confident zero. It dropped `meta-description-backfiller` on this very question.
- **⚠ "Total over every route" is not a safety property.** My wrapper was total in the wrong direction
  too, re-stamping Rule 3's own `cta` gap and destroying its dedup key — same `item_type`, so it
  looked right. State the property as a BICONDITIONAL over the whole universe with both arms asserted
  non-empty. `WRONG_CALLS.md`.
- **⚠ A REJECTED predicate is a WRAPPER** (`{verdict, reason, predicate:{…}}`) — its `field` is one
  level down, and reading it flat yields `""` silently. **This was load-bearing on day one: 1 of the 2
  live catches came through that branch.**
- **⚠ Grep for the MECHANISM *and* for the COLUMN.** `320` and `395` are the same rows described in
  vocabularies that share nothing, so neither lane's "grep before you file" could find the other.
- **⚠ A snapshot read expires.** `bugs_open/395` grew 106 lines beneath me mid-session; I was minutes
  from proposing work that was already committed.

## 5. The four mutations, so a later session can re-run them

Copy the files aside first — **`git stash` is FORBIDDEN and hook-blocked on this tree.**

1. remove the `withUnwritableFieldGuard` call → the worked case routes at `page-build-handler` again.
2. drop the `["predicate"]` unwrapping → `predicateTargetField` returns `""`, guard goes silently inert.
3. add a third field to `acceptancePredicateTextFields` → the vocabulary lockstep fires.
4. remove the empty-handler early return → fails **naming Rule 3's own `cta` gap**, whose dedup key
   the guard would have destroyed. (This one found a real defect.)

## 6. Residuals, stated plainly

1. **Nothing repairs a wrong page description.** Decision 1. The findings are recorded and permanently
   unrepairable until it is taken.
2. **The staleness audit is named and unbuilt.** Decision 3. Until it exists this mechanism's guarantee
   rests on a human re-running the census — the thing it was built to disprove the value of.
3. **Rule 3b is a PARTIAL guard.** It fires only where a finding names its field mechanically. The
   **6,919** items `[MEASURED 2026-08-25, live UNION archive]` carrying a prose-only `acceptance_test`
   are untouched, because prose is undecidable at this seam.
4. **The feed-forward defect is real and unbuilt.** `[MEASURED 2026-08-25]` **0** live agents read
   `spec.acceptance_test` as an input (demand control: **3** read `spec.suggestion`); **6,281** items
   closed `complete` graded against a criterion the writer was never shown. The bug's owner said
   "build it, don't hold it". It is NOT the fix for 395 and will not produce gate 1c's control, but it
   is the fix for every criterion naming something a writer CAN change.
5. **Gate 1c's evidence stream for this population is gone, by design** — see §7.
6. **`filing_mode=record` currently masks rule 3b's practical effect for this producer** — see §7.

## 7. Two interactions a later reader will otherwise misdiagnose

- **An empty gate-1c census now means "the upstream guard got there first."** Rule 3b parks these
  findings before dispatch, so they never complete and gate 1c never grades them. `[MEASURED
  2026-08-26]` 0 graded, 0 permitted. Do **not** read that as "nothing refutes" or "gate 1c is off".
- **`filing_mode=record` (`RFC_056`, `c440d5c5e`) parks every finding this producer files anyway.** So
  rule 3b's effect is currently masked for `offer-analysis`; it becomes load-bearing again the moment
  record mode is off, or for any of the other **6** agents that reach `write_audit_findings`
  `[MEASURED 2026-08-25]`. ⚠ **The two parks are deliberately DISTINGUISHABLE and that is a safety
  property:** the documented `release_recipe` filters `spec->>'filing_mode'='record'`, so a rule-3b row
  **cannot** be released. Had they stamped rows identically, running the recipe would have
  reintroduced this bug wholesale. Never "tidy" the two shapes into one.

## 8. Who owns what nearby

- **`vigilant_designer_offer_analysis`** owns `bugs_open/395`, gate 1c (`WII-033`), `CLM-024`, the
  emit-side `field_writable` stamp and `RFC_057`. Coordinated in full; they verified every load-bearing
  claim here first-hand. Their register file is a **high-traffic shared file** with this work — narrow
  pathspecs only.
- **`bugs_open/320`'s lane** owns `pages.meta_description` and decision 1's subject matter. A CONTRIB
  is filed in `docs/agent_docs/docs024_key_docs_latest/meta_description_never_backfilled/`.
- **`bugs_open/375`'s lane** owns the claim-timeout exclusion migration (needed by gate 1c's promotion,
  **not** by rule 3b). They accepted ownership; the ordering asymmetry is settled in their handoff.
- **`bugs_open/345`'s lane** owns `retry_feedback`. They are putting a skip-empty guard to their user.
- ⚠ **HEAD carries one failing test that is nobody's here:** `TestFindingCodeScanEveryWriteIsRegistered`
  — `WORK_ITEM_STATUS_OVERRIDE_REFUSED` was added at `2b46afbe6` without its registry declaration.
  Verified as pre-existing and not caused by this work; the 396 lane owns the declaration choice.
