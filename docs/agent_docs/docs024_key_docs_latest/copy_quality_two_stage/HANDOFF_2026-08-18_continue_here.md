# HANDOFF 2026-08-18 — continue here

**Lane:** `copy_quality_two_stage`. **Supersedes `HANDOFF_2026-08-17_continue_here.md`**
(kept for its worked example and the apply recipe; its state lines are stale).

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
- **Mine:** it is not a fossil. Same `agent_type`, split at the v2 flip, normalised per 1,000
  words with mean response length identical (222 vs 223): **2.72 → 2.85.** The carrier did not
  reduce the construction. Uncomfortable, because CQ-022 is this lane's own delivery.

**Root cause is NOT asserted** — `090` run **`57b2dcd2-2ded-473c-9f2e-617176f39c15`**, filed
per the 2026-07-31 ruling. **The next session should read that verdict first** (it was still
at its `verdict` step when this was written):
```sql
SELECT status, result FROM site_work_items WHERE item_type='needs_diagnosis'
  AND spec->>'dispatch_correlation_id'='57b2dcd2-2ded-473c-9f2e-617176f39c15';
SELECT body FROM doc_notes WHERE body ILIKE '%57b2dcd2%' ORDER BY created_at DESC LIMIT 1;
```
⚠ **Do NOT edit the voice block on the strength of the hypothesis in `305 §4`** (that naming
the construction licenses it). It is labelled a hypothesis. Editing on that guess is this
lane learning "exemplars beat rules" a third time by doing it wrong again.

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

PLAN (§11 = delivery + three corrections) · NOTES (read the 08-17 → 08-18 tail; corrections
marked in place) · README_where_we_are (owner's plain-prose log) · SUMMARY series
(08-12 / 08-14 / 08-15 / 08-17 — **08-17's central claim is narrowed by 08-18 and a new one
is due**) · this HANDOFF. Two inbound CONTRIBs dated 2026-08-18 are unanswered.
