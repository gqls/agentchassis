# 440 — a routing key nobody understands completes green: the refusal half of 410's candidate 1, and the `spec.reason` split that makes it possible

Spun out of `bugs_closed/410_…three_seams…` (candidate 1) by owner decision 2026-09-02, so the
pattern file could close with its diagnostic work done. Owned by the (new) lane
`docs/agent_docs/docs024_key_docs_latest/bugfix_440_unknown_routing_key/`. Continues — and must
not compete with — `bugs_open/404`, whose own material names this bug's fix as "the real repair"
and defers it as RFC-scope (`platform/livespec/rerender_reasons.go`, header: *"Splitting
annotation from routing key is the real repair and is RFC-scope; this file is the half that can
ship now"*). 404's council round 4 is `complete_approved` as of 2026-09-02; nothing here touches
their shipped surface until that lane has read and recorded its verdict.

## The mechanism (all verified 2026-09-02 against live code and DB, citations by symbol)

`page-rerender`'s `check_rerender_mode` step routes to `rerender_sections` on a five-value
allow-list and takes `else_step: render_page` — assemble, re-shipping stored `rendered_html`
verbatim — for EVERYTHING else. The Go item creator (`create_rerender_items_action.go`,
`RerenderSectionReasonByName` branch) WARNS on an out-of-vocabulary reason and proceeds:
LOUD-but-assemble, shipped by 404 and correct as far as it goes. So today, end to end: **a
routing key nobody understands completes green and changes nothing** — 410's quiet-default
pattern, on the fleet's busiest pipeline.

**Refusal is IMPOSSIBLE today, and that is the actual defect.** `spec.reason` is two fields
wearing one name (404's measurement, re-confirmed below): a routing key the gate branches on AND
a free-prose annotation humans write for humans. You cannot refuse unknown values of a field
whose legitimate values include arbitrary sentences. The fix is not "refuse harder"; it is the
SPLIT — after which "present but unknown" becomes a meaningful, refusable state distinct from
"absent" (annotation-only, assemble, forever legal).

## Evidence, `[MEASURED 2026-09-02]`, each item disconfirmable

1. **The two-fields-one-name premise is being minted TODAY.** Census over live `page_rerender`
   items (`SELECT COALESCE(spec->>'reason','(none)'), count(*) … GROUP BY 1`): 11 distinct
   values. In-vocabulary: `cta_links_stale` 2,164 · `section_data_resolved` 608 ·
   `template_changed` 92. Deliberate assemble-only key: `verbatim_adoption_deploy` 24. Free
   prose, dated 2026-09-02: *"FCA rule citation corrected by migration 696 (owner decision
   2026-09-02)"* ×11 and *"component adopted by migration 693 — first rerender since the
   component_id was NULL"* ×3. Routing-SHAPED but unknown to the gate: `tool_retirement` ×16
   (08-31), `light_palette_chrome_replaced` ×13 (08-25), `listing_stale` ×1 (08-24) — all
   assembled silently.
2. **The creator's warning has fired ZERO times in production** — and the guard only watches one
   door. The free-prose items above were INSERTed by raw-SQL migrations
   (`sql_for_agents/696_…`, `693_…`), which bypass the Go creator entirely; the warning
   capability is probe-verified live in today's chassis binary (`not in the sections-rerender
   vocabulary` present in `/proc/1/exe`) and no production row carries its output. (First
   measurement of this returned "2 firings" — both were council payloads QUOTING the string;
   WRONG_CALLS 2026-09-02, the query-side twin of prompt-text-poisons-its-own-detector.)
3. **A correct assemble and a wrong assemble are indistinguishable IN THE DATA.** Migration 696
   updated `content_data` AND `rendered_html` in lockstep, so its 11 assemble-only items are
   CORRECT — the author knew assemble semantics and used the field as annotation. A migration
   that edited only `content_data` would produce identically-shaped items whose assemble
   re-ships the stale artefact. Nothing in the row says which happened. Intent has nowhere to
   live: that is the field to add.
4. **The doors are many and the guard is at one.** Producers of `page_rerender` items observed
   in the census: the Go creator (guarded), migration INSERTs (unguarded, the dominant
   free-prose source), the 615-shape fixer fan-out, `adopt_verbatim.go`, and hand dispatches
   (the 404 lane's own concession, guardian round: a config scan cannot see a kcat dispatch).
   A creator-side refusal would therefore close one door of five; the gate is the only seam all
   five pass through.

## Fix candidates, ordered by what closes the door

1. **Split the field; refuse present-but-unknown at the seam every producer crosses.** A new
   spec key (working name `spec.routing_reason`) carries ONLY vocabulary values, validated
   against `livespec.RerenderSectionReasons`; `spec.reason` remains the annotation and is never
   validated. Absent routing key → assemble (today's safe default, preserved — annotation-only
   items stay legal forever). Present-but-unknown routing key → REFUSE: fail the item toward
   `needs_human_review`, never silently assemble. Ships opt-in, unsafe-side default OFF (owner
   ruling 2026-08-02 §2); flipping refusal ON changes what the shared gate GUARANTEES and is
   therefore RFC-scope (owner ruling 2026-07-29 §1) — RFC_062 is that document. **This is the
   door-closing fix**: it guards at the gate, which raw-SQL producers cannot bypass.
2. **Interim, cheap, no design decision: put the missing-guard rule where the unguarded
   producers actually work.** A `scripts/pattern-check.py` advisory for migration files that
   INSERT `page_rerender` items with a routing-shaped-but-unknown `reason` (in-vocabulary or
   free prose both pass; `snake_case_unknown` warns, pointing at candidate 1's split). Catches
   the `tool_retirement` class at authoring time; catches nothing at runtime.
3. **Document-only**: state the assemble semantics at every producer. Rejected as primary — a
   doc comment is not a control on a tree this many sessions share (CLAUDE.md, 2026-08-02 §2).

## Verification, and the trap in it

Whatever ships, prove it can fail BOTH ways: a fixture with an unknown `routing_reason` must
REFUSE (and a mutation restoring silent-assemble must turn the test red); a fixture with
annotation-only `reason` prose must ASSEMBLE unwarned. A test asserting only over today's five
values passes the day it is written and can never do anything else — 410's verification trap,
inherited verbatim.

## 090 substitution, stated (owner ruling 2026-07-31)

Not run through the diagnosis loop. Substituted: the mechanism was established first-hand by two
lanes (404's four council rounds — the design drew no objection in any of them; 410's approved
rounds `c8385154`/`a69d82f2`), and every figure above was measured fresh against the live DB and
binary today, with the one mismeasurement caught by reading a member row and logged in
WRONG_CALLS. What the loop could still add: an independent read of whether `tool_retirement` and
`light_palette_chrome_replaced` items WANTED re-renders (their authors' intent is the one thing
no query here can recover).

## Relations

`bugs_closed/410_…three_seams…` (parent pattern file; candidate 1 verbatim at its §"Fix
candidates") · `bugs_open/404` (the shipped loud-half and the deferral this bug takes up; ACTIVE
lane, r4 approved — coordinate, never compete) · `bugs_open/384` (instance 1; its
chrome-invisible-to-content-hash lore is why `light_palette_chrome_replaced` may be a correct
assemble) · `platform/livespec/rerender_reasons.go` (the vocabulary and its declarations) ·
RFC_062 (this bug's design document) · owner rulings 2026-08-02 §2 and 2026-07-29 §1 (the
shipping shape and the RFC trigger).

---

## 2026-09-03 — THE PRODUCER CENSUS, and it resizes the fix: the creator I fixed mints ~1 of 3,172 reason-bearing items

Raised as a medium objection by the council's `bug_historian` seat against phase 1b's claim of
"the first and only producer" (correlation `934327db`, APPROVED). The claim was true as written —
nothing else writes `spec.routing_reason` — but the seat's underlying question was the right one
and the answer is material. `[MEASURED 2026-09-03]`, by `created_by` over live `page_rerender`
items:

| producer (`created_by`) | items | carry `spec.reason` | carry `routing_reason` |
|---|---|---|---|
| `rerender-pages` (**the Go creator phase 1b fixed**) | 8,918 | **1** | 0 (will stamp from next roll) |
| `completeness-discovery-agent` | 1,882 | 1,882 | 0 |
| `generic` | 388 | 388 | 0 |
| `derive_card_asset` | 313 | 313 | 0 |
| `render_news_section` | 275 | 275 | 0 |
| `component-template-fixer` | 94 | 94 | 0 |
| lane/migration producers (`bugs_open/425`, `loanzy_uk_example_site`, `agritec-workstream-…`, `lendzy_co_uk lane (migration 696)`, `bugfix_357_… (migration 701)`, …) | ~120 | ~120 | 0 |

**Consequence, stated plainly: phase 1b is correct and nearly inert.** The action it fixed almost
never stamps a reason (1 of 8,918 — its normal traffic is the assemble-only site-wide refresh).
The in-vocabulary reasons that actually drive the sections branch are written by OTHER producers,
overwhelmingly, and **`[MEASURED 2026-09-03]` 13 Go files write an in-vocabulary reason directly**,
most as raw JSON string literals that never touch the vocabulary constants — e.g.
`render_news_section_html.go`, `refresh_evidence_base_action.go`, `render_directory_action.go`,
`reconcile_section_data_action.go`, `flag_page_image_rebuild_action.go`,
`store_generated_component_action.go`, and three `discovery_checks/*` files
(`check_misdirected_cta.go`, `check_literal_markdown.go`, `check_contact_form_undeliverable.go`). ⚠ **CORRECTED 2026-09-03 (same day, by the phase-2 enumeration):** *13 Go files write an in-vocabulary reason* is true and is NOT the conversion set — reading each site's ITEM TYPE found only **five** are `page_rerender` producers. `render_directory_action.go` and `reconcile_section_data_action.go` file `needs_page`, `store_generated_component_action.go` files `needs_rerender` (its reason is propagated into page_rerender items BY the creator, which stamps there), and `discovery_checks/check_literal_markdown.go` files `literal_markdown` — matching 404's own finding that `literal_markdown` never appears on a `page_rerender` item. A blanket sweep over the 13 would have stamped page-rerender routing decisions onto items no rerender gate reads.

**What this changes (carried into RFC_062):**
1. **Phase 2 is not "migration authors" — it is a producer conversion programme** across those 13
   Go writers plus the agents and migrations. Each is a `{"reason":"x"}` literal that must gain
   the routing key, ideally via the vocabulary constants rather than a second raw literal.
2. **The transition clause is LOAD-BEARING, not a drain-window nicety.** Narrowing the gate to
   `routing_reason` alone before every producer stamps would send all ~3,100 reason-bearing items
   to assemble — this bug's own shape, fleet-wide, in the change meant to fix it.
3. **A new phase-3 gate condition:** narrowing requires a census showing zero reason-bearing items
   from unconverted producers, not merely a drained queue.

The seat was right to press, and the honest reading of phase 1b is: the seam is proven and the
first producer is converted; the population is still ahead of us.

---

## 2026-09-03 (late) — phase 2 is LIVE AND PROVEN IN PRODUCTION, and the phase-3 precondition now has a number

**Live proof, better than a probe** `[MEASURED 2026-09-03]`: **12 `page_rerender` items now carry
`spec.routing_reason`**, first seen 12:13, all from `completeness-discovery-agent` — the runner of
`check_misdirected_cta.go`, one of the sites phase 2 converted. A build containing that conversion
(`d44644635`) is live and is the fleet majority (`d0252fd4dab2`, 98 pods; the fleet is straddling
two builds, 43 pods still on `7bf1ff674021`). So the producer conversion is not merely committed —
it is writing the key into real items. `[UNMEASURED]` which exact build was running at 12:13; the
claim here is that a build carrying the conversion runs now and the data exists, not that a
particular pod produced it.

**The phase-3 precondition, quantified — this is the tracking the council's `bug_historian` and
`editquality` seats asked for** `[MEASURED 2026-09-03]`:

| population | count |
|---|---|
| pending `page_rerender` items carrying an in-vocabulary `reason` | **1,803** (unresolved 1,704 · failed 78 · deferred 12 · triaged 9) |
| …of those, carrying `routing_reason` | **12** |

**So 1,791 pending items would route to assemble the moment the gate narrows to the routing key
alone.** That is not a reason to hurry the flip — it is the measurement that makes the transition
clause (accept EITHER key) load-bearing rather than a nicety, and it is the number phase 3's
narrowing census must drive to zero (or accept as drained) before the compat disjuncts come out.

**On the 11 corpus migrations the phase-2b sweep found (2 unapplied `_HOLD`, 9 applied):**
- `_HOLD` files ARE lintable — `migration_is_lintable()` returns True for both
  `683_…_HOLD.sql` and `701_…_HOLD.sql` (verified by executing the predicate, answering
  `editquality`'s doubt directly), so the advisory fires the moment either is touched, which is
  what applying one requires.
- **The 9 applied migrations are deliberately NOT edited.** An applied migration is append-only
  history frozen by its checksum; editing one to add a routing key would rewrite the record of
  what was actually run and could break the runner on replay. The defect they left is not in the
  file, it is in the ITEMS they minted — and those items are inside the 1,803 above, covered by
  the transition clause and gated by the narrowing census. Recorded here so the absence of a fix
  is a decision with a reason, not an oversight.

## 2026-09-03 (evening) — the FIX for the headline defect is BUILT and PROVEN, held at one signature

`741_refuse_unknown_rerender_routing_key_HOLD.sql` (+ `_ROLLBACK`, `_VERIFY`) is the refusal.
Written, then **executed against the live database inside a transaction and rolled back** — twice,
then a full round trip: apply 741 → run its `_VERIFY` companion → apply the `_ROLLBACK` → compare
to the original → discard. Every assertion the migration makes about itself passed, and the
rollback restored the pre-flip state byte-exact (step count 13 → 10, gate condition identical).

~~**Why it is not applied:** owner ruling D2 — the 404 lane co-signs the gate migration. That lane
has been dormant since 2026-08-26 and the CONTRIB asking for the co-sign (2026-09-02) is
unanswered.~~ Owner decision 2026-09-03: build it all, stop before applying. The release condition
is the co-sign and nothing else.

> **⚠ SUPERSEDED 2026-09-04 — THE CO-SIGN IS GIVEN, and the "dormant" premise was false** (the
> 440 lane retracted it themselves on 2026-09-03; that lane's r4 verdict had been APPROVED and
> unread since 2026-09-02 16:33Z, and is now read and recorded — `bugs_open/404` is CLOSED).
> **The co-sign carries ONE condition:** 741's applier checklist step (c) declares
> `check_routing_key_known` as a `FragmentMatch` only, which is BLIND TO ADDITION for exactly the
> reason step (b) exists — mutation-proved with both controls (a value REMOVED live → 1 finding,
> so the guard is armed; a sixth value APPENDED live → **0 findings, silent**). Add a paired
> `CountEqual` on that clause, `ExpectCount` derived from `CheckRoutingKnownConditionClause()`
> (**7**, not 5 — it carries `== null` and `== ''` too) and the co-sign stands with no further
> round. Written up in
> `docs/agent_docs/docs024_key_docs_latest/bugfix_440_unknown_routing_key/CONTRIB_2026-09-04_from_the_404_lane_cosign_GIVEN_with_one_condition.md`
> and repeated in both `_HOLD` migration headers, where an applier will actually read it.
>
> ⚠ **CORRECTED 2026-09-04 (same day):** the blindness is real and mutation-proved, but the shape **cannot ship unnoticed** — `TestEveryFragmentMatchDeclarationIsGainVisibleOrWaived` (`platform/livespec/livespec_test.go:364`) already refuses any `FragmentMatch` Declaration with no paired `.value_count` and no waiver (verified by execution). So the condition is **"when that test stops you, take the COUNT door, not the WAIVER door"** — the clause IS an enumerable vocabulary, so a waiver would be wrong — and the part no test gives you is **`ExpectCount` = 7, not 5**. ⚠ Residual risk: `platform/livespec` has been RED at HEAD for nine days on another lane's file, so that new correct failure arrives camouflaged.

### What closing this bug will need (unchanged in shape, now concrete)

The bar is fixed AND live, and a census of zero refusals will NOT establish it: zero is equally
consistent with "nothing bad has been written yet" and "the refusal branch is unreachable". So
closure needs an INDUCED unknown routing key landing in `needs_human_review` with the message
naming it — recipe in the lane RUNBOOK — plus annotation-only prose still assembling unwarned.

### Flip-day safety, measured rather than argued (all `[MEASURED 2026-09-03]`, live)

| question | answer |
|---|---|
| pending items that would be REFUSED on flip day (present-but-unknown routing key) | **0** |
| rows in the WHOLE table (every status) that would fail the CHECK | **0** — so it can be VALIDATEd immediately after apply, not eventually |
| pending items carrying `routing_reason` | **14** (12 earlier the same day — the converted producer is live and stamping) |
| all of them, routing key and `reason` | `cta_links_stale` under both — lockstep holds live |
| pending items carrying an in-vocabulary `reason` | **1,804**, so the transition clause stays load-bearing and narrowing waits |
| items where the two keys DISAGREE | **0**, and nothing enforces it — see the residual below |

### A NEW blind spot the flip creates, in this bug's own drift class

`[MEASURED by execution]` both of the 404 lane's live livespec Declarations stay **GREEN** through
the flip, unchanged: the `FragmentMatch` holds because the old five-value clause is a substring of
the transition clause exactly once, and the paired count still reads 5 because it counts
`input_data.spec.reason ==`, which `input_data.spec.routing_reason ==` does not contain.

So five new `routing_reason ==` disjuncts arrive asserted by **nothing** — and a sixth routing
value appended to the live gate without touching Go would drift exactly the way `bugs_open/404`
drifted, inside the change built to fix it. The remedy (a `routing_reason ==` count Declaration,
plus entries for the guard step and the CHECK) is enumerated in 741's own header as work the
APPLIER owes in the same commit, because committing declarations for a HELD migration would turn
the daily auditor's clean `0 finding(s)` red every morning and mask real drift.

### The residual, stated rather than left implicit

`keys_disagree` is 0 today **as a property of the producers, not of the gate.** An item carrying
`routing_reason='literal_markdown'` with free-prose `reason` would be routed by the transition
clause (it matches either key) while `rerender_sections` is still handed `input_data.spec.reason`
— so `shouldStripLiteralMarkdown` and the CTA recompute would see the annotation and silently
under-deliver. Those readers move to `routing_reason` when the gate NARROWS (lane PLAN); until
then the `_VERIFY` companion counts the state so it cannot arrive unnoticed.

### Two traps found while building it, both now LANDMINES entries

- `ALTER TABLE site_work_items ADD CONSTRAINT ... NOT VALID` skips the SCAN, not the LOCK.
  Measured both ways on the SAME statement: **2 ms** once, **still waiting after 2 minutes**
  earlier. A queued `ACCESS EXCLUSIVE` blocks every later reader and writer behind it, so the bad
  case stalls the work-item pipeline. `SET LOCAL lock_timeout` is now in the file. One dry run
  certifies whichever answer it happened to get.
- `python3 scripts/pattern-check.py <file>` ignores the filename and lints the git INDEX, so it
  printed nothing on a deliberately bad migration — a clean pass from a checker that never looked,
  which is the shape phase 2b existed to prevent. And `migration_is_lintable()` takes a BASENAME:
  passing a path returns False and nearly produced a correction against a true claim.

---

## CONTRIB 2026-09-04, from the bug sweep lane (442) — `83407cd37` leaves **TWO** guards RED at HEAD, not one

Not this lane's bug and not touching it — reporting only, because a lane fixing the half it was told
about will still be red and will not know why.

`[VERIFIED 2026-09-04, `scripts/verify-head-builds.sh --test ./platform/orchestration/actions/...`,
HEAD `541193665`; re-checked at `a8451c251` — **no commit has touched either guard, the registry, or
`fail_work_item_message_template.go` since**, so both still stand]`

The **build** half is green. Two tests fail, both tripped by `83407cd37` ("440 phase 3 BUILT and
HELD"):

```
--- FAIL: TestTemplateExecutorsAreDeclared
    render_seam_one_spelling_test.go:246: UNDECLARED template executor(s) [renderFailWorkItemMessage]
      — a new executor is a new DIALECT … Declare it in declaredTemplateExecutors and say what its
      language is.

--- FAIL: TestFindingCodeScanEveryWriteIsRegistered
    findingcodes_scan_test.go:284: error code "FAIL_WORK_ITEM_MESSAGE_TEMPLATE_FALLBACK" is written
      by this package, is not declared in …/finding_code_registry.json, and is not in
      `_scan_baseline` — so it is NEW. Declare it (consumed / instrumented / human-evidence /
      operational, or `unruled` …) in the same commit that adds it.
```

**The `theme_kits` lane reported the FIRST on 09-03** (their handoff,
`docs/agent_docs/docs024_key_docs_latest/theme_kits/HANDOFF_2026-09-03_continue_here.md`, §4) and
was right to tell readers not to patch another lane's guard. **The second appears nowhere** — it is
not in this file, not in that one, and `grep` across `docs/` and `bugs_open/` finds no report of it.
Worth saying explicitly because the two arrive from one commit but read as unrelated failures: one
is about a template dialect, the other about an error-code registry.

Both guards are behaving exactly as designed, and both are asking a question only this lane can
answer — the registry entry in particular is a **ruling** (`consumed` / `instrumented` /
`human-evidence` / `operational` / `unruled`), not a formality, and the register entry for **WII-038**
already argues the case: the fallback "answers *the message a human is reading is the FALLBACK, not
the intended one*", which reads like `human-evidence` or `operational` — but that is this lane's call
to make, not mine.

Note the second guard's own error text names the reason it exists: *"LINK_CONTEXT_UNAVAILABLE reached
the live table on 2026-08-24 past a source-side early warning that could not see it"* (`bugs_open/358`).

⚠ **A caution for whoever runs the check.** `./platform/orchestration/actions/...` FAILs as a
package, so a `grep FAIL` over a whole-repo run tells you nothing about **which** lane owns it. Run
the two tests by name to separate them, and run your own by name to prove yours are clean — that is
how this lane established its 10 tests are green on the same red package.
