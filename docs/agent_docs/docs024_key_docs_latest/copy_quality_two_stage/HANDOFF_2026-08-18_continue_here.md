# HANDOFF 2026-08-18 (current as of 2026-08-19 evening) — continue here

**Lane:** `copy_quality_two_stage`. **Supersedes `HANDOFF_2026-08-17_continue_here.md`**
(kept for its worked example and the apply recipe; its state lines are stale).

> ## ▶ START HERE, IN THIS ORDER
> 1. **`SUMMARY_2026-08-19_the_fault_was_never_in_the_writer.md`** — 5 minutes, plain prose,
>    tells you what this lane now believes and why the 08-17 summary's headline is wrong.
> 2. **`bugs_open/305`** — read its §3 CORRECTION block and the ROOT CAUSE section at the end.
>    Skip its title, which overstates its own evidence and is left standing on purpose.
> 3. **This file's "Next work"**, below. Everything above that heading is context.
>
> **One-line state:** stage 2 is built, live, and has applied owner-approved edits on two
> sites; the fault the owner complained about is **in the briefs, not the writer**, fleet-wide
> (24 of 25); nothing is in flight; no decision is waiting on the owner.
>
> ⚠ **Two things that changed under us and will change again:** the chassis rolled three times
> in two days (currently `v1.0.1314` = `d3590ca46`), and migration **474 went live at
> 2026-08-19 10:34:35Z**, so a stage-2 apply now runs the write-time markdown strip — read
> `stripped_markdown_fields` on the apply result. **Re-verify both before trusting anything
> dated here.**

## State in five lines

- **Stage 2 is BUILT, LIVE, and has applied edits on TWO live sites.** `copy-editor`,
  config-only (migrations **447** seed + **462** budget), register **CQ-024**.
- **It cannot write to a page.** No step in it can; the migration RAISEs if one is added.
  Output goes to `copy_edit_proposed` at `needs_human_review`. **2 proposals to date, both
  approved by the owner and applied by hand.**
- **The gate is `gate_stage2_edit.py`** — 5 induced controls + a dialect control + a
  prose-URL control, all fire. **Three holes were found in it on 2026-08-18 and fixed**;
  see "what the gate learned" below before trusting it further.
- **Nothing dispatches `copy-editor`.** No item_type routes to it, by choice. Every run so
  far was a hand-fired canary.
- **The lane now has TWO owner-directed inbound asks** (below). They outrank anything I
  would have picked myself.

## ⚠ START HERE — read the `57b2dcd2` diagnosis verdict (above), then the two asks below (now ANSWERED, kept for context)

### 1. `portfolio_positioning`'s CONTRIB — the owner says framework copy "looks like it didn't go through the framework"

`CONTRIB_2026-08-18_the_negative_default_survives_a_POSITIVE_identity_spec_on_directory_pages.md`

**The owner gave this lane BOTH halves: *"ensure that that sort of copy never leaves this
framework again"* AND fixing the affected pages.** He reviewed three live directory pages on
`ai-agent-orchestration.com` (`model-directory`, `adoption-tracker`, `protocol-tracker`) and
quoted:

> *"The registry shows you what's possible, not what survives production."*
> *"…tells you which agents exist. It doesn't tell you how they…"*

**That is the define-by-negation construction, and it matters for how you read today's run
3.** I measured 19 such constructions on that site's index page in the morning and recorded
that stage 2 left them nearly untouched (19 → 15), noting I could not tell whether the agent
was right to spend its budget on restatement instead. **The owner has now independently
answered that: the negation construction is what reads wrong to him.** So the honest reading
is that stage 2 de-prioritised the fault the owner cares most about. Do not treat that as
settled either way without checking — but do not repeat my "cannot separate the two
readings" line as though the owner had not spoken.

⚠ The CONTRIB states plainly that **this lane's 2026-08-12 root cause does not explain this
case** — the site's identity spec is entirely POSITIVE, so "the default is negative because
the identity spec says so" cannot be the mechanism here. Read their §2–§5 before theorising:
they located it in `page_components.rendered_html` for the `call-to-action` component on
`model-directory`, updated 2026-08-17 20:08Z, and deliberately did NOT rerender or touch the
writer, so the evidence is intact.

### 2. `finetuning_uk_service`'s CONTRIB — three questions, owner told them to ask us

`CONTRIB_2026-08-18_finetuning_offer_page_needs_your_register_machinery.md`

A NEW page on a site stage 2 has never run on, with an owner-specified register
("friendly and EXPANSIVE, not dense; a techie thing that doesn't sound techie"; possibly a
glossary). Their three questions: what our machinery needs from them in `site_specs`/identity
fields; whether Stage 2 should run on offer pages; and how the edit budget behaves at page
BIRTH rather than on an existing page. **That last one is a real gap** — everything measured
so far is stage 2 editing a page that already exists.



## ⚠ UPDATED 2026-08-19 — ROOT CAUSE FOUND, and it is the BRIEF. Plus a pending change to stage 2's write path.

### The cause of the owner's directory-page complaint: `content_direction` is written in the construction

Full evidence in `bugs_open/305`'s final section. In short `[MEASURED 2026-08-19]`:

- **All three flagged pages are PRE-v2** — adoption-tracker and protocol-tracker CTAs written
  **2026-07-26**, model-directory 08-06/08-08, and the hero phrase first seen **2026-04-10**
  across 251 calls. The symptom contains no post-v2 output at all, which is why every
  v2-based explanation failed, mine included.
- **The site's brief carries the shape seven times and hands down the tagline verbatim** —
  *"deployed to production in days, not months"* appears in **1,348 rendered prompts and 408
  responses** across 21,078 writer calls. Literal chain, not a rate.
- **Fleet census: 24 of 25 `content_direction` specs carry it**, at 24–38 instances —
  `ai-agent-orchestration.com` has SEVEN, i.e. the complaint came from one of the LEAST
  saturated briefs. ⚠ Top of the list is `remortgagecalculator.uk`, the Phase C pilot
  `portfolio_positioning` offered to rerun — warned them: rerunning against that brief
  regenerates the register from a worse source and would read as the fix failing.
- **The fix is the SPEC, not the writer.** A detector pointed at specs
  (`count_negation_tells.py` aimed at `content_direction` instead of a page) catches it before
  a page is written. ⚠ Site-config on other lanes' sites — coordinate, do not edit unilaterally.
- **Not proven:** that the briefs' *instructional* uses of contrast transfer to output. One
  literal transfer is proven (the tagline). Design that measurement so it can come out either
  way — see `305 §3`'s correction for why that sentence is there.

### ⚠ A pending migration changes stage 2's write-path guarantee

> **⚠ SUPERSEDED WITHIN THE HOUR — 474 IS APPLIED. `2026-08-19 10:34:35Z`** (473 at 10:34:26Z).
> My "PENDING / inert today" reading below was true when measured and false minutes later; the
> 184 lane told me, and I verified independently: both rows in `schema_migrations`, and
> `strip_literal_markdown = true` on **all three** agents (`section-editor.apply_edit`,
> `page-rerender.rerender_sections`, and `page-content-writer`'s nested sub_workflow step —
> which a top-level-steps query does NOT see; use `jsonb_path_query_array(default_config,
> '$.**.strip_literal_markdown')`).

**So, from 10:34:35Z, `"what the gate graded is what lands"` no longer holds unconditionally.**
`section-editor`'s `apply_edit` runs the merged content map — LLM `field_updates` included —
through `StripLiteralMarkdownFromContentData` before render and persist.

**The exemption is per-VALUE and asymmetric — an HTML field is NOT exempt** (supplied by the
184 lane, confirmed by me in `literal_markdown.go:112-121`): `mdLinkStripRe` and
`mdCodeSpanStripRe` are gated behind `includeCodeSpan = !HTMLMarkupRe.MatchString(value)`, but
**`mdBoldStripRe` and `mdHeadingStripRe` run UNCONDITIONALLY.** So `ported-prose.content`
skips two of four transforms and is exposed to the other two.

**What was done about it, so the burden stays ours:**

- `gate_stage2_edit.py` now prints a `⚠ strip <field>` advisory when a proposed value carries
  markers the strip will act on — **advisory, never a failure** (a strip firing on stage-2
  output is arguably fixing the agent's mistake). It is a heads-up keyed on markers, and
  deliberately **not** a reimplementation: a second copy of production's regexes would drift.
  Nine-case control committed.
- **The authority remains `stripped_markdown_fields` on the apply result** — non-empty iff it
  fired, with the exact field paths. Read it after every apply.
- Their canary found a deeper layer (news-listing items are query-resolved, and the resolver
  overlay merges raw markdown over the stripped copy). Fix committed, pending roll. **Stage 2
  is unaffected by that layer** — `apply_edit` has no resolver overlay.

⚠ **The dangling-address landmine FIRED, within 24h of being filed.** Re-grading run 3's
proposal (`d2378b77`) now exits `no page_component 3eb70551…` — those components were
re-rendered and carry new ids. Resolve a parked proposal by `(page_name, slot_name)`.

### From the `bugfix 083` / 277 lane (live exchange, both directions answered)

- **Their promoter cannot reach `copy_edit_proposed`, structurally.** `checkpoint_for_review_action.go:202`
  hardcodes BOTH `handler_agent='human-review'` and `status='needs_human_review'` in one INSERT,
  so the type cannot be born `detected` by its only producer. Their suggested
  `handler_agent=''` fix is therefore **not available to me** — it is a Go change on a shared
  action (4 agents, 9 items), council + roll.
- ⚠ **Their `held-pair-canary-escalation` would, after 3 days, ask a human to hand-canary any
  `copy_edit_proposed` row that ever did reach `detected`** — inviting exactly the D2 breach.
  They are landmining it; I accepted their offer of an explicit `item_type` exclusion + D2
  citation in the promoter's `pre_query`, as a guard on that narrow residual.
- **`apply_section_edit` has an RFC_015 citation gate** (`acknowledges_decision` /
  `supersedes_decision`) that a hand-filed `section_edit` can silently omit. Mine omitted it —
  **no harm done, and provably so**: the gate returns a SKIP, and all four of my applies
  changed content, so no decision covered those slots.
- **Answered their design question** (`bugfix_277_required_fields_repair/CONTRIB_2026-08-19_reply_…`):
  a one-component-one-defect repair is a **different agent** — stage 2's page-scoped read is
  its whole reason to exist and is pure cost for a named defect — and they should check
  **migration 473** first, which makes a page-rerender the mechanical repair for
  `literal_markdown` with no LLM at all.

### Two self-inflicted evidence losses, worth not repeating

1. **Closing an item by writing `result` DESTROYS the handler's own record** — my four applies'
   `content_edit_mode` / `updated_field_count` are gone. Use `result = result || jsonb_build_object(…)`.
2. **Orchestration rows are not an archive.** The 08-17 apply orchestration **no longer exists**
   (count 0) while rows from 2026-07-13 survive, so it was not age-based. Between (1) and (2)
   there is now no record of those applies except the artefact and this lane's docs.

### Chassis today

**`v1.0.1314` = `d3590ca46`** (both replicas, binary-probed with a negative control; the
startup line had scrolled). Mode-split still in. The 07:51:10Z bulk update to 195 agent rows
was the **release stamping `image_tag`** fleet-wide — `copy-editor`'s `default_config` is
otherwise exactly as 462 left it (max_tokens 32000, budget and voice references intact).


## ⚠ UPDATED 2026-08-18 (evening) — BOTH inbound asks are ANSWERED, and one produced a bug + a landmine

**Neither ask is outstanding any more.** What replaced them:

### `bugs_open/305` — the v2 house voice did NOT reduce define-by-negation

Filed from `portfolio_positioning`'s report. **Two comfortable readings refuted, one on each
side:**

- **Theirs:** the offending copy was written **2026-08-08**, before BOTH the identity-spec fix
  (08-12) and the v2 carrier (08-13). `page_components.updated_at` (08-17) dates the
  RE-RENDER. Neither fix failed on this case. **Any live sentence can be dated:**
  `SELECT id, agent_type, created_at FROM llm_call_log WHERE response_text ILIKE '%<sentence>%'`
  — a component timestamp dates the render; only the call log dates the words.
- **Mine — and then MY OWN CORRECTION, same session.** I reported **2.72 → 2.85** per 1,000
  words ("the carrier did not reduce it"). **Withdrawn.** Adjacent equal-length windows give
  **4.35 → 2.85** — a 34% FALL, the opposite conclusion from the same table — and the weekly
  series runs **0.38 to 4.27 with no trend**, so neither claim is detectable at this sample.
  **Do not quote either figure.** The supportable statement is: *the writer still produces the
  construction at a non-zero rate, and whether v2 moved that rate is not currently
  measurable.* Full correction in `305 §3` and `WRONG_CALLS` 2026-08-18; both other lanes were
  told which sentence to stop repeating.
  - **The cheap check I skipped:** plot the metric over time before quoting two points from
    it (`GROUP BY date_trunc('week')`, seconds). Marking it `[MEASURED]`, stating the method
    and normalising for length all made a wrong number look MORE trustworthy.
  - ⚠ **The `090` symptom was authored from the refuted framing**, so read its verdict against
    this correction rather than at face value.
  - **A within-site pre/post query is the right shape** and was too slow to finish inline —
    it is the measurement someone should actually complete.

**`090` verdict is IN (`57b2dcd2`): `UNVERIFIABLE` — "stopped: scope-not-narrowing", no fix
proposed, "hand to a human, do NOT auto-conclude." And it REFUTED the hypothesis, which is
the run paying for itself.**

- It found the same rhetorical shape on **2026-08-06** (a week before the v2 flip), via a
  different generation than the 08-08 one I found. So v2's wording did not introduce it.
- It asked for the voice text active on 08-06 and could only see the present config. **I
  closed that gap:** the pre-v2 block is preserved in
  `agent_default_configs_bak_20260813_voicecarrier` (a backup THIS lane took shipping v2).
  **Pre-v2: 2,499 chars, does NOT contain "contrasting pair". Current v2: 6,032 chars, does.**
  The construction was produced under a block that never mentions it → **naming it cannot be
  the licence. `305 §4` is dead by two independent routes.**
- ⚠ **Do NOT edit the voice block.** The obvious fix is now positively refuted, not merely
  unproven.

**What the loop could NOT close, and is the next real step:** the same before/after for
**`adoption-tracker` and `protocol-tracker`** — its bundle had no generation history for
either (symbol bodies "unavailable", a tooling failure, and its own data request died with
`canceling statement due to statement timeout`). ⚠ I was running an expensive correlated
query on the same database in that window; I cannot show it caused the timeout, but **do not
run heavy exploratory queries while a diagnosis run is in flight.**

Reading the artifacts (the item's `result` is a SPAWN RECORD, not the diagnosis — follow the
child):
```sql
SELECT result->>'child_orchestration' FROM site_work_items
 WHERE item_type='needs_diagnosis' AND spec->>'dispatch_correlation_id'='57b2dcd2-…';
SELECT collected_data->'call_diagnoser'->'response'->>'conclusion'
  FROM orchestration_states WHERE orchestration_id='74c5ca26-6468-4f89-845c-a72f6d442348';
```

### The landmine that came out of answering `finetuning_uk_service`

**A site's `voice` spec never reaches the writer — it feeds the DETECTOR.** `[MEASURED]`
`tone_guardrails` appears in **0 of 1,338** post-v2 `page-content-writer` prompts, while
`key_differentiators` appears in **214 of the same 1,338** (the positive control that makes
the zero mean something). In code: written by `write_site_plan_action.go:161`, read by
`check_voice_tells.go:238`, no generation-time read. **So register written there changes what
gets FLAGGED, not what gets WRITTEN.** Where it actually lands, in measured order of force:
`identity.key_differentiators` (the lead) → `content_direction.example_phrases.characteristic`
(exemplars beat rules; and the writer reads `.formatted`, not the array) → `strategy`.
⚠ `tone_of_voice` and `voice_and_tone` have **zero code references anywhere** and sites carry
them as current rows.

### Replies filed (both lanes are waiting on nothing from us)

- `portfolio_positioning/CONTRIB_2026-08-18b_reply_your_urgency_holds_but_the_timeline_does_not.md`
  — their urgency stands, their timeline does not; we accepted their offer to rerun the pilot
  pages AFTER a fix is live, with two conditions (date the copy first; bank the before-image).
- `finetuning_uk_service/CONTRIB_2026-08-18_answers_from_copy_quality_two_stage.md` — all
  three questions answered. Notable for us: stage 2's no-invented-figure gate is a real
  guarantee for a priced page, and **the required-links arm is VACUOUS on a page that declares
  none** ("all 0 declared links present" is a pass that checked nothing).

## What happened 2026-08-17/18 (the short version; NOTES has the evidence)

**2026-08-17 — built, and the proof case closed.** Prior-art sweep found the only missing
piece was the page-scoped READ, reachable in config via `query_database`, so no Go, no roll,
no council round. Ran against the owner's preserved proof case (6 guide links missing from
LMC's index since 08-12), proposed six added `<li>` and nothing else, gate passed, owner
approved, applied. `gate_page_links.py` exits 0 and the six links are live.

**2026-08-18 — the harder page broke it, twice, and both faults were real.**
`ai-agent-orchestration.com/index` (8 components, 78,302 chars, a DIFFUSE fault) made the
agent attempt a whole-page rewrite; it truncated at `stop_reason=max_tokens`.
**So 08-17's restraint was a property of a LEGIBLE DEFECT, not of the design** — the
`SUMMARY_2026-08-17` title ("…and it showed restraint") is narrowed by this and a corrected
summary is DUE (I did not write one; the five headings would now differ materially).
Fixed by migration **462**: `max_tokens` 32,000 **and a hard ceiling of THREE edits**,
because an edit set bounded at the source cannot truncate. Re-run: 8,181 tokens, 3 edits,
and the judgement was what page-scoped read exists for — the same pitch restated in FIVE
sections, one resource under FOUR names. Owner approved; all three applied and verified.

## What the gate learned (read before trusting it)

Three holes, all found by USING it on something harder, all the same family — **it reported
"checked" for something it had not looked at**:

1. **`--item` read a shape I had guessed at.** A real proposal nests `review_data.edits[]`,
   one entry per component. Would have failed on every real proposal.
2. **The volume floor could not tell a GUTTED section from deliberate DE-DUPLICATION** —
   which is half of stage 2's remit. **Made to discriminate, not relaxed:** a shrink passes
   only if every removed figure and link is still reachable elsewhere ON THE PAGE; under 25%
   fails outright. The discriminator is mechanical on purpose — keying it on the agent's
   rationale would let it talk past its own gate.
3. **Array fields were unchecked beyond their type**, while printing "1 of 1 type-checked".
   Arrays are now flattened so links/markup/facts/volume apply; item deltas print (`10 → 9`).
4. **A URL written as PROSE was invisible** (the `href` check has no attribute to read).
   Now checked on the same page-scoped standard. ⚠ **Its live control PASSES and proves
   nothing** — on the page that motivated it the URL legitimately survives elsewhere, so the
   FAIL arm is exercised directly instead.

**Rule this lane now works to: after changing a check, re-run the controls.** Every fix above
was followed by proving the gate can still fail. "It complained, so I changed it" is how a
check becomes decorative.


## NEXT STEP, designed but NOT built — the spec-side detector, and the distinction that makes it honest

The owner's second half (*"ensure that that sort of copy never leaves this framework again"*)
points at a detector on the BRIEF, not the writer — `bugs_open/305`'s root-cause section says
why. Design work done 2026-08-19, **nothing built**:

**`content_direction` mixes two kinds of text, and only one is evidenced.** Inspected on
`ai-agent-orchestration.com`:

- **Supplied-for-reuse fields** — the canonical tagline, `avoid_phrases`, `key_terms`,
  `example_phrases` elsewhere. These are handed to the writer to USE. **Proven transfer:** the
  tagline appears in 1,348 rendered prompts and 408 responses. A contrastive construction here
  is a defect with literal evidence behind it.
- **Instructional prose** — `cta_style.approach` (*"…a technical conversation, not a sales
  process"*), `terminology.approach`, `voice`, `emphasis`. These tell the writer HOW to write;
  a contrast here is a reasonable way to give guidance. **Transfer NOT established.**

**So the detector should FLAG the first kind and merely REPORT the second**, and say which is
which in its output. Building one that treats all matches alike would fire on 24 of 25 sites
at 24–38 instances each — a number nobody can act on, and mostly aimed at text whose transfer
is unproven. ⚠ `count_negation_tells.py` as written does exactly that (it counts a whole
document), so it is the wrong tool pointed at a spec until it learns the split.

**Before building it, the falsifiable measurement is still owed** (`305 §D.4`): do the
INSTRUCTIONAL uses transfer to output at all? Design it so it can come out either way —
per-site brief-instance count vs per-site output rate, with the refutation condition stated
BEFORE looking, and with enough calls per site to beat the 0.38–4.27 weekly variance that
already produced one withdrawn finding. ⚠ Attributing calls to sites is the awkward part:
`llm_call_log` has no `site_id`, the join through `orchestration_states` timed out at ~7
minutes, and `prompt_rendered ILIKE '%<domain>%'` is the cheap substitute.


## Next work, in the order that closes doors

1. ~~**The two inbound asks**~~ **ANSWERED 2026-08-18 evening.** What replaced them:
   **(a) read the `57b2dcd2` verdict and act on it** — it is the writer-side half of an
   owner directive, and `bugs_open/305` is open until it lands; **(b) the three directory
   pages are still uncleaned** (`ai-agent-orchestration.com`, another lane's site — coordinate;
   we already edited its index today), and per the owner's instruction the writer fix comes
   first or ~140 planned directory pages inherit it.
2. **A fourth run on `ai-agent-orchestration.com/index`** — the 3-edit budget is a known
   UNDER-fix on a page with five-fold restatement. Does a second pass find the two remaining
   restatements and the negation density, or re-propose what it already did? Cheap, and it
   tests whether the budget is self-correcting or just a ceiling.
3. **Dispatch.** Wiring `content-quality-auditor`'s findings to `copy-editor` is the
   `css-patch-agent` shape the PLAN cites. Not before (1) and (2); a new (item_type, handler)
   pair is held for a human canary anyway.
4. **`bugs_open/033`** (another thread's) still gates ROUTINE operation. It does not gate
   one-off runs — decision 4 (2026-08-15) established that.

## The apply recipe (unchanged, and it works — twice)

Full version with the SQL: `HANDOFF_2026-08-17_continue_here.md`. The short form:

```
0. queue health first     SELECT healthy FROM ai_endpoint_health WHERE name='claude';
1. staleness + re-grade   gate_stage2_edit.py --item <review item>      # reads LIVE rows
2. file section_edit      spec COPIED BY SQL from the proposal row, never retyped;
                          born 'triaged', then CLAIM IT YOURSELF before dispatching
3. publish to section-editor (kcat -P exits 0 having sent nothing — the orchestration row
                          is the only proof of dispatch)
4. verify AT THE ARTEFACT  — 'complete' also means a lock/decision-gated REFUSAL
5. close both items with the evidence on the row — nothing in the platform will
```

**Apply SEQUENTIALLY when a proposal touches several components of one page** — concurrent
`section-editor` runs would race on the render and deploy. Three edits took ~2 minutes.

## Standing cautions (the fresh ones first)

- **A served-page check immediately after a deploy tests PROPAGATION, not the edit.** Mine
  read as a failure and was wrong: the page still showed old text while the stored artefact
  was already correct. Re-fetch, and believe the DB when the two disagree.
- **A parked proposal's `page_component_id` DANGLES.** A rerender REPLACES the component row
  — the proof case's id died within a day (`d6c9198b…` → `f05d59e5…`, content intact). The
  gate then says `no page_component <id>`, which reads like a typo. Resolve by
  `(page_name, slot_name)`; treat a stored id as a hint. LANDMINE filed.
- **⚠ Migration number 462 is DUPLICATED in the directory**: `462_copy_editor_edit_budget.sql`
  (ours, applied 11:59:01Z, in the ledger) and `462_fixer_rerenders_skip_owned_pages.sql`
  (the 283 lane's, committed 13:43 vs our 13:05). The ledger keys on filename so both can
  coexist, but a bare "462" is now ambiguous in conversation. Not ours to renumber.
- **Chassis stamp is a DATED OBSERVATION, never a current fact.** It moved THREE times in two
  days: `v1.0.1305`/`6a782274b` → `v1.0.1308`/`e7e5e4d53` → **`v1.0.1309`/`f0117fb8b`
  (deployed 2026-08-18 15:45Z, both replicas)**. Mode-split ancestry has held throughout, and
  **1309 changes nothing this lane depends on** (no diff in `section_editor_actions.go`,
  `ai_actions.go`, `voicestyle/`, `checkpoint_for_review_action.go`) — `copy-editor` is
  config-only, so a roll cannot affect it either way.
  - ⚠ **Read the log FIRST when the pods are fresh; do not probe candidate commits.** The
    binary stamps exactly ONE sha — the build HEAD — so every ancestor you try returns 0 and
    a list of plausible candidates all read "absent". I burned a round on that today with
    four candidates, one of them the *previous confirmed build*. `logs --limit-bytes=600000
    | grep -m1 '"msg":"build provenance"'` answered it in one call.
  - ⚠ **`grep -acF 000…0` is NOT a negative control — it returned 2.** Forty zeros match Go's
    internal digit table, so a control that matches everything proves nothing. Use a real sha
    that must be absent (current HEAD, made after the build).
- **Re-verify "X does not exist" against the live DB before building X.** This lane's whole
  history is that lesson and it paid three times in two days.
- LMC: never fire `run_improvement_sweep_once.sh`. Check lane activity before writing to
  any site you do not own.

## The five living docs

- **PLAN** — §11 records delivery + three corrections.
- **NOTES** — the evidence log. Read the 08-17 → 08-19 tail; corrections are marked in place.
- **README_where_we_are** — the owner's plain-prose log, newest at the bottom.
- **SUMMARY series** — 08-12 · 08-14 · 08-15 · 08-17 · **08-19 `the_fault_was_never_in_the_writer`
  (newest — start here)**. ⚠ 08-17's headline claim ("it showed restraint") is refuted by 08-19
  as a general claim; the series is the record, so it stays.
- **this HANDOFF.**

**Tooling this lane owns:** `gate_stage2_edit.py` (grades one proposal; `--self-test` runs
every control and MUST fail) · `count_negation_tells.py` (an OBSERVATION, never a gate) ·
`loanandmortgagecalculator_couk/gate_page_links.py` + its `acceptance/` baselines.

**Migrations this lane owns:** `447` (seed `copy-editor`) · `462` (3-edit budget + 32k cap).
Both have `_ROLLBACK` files; 462's restores from `agent_definitions_backup` by
`snapshot_reason`.

**Peer lanes with live threads, all answered, none waiting on us:** `bugfix 083` / `277` (the
promoter; owes us an `item_type` exclusion + D2 citation, deliberately not done on a peer's
say-so) · `bug 184` (owns 473/474 — **do not touch their bug file or migrations**;
contribute in) · `portfolio_positioning` (raised the symptom; warned about their pilot's brief)
· `finetuning_uk_service` (three questions answered).
