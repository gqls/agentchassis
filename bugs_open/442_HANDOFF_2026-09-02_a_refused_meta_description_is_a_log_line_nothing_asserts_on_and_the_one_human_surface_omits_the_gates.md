# 442 — a REFUSED meta description is a log line nothing asserts on, and the one human-facing surface omits the three gate reasons that actually fire

**Filed** 2026-09-02 by the `bugsweep_2026_08_26` lane. **Status: OPEN.**
Needs a config change (and possibly a Go change — two candidates below).

> ⚠ **UPDATED 2026-09-03 — read §9 before acting on §5 or §6.** Candidate 3 is **SHIPPED,
> applied and live** (migration `728`, commit `5a8728db9`) — do not redo it. §2's attribution
> of its own zero is **corrected** in §9b, which also retires §6's "do not verify at
> `orchestration_states`". And §9d records a **second silent path** in this same workflow that
> candidate 1 does not cover. Candidates 1 and 2 are open; the bug is open.

> **Resolve by SLUG** (`refused_meta_description_is_a_silent_log_line`) — bug numbers
> collide on this tree, and `git log` the FILE PATH, not the number.

**Found by the council gate reviewing someone else's fix — mine.** The `bug_historian`
seat raised it as a [medium] advisory objection against the `bugs_open/338` submission
(`106802fc-ad14-4beb-b622-147c3a0ab982`, verdict APPROVED). The round approved and still
found a real hole, which is the case for reading the objections rather than stopping at
the verdict. Its exact ask is §3 below; investigating it made the finding **sharper than
the objection knew** (§4).

---

## 1. What a reader needs first

`save_page_meta_description` (register **SEO-004**, `bugs_open/320`) writes a search-result
sentence onto a page. Before writing, it runs the site's **copy gates** — the voice gate
and the banned-claims sweep. When a gate fires, the action **refuses**: it writes nothing.

Refusing is correct behaviour. **The defect is what a refusal LOOKS LIKE from outside.**

## 2. The mechanism, read first-hand

`save_page_meta_description_action.go:192-199` returns `(map, nil)` — a **nil error**:

```go
if reason, detail := metaDescriptionFailsCopyGates(ctx, params, candidate, logger); reason != "" {
    logger.Warn("candidate refused by the copy gates — nothing written",
        zap.String("reason", reason), zap.String("detail", detail))
    return map[string]interface{}{"updated": false, "reason": reason, "detail": detail}, nil
}
```

So the step **succeeds**. Downstream, the live `meta-description-backfiller` config
(read 2026-09-02 from `agent_definitions`, not from a repo seed):

- `backfill_loop.sub_workflow` is `save_description → done` (`loop_complete`). **There is no
  conditional on `save_result.updated` anywhere in the workflow.**
- `continue_on_error: true` — moot, since the action does not error.
- Nothing reads `save_result` again.

**Net: a refusal is one `logger.Warn` plus a field in `collected_data` that nothing asserts
on.** No work item, no retry, no error, no failed status. The orchestration COMPLETEs and
the scheduled task stamps a clean run. `[MEASURED 2026-09-02]` `orchestration_states`
returns **zero rows** carrying a `save_result.reason` fleet-wide — the rows age out, so
even the field that does exist is not readable after the fact.

This is the estate's most-repeated failure shape (016b §9, *"a silent fallback deploys a
hollow section as success"*), on a page that then stays blank on an hourly schedule.

## 3. The objection, verbatim, so the ask is not paraphrased

> "The plan should either confirm an existing failure-surfacing mechanism (work item,
> error log, retry) already exists downstream of `metaDescriptionFailsCopyGates`'s
> non-empty return, or add one — otherwise this is a symptom fix on a known-recurring
> silent-loss mechanism, not a fix of the mechanism."

**Answer: it does not exist.** §2 is that confirmation, and it is the reason this file
exists rather than a line in `bugs_open/338`.

## 4. ⚠ AND THE ONE HUMAN-FACING SURFACE OMITS THE GATES — a stale list, by addition

The workflow's `complete` step carries a `result_message`, which is the only place anything
tells a person how to read the outcome. Quoted verbatim from the live row:

> "Meta description backfill finished. Read each save_result: `updated` true is a write,
> false carries a named reason (**empty_candidate / candidate_looks_internal /
> candidate_too_long / already_has_description**)."

**Those four are the cheap refusals. The list omits all three copy-gate reasons —
`voice_tell`, `banned_claim`, `voice_gate_unreadable`** — which are the expensive ones, the
ones that need a human, and the ones that caused `bugs_open/338`. A reader following this
message's own instruction would conclude the gates cannot refuse.

`[MEASURED 2026-09-02]` the action returns **7** distinct reasons; the message names **4**.
⚠ **It takes TWO greps, and that asymmetry is the whole point** — the four the message
names are string literals in the action, and the three it omits are returned as bare
strings from `metaDescriptionFailsCopyGates`, so the obvious single grep finds exactly the
four already documented and reports the list as complete:

```bash
# the 4 the message names — literals in the result map
grep -oE '"reason": *"[a-z_]+"' save_page_meta_description_action.go | sort -u   # -> 4
# the 3 it omits — returned from the gate helper, invisible to the grep above
grep -nE 'return "[a-z_]+",' save_page_meta_description_action.go               # -> :316 voice_gate_unreadable, :334 voice_tell, :341 banned_claim
```

(I wrote the single-grep version into this file first and it did not reproduce my own
number — logged in `WRONG_CALLS.md` as the same fault one step along: **a citation of a
command is not its output**.)
The gates were added to the action by the owner requirement of 2026-08-19 (`bugs_open/320`)
and **the message was not updated**. Correct when written, wrong by addition, reading as
current ever since.

> **This is the THIRD stale-by-addition list found in one task**, after `bugs_open/338`
> §4's check list (missing `negation_density`, from `bugs_open/305`) and the LANDMINES
> entry's remedy. The owner ruling of 2026-08-22 requires a COUNT to carry the date it was
> counted; **these are not counts, they are ENUMERATIONS, and they rot the same way with no
> equivalent rule.** Worth considering as its own norm.

## 5. Fix candidates, ordered by what closes the door

1. **Make a gate refusal a filed work item.** The strongest: it moves the failure onto a
   queue with a surface, and the copy is exactly what a human must judge. ⚠ Check the
   queue's volume first — `bugs_open/033`/`083` record `voice_tells` items parked in a
   review queue that has closed ~one item ever, so this must not simply relocate the
   silence. Cheapest honest version: file only on `voice_tell`/`banned_claim`, not on the
   four cheap reasons.
2. **Branch in the workflow on `save_result.updated`.** Config-only, live immediately, no
   roll. Weaker — it needs somewhere to send the branch, so it is really candidate 1's
   plumbing.
3. **Fix the `result_message` list.** ~10 minutes, config-only, closes §4. It does not make
   a refusal loud, so it is a documentation fix, not a mechanism fix — but it removes an
   actively misleading surface and should ship regardless of 1 and 2.
4. ~~Make the action return an error~~ — **rejected, name it so nobody re-proposes it.**
   A refusal is correct behaviour; erroring would fail the whole batch of up to 25 pages
   for one bad sentence, and `continue_on_error: true` would swallow it anyway.

## 6. How to verify a fix

**Induce both arms, or the test proves nothing:**
1. A candidate containing a banned phrase must produce a **visible** artefact — a work item
   row, or a distinguishable status — not merely a `logger.Warn`.
2. A clean candidate must still write silently, with **no** spurious item filed.

⚠ **Do not verify at `orchestration_states`** — it ages out (§2) and returns zero rows for
this today. Verify at whatever durable surface the fix adds.

## 7. Blast radius

`save_page_meta_description` is today the only caller of the copy gates on a single-value
field, so the silent path is one action wide.

> ⚖ **MEASURED AND CLOSED 2026-09-03 — this was an assumption when written and is now a finding.**
> `bugs_closed/464` read **every** copy-gate caller on the tree (whole-tree sweep: `platform/`,
> `internal/`, `pkg/`, `cmd/`, with a control). Ten call sites. **This action was the only one that
> returned a refusal as `(map, nil)` with nothing asserting on it.** Every other one errors, blocks
> to human review, records into a structured `rejected`/verdict list, files a work item, or writes
> a durable record by design. So the blast-radius claim below is **forward-looking, not realised**:
> the shape travels with the gate and will bite the next single-value caller, but today there is
> exactly one and it is fixed. ⚠ 464 §7.1 also corrects that bug's own five-file population, which
> was wrong in BOTH directions — read the table there, not the population. **It widens the moment the gates are reused
on a title, nav label or alt text** — which `bugs_open/338` §5 and CQ-035 both expect —
because the refusal shape travels with the gate, not with the field.

## 8. Provenance

Not run through `090`. **Substituting first-hand verification per the owner ruling of
2026-07-31, and stating the substitution rather than omitting it:** the deciding code path
was read (`save_page_meta_description_action.go:192-199`, the nil return); the live
workflow config was read from `agent_definitions` rather than a repo seed (the loop step,
its `continue_on_error`, the absence of any conditional on `save_result`, and the
`result_message` quoted verbatim); the seven-vs-four reason count is a grep over the
action. The claim is narrow — one action's refusal has no downstream reader — and asserts
no cause outside the symptom. It was independently raised by the council's `bug_historian`
seat from the submission alone, which is corroboration from a reader with no access to
this tree.

**Related:** `bugs_open/338` (the false-positive trigger, fix committed `425398a01`, inert
until the next roll) · `bugs_open/320` (SEO-004, the owner requirement that added the
gates) · CQ-035 (the single-value classification) · 016b §9 silent-fallback family.

---

## 9. CONTRIB 2026-09-03 (same lane) — candidate 3 SHIPPED, one correction to §2, and a second silent path candidate 1 does not cover

### 9a. Candidate 3 is done, applied and live
Migration `728_meta_description_backfill_result_message_names_the_copy_gates.sql`
(+ `_ROLLBACK`), commit `5a8728db9`, applied and recorded in `schema_migrations`, council
`2ed33c57-b49a-4b1b-ad1e-7e23ce6c477a` (`Council-Submitted:`, verdict owed a read).

Verified at the artefact, not at the migration's own verify block — the live row read back
independently, with a must-be-absent control:

```
empty_candidate t · candidate_looks_internal t · candidate_too_long t · already_has_description t
voice_tell t · banned_claim t · voice_gate_unreadable t · THIS_REASON_DOES_NOT_EXIST f
```

The new message names the seven, **splits them by what they ask of a reader** (four that need
nothing, three that need a person), and — because enumerating seven fixes today and rots
tomorrow by exactly the route §4 describes — says the list is a copy, names
`save_page_meta_description_action.go` + `metaDescriptionFailsCopyGates` as the authoritative
set, and says finding it takes **two greps, not one**.

Guards proven rather than assumed: rehearsed under `ROLLBACK` against the live row, then three
mutations proven RED — dropping `voice_tell` from the message, a second `jsonb_set` inside
`default_config`, and a misspelt `jsonb_set` path. ⚠ A **fourth mutation PASSED**, and it is
recorded rather than hidden: it wrote a *different column* (`task_workflow`), which the
positive control is not scoped to. The control compares `default_config` only, which is the
only column the UPDATE writes, and claims no more than that.

**§5 candidate 3 is CLOSED. Candidates 1 and 2 are not, and this bug stays OPEN.**

### 9b. ⚠ CORRECTION to §2 — the zero had TWO sufficient causes and was attributed to one
§2 says: *"`orchestration_states` returns **zero rows** carrying a `save_result.reason`
fleet-wide — the rows age out, so even the field that does exist is not readable after the
fact."* The zero is real; **the attribution is wrong**, and §6's "do not verify at
`orchestration_states`" follows from it.

`[MEASURED 2026-09-03]` the field **is** readable, for about 26 hours (oldest row in the whole
table `2026-09-02 09:41`). Five backfiller runs survive in the window — boxingonline.com,
finetuning.uk ×2, gamesdesign.co.uk, vetcomparison.uk — and **all five carry a `save_result`**:

| runs in window | carry `save_result` | `updated: true` | carry a `reason` |
|---|---|---|---|
| 5 | 5 | 5 | **0** |

So the zero is what **no refusal happened** looks like, not what **unreadable** looks like. The
action only puts `reason` in the map when it refuses, so a window with no refusal returns zero
rows either way — two sufficient causes, one measured. The discriminator is one column: ask
whether `save_result` is present **at all**, which is the demand control §2 never ran.

That does **not** rescue the mechanism — a refusal readable only by someone who queries within
26 hours and already knows to look is still a surface nobody reads — but it changes what a fix
must supply. It needs to be **durable past the window, or actively delivered**; it does not need
to invent the record.

### 9c. The volume objection to candidate 1, measured, and it holds
`[MEASURED 2026-09-03]`, counting the **archive** as well as the live table (a closer census
cannot see what it succeeded at — `site_work_items` is a rolling window):

| item_type | needs_human_review | complete | window |
|---|---|---|---|
| `voice_tells` | **66** | **5** (3 live + 2 archived) | 2026-07-17 → 2026-08-27 |

Five closed against 66 parked, and nothing filed since 08-27. §5 candidate 1's own caveat was
right: **filing there relocates the silence.** The route needs a stated reader before it is
worth taking, which is the open owner question in the lane handoff §0.4.

### 9d. ⚠ NEW — there are TWO silent paths in this workflow, and candidate 1 covers ONE
Read first-hand from the live `agent_definitions` row, 2026-09-03. The writer step's own prompt
says, verbatim:

> "Ground it in the content given. If a page's content does not support a specific description,
> **omit that page entirely** rather than inventing one. **Returning fewer entries than you were
> given is a correct answer.**"

And the workflow never compares the two counts. `check_has_pages` tests
`pages_missing_meta.count > 0`; `backfill_loop` iterates `written.result.descriptions`;
`complete` prints a message. **Nothing anywhere reads both.** So a page the LLM silently drops
leaves exactly the trace a gate refusal leaves: nothing — no `save_result` at all, not even a
`reason`, because the loop never reaches it.

**This matters for the route choice.** Candidate 1 files a work item from inside
`save_page_meta_description`, so it fires only on the pages the writer *did* return. A page
dropped by the writer never reaches the action and would never file. A fix that reads
`offered vs written` catches **both** paths and needs no queue: the comparison is two integers
already sitting in `collected_data`.

⚠ **And the sample cannot tell you whether this fires.** All five surviving runs were
`offered 1 / written 1` — but five single-page runs is a sample that could only have detected an
omission rate high enough to hit at least one of five, i.e. it rules out nothing below roughly
45%. **0 of 5 is not evidence the writer never omits.** It is evidence that nothing in the
window omitted, on runs of one page each, where the instruction has least occasion to fire.

### 9e. Damage today is ZERO, and that is a demand-controlled measurement, not an absence
`[MEASURED 2026-09-03]` fleet-wide, active pages:

| | pages | avg visible text | clearing the backfiller's `> 200` gate |
|---|---|---|---|
| blank `meta_description` | 37 | 8 chars | **0** |
| has a description | 1,171 | 4,381 chars | 1,137 |

**Zero pages are both eligible and blank.** Every remaining blank is a near-empty page the
backfiller structurally cannot select (lane handoff §3), and the owner has ruled those get no
description. The described-pages row is the demand control that makes the zero mean something:
the instrument plainly can see eligible pages, it sees 1,137 of them.

So both silent paths are **latent today, not costing a page**. That is an argument about
priority, not about whether the mechanism is broken — the next refusal on an eligible page is
silent exactly as described, and §7's blast radius (the shape travels with the gate, not the
field) is unchanged.

### 9f. What is left, for whoever takes this next
- **Candidate 1 / Route B** — needs the owner's route decision *and* a stated reader for the
  queue (9c). If taken, it should also answer 9d, or it fixes half the mechanism.
- **A fifth candidate, from 9d, not in §5 when it was written:** compare
  `pages_missing_meta.count` against `jsonb_array_length(written.result.descriptions)` and make
  the difference visible. Config-shaped rather than Go-shaped, covers both silent paths, files
  nothing into a queue nobody reads. Not costed here; not this session's to choose.

### 9g. Council verdict on the 728 submission: APPROVED — and the one thing a seat asked for, answered

`2ed33c57-b49a-4b1b-ad1e-7e23ce6c477a`, 2026-09-03 10:42Z. **`approved`, `decided_by: all
reviewers approve`**, 10 seats fired, 7 abstained, `gated_by_truncation: false`. No objections.

One seat (`prior_art_librarian`) filed a `missing` item rather than an objection, and it is right:

> "Verification that `metaDescriptionFailsCopyGates` exists and actually returns bare strings for
> voice_tell/banned_claim/voice_gate_unreadable, as opposed to the four literal result-map reasons
> — the plan's whole justification for 'this took two greps not one' rests on that asymmetry
> existing in code today, not just in the rationale."

**This is §4's own footnote one step along, for the third time in this file's life:** §4 already
records that I wrote a grep into this file whose output I had not read (`WRONG_CALLS.md`, *"a
citation of a command is not its output"*), and the submission then carried the resulting line
numbers as if the numbers were the evidence. They are a citation. Answered here at **committed
HEAD**, not the working tree, with the output and a control:

```
$ git show HEAD:…/save_page_meta_description_action.go | grep -n "func metaDescriptionFailsCopyGates"
301:func metaDescriptionFailsCopyGates(ctx context.Context, params ActionParams, candidate string, logger *zap.Logger) (reason, detail string) {

$ … | grep -nE '"reason": *"[a-z_]+"'          # grep 1 — literal reasons in the result maps
161: empty_candidate   175: candidate_looks_internal   182: candidate_too_long   233: already_has_description

$ … | grep -nE 'return "[a-z_]+",'             # grep 2 — bare returns from the gate helper
316: voice_gate_unreadable   334: voice_tell   341: banned_claim

$ … | grep -E '"reason": *"(voice_tell|banned_claim|voice_gate_unreadable)"'   # CONTROL
(empty — the asymmetry is real)
```

**And the mechanism of the asymmetry, which is sharper than "two greps".** The gate path does not
omit the `"reason"` key — it writes it with a **variable**, at `:192-199`:

```go
if reason, detail := metaDescriptionFailsCopyGates(...); reason != "" {
    return map[string]interface{}{"updated": false, "reason": reason, "detail": detail}, nil
}
```

`"reason":  reason` is invisible to any grep shaped for `"reason": "<literal>"`. So grep 1 does not
merely *miss* three cases — it matches every place a reason is written **as a literal**, which is
exactly the set already documented, and returns a complete-looking list. **A grep keyed on the
literal form of a value cannot see the call site that passes it through a variable**, and that is
the transferable form: it is not about this file, it is about auditing a vocabulary by grepping its
members instead of its writers.

---

## 10. ⚖ OWNER RULING 2026-09-03 — "make them loud". Candidate 1 BUILT, and why it is not the version §5 describes

Owner, verbatim, in answer to §9c's measurement: **"yes, make them loud"**.

Built as commit `776511e70`. Config half **applied and live** (migration `734`); Go half
committed and **inert until the chassis rolls**. Council `76288ff9-3cde-46e6-b65a-22564fac8f6d`
(`Council-Submitted:`, verdict owed a read).

### 10a. The measurement that changed the design, and it is the finding of this section
§5 candidate 1 says "file a work item". §9c warned the obvious queue is a graveyard. Before
building, the question was made answerable: **is the graveyard about humans, or about the shape
of the row?**

`[MEASURED 2026-09-03, site_work_items UNION site_work_items_archive]`

| shape | items | complete | % | parked |
|---|---|---|---|---|
| **WITH a `handler_agent`** | 56,315 | 46,465 | **83%** | 407 |
| **NO handler (flag-only)** | 6,699 | 1,142 | **17%** | 989 |

And `voice_tells`, the queue this refusal would naturally have joined: **69 rows, every one
`handler_agent = ''`** — 3 complete, 66 parked, nothing filed since 2026-08-27.

**So the graveyard is not a fact about busy humans. It is what filing without an actor looks
like.** A `needs_human_review` row with no handler would have looked like a fix and been one more
row in the 17%. That single comparison is what candidate 1 was missing, and it is cheap: one
query, both arms, the demand control built in.

### 10b. What was built
`save_page_meta_description` files `meta_description_refused` **at an actor** —
`meta-description-repair` (migration `734`), status `triaged`, page-scoped dedup key. That agent
re-asks for the sentence **with the refusal quoted back** and saves through the **same gated
action**, so the same voice gate and banned-claims sweep judge the retry. The gate stays inside
the action, never in the workflow (the 2026-08-02 §2 rule: a gate a workflow author can forget to
wire is a comment).

**The reason the first attempt failed is the one piece of information the hourly backfiller never
had.** It has re-offered the same page every hour with the same instructions and no knowledge of
why the last sentence was rejected.

⚠ **A second refusal parks at `needs_human_review`, deliberately, despite the 17% above.** The
alternative is `fail_work_item`'s ordinary ladder, which brands a two-striker `unresolved` —
terminal and silent, which is the defect this file exists about. **Parking is the terminal state
AFTER a genuine automated attempt, not instead of one**, and the row then carries the original
candidate, the rule that refused it, and the rewrite that also failed.

### 10c. What is deliberately NOT filed, each for its own reason
- **The four cheap reasons.** Nothing was published and nobody must judge anything.
- **`voice_gate_unreadable`.** The tempting one, and wrong here: it means the gate could not be
  LOADED — an infrastructure fault, not a copy judgement. A rewrite handler is the wrong actor;
  the retry would produce a new sentence, the gate would still be unreadable, and the item would
  churn against its own dedup key for as long as the fault lasted. **Residual, recorded rather
  than folded in: a persistently unreadable gate still leaves a page blank for ever and is still
  only a `logger.Warn`.** It needs an operational surface, not a rewrite.
- The classifier is an **allow-list**, not "everything except the cheap four", so an **eighth
  reason is silent by default rather than loud by default**. §4 is the evidence that this
  vocabulary grows by addition without anyone noticing.

### 10d. ⚠ A MUTATION FOUND A REAL DEFECT IN MY OWN TEST — reported, not hidden
The arm asserting §6's second requirement — *"a clean candidate must still write silently, with
no spurious item filed"* — was written as: register **no** sqlmock expectations, then assert
`ExpectationsWereMet() == nil`, with a comment claiming this "cannot pass by accident".

**It passes unconditionally.** `ExpectationsWereMet` reports UNFULFILLED expectations; with none
registered it is nil whatever the code does. Deleting the classifier's early return — so that
*every* reason files — left the test GREEN, because the unexpected `BeginTx` merely returned an
error to the code under test, which logged it and returned.

Rewritten as an **inverted assertion**: register `ExpectBegin()` and require it to be
**unfulfilled**. Now red for all six reasons under that mutation. Five Go mutations proven red in
total (no handler · `needs_human_review` · delete the early return · drop the page id from the key
· allow-list → deny-list), plus two on the migration (the save step bypassing the gated action ·
both branches routing to the same step).

**This is §6's own warning landing on the person who wrote it**: "induce both arms, or the test
proves nothing" — and the arm I got wrong was the negative one, which is always the easier one to
write vacuously.

### 10e. Stated gaps, so they are not discovered as surprises
- **No verifier is registered for `meta_description_refused`.** `complete_work_item` completes
  without one, as it does for `content_rewrite` and most types (23 `RegisterVerifier` calls
  fleet-wide). Registering one is **not** a one-line change: it fails five build guards and needs
  a live migration amending the claimed-item-timeout sweep's `pre_query`, merged with any other
  lane's pending amendment — `discovery_checks/verify_required_fields_missing.go` documents the
  exact sequence. **Named follow-up, not smuggled in.**
- **§9d's second silent path is still open.** The writer omitting a page never reaches the action,
  so this filing cannot see it. The fifth candidate (compare `pages_missing_meta.count` against
  the length of `written.result.descriptions`) is unbuilt.
- **Nothing has exercised this yet**, and it cannot be exercised on demand: `[MEASURED
  2026-09-03]` zero active pages are both blank and clearing the backfiller's `> 200` gate, so
  there is no page to refuse. The first real proof will be the first `meta_description_refused`
  row. ⚠ **Do not read "no rows" as "it does not work"** — read it with the demand control:
  `SELECT count(*) FROM pages WHERE status='active' AND COALESCE(meta_description,'')='' AND
  page_visible_text_len(id) > 200;` If that is 0, nothing could have filed.

### 10f. ⚠ THE PRIOR RULING THIS COMES CLOSEST TO, AND THE CHECK THAT IT IS NOT CROSSED

Found *after* the council round was submitted, by reading
`docs026_concept_register/102_coverage_ratchet.txt` line 105 while registering the mechanism.
It names, as a thing that must NOT be done casually:

> "a work-item-driven route to `save_page_meta_description` would be **NEW STANDING AUTHORITY to
> rewrite published copy on an automated finding** — the authority the owner explicitly withheld
> on 2026-08-21 (`bugs_open/320` §15) — and is architecture-scope with its own council round, not
> a ratchet line"

**This IS a work-item-driven route to `save_page_meta_description`.** So the question is exact:
does it carry that authority? **No, and it is checkable rather than arguable.**

The withheld thing is specifically `overwrite_existing: true`. `bugs_open/320` §15 records that
the flag was granted **for one act only** — set inline on a one-off dispatch script — and that
*"the seeded agent was never armed"*, verified afterwards by probing the step config.
`meta-description-repair`'s `save_description` step **does not declare the key**, so it defaults
`false`: it can fill a blank and nothing else. It cannot touch a page that already has a
description, which is the entire content of the withheld authority.

Verified 2026-09-03 with the same probe the 320 lane used, and with a control proving the probe
can see a declared key at all:

```
meta-description-repair | declares overwrite_existing: f | (absent -> defaults false)
control: jsonb_build_object('overwrite_existing',true) ? 'overwrite_existing' -> t   (must be true)
         jsonb_build_object('other',true)              ? 'overwrite_existing' -> f   (must be false)
```

⚠ **Two things follow, and both are for whoever touches this next.**
1. **Adding `overwrite_existing: true` to this agent IS the withheld authority.** It is a
   one-line config edit that would look completely innocuous in a migration diff, and it is an
   owner decision plus a council round — not a config edit. It is now flagged in the register
   entry (**SEO-008**) as well as here.
2. ⚠ **The council round `76288ff9` did NOT see this fact** — my submission never cited
   `bugs_open/320` §15 or the ratchet line, because I found them afterwards. Whatever that round
   returns, it did not weigh this, and the verdict should be read with that gap named rather than
   as having cleared it. The constraint itself is verified above, independently of any seat.

### 10g. Council round 1 = REVISE, and the objection that changed the code

`76288ff9`, round 1: **revise**, gated by a **guardian [HIGH]**. Round 2 resubmitted under the
same correlation. **The gating objection was correct and so were most of the rest** — this is the
second time this week a round found a real defect, and the first cut of this very change is what
it found it in.

**The gating objection (guardian [HIGH]).** *"A shared action is being silently modified outside
the reviewed edit list — that is a scope-enumeration gap, not a budget trim."* Round 1 declared
only the NEW file and showed the wiring to `save_page_meta_description_action.go` inside a sketch;
my own "files not in the edit list" section justified three omissions and never mentioned this
one. `editquality` raised it independently at [medium]. Fixed: the action file is now edit 1.

**⚠ THE ONE THAT CHANGED THE CODE, and it is the sharpest thing either round produced.**
`bug_historian` [medium]:

> "`fileMetaDescriptionRefusal` itself has ~5 silent early-return branches … that only
> `logger.Warn` and return … If the DB write of the 'loud' record fails, the refusal is right back
> to being a log line, **i.e. the exact defect this plan exists to close, now one hop deeper and
> harder to notice because the design narrative says it's already fixed.**"

It named `bugs_closed/034` (validation errors dropped with no durable record) as the same shape.
**It is right, and I had reintroduced the defect inside its own fix.** The filing now returns
`(filed bool, fileError string)`; the action puts both in its result map, which is the surface
migration `728`'s operator message already tells a reader to read. Control flow unchanged — a
bookkeeping fault must never fail a correct refusal. `inserted:false` (dedup key already holds an
open row) counts as **filed**: the question is whether the refusal is RECORDED, not whether this
call wrote a new row. Mutation proven RED, and it kills **only** the new test — nothing else in
the suite can see it.

**Two claims I had asserted, now measured** (both were fair hits):
- **guardian [medium], other callers.** Round 1 said "single-producer" without measuring.
  `[MEASURED 2026-09-03, live config, top-level AND `sub_workflow` steps]` **exactly two callers,
  both this lane's** — `meta-description-backfiller` and `meta-description-repair`. Control: the
  same query for `ensure_site_record` returns **50**, so it can see many callers. No third-party
  pipeline gains a side effect.
- **reuse_agent [medium], a fourth bespoke refusal-parker.** Audited the three it named:
  `emitOwnedPageReviewItem` (`owned_page_guard.go:393`), `parkPageBuildFailure`
  (`page_build_failure_guard.go:162`), `emitRequiredFieldsMissing` (`work_items_common.go:705`).
  **All package-private, none exported, no common signature** — one takes a page name and a class,
  one a strike count, one a page context. A grep for an exported work-item refusal helper finds
  none. So: **there is nothing to reuse at that level**, and what IS shared is used —
  `insertWorkItem`/`writeWorkItem`, the door carrying the policy probe, registration probe,
  anti-churn and `idx_swi_dedup`. The seat's underlying point stands and is recorded in **SEO-008**
  as an extraction candidate: this shape has now been independently reinvented **four** times.

**Answered honestly rather than claimed clear** (`bug_historian` [medium]): do sibling gated-save
actions share the `(map, nil)`-on-refusal shape? Five files call the copy gates; cross-referenced
against files with refusal-shaped returns, **only this action is in both sets** — but that is a
**grep intersection, not a read**, and a differently-named result key would be invisible to it.
**Not claiming the class is clear.** Reading the other four is its own piece of work.

**⚠ And the fourth instance of my own worst habit.** `prior_art_librarian` objected [medium] to
three load-bearing claims being asserted rather than shown — and **I had personally verified all
three earlier the same session**, and had logged the remedy for exactly this fault, myself, hours
earlier (§9g). The seat cannot tell a verified claim from an unverified one. `WRONG_CALLS.md` now
carries it as instance four, with the note that **logging a lesson is not adopting it.**

### 10h. ⚠ THE ROLL HAPPENED AND MY CODE IS **NOT** IN IT — measured at the binary, with two controls

`agent-chassis` rolled to **`v1.0.1358`** at **12:06:47Z / 12:07:16Z** on 2026-09-03. Both my
commits (`776511e70` r1, `356196fe9` r2) were already ancestors of HEAD before the pods restarted.
**The Go half still did not ship.** The build was cut from an earlier HEAD.

**This is the exact landmine — "a roll is not evidence your fix shipped" — and the inference that
would have been wrong is the obvious one:** *my commit is in HEAD, a roll happened after it,
therefore it is live.* It is not.

**How it was measured.** ⚠ The prescribed `grep 'build provenance'` recipe **failed exactly as its
own landmine says it does**: on this service it matched the chassis's logged landmine corpus
*about* build provenance and returned pages of unrelated text. The fallback with no shelf life is
the binary probe of a KNOWN value, and it is only evidence with both controls:

| probe string | meaning | result |
|---|---|---|
| `meta_description_refused` | my new item type | **absent** |
| `meta-description-repair` | my new handler | **absent** |
| `candidate_looks_internal` | pre-existing, must be present | **PRESENT** |
| `ZZZ_string_that_cannot_exist_9f3a` | must be absent | absent |

The third row is what makes the first two mean something — it proves the probe is reading the
right binary and can say yes. The fourth proves it can say no. **A probe with neither control
returns "absent" for a broken probe and a missing feature identically.**

**So the status lines stay accurate but are now SPECIFIC:** it rides the *next* roll, and
`v1.0.1358` is measured NOT to carry it. Re-probe with the table above rather than reasoning from
ancestry: `git merge-base --is-ancestor` answers "is my commit in the source HEAD", which is a
different question from "is my code in the binary", and this is the case that separates them.

**Knock-on: the roll killed the council round.** Round 2 was at `review_guidelines` when the pods
were replaced at 12:06 — `[MEASURED]` 7 of 11 in-flight orchestrations fleet-wide froze in the
same window, so it was the roll and not this submission. Resubmitted under the same correlation
(`RUN_ORCH_ID=0eed76cd`). ⚠ **A council round is not durable across a roll** — the landmine says
so and this is a second instance. If a roll is expected, either submit after it or expect to
resubmit; the correlation survives, the run does not.

### 10i. ⚖ LIVE on `v1.0.1359` — and the round-2 verdict, answered

**The mechanism is fully live as of 2026-09-03 13:28Z.** Config (`728`, `734`) applied earlier;
the Go shipped on `v1.0.1359`, verified at the binary on **both** pods:

| probe | meaning | v1.0.1358 | **v1.0.1359** |
|---|---|---|---|
| `meta_description_refused` | the new item type | absent | **PRESENT** |
| `meta-description-repair` | the new handler | absent | **PRESENT** |
| `candidate_looks_internal` | pre-existing, must be present | PRESENT | PRESENT |
| `ZZZ_string_that_cannot_exist_9f3a` | must be absent | absent | absent |

**Nothing has EXERCISED it.** `[MEASURED 2026-09-03]` zero active pages are both blank and past
the backfiller's `> 200` gate, so no page can be refused. Read "no rows" with that control.

### Round 2: REVISE (gated by `prior_art_librarian` [HIGH]). Round 3 submitted.

**The gating objection was right about the ambiguity and wrong about the risk — and the fix was
evidence, not code.** It read: *"Rationale claims the new filing 'uses the shared writeWorkItem
door' … but the sketch calls `insertWorkItem(...)` directly — a different symbol."* True: round 2
named one symbol in prose and called another in code, and gave the reviewer no way to connect
them. **They are not two doors** — `load_work_item_actions.go:1901-1904`, verbatim:

```go
func insertWorkItem(ctx context.Context, tx *sql.Tx, item workItem, logger *zap.Logger) (bool, error) {
	w, err := writeWorkItem(ctx, tx, item, dropOnConflict, logger)
	return w.Inserted, err
}
```

A two-line wrapper with the default conflict policy (`dropOnConflict`, `:1817`). Every probe runs.
⚠ **This is the same fault in its third consecutive round: citing instead of showing.** Four rows
in `WRONG_CALLS.md` now. The fault is not being wrong — all three claims were true — it is that a
reviewer cannot tell a verified claim from an unverified one, and this time it cost a HIGH.

**`editquality` [medium]: `summary` undefined.** A **sketch** defect, not a code defect —
`summary := fmt.Sprintf(...)` is at line 190 of the real file, and the independent proof is that
the code is *running*. Sketch corrected in round 3.

**`bug_historian` [medium] + MISSING: the four unread call sites → `bugs_open/464` FILED**
(`9e72f75b0`), naming `section_editor_regulated_guard.go`, `save_sections_claims_guard.go`,
`rewrite_negations_action.go` and `validate_page_content.go`, stating they are **unread rather
than clear**, and recording why the grep intersection was unsound: it keys on four result-key
spellings, and `rewrite_negations_action.go` returns `{"status": …}, nil` — invisible to it.

**`reuse_agent` [low]: is there a shared item-key convention? MEASURED — no.** Eight item-key
builders under `platform/orchestration/actions/`, every one package-private with a different
signature (`silentItemKey(check, siteID)`, `deadURLControlItemKey(pageName, slot, deadFields)`,
`triageItemKey(itemType, handler, errSig)`, …). Nothing to reuse — same answer as the parkers,
now measured rather than assumed.

**`prior_art` [low]: attach the SQL for "exactly two callers".** Done in round 3. ⚠ Its
`sub_workflow` arm is load-bearing: the backfiller's own call is NESTED, so a top-level-only query
reports ONE caller and misses the incumbent.

**Nine seats approved cleanly** in round 2, including `guardian` and `architecture` — the two that
objected in round 1 — so the r1 fixes held.

### 10j. ⚖ COUNCIL **APPROVED** (round 3) — and the three advisories, answered with output

`76288ff9-3cde-46e6-b65a-22564fac8f6d`, round 3: **`approved`** — *"approved with 1 advisory
objection(s) — none high-severity"*, 5 abstained. Seven seats clean, including `editquality`,
`architecture`, `debug_historian` and `constitution`.

**Three rounds, and each found something real:** r1 the missing edit declaration **and** the
silent-failure branches inside the fix; r2 the `insertWorkItem`/`writeWorkItem` ambiguity and the
unread call sites (→ `bugs_open/464`); r3 only advisories. That is the argument for revising
rather than defending.

**`guidelines` [low] — `recurrenceExpected` unset. Answered from the code, not from intent.**
The seat worried a legitimately recurring refusal would be mishandled by the two-strike brake.
`load_work_item_actions.go:2075-2088`, the brake's own counting query:

```sql
WHERE site_id = $1 AND item_key = $2
  AND status IN ('complete', 'failed')          -- ← the strike set
  AND created_at > NOW() - INTERVAL '7 days'
```

Two things follow, and both are why leaving it unset is right:
1. **A parked repair is NOT a strike.** The repair agent's failure path is `fail_work_item` with
   `status_override: 'needs_human_review'` — not `failed` — so it never enters that count. It is
   an OPEN row, so it **holds the dedup slot** and the next hour's refusal collapses onto it.
   One row, one unfixed defect, waiting for a person. Exactly the intent.
2. **Even a strike does not lose anything now.** The within-cycle arm writes the row `deferred`
   with `retry_after` and holds the slot ("*the old arm had no legitimate use: no caller wants its
   request destroyed with nothing recording that it existed*"), and a two-striker "*falls through
   and is branded — recorded, like every other two-striker, instead of vanishing*".

Setting `recurrenceExpected` would **skip the brake entirely**, which is the opposite of what is
wanted: an hourly re-refusal of the same page is the SAME unfixed defect, not a fresh request.

**`tooling_provenance` MISSING — a `doc_notes` entry for the action/agent. It already exists**, via
the sanctioned channel (`landmines-sync.py`), and CLAUDE.md forbids hand-writing landmine rows:

```
landmine | save_page_meta_description
landmine | agent_definitions where type='meta-description-repair' or 'meta-description-backfiller'
landmine | pages.meta_description
```

**`reuse_agent` [medium] — the real remaining reuse question is `bugs_open/464`.** The seat is
right that the sharper question is not item-key helpers but whether the four other copy-gate
callers need the same treatment. ⚠ **Its wording asserts they "already refuse saves silently",
which is precisely what 464 says is UNKNOWN** — they are unread, not established. That bug is the
work; this entry does not inherit the seat's assumption.

---

## 11. CONTRIB 2026-09-04 (same lane) — the mechanism survived a second roll, and §3's zero was read wrong

Four things, in descending order of how much they change what the next reader should do.

### 11a. Still live after `v1.0.1360` — and this needed checking, it was not a formality

`[MEASURED 2026-09-04 11:17Z]` The pods rolled to **`v1.0.1360`** at 22:06Z on 09-03, *after* the
`v1.0.1359` that §10i verified. §10h is the standing reason not to assume a later tag carries an
earlier tag's code, so it was re-probed at the binary, on **both** pods, with both controls:

```
agent-chassis-ffc9ddff9-jvw92 / -k866t   (image v1.0.1360)
  PRESENT meta_description_refused
  PRESENT meta-description-repair
  PRESENT candidate_looks_internal      <- positive control
  absent  ZZZ_cannot_exist_9f3a         <- negative control
```

And the agent row is unchanged: `is_active = t`, `declares_overwrite = f`. **§4.1's owner ruling
still holds on the live row** — that check is worth repeating on every roll, because the thing it
guards against is a one-line config edit, not a code change.

### 11b. ⚠ CORRECTION to §9e and to the lane handoff's §3 — "no page can be refused" is a DRAINED QUEUE, not absent demand

§9e measured zero eligible-and-blank pages and concluded the silent paths were "latent today". The
lane handoff turned that into the stronger sentence **"no page can be refused and no item can
exist"**. The zero is real and I reproduced it (`0` again today). **The conclusion drawn from it is
wrong**, and it is wrong in the direction that tells the next session not to expect the acceptance
evidence.

The zero is measured *at rest*, after the hourly drainer has run. Read at the deciding arm
(`cmd/scheduler/main.go`, the `dynamicData == nil` branch): when a task's `pre_query` finds nothing
the scheduler stamps `last_triggered_at`/`last_completed_at` and `continue`s — **no message, no
orchestration, but the stamp still advances**. So an empty tick is invisible in
`orchestration_states` and visible only as the stamp.

And `meta-description-backfill`'s `pre_query` **is** §3's demand-control query, verbatim, grouped
by site with `LIMIT 1`. So the two are the same instrument read at different moments.

`[MEASURED 2026-09-04, `orchestration_states`, whole retention window 09-03 10:21Z → 09-04 11:24Z ≈ 25 h]`

| | |
|---|---|
| backfiller dispatches in the window | **4** (all reached `complete`; none `complete_nothing_to_do`) |
| pages offered to the gated action | **5** (1, 1, 1, 2) |
| pages written | **5** |
| pages **refused** | **0** |
| eligible pages at rest, now | **0** |
| schedule | `interval_seconds = 3600`, `enabled = t`, last stamp 10:56Z — an **empty** tick (no 09-04 10:56 orchestration exists, and demand was 0) |

The five are five **distinct** pages on four domains (loanzy.uk, remortgagecalculator.uk ×2,
leopardessconsulting.co.uk, mortgagecalculator.co.uk), all now carrying 78–111 char descriptions.
All five were **created weeks earlier** (08-02 … 08-18), so these are pages that became *re-eligible*,
not new pages. `[INFERRED, not checked]` what made them eligible — crossing the `>200` visible-text
gate on a re-render, or a description being cleared. Worth one query by whoever cares; it decides
whether arrivals continue.

**So the correct statement is: the gated action is exercised roughly five times a day and passes
every time. The refusal branch is untaken, not unreachable.** The first `meta_description_refused`
row will arrive on its own; nothing needs to be forced, and no one should read §3 as "this cannot
be tested". This is the **same shape as §9b's misstep** — a zero with two sufficient causes
(nothing eligible / nothing refused) attributed to the one that stops you looking — one section
along, in a file that already carries the warning.

### 11c. §9d's offered-vs-written measurement, redone with a stated denominator

§9d could not say whether the writer silently drops pages. Redone across every surviving run:

| run | offered | written | dropped |
|---|---|---|---|
| 09-04 06:43 | 1 | 1 | 0 |
| 09-04 04:36 | 1 | 1 | 0 |
| 09-03 17:05 | 1 | 1 | 0 |
| 09-03 14:03 | 2 | 2 | 0 |

**0 of 5 pages dropped, 0 of 4 runs.** The power statement from §9d is unchanged and must travel
with the figure: 0 of 5 rules out nothing below roughly a 45% omission rate, and four of the five
were single-page runs, where "omit that page entirely" has least occasion to fire. **This is not
evidence the writer never omits.** It is a second clean sample with its window and denominator
written down, so a third can be compared against it.

### 11d. ⚠ NEW — a loop's output field ALSO appears UN-SUFFIXED, holding the LAST iteration, and §9a's completion message points an operator straight at it

The message this lane shipped tells the operator: *"Read each save_result"*. There is a trap in
that instruction that I put there and did not see.

`[MEASURED 2026-09-04]` A loop sub-workflow's `output_field` lands in `collected_data` **twice**:
as the per-iteration series `save_result_0`, `save_result_1`, … **and as a bare `save_result`
holding the last iteration's value.** On the 09-03 14:03 two-page run:

```
save_result_0 -> page 2bcf3e28…  updated true   (102 chars)
save_result_1 -> page 34d8d807…  updated true   (105 chars)
save_result   -> page 34d8d807…  updated true   ==  save_result_1
```

**It is not this workflow's quirk.** On `page-content-writer`, over the 20 most recent loop runs,
bare `copy_gate` equalled `copy_gate_<max N>` — the genuine last iteration — **20 times out of 20**,
and never the first. (An earlier cut of that query compared against a hard-coded `_4` and reported
two mismatches; those were runs of 3 and 4 sections, i.e. my comparand was wrong, not the rule.)

**And it is not uniform, which is what makes it a trap rather than a convention:** in that *same*
loop, `section_output` has **no** bare form at all. So neither the presence nor the absence of a
bare key tells you anything, and a reader cannot learn the rule from one example.

Why it matters here: a multi-page run where page 0 is **refused** and page 1 is **written** leaves
`collected_data->'save_result'` reading `{"updated": true}`. An operator who follows the message to
the obvious key sees a clean run. 1 of the 4 runs in the window was multi-page, so this is not
hypothetical arithmetic. The refusal is **not lost** — §10b's work item is filed regardless, which
is precisely the property that makes the fix worth having — but the human-facing surface, the one
this whole bug is about, still under-reports.

**Scope of the claim, stated honestly:**
- `[MEASURED]` the bare key exists and equals the last iteration — two agents, 24 runs.
- `[MEASURED]` **no live step outside the loop reads the bare form on `page-content-writer`** (zero
  steps whose config matches `\m(copy_gate|generated_content)\M`), so no realised *config* damage
  was found there. The exposure measured here is to a **human or a query**, not to the workflow.
- `[UNTRACED]` **which code writes the bare key.** `loop_expansion_handler.makeIterationOutputField`
  rewrites each injected step's `OutputField` to `{field}_{N}`, so the injected steps are *not* the
  writer of the bare key. I did not find what is. **Nobody should assert a mechanism for this
  without tracing it** — and note it sits close to, but is not the same as, `bugs_closed/287`'s
  landmine, which describes an un-suffixed reference finding **nothing**. Here the un-suffixed key
  is *present and plausible*, which is the worse failure.

**Candidate, not shipped:** amend §9a's `result_message` to name the series
(`save_result_0`, `save_result_1`, … — the bare `save_result` is only the LAST page). One migration,
DB-config, live immediately, council-scope. Deliberately left for a decision rather than appended to
an already-approved message at the end of a session.

### 11e. Adjacent — HEAD's `actions` package is RED, and both failures are the `bugs_open/440` lane's

`[VERIFIED 2026-09-04, `scripts/verify-head-builds.sh --test ./platform/orchestration/actions/...`,
HEAD `541193665`]` The **build** half is green; two tests fail, both tripped by `83407cd37`
("440 phase 3 BUILT and HELD"):

- `TestTemplateExecutorsAreDeclared` — undeclared template executor `renderFailWorkItemMessage`.
- `TestFindingCodeScanEveryWriteIsRegistered` — error code `FAIL_WORK_ITEM_MESSAGE_TEMPLATE_FALLBACK`
  written by the package, absent from `finding_code_registry.json` and from `_scan_baseline`.

The `theme_kits` lane flagged the **first** on 09-03 and correctly told readers not to patch another
lane's guard. **The second appears unreported** — worth saying, because both come from one commit and
a lane that fixes only the one it was told about will still be red. Both guards are doing exactly
what they were built to do; the fix is a declaration, and it is the 440 author's to make.

**This lane's own tests are green at HEAD:** all 10 of
`TestSavePageMetaDescription_*` / `TestMetaDescriptionRefusal*` pass when run by name against
`541193665`. Do not read the package FAIL as this lane's.
