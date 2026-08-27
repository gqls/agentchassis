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
> `v1.0.1341` at 2026-08-25 23:11:52Z and ~~fired twice within 33 minutes~~ **[CORRECTED 2026-08-26 17:00Z: 25 firings across 19 sites — see §9]** on sites nobody had
> measured. TWO OF THE FOUR DECISIONS ARE NOW RULED (§2): the overwrite authority is GRANTED but only
> over machine-written descriptions, and the test vocabulary is to be WIDENED. ⚠ The first cannot be
> implemented yet — `pages` carries NO provenance, so the machine/human distinction the ruling rests
> on cannot be made by the system, and the marker that would make it is `bugs_open/403`'s
> (leopardess lane, active). Coordinating, deliberately not building. Decisions 3 and 4 are open.**
>
> **⚠ ADDED 2026-08-26 22:12Z, AND IT IS THE FIRST THING A NEW CHAT SHOULD READ.** Two things moved
> after the above was written. **(1) A FRESH CHASSIS ROLLED — `v1.0.1345`, pods up 20:25Z — and rule
> 3b is verified live on it with both controls; it has fired 32 times, 3 of them post-roll (§10).
> (2) THE ROSTER RULE 3b READS HAS A FALSE ENTRY (`title`), found by running §3.5's own staleness
> check (§9).** The defect is REAL, VERIFIED and LATENT: no shipped park is wrong, because all 32
> firings displaced handlers that genuinely cannot write the field. It is RECORDED, NOT BUILT — the
> map is a shared seam with another lane's caller. **§9 also changes decision 3's answer**: the
> staleness audit as filed would never have caught this, because this was not staleness.

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

## 2. ⚠ OWNER RULINGS, 2026-08-26 — READ THIS BEFORE §2.1; TWO OF THE FOUR ARE ANSWERED

> **The four decisions below are kept in full as originally put, because the ruling only makes sense
> beside the question. What changed is recorded here.**

### RULED — decision 1: **option (c).** An automated finding MAY cause a published page description to
be rewritten, **but only where the description was machine-written, never human-authored copy.**
The owner added, unprompted: ***"I haven't yet written any manually."***

⚠ **AND THE RULING CANNOT BE IMPLEMENTED AS STATED, YET — this is the live blocker.**
`[MEASURED 2026-08-26]` the `pages` table carries `meta_description` **and no provenance of any kind**
— no source column, no author, no stamp. So the distinction the owner drew **cannot currently be made
by the system**; it exists only in his sentence. With **838** live descriptions (and 40 blank) and
none hand-written, **option (c) today covers all 838** — it is not narrower in reach than granting the
authority outright. What it buys is protection from the moment anyone *does* write one by hand.

**So the build order is provenance FIRST, authority SECOND, gated on the stamp** — because the owner's
own standing rule is that a statement is not a control on a shared tree. This was put to him and he
said to proceed.

⚠⚠ **DO NOT BUILD THAT PROVENANCE HERE.** `bugs_open/403` (the **leopardess** lane, active
2026-08-26) is the same question — *"did a human or the machine write this value?"* — and its fix
candidate 1 proposes an `__authored` marker, the inverse of the existing `__cta_minted`
(`platform/orchestration/datahelpers/cta_provenance.go:57`). Building a second, differently-shaped
answer for `pages.meta_description` is precisely the drift the council objected to on this very
change. **Coordination message sent; awaiting their call on three things:** which direction the marker
takes, whether the convention covers a plain COLUMN as well as a key inside `content_data`, and
whether `cta_provenance.go` is the right home.

**The fork that matters, stated so whoever picks this up does not re-derive it:**

| direction | "safe to overwrite" means | fails how | consequence for decision 1 |
|---|---|---|---|
| mark the **MACHINE** (`__cta_minted` shape) | carries the machine mark | **fail-SAFE** — unmarked is treated as possibly-human and left alone | **nothing is overwritable today**, since all 838 are unmarked, until each has been rewritten once by a marked path |
| mark the **HUMAN** (403's `__authored`) | no human mark present | **fail-OPEN** — a human write that misses the mark is destroyed | works on day one, and matches "none are manual" exactly |

403 is a bug about authored values being **destroyed**, so that lane will likely want fail-safe; this
decision wants the other. Probably not a conflict — both markers with different meanings (machine-minted
*licences* an overwrite, human-authored *forbids* one, neither present = the 2026-08-02 default-OFF,
i.e. today's behaviour) — **but it is a decision, and it must be made once by the marker's owner.**

### ⚠ 2.1a — AND THE MARKER WOULD HAVE NO WRITER: option (c) is STRUCTURALLY option (b) today

Measured after leopardess ruled the marker design (§2.1b), before building anything. **There is no
interface through which a person can write a page description at all.**

`[MEASURED 2026-08-26]` **Demand control first, so the negative means something:** a human CAN edit
page rows — `internal/core-manager/admin/page_admin_handlers.go` has **four** `UPDATE pages` sites,
writing `suppressed_sections` and `page_spec`. So the admin surface exists and reaches this table.
**`meta_description` is not among the columns it writes: ZERO mentions of the column in the whole of
`internal/` and `frontends/`.** The only non-agent writer anywhere is `cmd/webdesignport/import.go`,
a port tool, and even that is `COALESCE(NULLIF(EXCLUDED,''), existing)` — fill-blank.

**So the owner's "I haven't yet written any manually" understates it: he could not have.** And the
consequence for his ruling is structural, not incidental:

- A `meta_description_authored_by` column would have **no producer**. Permanently NULL.
- An authority gating on `authored_by IS NULL` therefore **permits everything, for ever**.
- **Option (c) is not "currently as wide as (b)" — it IS (b), until a human-write path exists.**
- Building the marker now is a mechanism with no writer: the "built but never exercised" residual
  this estate keeps paying for (CLM-023's shape, three instances already on record).

⚠ **DO NOT BUILD THE HUMAN MARKER FOR THIS COLUMN YET.** What actually protects human copy today is
**the absence of the capability to write it**, not a mark. That is a real guarantee and it is the same
KIND of guarantee rule 3b already reasons about — and it goes stale the same way, **by addition**.

**So the thing to record now is a CONDITION, not a mechanism:** whoever builds a human-facing editor
for `pages.meta_description` must add the `__authored`-equivalent mark **in the same change that ships
the editor** — the 2026-07-28 condition (2) discipline, register the seam in the commit that creates
it. Until then the overwrite authority's guarantee rests on a dated absence, which must be re-checked
with the control above rather than assumed.

**What IS worth building, and is leopardess's own recommendation:** the MACHINE mint stamp, which has
a real writer today (`save_page_meta_description_action.go:211`). It accumulates a positive record
from its first write, costs nothing, and is the only thing that keeps a future tightening to
fail-SAFE available — without it the estate can never move off "unmarked means overwritable".

**⚠ CITE THE FILE, NOT THIS HANDOFF, AND NOT A CHAT MESSAGE.** The marker design and the no-writer caveat are both recorded in `bugs_open/403` by its owning lane (`0049b10d9` — the ruling; `24cc44ed1` — the caveat, credited to this measurement). That was deliberate on both sides: this whole thread exists because a rule that lived only as prose in one lane's file did not bind on another, and a rule that lives only in a cross-session message is strictly worse than that. The three-state vocabulary (`__authored` FORBIDS · minted LICENSES · neither = today's behaviour) is 403's to define; this handoff is a consumer of it.

### RULED — decision 2: **widen it.** The owner's words: *"it's ok for the tests to read real content
but not to write it."*

That constraint is already satisfied by construction — a predicate IS a read-only condition with no
power to change anything — so what he is approving is what a test may **READ**. Relayed to the
`vigilant_designer_offer_analysis` lane, whose seam it is (`acceptancePredicateTextFields`,
`features_open/030` §10 v2(a), the bounded head-of-hero excerpt).

⚠ **Why this is the unblocker for gate 1c and not merely a widening:** `meta_description` and `title`
are both unwritable on the audit-routed path, so *every* predicate that producer can currently emit is
doomed at birth. **Page body content is different — `page-content-writer` can actually write it** — so
a predicate over body content is the first one that could be SATISFIED after a fix, i.e. the first
route to `outcome='permitted'` and to promoting gate 1c from recording to refusing.

⚠ **AND IT WILL BREAK THIS LANE'S BUILD, BY DESIGN.** `TestPageFieldWritersCoversThePredicateVocabulary`
fails the moment a third field joins `acceptancePredicateTextFields`. Two ways to satisfy it, and the
choice was offered to that lane: add the new field to `pageFieldWriters` with
`WritableBy: {"page-build-handler": true}` and its dated measurement, or relax the lockstep to exempt
writable fields. **Prefer the first** — it keeps every field in the vocabulary carrying a dated
statement of who can write it, which is the property that made rule 3b possible at all.

### ⚠ 2.5 — TWO THINGS NOW WAIT ON THE OWNER **DIRECTLY**, AND A RELAY WILL NOT DO

Both rulings above reached the lanes that own the work **through this session**. One of those lanes has
correctly refused to act on that, and it is worth understanding rather than routing around.

**Decision 2 is RULED but STALLED.** The `vigilant_designer_offer_analysis` lane's reply, recorded
verbatim because the reasoning is right: *"A peer relay is not the owner's approval for me to open a
new piece of work of that size, however well-founded, so I have surfaced it to him as a decision to
take rather than acting on it. That is a constraint on me, not a doubt about your relay."*

v2(a) is substantial — config-only but migration `602` is unwritten, it GROWS the offer surface, it
widens what a predicate can address, and the truncation check needs re-running on webdesign.co.uk
afterwards. **So the owner's "go ahead" needs to reach that lane from him, not from here.** Until it
does, decision 2 is ruled and unstarted, and gate 1c's route to a negative control stays closed.

**Decision 1 is RULED and BLOCKED on another lane's design** — `bugs_open/403`'s marker (§2 above).
Not an approval problem; a dependency. Coordination sent, awaiting their call.

**Their answer to the lockstep question, so nobody re-asks it:** take option 1 — when the vocabulary
widens, add the new field to `pageFieldWriters` with `WritableBy: {"page-build-handler": true}` and its
dated writer census. Their reasoning matched this lane's: option 2 trades a coverage guarantee over
exactly the population the roster exists to describe, for the cost of writing down one measurement.

⚠ **AND THE TRAP THEY SPOTTED IN MY OWN TEST, which I had not stated:** *do not add the roster entry
speculatively AHEAD of the widening.* `TestPageFieldWritersCoversThePredicateVocabulary` is
**bidirectional** — a roster entry naming a field no predicate can name fails it in the other
direction. The two edits must land together, vocabulary first or same commit.

---

### STILL OPEN — decisions 3 and 4. The owner asked for a plainer explanation of both and it was given;
no ruling yet. §2.3 and §2.4 below are the questions as put. For 3 the ask is narrow: agreement to
**defer with a trigger** (build the audit only when the roster gains a third entry) rather than a
yes/no on building it now.

---

## 2. ⚠ THE FOUR DECISIONS. Everything below is blocked on 1; 2–4 are independent of it.

### DECISION 1 — may an automated finding cause a published page description to be REWRITTEN? (OWNER)

**This is the one that matters and it is not a lane's to take.**

*What is true today.* Nothing in the estate can rewrite a non-empty `pages.meta_description`.
`[MEASURED 2026-08-25]` every writer is create-or-fill-blank
(`site_db_actions.go:1235`, `apply_adoption_plan_action.go:84`, both
`COALESCE(NULLIF(EXCLUDED,''), existing)`) except `save_page_meta_description_action.go:211`, which is
reachable from one agent whose scheduled `pre_query` selects `COALESCE(meta_description,'')=''`.
> **⚠ CORRECTED 2026-08-27 — "the only unconditional UPDATE" is WRONG, and the correction makes this
> decision CHEAPER, not harder. See §12.** That statement is imprecise here, and in the roster's own
> `Why` string in shipped code. The write is gated by an **opt-in `overwrite_existing` field whose
> default is FALSE**, enforced inside the WHERE clause. There are **two guards in series** — that
> flag AND the agent's `pre_query` — not one.

*Why it is yours.* `bugs_open/320` §15 records you granting `overwrite_existing: true` for a **one-off**
681-page regeneration and **explicitly withholding it for the standing mechanism** — verified
afterwards that the seeded agent was left unarmed. A standing path that rewrites published copy on an
automated finding is precisely that withheld authority.

*The cost of leaving it.* The producer keeps generating these — ~~**two in 33 minutes**~~ **25 in 18 hours [MEASURED 2026-08-26 17:00Z]** on the first
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

> **⚠ UPDATED 2026-08-26 BY §9 — and the update CHANGES WHAT TO BUILD, so read it before costing this.**
> The roster has **two** entries and one of them is **already false** (`title`; §9). So the argument
> above is no longer a prediction about ten entries — the map is wrong at N=2, on the day it shipped.
>
> **But the audit as specified here would NOT have caught it.** A staleness audit answers *"has a
> writer been ADDED since the Measured date?"* The `title` entry was incomplete **at the moment it
> was written** — its missing writers were three weeks old — so a `--since` check returns clean and
> the entry stays false for ever. I found it by hand, and only because I re-read the code rather
> than re-running the check.
>
> **So decision 3's question is now two questions, and the second is cheaper AND catches more:**
> **(i)** the staleness audit as filed (writers added since `Measured`), and **(ii)** a
> **totality** check — every handler `classifyFindingRoute` can emit must carry an explicit
> per-handler verdict in `WritableBy`, so a handler the router gains cannot inherit a silent "no".
> (ii) is a build-time test over two greps, needs no cluster, and would have failed on day one.
> **If only one gets built, build (ii).**

### DECISION 4 — `RFC_057`, already filed and awaiting your ruling

Two questions in it: whether `HandlerCanWriteField` as a new shared contract needs more than the
written contract note now in the file; and **§6 Q2 — whether the estate wants a seam for "these two
changes must land together" at all.** The second is the general form of a real problem: the council
objected that our cross-lane agreement was *a chat message*, which is the mitigation for the very
failure mode the founding incident describes. A build-time test naming one symbol fixed this instance
and does not generalise.

---

## 3. What the next session should do

1. ~~**Nothing on rule 3b unless a decision above lands.**~~ **AMENDED 2026-08-26 by §9.** Rule 3b's
   BEHAVIOUR still needs no work — it is live, proven post-roll, and no shipped park is wrong. But
   its ROSTER has a false entry, so "resist tidying it" no longer covers the whole file. The open
   piece is §9f's fix (make the roster TOTAL over the handler universe + a test that would have
   caught this), and it wants the `vigilant_designer_offer_analysis` lane's agreement first because
   their emit gate reads the same map. **Do not add `content-gap-planner` as a one-line patch** —
   that fixes the instance and leaves the silent default that produced it.
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
   ✅ **DONE 2026-08-26: it returned 43 commits, the re-measure was carried out, and it found §9.**
   ⚠ **But note what it did NOT find, because this is the useful half:** the defect it surfaced was
   an ORIGINAL OMISSION, not an addition — this check cannot see that class, and a clean result from
   it must never be read as "the roster is sound". That is exactly why §9f's totality test is the
   piece worth building.

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
- ~~⚠ **HEAD carries one failing test that is nobody's here:** `TestFindingCodeScanEveryWriteIsRegistered`~~
  **[RESOLVED 2026-08-26 17:00Z — verified PASSING at HEAD with `-v`, so it genuinely ran and is not a `-run`-matched-nothing vacuous ok. The 396 lane declared it. Original note kept below for the trail.]**
- ⚠ `TestFindingCodeScanEveryWriteIsRegistered`
  — `WORK_ITEM_STATUS_OVERRIDE_REFUSED` was added at `2b46afbe6` without its registry declaration.
  Verified as pre-existing and not caused by this work; the 396 lane owns the declaration choice.

---

## 9. ⚠ FOUND ON RE-CENSUS, 2026-08-26 — the roster's `title` entry is FALSE for one of the four handlers

**Found by doing what §3.5 tells the next session to do, and it fired.** `git log --since=2026-08-25
--diff-filter=AM -- platform/orchestration/actions` returns **43 commits**, so the roster was due a
re-measure before anything quoted it. Re-measuring found a defect that is not staleness at all — the
`title` entry was **incomplete on the day it was written**.

### 9a. What the thing is, before what is wrong with it

`pageFieldWriters` (`platform/orchestration/actions/write_audit_findings_field_capability.go:141`) is
a **roster of negative capability claims**: for each field a finding's acceptance criterion may name,
it states which handler agents can actually WRITE that field. Rule 3b reads it, and when the answer is
"none of them can", it parks the finding as a `capability_gap` instead of routing it at a handler that
would rebuild the page, report success and close green with the field untouched.

The rule the roster must satisfy is its own doc comment: *"this entry starts refusing findings that
have become fixable, and nothing here can notice."* A wrong **negative** claim does not fail loudly —
it silently parks work that could have been done.

### 9b. The claim, and the measurement that refutes it

The entry reads `"title": {WritableBy: map[string]bool{}}` — **no handler can write it** — licensed by:

> *"pages.title has one UPDATE writer, `apply_gap_plan_action.go:652`, which is reached from the
> gap-plan path and **from no audit-routed handler**"*

**Both halves of that clause name the same agent.** `[MEASURED 2026-08-26]`:

1. `classifyFindingRoute` routes findings at **`content-gap-planner`** — `write_audit_findings_action.go:696`
   (Rule 5: category gap/content/differentiation/structure) and `:712` (Rule 6: tone/content_rewrite on
   a placeholder page). So content-gap-planner **is** an audit-routed handler.
2. Its live workflow **carries `apply_gap_plan` as a real step** — resolved nesting-safe through the
   action key, not a config-text `LIKE`, per `RFC_057` §3:
   ```sql
   SELECT jsonb_path_query_array(default_config,'strict $.**.action ? (@ != null)')
     FROM agent_definitions WHERE type='content-gap-planner' AND is_active
      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
   -- apply_gap_plan, complete_workflow, ensure_site_record, execute_llm_prompt,
   -- load_site_pages, query_database, read_site_spec
   ```
3. `apply_gap_plan_action.go:652` is a **bare unconditional overwrite** — `UPDATE pages SET title = $3,
   sections = $4::jsonb, updated_at = NOW()` — with no COALESCE/NULLIF guard.
4. It is a live, completing route: **989 `needs_content_planning` items complete at
   `content-gap-planner`** (live UNION archive).

**So `HandlerCanWriteField("content-gap-planner","title")` returns false and the true answer is true.**

### 9c. The census behind it was also incomplete — 1 writer named, 5 exist

`pages.title` has **five** UPDATE writers as of 2026-08-26, not the one the entry names. The other four
are one helper reached from four call sites: `UpsertPageForRole`'s `Refresh` list →
`updatePageColumns` (`page_role_upsert.go`), which emits a bare `UPDATE pages SET title = $n`:

| call site | `Refresh` |
|---|---|
| `create_report_page_action.go:178` | `url, title, sections` |
| `deploy_tool_action.go:464` | `url, title, sections` |
| `deploy_tool_action.go:636` | `title` |
| `create_tool_component_action.go:653` | `title` |

⚠ **This is NOT stale-by-addition** — `page_role_upsert.go` was born **2026-08-02**, three weeks before
the census. It was missed. The `meta_description` entry directly above it did the action-registry
census properly and carries the whole argument in a comment block; the `title` entry got a one-line
code read and no config census at all. **The two entries in one map were measured to different
standards, and only the shape tests ran over both.**

### 9d. Why it has not bitten — stated with the control, because a bare zero here proves nothing

**Latent, not live.** `[MEASURED 2026-08-26 17:00Z, live UNION archive, predicate era only]`:

| item_type | with a predicate | filed since 2026-08-25 |
|---|---|---|
| `content_rewrite` (→ page-build-handler) | **6** | 411 |
| `needs_content_planning` (→ **content-gap-planner**) | **0** | 152 |
| `needs_copy_edit` (→ copy-editor) | 0 | 16 |

The `content_rewrite` row is the **demand control**: predicates do get stamped on routed findings, so
the zero on the content-gap-planner route is a real absence and not a dead instrument. Restricting to
the predicate era matters — all-history reads `0 of 1,177`, which is vacuous, since predicates are
days old.

A second, independent narrowing: Rules 5 and 6 set `PageName: ""` and `PageID: nil`, so a finding
routed at content-gap-planner **names no page** — a `title` predicate has nothing to evaluate against.
The wrong answer is real; the path to it is doubly narrow.

### 9e. What is NOT wrong — the 25 firings are all correct

Every firing so far displaced `page-build-handler` (23) or `copy-editor` (2). Step-level census of
both, same method as above:

- `page-build-handler` → `call_agent, complete_workflow, conditional, ensure_site_record,
  fail_work_item, load_current_section_content, load_existing_content, load_page_record,
  load_page_sections_from_spec, plan_sections, save_page_sections, spawn_agent,
  update_work_item_status, validate_page_content` — no `pages` scalar-column writer.
- `copy-editor` → `checkpoint_for_review, complete_workflow, conditional_branch, ensure_site_record,
  execute_llm_prompt, query_database` — no page writer at all.
- `spec-updater` → `update_site_spec_from_item` only.

**So no live park is wrong, and rule 3b's shipped behaviour to date is sound.** The defect is a false
entry waiting for a route that has not yet carried a predicate.

~~⚠ **One caveat I did not close:** `page-build-handler` carries `call_agent` and `spawn_agent`. A step
list therefore bounds what a handler does *itself*, not what it can cause. The roster's method — and
this section's — inherits that gap.~~

✅ **CAVEAT CLOSED 2026-08-27 — the full spawn closure is clean, so the 23 page-build-handler parks
are sound at every depth, not just at depth 1.** Prompted by the `bugs_open/414` lane's sub_workflow
measurement (§11f). Every hop resolved with `jsonb_path_query_array($.**.agent_type)` and
`$.**.action`, i.e. recursive descent, so a target nested in a loop's `sub_workflow` cannot hide from
it — which is exactly the trap §4 already warns about and the reason a `workflow.steps` walk was not
used:

| depth | agent | can it write `pages.title` / `meta_description`? |
|---|---|---|
| 1 | `page-build-handler` | no — 14 actions, none touches those columns |
| 2 | `page-content-writer` | no — 16 actions (incl. `loop`); no `sync_pages_to_db` / `save_page_meta_description` / `apply_gap_plan` / `create_*` / `deploy_tool_*` |
| 2 | `page-rerender` | no — `rerender_single_page` READS `meta_description` (`:529`, a SELECT); every `UPDATE pages` in `update_page_status` sets only `build_status` / `built_from_plan_version` / `updated_at` |
| 3 | `internal-link-resolver` | no — `resolve_internal_links`, zero `UPDATE pages` |
| 3 | `research-agent` | no — `InsertResearchResultAction` scoped to its own 225 lines: zero `UPDATE pages`, no mention of either column |
| 4 | — | neither depth-3 agent spawns anything (`$.**.agent_type` → `[]`) |

⚠ **One measurement was loose on the way and is worth recording:** `grep -c 'UPDATE pages'` on
`v3_site_actions.go` returned **3** for `InsertResearchResultAction` — a FILE-wide count attributed to
one function in a 6,000-line file. Scoped to the function's own 225 lines it is **0**. A count taken
at file granularity and reported at function granularity is the same error class as everything else in
this section.

### 9f. The fix, and why it is NOT applied here

The fix is not "add content-gap-planner to the map". The defect is that **absence means "cannot
write"** — a silent default — so a handler the router gains is unmeasured and reads as incapable.
Ordered by what closes the door (not by effort):

1. **Make the roster total over the handler universe.** Every handler `classifyFindingRoute` can emit
   (`page-build-handler`, `copy-editor`, `content-gap-planner`, `spec-updater`, plus `designRouting`'s
   `webdesign-agent`, `component-template-fixer`, `site-component-linker`, `css-patch-agent`) must
   appear in `WritableBy` with an explicit `true`/`false` and its dated measurement. A handler added to
   the router then fails a test instead of inheriting a silent "no".
2. **A test that would have caught this one.** The existing suite checks the roster's SHAPE only —
   vocabulary lockstep, a `[MEASURED` marker, a `Measured` date. Nothing checks its CONTENT against the
   router. Assert the universe is enumerated, and mutation-prove it by deleting one handler.
3. Correct the `title` entry's `Why` to the five-writer census, and re-check `meta_description`'s
   entry against the same handler universe.

**Not applied, deliberately, and this is the same reasoning §2.5 records the other lane using on me:**

- `HandlerCanWriteField` is a **shared seam with two callers** — this routing guard and the other
  lane's emit gate (`CLM-024`). Changing `WritableBy` changes what THEY stamp at source. The
  2026-07-29 ruling (3) says a shared mechanism's other consumers must be **told, not merely
  measured**, and `RFC_057` is the open question about this very contract.
- It is a platform-code change to a live guard → council gate before or alongside the commit.
- It is latent, so nothing is burning, and the handoff's §3.1 says do not tidy rule 3b unopened.

**So: recorded, not built.** Whoever takes it owns telling the `vigilant_designer_offer_analysis` lane
first, because the emit gate reads the same map.

### 9g. The transferable lesson

**Two entries in one roster, written in one commit, can be measured to different standards — and every
test over them was a SHAPE test, so both passed identically.** `[MEASURED` markers, dates and a
lockstep test all fired correctly on an entry whose content was wrong. A marker proves a measurement
was claimed, never that it was complete. Belongs in `WRONG_CALLS.md`; the "resolve handler capability
through the action registry, never a config-text search" half is already a `LANDMINES.md` entry, and
this is the second instance of it — **inside the file that documents it**.

---

## 10. POST-ROLL VERIFICATION — `v1.0.1345`, measured 2026-08-26 22:12Z

**A fresh chassis rolled after §9 was written.** Nothing in this lane was waiting on it (rule 3b has
been live since `v1.0.1341`), but a roll changes what every measurement in this file means, so all of
it was re-taken rather than carried forward. **Everything below re-measured. Nothing inherited.**

### 10a. What is running, proven at the artefact with BOTH controls

```
agent-chassis-5864bf97c5-5l8xd   v1.0.1345   started 2026-08-26T20:25:20Z   ready
agent-chassis-5864bf97c5-68t5h   v1.0.1345   started 2026-08-26T20:24:56Z   ready
```

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system exec "$POD" -- grep -aq 'no_writer_for_page_field'  /proc/1/exe  # PRESENT — rule 3b
kubectl -n ai-persona-system exec "$POD" -- grep -aq 'handler_reported_no_change' /proc/1/exe  # PRESENT — positive control
kubectl -n ai-persona-system exec "$POD" -- grep -aq 'zzz_invented_control_395'   /proc/1/exe  # absent  — negative control
kubectl -n ai-persona-system exec "$POD" -- grep -aq 'model_opinion'              /proc/1/exe  # PRESENT — dates the build
```

All four ran. The fourth is the one worth keeping: `model_opinion` is `ffa1707b3`'s literal, committed
**after** rule 3b shipped, so its presence dates this binary **forward of `v1.0.1341`** rather than
merely "carrying rule 3b" — which a cached same-tag image would also have done.

⚠ **A probe run may time out part-way and still be sound — but only if you say N out loud.** The first
attempt returned 3 of its 4 answers before the harness timeout. Three is a complete three-way probe
(claim + both controls) and was reported as such; the fourth was re-run alone. **What is NOT allowed
is quietly reporting three when you asked four.**

⚠ Still do **not** use CLAUDE.md's `grep -m1 'build provenance'` — that string is in no Go source in
this repo (`LANDMINES.md`).

### 10b. Rule 3b post-roll: firing, and still correct

**32 firings, 3 of them since the roll** — so the guard is live and working on the new binary, not
merely present in it.

| field | displaced handler | firings | latest |
|---|---|---|---|
| `meta_description` | `page-build-handler` | 23 | 2026-08-26 21:50:52Z |
| `title` | `page-build-handler` | 7 | 2026-08-26 14:14:41Z |
| `title` | `copy-editor` | 1 | 2026-08-26 12:13:21Z |
| `meta_description` | `copy-editor` | 1 | 2026-08-26 11:35:52Z |

**`content-gap-planner` still appears ZERO times** — which is what keeps §9's defect latent.

### 10c. The handler universe did NOT move across the roll — so §9 stands exactly as written

The obvious post-roll risk is that a router change widened the handler set and turned §9's latent
defect live. It did not. `write_audit_findings_action.go` took 2 commits in the window (`ffa1707b3`,
`7dd64550c` — both record-mode/origin work), and the handler universe is **byte-identical** between
rule 3b's ship and HEAD:

```bash
git show f4aa19ae7:platform/orchestration/actions/write_audit_findings_action.go \
  | grep -oE 'HandlerAgent:[[:space:]]*"[^"]*"' | sort | uniq -c   # compare with the same over HEAD
```
Both sides: `"" ×2, content-gap-planner ×2, copy-editor ×1, page-build-handler ×2, spec-updater ×1`,
and `designRouting` identical. The roster file, the predicate vocabulary, `apply_gap_plan_action.go`
and `page_role_upsert.go` took **zero** commits in the window. All three of §9's load-bearing lines
re-asserted present at HEAD.

### 10d. §9's latency control, re-run — the zero held while the denominator grew

`[MEASURED 2026-08-26 22:12Z, live UNION archive, predicate era]`

| item_type | with a predicate | filed since 2026-08-25 | was (17:00Z) |
|---|---|---|---|
| `content_rewrite` (→ page-build-handler) | **6** | 557 | 6 of 411 |
| `needs_content_planning` (→ **content-gap-planner**) | **0** | 246 | 0 of 152 |
| `needs_copy_edit` (→ copy-editor) | 0 | 23 | 0 of 16 |

**The denominator grew by 94 in five hours and the zero held.** That is a stronger result than the
morning's, and it is stronger *because* it could have come out otherwise: 94 more chances to produce
the first counter-example, and none did. The `content_rewrite` row remains the demand control — 6
predicates prove the stamping mechanism works, so the zero is a real absence and not a dead instrument.

### 10e. What this does and does not license

- **Licensed:** rule 3b is live, firing and correct on `v1.0.1345`; §9's defect is unchanged by the
  roll and remains latent; no shipped park is wrong.
- **NOT licensed:** any claim that the roster is now safe. Nothing was fixed. The latency is a
  property of what the analyser happens to emit, not a guarantee — the day a `needs_content_planning`
  finding carries a `title` predicate, rule 3b files a `capability_gap` that names a handler which
  could have done the work. **There is no alert on that transition.** The cheapest tripwire, if
  someone wants one before §9f is built, is the §10b table: a `content-gap-planner` row appearing in
  the `would_have_routed_at` column is the signal.

### 10f. The landmine verifier dispatch reported FAILURE on a message that LANDED — do not retry it

`./scripts/landmines-verify-dispatch.sh` synced the new §9 landmine (4,799 owned rows present) and then
printed, for correlation `1a358ed1-8dba-4f52-bd07-316ec0e8a366`:

```
RECEIPT INDETERMINATE  topic=system.agent.generic.requests
VERIFICATION NOT DISPATCHED — no verdict will ever arrive for this run.
Dispatched 1, 1 failed to publish.
```

**It had published.** The script's own instruction is the right one and it was followed —
`kafka_verify_landing 1a358ed1-…` returns `LANDED … EXECUTING_STEP|spawn_verifier`. The run was
already executing while the dispatcher was reporting it lost.

⚠ **"1 failed to publish" is the strongest possible invitation to re-run the command, and re-running
would have duplicated a live orchestration.** The script is behaving correctly — it refuses to claim a
publish it cannot prove, which is the estate's `a-receipt-nobody-asserts-on-is-a-log-line` rule working
as designed. The failure mode to guard is the READER's: **an unproven publish is not a failed publish.**
Always run `kafka_verify_landing <corr>` before concluding anything, in either direction.

---

## 11. CROSS-LANE, 2026-08-27 — a red tombstone guard at HEAD, fixed here rather than routed

**Not this lane's work and recorded anyway, because the next session in this file will see the commit
and wonder why it is here.** A peer session (the `bugs_open/414` lane) reported that
`go test ./platform/orchestration/datahelpers/` had been **red at committed HEAD, fleet-wide**, since
`bc8167100` (2026-08-26 20:44). They could not identify the owning lane and asked either for a fix or
for the owner's name.

### 11a. What it was

`TestNoHandSpelledTombstonePredicate` walks `platform/orchestration` and fails on any non-comment
hand-spelling of the tombstone clause. `component_hierarchy_walk.go:397` hand-spelled it inside
`hierarchyChildrenOf`. Fixed at `8cf0c2f59` by using `datahelpers.NotRemovedSQL`.

**The peer's diagnosis was correct in full** — reproduced first-hand before touching anything, which
is the rule for a peer report as much as for a bug file.

**Two things they had slightly wrong, both worth stating because both are the interesting part:**

1. They suggested `datahelpers.NotRemoved("pc")`. **The bare constant is correct here** — the query is
   single-table `FROM page_components`, so there is no alias to qualify and `NotRemoved("pc")` would
   have emitted invalid SQL.
2. They called it "someone else's guard-protected predicate" and declined to touch it on that basis.
   Reasonable, but the equivalence is *checkable*: the hand-spelling was **already the NULL-safe form**
   and `NotRemovedSQL` is the **byte-identical string**, so the emitted SQL does not change. There was
   no design decision to take on the owning lane's behalf — which is what made fixing it the smaller
   act, not the larger one.

### 11b. Why fixed here rather than routed onward

The owning lane is `features_open/035` (`editorial_design_uplift`), which took that file through three
council rounds. Ordinarily: contribute, do not compete. But the peer's own argument is right and
decides it — **that test is the estate's only mechanism stopping the tombstone clause drifting from the
assembler's, and a red guard that nobody present caused is exactly how a guard stops being read.**
Every session running the datahelpers suite for its own change was getting a failure it did not cause.

CONTRIB filed into that lane's NOTES (`b655d76ba`) stating plainly that no decision of theirs was
taken and why the bare constant is right.

### 11c. ⚠ THE MEASUREMENT LESSON, and it is the reusable half: HEAD IS NOT GREEN, so "the tests pass" proves nothing here

`scripts/verify-head-builds.sh --with <file> --test` came back **FAILED**, and the first instinct — that
my one-line change broke something — was wrong. **The control run is what settles it:**

| run | packages failing |
|---|---|
| plain HEAD (`--test`, no `--with`) | **14** |
| HEAD + this one file | **13** |

- **FIXED: exactly `platform/orchestration/datahelpers`.**
- **INTRODUCED: nothing.**
- The other 13 are pre-existing at HEAD and none belongs to this thread — mostly integration/e2e
  suites wanting a live Kafka or database, plus `cmd/config-key-audit`, `platform/livespec` and
  `test/unit/actions`, which look real and are unowned here.

⚠ **The diff was wrong the first time and looked plausible.** `comm` over the raw `FAIL` lines reported
seven packages both "fixed" and "introduced" — because the line carries a **duration** (`0.425s` vs
`0.471s`), so identical failures differ as strings. **Diff on package NAMES:**
`awk '/^FAIL\t/{print $2}' … | sort -u`. A per-run varying field inside the key silently doubles your
result set, and the shape it produces — a long symmetric fixed/introduced list — reads like a real
regression.

⚠ **And `gofmt -l` EXITS 0 WHILE LISTING FILES**, so `gofmt -l <file> && echo clean` prints "clean"
about a file that needs formatting. It did, here. **Empty output is the signal, not the exit code:**
`out=$(gofmt -l <file>); [ -z "$out" ]`.

### 11d. What this cost and what it did not

One line of Go, plus the control run that made the claim honest. **It changes no SQL, adds no
capability, and touches nothing this lane owns.** It is recorded here only so the commit in this
lane's history has an explanation, and because §11c's two instrument traps are worth more than the
fix was.

### 11e. Two follow-ups from the same exchange, both fleet-wide rather than lane state

**1. The §9 landmine's verifier verdict is `NEEDS_HUMAN_REVIEW` and that is INDEX STALENESS.** Verdict
2026-08-26 22:19Z: four of five footprint items returned 0 rows because the code index was pinned to
commit `a7f76fa8` of **2026-08-25 19:17 UTC** — before the entry, and before
`write_audit_findings_field_capability.go` was in the index. Only `acceptancePredicateTextFields`
resolved, and it agreed. **A zero from a stale index is not an absence.** The entry now carries that
note plus the first-hand review it asked for, line by line, so nobody reads the verdict as doubt about
the claims. (It also confirms §10f: the verifier ran **four minutes after** the dispatch the script
reported as "failed to publish".)

**2. `who-owns.py NNN` answers for the BUG namespace only** — new `LANDMINES.md` entry. Resolving the
owner of `bc8167100` exposed it: `BUG_DIRS = ["bugs_open", "bugs_closed"]` (`who-owns.py:53`) is the
entire search space, so **`features_open/` is structurally invisible**, while the subject search matches
the bare number every other namespace also uses. `who-owns.py 035` returned an unrelated
`bugs_closed/035` and a commit list mixing `features_open/035` with **this lane's own `WII-035`**, and
never named the owning lane. Worse than the bare-number ambiguity CLAUDE.md documents, because that
warning is about two *bugs* sharing a number and its remedy (resolve by slug) presumes you are in the
bug namespace at all. **Resolve an owner from the commit's touched paths, never from the number.**

⚠ **Filed in `LANDMINES.md`, deliberately NOT in CLAUDE.md.** The `bugs_open/414` lane suggested
putting it "where who-owns.py's users will meet it", which was a good suggestion and is what the
landmine footprint does — but a peer's suggestion is not authority to edit the standing instructions,
and CLAUDE.md is the one file where that distinction has to hold absolutely.

### 11f. The `sub_workflow` blind spot now has a NUMBER, from the `bugs_open/414` lane — and it is the worst possible one

§4 has warned since 2026-08-25 that *"a `workflow.steps` walk is a LOWER BOUND — it misses steps
nested in a loop's `sub_workflow` and returns a confident zero"*, on the evidence that it dropped
`meta-description-backfiller` on this lane's own question. That entry had no count. It has one now,
measured independently by the `bugs_open/414` lane on a different question (a fleet-wide spec-surface
census), and relayed here:

> **Exactly ONE live agent has a `site_specs` reference nested inside a `sub_workflow` — and it is
> `page-content-writer`, whose refs all live in its process-sections loop.**

**A step-walking implementation would have gone blind to the single most load-bearing prompt in the
estate while reporting a clean fleet.** That is the sharpest available statement of why this trap
matters: the population it hides is not a random 1-in-N, it is the one that carries the most weight,
because *deep nesting is what a heavily-developed agent looks like*. A blind spot correlated with
importance is not a sampling error.

Their census avoids it by regexing `default_config` as one document rather than walking steps; this
lane avoids it by resolving through `jsonb_path_query_array($.**.action)`, which is recursive descent.
**Both work. `workflow.steps` iteration does not.** Their finding is pinned by a test whose failure
message names `platform/validation.WalkSteps` for whoever converts it.

⚠ **Consequence for this lane, and it is why the relay was worth acting on rather than filing:** the
`page-build-handler` spawn closure in §9e goes through **`page-content-writer`** — the exact agent
whose steps are nested. Had that closure been walked with `workflow.steps` it would have returned a
clean, confident, wrong answer about the one hop that mattered. It was not; it used `$.**`. But the
margin was a method choice made for a different reason, which is the kind of near-miss worth writing
down.

---

## 12. ⚠ CORRECTION 2026-08-27 — the overwrite authority decision 1 is about ALREADY EXISTS as an opt-in field, default OFF

**Found by accident, which is worth saying: another session had
`save_page_meta_description_action.go` dirty, I looked at the diff to check it did not move the line
number this file cites (it did not — a blank-line removal at `:283`), and read the surrounding SQL
while I was there.** The MEMORY rule "a DIRTY file's symbol is someone's PLAN" says look; what it does
not say is that looking will correct your own claim.

### 12a. What this file said, and what the code says

This handoff (§2) and — more importantly — **the roster's own `Why` string in shipped code** both call
`save_page_meta_description_action.go:211` *"the only **unconditional** UPDATE"*. It is not
unconditional:

```go
overwrite := datahelpers.GetBoolField(config, "overwrite_existing", false)
// One statement, so the decision cannot race a concurrent writer: the WHERE
// clause carries the overwrite policy rather than a read-then-write in Go.
const q = `
    UPDATE pages
    SET meta_description = $2, updated_at = NOW()
    WHERE id = $1
      AND ($3::bool OR COALESCE(meta_description, '') = '')
    RETURNING id`
```

**Two guards in series, not one:** the `overwrite_existing` flag (default **false**) AND the
backfiller's `pre_query` restricting to blanks. §2 credited only the second. `sql.ErrNoRows` is
handled as a *refusal* — `{"updated": false, "reason": "already_has_description"}` — not an error, so
a caller that forgets the flag gets a clean no-op rather than a surprise write.

### 12b. Why this makes DECISION 1 cheaper, and changes what it is a decision ABOUT

§2's option table costs (b) and (c) as *"automation edits published copy; needs its own council round
and an RFC"* — i.e. as though the overwrite capability must be BUILT. **It is already built, and built
in exactly the shape the estate's own ruling prescribes for this situation:** the 2026-08-02 owner
ruling §2 says new authority on a shared seam ships as an **opt-in field with the unsafe default OFF**,
because "a comment is not a control on a tree this many sessions share". That is precisely
`GetBoolField(config, "overwrite_existing", false)`, enforced in SQL rather than in Go.

So decision 1 is **not** "may we build a path that rewrites published copy". It is: **may an
automated finding SET a flag that already exists, on a work-item-driven route that does not yet
exist.** What remains to build is the route and the provenance gate the ruling requires — not the
authority.

⚠ **This does NOT reopen the decision or reduce what it is worth pausing over.** `bugs_open/320` §15
records the owner granting `overwrite_existing: true` for a **one-off** 681-page pass and *explicitly
withholding it for the standing mechanism*, then verifying the seeded agent was left unarmed. **The
flag existing is exactly why that withholding was a real act rather than a theoretical one** — there
was something concrete to leave switched off, and it was left off. The finding sharpens the cost, and
leaves the authority question exactly where the owner put it.

### 12c. The lesson, and it is about my own roster finding

**§9 found one false entry in that roster. This is the OTHER entry's `Why` being imprecise in the same
direction** — overstating a capability's reach ("unconditional") where §9's `title` entry overstated a
census's completeness ("one UPDATE writer"). The *conclusion* of the `meta_description` entry still
stands (`page-build-handler` cannot write the column, by any route — §9e now proves that across the
whole spawn closure). But **both entries' evidence strings have now been found wrong on inspection,
and neither was caught by any of the three shape tests.** That is no longer a one-off; it is the
roster's character, and it strengthens §9f's case for a totality test over guessing which entry to
trust next.
