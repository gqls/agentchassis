# HANDOFF 2026-08-12 — contrast front CLOSED; two NEW fundamentallyai.com items are what's live

**Continues:** `HANDOFF_2026-08-10_contrast_front_continue_here.md` (read its ADDENDUM 5 and 5a
for the repair itself). **This file is the cold-start.**

**In one line:** `bugs_open/113` is repaired and pod-verified end to end, the council trail is
closed APPROVED, and the owner has since opened **two new pieces of work on
fundamentallyai.com** — a copy defect and a navigation gap — which are the actual next task.

---

## PART 1 — the 113 front: DONE. Nothing is owed.

| item | state |
|---|---|
| council `b8e341b9` | **REVISE → REVISE → APPROVED** (round 3, 2026-08-12 14:17:57Z) |
| the code | live and **pod-verified** — chassis `v1.0.1291`, stamp `da5a7eb8f`, MATCH on both replicas with a bogus-sha control absent; `git merge-base --is-ancestor 9d4fbb4f7 da5a7eb8f` → true |
| the site | `ai-agent-orchestration.com` repaired at the served artefact — `--color-card-bg` `#ffffff` → `#0D1117` |
| the measurement | 58 → **40** against a pre-registered ~15. The miss was the finding |
| register | DES-082 → **live and used**; DES-083 → **observed emitting and dispatching** |

**Commits:** `a36ce9410` (r3 code) · `dbf74bc71` (repair record) · `9d4fbb4f7` (r3 advisory fix,
`Council-Reviewed:`) · `65aaaf7e2` (register) · `a1eb363f7` (sweep: orphaned r2 submission).

### The two leftovers, both genuinely unowned

1. **`extractContentWithFallbacks` and page-content-writer's `extractSiteID`** still carry the
   heuristic two-branch envelope readers that round 3 replaced in
   `install_site_composition_action.go`. Raised by the `bug_historian` seat, which named the
   shape exactly: *"one call site of a shared judgement gets the rigorous fix; the sibling
   stays heuristic."* Small, well-scoped, and the pattern to copy is in that file.
2. **`.team-member { background: #fff }`** — 20 of the 40 residual AA failures on
   ai-agent-orchestration.com, a hard-coded literal in a component template that no palette
   change can reach. Same family as D4 and `features_open/026`.

### Do not re-derive these

- **A `needs_composition` repair is TWO work items.** The composition half completes, changes
  the DB, and queues nothing; `styles.css` is rendered by `webdesign-agent` off the
  `needs_design` half. Every DB check reads green while the site serves the old sheet. In
  `LANDMINES.md` with 12 footprint keys.
- **`grep 'build provenance'` on chassis logs matches the phrase inside AGENT PAYLOADS.** Our
  own landmine text about build provenance is synced to `doc_notes`, council seats read
  `doc_notes`, and their `collected_data` goes through the logs — so the grep returns a 1.4 MB
  orchestration dump and looks like a hit. Anchor on the field:
  `grep -o '"msg":"build provenance","git_commit":"[0-9a-f]*"'`. **Found today, not yet a
  LANDMINE entry — worth adding.**
- **`landmines-sync.py` splits footprints on COMMAS/semicolons only** (`landmines_lib.py:51-69`),
  **not** on the `·` several recent entries use. A `·`-separated footprint list becomes ONE
  run-on `subject_key` the SessionStart hook can never match against a dirty file. Mine is
  fixed; **the other `·` entries are worth a sweep.**

---

## PART 2 — THE ACTUAL NEXT TASK: fundamentallyai.com

Raised by the owner on 2026-08-12 while reading the live site. Both are diagnosed; neither is
fixed.

### 2a. The copy reads AI-generated, and the defect is in the SPEC, not the sentences

**Owner:** *"please relook at the copy as it looks a bit AI generated with the negative framing.
We have solved this elsewhere so there should be nothing to invent here."*

**The symptom, measured on `/platform-log/index.html`** — seven negative-definition
constructions, one in almost every entry:

| where | the clause |
|---|---|
| intro | "the decision record is kept, **not summarised** after the fact" |
| intro | "write about the mechanism **rather than** the message" |
| LLM Cost Calculator | "a number to argue with, **not a demo to admire**" |
| Review Council Simulator | "a teaching tool **rather than** the production council… illustrative, **not a guarantee**" |
| AI Readiness Checker | "a starting conversation **rather than a verdict**" |
| Automation Savings Estimator | "a starting estimate, **not a forecast**" |
| Model Approach Selector | "meant to narrow the decision, **not make it for you**" |

**THE RULE ALREADY EXISTS — do not invent one.** Settled by the owner on 2026-08-11 in the
`mortgagecalculator_couk_adoption` lane (commit `82d510bb5`, NOTES §21):

> **Say what a thing IS, not what it is not.** A negative definition makes the reader do
> subtraction and reads colder, because it withholds. Added to `things_to_emulate`.

**THE CAUSE, and it is the same shape as that lane's finding — *"the defect was in a RULE I
wrote, not in the sentence."*** `fundamentallyai.com`'s `content_direction` spec **teaches the
X-not-Y construction by example**, and examples outweigh rules for a writer:

- `example_phrases.characteristic` — **3 of 4** use it: *"the decision record is real, **not a
  log entry**"* · *"That is **not a marketing story** — it is what happened"* · *"a demonstrated
  property of this platform, **not a promise**"*
- `heading_style` worked example: *"A review council that ships decisions, **not opinions**"*
- `persuasion.method`: *"Evidence-first, **not assertion-first**"*
- `explanation_pattern`: *"…**Acknowledge any current limitation plainly.**"* ← **this is the
  honesty requirement and it must survive.** The tool guides are being honest, correctly. The
  defect is that the only construction they have been shown for honesty is X-not-Y.

**So the fix is not to delete the caveats — it is to give the positive form.** "Treat the
result as a starting conversation" already says everything "not a verdict" says, warmly.

**⚠ THE TRAP THAT WILL WASTE YOUR EDIT.** The writer reads **one** field —
`{{.site_specs.specs.content_direction.formatted}}` — a flattened string auto-generated from
`data` when the spec is written (`site_spec_actions.go:206-215`). **Editing `data` alone is a
dead edit**; the prompt keeps reading the old `formatted`. Write the spec through the action
that re-formats, or update both together and verify `formatted` actually changed.

**And obey the standing ruling:** *the FRAMEWORK writes the content, not you* (owner 08-06).
Fix the spec and let it regenerate — do not hand-edit the served pages.

**Suggested spec change (not yet applied):**
1. Add to `things_to_emulate`: the positive-definition rule, in the owner's own words.
2. Rewrite the three `characteristic` example phrases into positive form — they are few-shot
   examples and they are doing most of the damage.
3. Leave `explanation_pattern`'s "acknowledge any current limitation plainly" **intact**, and
   add a positive worked example beside it so the writer has a pattern to copy.

**Verify at the served page, and the whole page set** — this is a template tic, so it will be
on more than the Platform Log. `curl -s <page> | grep -oE "[^.]*\b(not a|rather than)\b[^.]*\."`

### 2b. The tools are unreachable from the site's own writing — FILED

**Owner:** *"we need links to the tools from the platform log and elsewhere, and a tools entry
in the top nav would probably be nice."*

**Measured at the served page, not inferred.** `/platform-log/index.html` carries six *"How to
Use the &lt;tool&gt;"* guides. Each names its tool in the first sentence. The page contains
**exactly one** tools link — `/tools/llm-cost-calculator.html` — and it is in the **footer**
quick-links, where it renders as **"Llm Cost Calculator"** (a title-case helper applied to an
acronym; a small separate defect worth fixing in passing). **The six guide cards link to none
of their tools.** Top nav is Home / About / Contact / Platform Log / News / Capabilities —
**no Tools entry.**

**FILED as the experience-loop intake** — work item
`458f53a1-ab47-4c14-95aa-396bd13d60af`, `item_type='needs_experience_plan'`,
`item_key='needs_experience_plan:tools-are-unreachable-from-the-writing'`,
`pipeline='experience'`, `status='triaged'`, shape copied from the existing rows. The spec
carries all three asks and the measurement above.

**⚠ FILING IT DOES NOT RUN IT — this is the half still owed.** There is **no enabled
`scheduled_task`** targeting the experience loop (checked: zero rows matching
`%experience%` on name or target_agent_type). The loop *does* work — 2 of the 5 historical
`needs_experience_plan` items reached `complete` — because the convention is a **durable
intake item PLUS a kcat `orchestrate` envelope on the same correlation_id**
(`RUNBOOK_experience_loop.md` G17, on the 090 template). **Next session: fire that envelope.**
Note the standing landmine — **`kcat -P` sends NOTHING at exit 0**, so foreground-test it and
confirm the orchestration row appears; do not read a clean exit as a dispatch.

---

## Cold-start order for a fresh session

1. **Read Part 2 of this file.** Part 1 is closed and needs no action.
2. **2a first** — it is the owner's direct request, the rule is already written, and the
   `formatted`-field trap above is the only thing that makes it non-trivial.
3. **2b second** — the item is filed; it needs the dispatch envelope, then verification at the
   served nav and the served guide cards.
4. Both touch the same page set, so **sequence them so one re-render serves both** rather than
   rendering fundamentallyai.com twice.
5. Do **not** hand-write page copy (owner ruling 2026-08-04 / 08-06). Spec, then regenerate,
   then verify at the artefact.
