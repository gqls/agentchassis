# NOTES — bugfix 238

Append-only, newest at the bottom. The missteps are the point, not an appendix.

## 2026-08-10/11 — session 1 (the whole lane, so far)

### Picking it up

`who-owns.py 238` returned **OWNED or recently active** — the finetuning lane
filed it and has commits within 24h. That is a lagging signal by construction, so
I read the owning lane's handoff instead of trusting the verdict either way, and
its §8.4 says in terms: *"`bugs_open/238` is the whole of the remaining imagery
work … Fix candidate 1 there … is the durable answer"*, addressed to a fresh
thread. Not competing. Also checked `git status --short` for in-flight Go work in
the bug's code paths (the LANDMINE about commit-blind ownership checks) — clean.

### The finding that changed the whole shape of the fix

The bug file says the generator "reproduces the ones that look like copy and
drops the ones that look like plumbing". **That is wrong**, and I nearly built
the wrong fix on it — a "preserve structural keys through the LLM round-trip"
design, which would have been machinery for a journey those keys never make.

The 11 lost keys are exactly the component's **non-`llm`-sourced** fields. The
LLM is *forbidden* from emitting them ("exactly the keys listed"). What actually
happened: the declared sources resolved nothing, `on_missing` defaulted to
`skip_field`, and `save_page_sections` replaced the row wholesale.

The tell I should have reached for first, and did not: **the lost set was
suspiciously exact.** Not "some images", not "the ones the model forgot" — every
key ending `_url` and nothing else. A model dropping things it finds boring does
not produce a clean partition of the schema by a field nobody showed it.

### The measurement that made it certain

`page_component_history` still holds the last good row (58 keys, the snapshot
`save_page_sections` took *immediately before deleting it*). Diffed against the
live 47, cross-referenced against `input_schema`: the 11-key diff is precisely
and only the non-llm fields, and **zero** llm fields were lost. That is
disconfirmable — a generator that "drops plumbing" would have lost some alt texts
too, or kept some URLs.

### Wider than filed

While verifying the served page I checked for anchors as well as images. The five
"Read case study" links and the section CTA were **gone** — not empty, gone,
because the template gates them with `{{if}}`. Three of the four checks I would
naturally write (count `<img>`, count `src=""`, grep the CSS class) see nothing
wrong. **A gated field fails more quietly than an ungated one**, which is the
opposite of the intuition that gating is the safe pattern.

### Misstep 1 — I wrote SQL against a schema I had truncated

`\d page_component_history` piped through `head -25`. The column list fitted; the
CHECK constraints did not. My manual snapshot INSERT used `op = 'UPDATE'`, which
`pch_op_check` rejects (`'overwrite' | 'delete' | NULL`), and the transaction
aborted. Cost: one round trip. **`\d` piped through `head` truncates exactly the
block you are about to write SQL against** — the constraints print last.

### Misstep 2 — the FK named the opposite of what its name implies

Second attempt failed too: `page_component_history.component_id` is a FK to
**`page_components(id)`** — the row — not to `content_components(id)`, even
though `page_components.component_id` means exactly that. And it is
`ON DELETE SET NULL`, and `save_page_sections` archives a row then deletes it, so
**every historic row reads NULL** — which is why the column looked unused and
tempted me to fill it with the wrong id in the first place.

Both are now a LANDMINES entry. Both were caught by the transaction rather than
by me, which is the argument for `BEGIN` + `DO/RAISE` over a bare `UPDATE`.

### Misstep 3 — a test that passed twice and was a coin flip

`plan_sections_structural_carry_test.go` passed on its first run and on a
targeted re-run, then failed in the full suite. Not flaky infrastructure: **`planSection`
iterates the schema's fields MAP**, so the order in which sources resolve — and
therefore the order the queries arrive in — is randomised per run. My ordered
sqlmock expectations were a coin flip. Fixed with
`MatchExpectationsInOrder(false)` and a comment saying why.

Two clean runs did not establish stability, and I had already half-believed them.
The general form is in memory as *"2 clean runs cannot establish STABILITY"*; the
specific trigger worth remembering is **production code iterating a map ⇒ any
ordered mock expectation is probabilistic**.

### Misstep 4 — an inherited claim about the result payload

I wrote, in a doc comment, that the new result keys are "absent from the payload
entirely when nothing happened" — copying the shape of the neighbouring
`source_aliases_used` comment. **A nil typed map inside `map[string]interface{}`
is not `nil`**, and a nil map marshals to `null`, not to absence. My own test
caught it before it shipped (it asserted the wrong thing and failed), and I
corrected the comment to say `null`. Small, but it is the "inherited a claim
without re-deriving it" shape that `WRONG_CALLS` 2026-08-10 records one lane over.

### What I did NOT do, and why

- **No `090` diagnosis run.** The standing rule (owner, 2026-07-31) is that a
  `bugs_open/` file asserting a cross-cutting root cause is not filed until it has
  been through the loop *or the session states plainly why it substituted
  equivalent first-hand verification*. This is that statement: I am not filing a
  new cross-cutting claim, I am **correcting an existing file's claim** with the
  primary artefact — `page_component_history`'s own before/after rows, the
  component's `input_schema`, and the four code sites read end to end. The
  disconfirming result was available and did not occur (any llm field in the lost
  set would have refuted me).
- **No sixth guard in `save_page_sections`.** `bugs_open/178` left a standing
  instruction that a further floor there is the trigger for a unified
  content-loss detector as its own submission. The carry is not a floor and is
  not in that function.
- **Did not widen `sectionHasImageField`** — measured benefit zero, and it writes
  live-path data.
- **Did not re-scope `image_url_404`'s site-wide empty-src key**, though it is a
  real defect (finetuning's `blocked` row has held the fleet slot since 08-03).
  The tally is site-wide by construction — its query does not even select the
  page — so per-page keying is a restructure with its own volume question, on a
  check that cannot currently fire. Recorded as a follow-up rather than half-done.

### Repair, and the order it had to go in

Data first (`378`), then a **no-LLM** re-render (`379`, `reason:
section_data_resolved`). A regenerating rebuild is the operation that caused the
bug, and `needs_page` has a poor record on this site (5 failed / 4 wont_fix / 2
rejected against 20 complete). The repair regenerates nothing, so it could run
before the code fix and did.

Applied by hand rather than through `run-migrations.sh --apply`: the runner takes
**every** pending file, and other threads have several. Recorded afterwards with
`--record-only`.

### Council

Two submissions, deliberately separate: `bd38df2e` (prevention) and `98852baa`
(detection + the predicate widenings). Bundling them is the breadth RFC_016
slice 1 was rejected for, and they have genuinely different blast radii — one
adds a resolution source and can only re-supply bytes the page already served;
the other can refuse a live rebuild. Committed with `Council-Submitted:` on both,
per the trailer rule, because holding code for a verdict is not available on a
shared HEAD.

### My own LANDMINES entries were swept into another session's commit

`5c3322aa8` ("correct(168 sweep + 244)") carried my two new LANDMINES entries.
Nothing lost — they are at HEAD — and forward-only holds. Worth recording because
it is the *documented* hazard working exactly as documented: a pathspec commit
protects you from other sessions' files, and cannot protect your own edits inside
a file someone else commits. The 9 lines still uncommitted in that file when I
looked belong to the `diagnosis_schema_visibility` lane; I left them.

### CLAUDE.md changed under me, mid-session

The deploy-verification recipe was rewritten on 2026-08-11 (another lane's work,
`bugs_open/249`): binaries now carry a build-provenance stamp, and `strings` is
explicitly deprecated — it produced three confidently wrong readings in one day.
I had already written `verify-later` lines using the old `strings` recipe into
both register entries and into `380_..._HOLD.sql`. Rewrote all three, and marked
the PBP-039 one as a same-day correction rather than silently editing it, so the
next reader can see the recipe moved rather than assuming I chose badly.

### Council verdict A — APPROVED, and the objection that corrected me

`bd38df2e`, round 1, 12 reviewers, 8 advisory objections, none high-severity.

**`editquality` caught a real error in my own framing, not in the code.** I had
named ai-agent-orchestration.com `/index.html` as "the honest acceptance case"
for the carry. It cannot be: **the carry reads the CURRENT deployed row, not
`page_component_history`**, and that row is already stripped — so it yields a
`STRUCTURAL_KEY_CARRY_MISS` and no repair. **The fix is prospective only.** It
protects rows that still have their keys and remediates none of the damaged ones,
including the one in the bug's title. Had I left it, a correct null result on that
page would have read as a working fix — the exact shape of a check that cannot
come out false.

Corrected in the register (visibly, struck through) and in the bug file, and the
real population measured: **3 deployed `case-studies-grid` rows still hold
`card1_image_url`** — aao `/enterprise-reference-deployment.html`, finetuning
`/index.html` (repaired), leopardess `/who-we-help.html`. aao is the cleanest
test: 0 spec aspects, 0 current-plan imagery, so its sources cannot resolve and
the carry is the only thing that could preserve those values.

**`debug_historian` asked a question I should have asked myself**: is
`build_status='deployed'` load-bearing on `page_components`, given the fleet's
history of status columns that lie (`site_components.build_status` is `'rendered'`
and never `'deployed'`)? **Measured: deployed 1138, pending 171, approved 23,
removed 1.** It holds. But I had scoped the query on an assumption and shipped it
untested — the cost of being right was zero and the cost of being wrong would have
been a silently inert fix.

**Three seats — `reuse_agent`, `constitution`, `prior_art_librarian` — each
independently asked whether `load_current_section_content_action.go` already
provides this.** I had answered that during exploration (gated on
`mode == "edit_live"`; selects `rendered_html` only, never `content_data`) and
then **left it out of the submission**. Three seats spent effort rediscovering it.
**A deferral or a rejected alternative is only credited if it is stated** — the
answer is now in the register entry, since the code comment does not carry it
either.

**`guardian` raised a failure mode I had not considered**: the carry re-supplies
whatever the stored row holds without checking it is *correct*, so a previously
fabricated value (the `bugs_open/203` phantom `/contact.html` class) becomes
self-perpetuating while its source stays unresolvable. Bounded — the value was
already being served, and live resolution wins the moment the source works — but
genuinely new, and recorded rather than argued away.

**`architecture` returned `needs_rfc`**, on the ground that a new resolution tier
on a two-consumer shared function plus a deliberate cross-path semantic change is
RFC-shaped, and that this compensates for the REPLACE/MERGE split rather than
closing it. **Recorded, not rebutted — a scope objection is not answered by more
measurements.** It did not block, and review here is after-the-fact by design.

Lesson for the next submission: I wrote a long "risks" block and still omitted the
three things reviewers actually asked for — the alternative I had already ruled
out, the scoping value I had assumed, and the limit of what the fix can reach.
**The risks block was full of what I had thought about; the gaps were in what I
had checked and not written down.**

### Council verdict B — REVISE, and the objection was the one I had argued around

`98852baa`, 13 reviewers, gated by one high-severity objection from
`bug_historian`: **I guarded one call site and called the class handled.**

That is `bugs_closed/021` / `bugs_open/093`'s shape, and it is the seat's whole
remit. My submission *named* `RenderTemplateWithMap` as "exempt and named" and
stopped there — which is the move that reads as diligence and isn't one, because
naming one exempt sibling implies the others were checked. They had not been.

**Answered by audit, not argument. The score is 3 guarded of 11.** Unguarded:
`assembleComponents`, `applyContentEdit`, `applyComponentSwap`, `RenderHeader`,
`RenderFooter`, `RenderHead`, the legacy head template, and
`RenderTemplateWithMap`. The list is now in `dead_url_guard.go`'s own header, so
it is read where the guard is read rather than in a verdict artifact nobody opens.

I still did not widen to all eight, and that is a defensible answer only because
it is now a *stated* one with the reasoning attached — the three chrome renderers
have their own response one layer up, the two editor paths edit a row a human
holds, and making `RenderTemplate` itself the reporting form fleet-wide changes
the primitive every render flows through.

**Four seats objected that my censuses were unverifiable by them** — the tables
are outside their schema access. Fair, and the fix is to attach the numbers, not
to insist. All three re-run and attached: 26/12/2/2 by declared type, exactly one
live `render_component` agent, and both squatting `empty-src` rows (finetuning
`blocked` count 16 since 08-03, aao `detected` count 19). **Every one reproduced
exactly.** The lesson is not that the numbers were wrong — it is that a number a
reviewer cannot re-derive is an assertion to them however carefully I measured it.

**`debug_historian` found a real hole in the migration**: it updates by `type`
with no row tie-breaker, and four agent types here carry two active rows of which
only the higher version loads. `page-content-writer` has exactly one today, so it
does not bite — but the HOLD file now refuses rather than half-applying if a
second appears before the flag is lifted. Cheap, and I should have written it
that way first: the landmine exists precisely because "by type" looks total.

**`reuse_agent` asked for a shared emit helper instead of a sibling.** Not done,
with the reasoning recorded: the primitive that carries the semantics
(`insertWorkItem`, and its dedup/two-strike behaviour) is already shared by both
emitters; what differs is the item's shape and its response, which is per-surface.
The `architecture` seat named the right trigger — **a third dead-control emitter
is when this consolidates**, and that is now written where the next author will
be standing.

**Conceded outright:** record-only on the rerender path is worth more than
discarding the finding and less than detection, and I should stop describing it
as "the fleet's only live detection" while the rotations are paused. Nobody is
watching the queue it writes to.

The pattern across BOTH verdicts is the same and worth naming: my risks blocks
were full of what I had *thought about*, and every objection that landed was
about something I had *checked and not written down*, or *assumed and not
checked*. Length is not coverage.

### Council round 2 — REVISE again, and this time it found a real defect

`98852baa` round 2. Gated by `editquality` (HIGH) on the migration's `jsonb_set`
path being **asserted rather than checked** — and with `create_missing := true`,
a wrong path inserts a new branch and reports success, arming nothing while every
reader says "armed". I checked: **the path was right.** But the objection was
about process, not luck, and it was correct — so the file now asserts each path
resolves to a `render_component` step *before* any write, which makes mis-nesting
structurally impossible instead of verified-once-by-hand.

**The find of the whole lane came from `debug_historian`, and it came via my
METHOD rather than my number.** My coverage census was
`default_config::text LIKE '%render_component%'`. The seat pointed out that `_`
is a SQL wildcard — a fair but minor point. Re-measuring properly exposed the
real problem: **that query counts AGENTS (answer: 1), and the question was how
many STEPS. The answer is TWO** — `render_section` *and* `render_from_template`,
both inside page-content-writer.

So **"one config key is full live coverage" was false**, and my migration would
have armed half the render path while the register, the file header and the
coverage report all said "armed". Nothing in the surrounding prose — and there
was a lot of it — could have caught that. **The wrong instrument concealed the
number, and I quoted the number three times.**

Fixed: the migration arms both, and the verify block now **counts** every
`render_component` step and demands they all carry the flag, so a future third
step fails loudly rather than shipping unguarded.

**And I induced the failure to prove the block is not decorative.** Mutated the
file to arm only `render_section` and ran it: it raised
`1 of 2 render_component step(s) armed` and rolled back. A verify block that has
never failed is a claim, not a control.

**Three seats independently said the record-only emit should be gated** —
unconditional DB writes on a shared repair path, while the refusal one file over
was already opt-in. They were right and the inconsistency was inside my own
patch. Now behind `record_dead_url_controls`, default OFF.

**`bug_historian` also checked one of my exemptions and it was false.** I claimed
the chrome renderers "already have a dead-control response one layer up". True
for `render_site_components_action.go`; **false for `rerender_pages_actions.go:532`,
which renders the head template with its own bare `RenderTemplate` and never
routes through that guard.** Corrected in the header rather than dropped.

**Not resubmitting a third time on the remainder.** 8 of 11 call sites stay
unguarded; making `RenderTemplate` itself the reporting form is a change to the
primitive every render flows through, and that is RFC-shaped. The seat's own
phrase — "a documented remainder, not a closed exposure" — is the honest label,
and re-arguing it would be answering a scope judgement with more evidence, which
this repo has already ruled does not work.

**The lane's real lesson, across three verdicts.** Every single objection that
landed was one of two shapes: something I had **checked and not written down**, or
something I had **assumed and not checked**. Not one was about the design. My
risks blocks got longer each round and caught none of them — because they
enumerated what I had considered, and the failures were all in what I had
measured badly or not at all.

---

## 2026-08-20 — lane resumed. Armed the detector, filed the RFC, measured the population, and did NOT ship the code I planned to

Session working `bugs_open/238` end to end. Ownership re-checked first (`who-owns.py 238` → this
lane, dormant since 08-11; no open work item or diagnosis run on the target).

**Order of events, including the two things I got wrong.**

1. **Verified the bug is still real, and found the shape had changed.** Prevention live (merge-base
   against v1.0.1317's revision label, controls both ways). Detection code live but **armed
   nowhere** — 0 agent_definitions rows with either key, 0 `dead_url_control` items in all history.
   The register still said "INERT until the roll; config HELD", false in both clauses since 08-12.
2. **Declared the config key before arming it** (`bb6600e48`). `record_dead_url_controls` was
   undeclared on `RerenderPageSectionsInputSpec` — same omission `refuse_dead_url_controls` had on
   `RenderComponentInputSpec` until 08-19, same cause (read via a helper taking `config`, so a
   function-body grep cannot see it). Checked `CheckConfig` vs `StrictConfig` **at the deciding
   arm** first, because the same mistake on a StrictConfig spec cost the fleet 33 minutes of
   page-publishing the day before. Warn-only here. Test + mutation M1 (undeclare → fails with the
   exact report text the arming would have produced).
3. **Armed the record half** (migration `504`, applied; 1/1 steps fleet-wide; negative control
   confirming the refusal stays unarmed; recorded in `schema_migrations`).
4. **MISSTEP — the number.** Wrote it as `497`. Two other lanes took 497 AND 498 while I wrote;
   the tree was at 503. Caught by the runner's dry run printing the collision adjacently, not by
   any deliberate check. Renumbered to 504 before any apply. The council submission had already
   gone out naming 497 and forward-only forbids amending it, so the reconciliation is in 504's
   header. `WRONG_CALLS` 2026-08-20: allocate the number immediately before naming the file.
5. **Filed `RFC_042`** — the content_data write-discipline split, discharging the `architecture`
   seat's `needs_rfc` on PBP-039 from eighteen days ago that never became a file. Nine writers, one
   carried funnel. Flagged to be answered jointly with `RFC_008` (same question, sibling column).
   ⚠ RFC_041 was taken by another lane between reading the free number and writing — same
   collision as the migration, one hour apart. Claimed 042 on disk *before* writing it.
6. **MISSTEP — the history probe, wrong in both directions inside ten minutes.** Loose (page_id
   only) it over-counted by slot and nearly produced a fabricated "new regression" on
   gamesdesign; strict (slot + trigger source) it under-counted by writer and returned 0 for all
   25, contradicting §7's correct claim about aao. Full account in `WRONG_CALLS`. The working
   discriminator is content identity. **What caught the second one was the bug file disagreeing
   with my query** — and "0 of 25, nothing recoverable" is the comfortable answer, because it makes
   the remediation smaller.
7. **Measured the remediation population BEFORE building it, and it refuted the plan.** 0 open
   `required_fields_missing` items are resolver-sourced; 0 of the 25 carry-miss field slots have a
   `site_specs.*` source that resolves today. The class is "the source has never existed on this
   site". The router cannot help because the producer deliberately never files these — **the gap is
   in the detector, not the router.** Building the routes first would have shipped and moved
   nothing.
8. **Ran the `090`; it came back UNVERIFIABLE** (iteration-cap; it could not read `planSection`,
   `storedFieldValue` or `carryStored` at all — the symbol search missed them and it said UNKNOWN
   rather than concluding, correctly). Its substantive observation corroborated the census from the
   work-item side. Its "what would settle it" was a query, so I ran it.
9. **That query produced the session's best result** — and it is the acceptance test PBP-039's
   `verify-later` has been owing since 08-11. 66 non-llm field losses across archived generations,
   all `renderer`/`static`, all 08-11 → 08-14 18:36, **none since**, against 3,033 archived pairs.
   The `bugs_open/268` fix landed 08-14 09:13. **Proof on ordinary fleet traffic rather than one
   induced case.** Window stated (archive starts 08-09) because it bounds the claim.
10. **Did NOT ship the two carry-gap fixes.** Both are real in the code and I can cite the branches;
    neither has a single observed instance (0 loss events for those source families; 2 empty spec
    values fleet-wide; the 090 agreeing). Three lines of Go and a mutation-proved test would have
    looked like diligence and would have been sized from a code reading. Recorded in RFC_042 option
    (e) as "reachable by reading, unobserved in production", with the query that would justify
    shipping them written down.
11. **Council REVISE, acted on rather than argued with.** `reuse_agent` was right that I had not
    checked overlap with the existing detectors — and measuring it found something better than a
    defence: `check_image_url_404.go` has ZERO `href` handling, so the vanished-anchor class has no
    other detector at all, while on `<img src="">` the overlap is real and aao already holds an
    unworked `image_url_404:empty-src` item for the exact page. Also wrote the `doc_notes` decision
    row (`tooling_provenance`), named the 1-of-11 partial coverage explicitly (`bug_historian`), and
    the register correction (`editquality`'s "missing") which I had already committed before the
    verdict landed. Resubmitted on the same correlation.

**The through-line, and it is the same lesson twice:** the two things that would have wasted the
session — building the router, shipping the carry fixes — were both stopped by measuring the
population first, and the thing that produced the most value (proving 268 works) came from taking a
diagnosis loop's "what would settle it" literally and running it.

### 2026-08-20 addendum — the arming is WIRED-VERIFIED without touching a live site, and the demand control is NOT fired (owner call)

**The wiring question is the one that has bitten this exact file before.** Council round 1 on
migration `473` objected that `rerender_page_sections_action.go` might read
`params.ExecutionContext.Config` while the migration wrote to the STEP-level config — *"if step
config and ExecutionContext.Config are not the same map (they are two distinct maps), the flag is
written where nothing reads it"*. That objection applies verbatim to `504`.

**Settled by reading, and by a working sibling at the identical path:**
`recordDeadURLControls(params.StepConfig.Config)` (`:709`) reads the **same map** as
`shouldStripLiteralMarkdown(params.StepConfig.Config, reason)` (`:287`, `:613`) — and
`strip_literal_markdown` is already `true` at
`{workflow,steps,rerender_sections,config,strip_literal_markdown}`, visible in `504`'s own BEFORE
output, shipped by `473`, and documented as live and working. **So the key is in the map the code
reads, proven by a sibling flag that is already working at the same path in the same map** — not by
argument, and without a production dispatch.

**What is still unproven, stated as such:** the emit has never *fired*, because nothing has
exercised it. Measured after arming: **0 `page_rerender` items and 0 archive rows since 14:00 UTC**
— the fleet is simply quiet. So the standing "a sustained zero has two readings" warning is live
right now, and the reading is "no traffic", which the archive count independently establishes.

**The demand control is NOT fired, deliberately, and it needs the owner's word.** `504`'s WATCH
section proposes one 379-shape `page_rerender` at ai-agent-orchestration.com `/index.html` — the
page with five bare `card*_image_url` fields inside `src=`, which would force `missingBareFields`
non-empty and make the emit fire or prove it does not. Two reasons this thread did not do it:

1. **It is an outward-facing action on a live customer site**, not a read. It re-renders and
   redeploys a real page to test our own detector.
2. **`bugs_open/229` is the specific risk**: a rebuild silently discards hand-patched
   `rendered_html` with no divergence warning. Nobody has established that aao's homepage carries
   no hand patch, and the sibling lanes have hand-patched pages on other sites this month. The
   re-render path merges `content_data` and cannot lose a KEY — that much is structural — but that
   is not the same claim as "cannot overwrite hand-edited markup", and conflating the two is
   exactly the shape this bug family keeps punishing.

**So the honest position: the arming is verified at the config and at the code path, and unverified
at the artefact, with the one experiment that would close it named, costed and waiting on a
decision.** Check for it having fired naturally first — the fleet will not stay quiet for long:
`SELECT count(*), max(created_at) FROM site_work_items WHERE item_type='dead_url_control';`

---

## 2026-08-21 — the owner answered, and answering it turned up a live defect nobody had seen

**Owner ruling 1:** *"yes, that email should appear on contact pages"* → migration **525**
repoints `contact-block.contact_email` from the unreachable `site_specs.contact.email` to
`site_specs.identity.email`. Applied, verified, council `972a82ad`.

**Then I fetched the pages, and the row census had been wrong in BOTH directions.**

1. **The email was already there.** Both contact pages serve the `mailto:` today while
   `content_data` holds no `contact_email` key. `bugs_closed/140` recorded the identical shape on
   idea.uk: the stored artefact carries a value the data has lost. So the page was a *fossil* —
   correct on screen, unreproducible underneath, and the next regeneration would have deleted it.
   525's real effect is to make it reproducible, not to make it appear.
2. **The same component was serving `<a href="tel:"></a>`** — a dead telephone control — on
   **6 rows across 3 sites**, right now. That is this bug's own class on a different attribute, and
   **no row-level query I ran on 08-19, 08-20 or 08-21 could see it.** Only `curl` did.

**Ruling 2:** *"please go ahead and do both"* → migration **538**: gate all three
`cb-detail-item`s on their own fields AND repoint `contact_phone`/`contact_location` at the
identity spelling. Applied, verified, council `1c8aed61`. Six re-renders queued
(`reason=section_data_resolved`, all six pages checked `machine_made` first).

### Why both halves, since each looked sufficient on its own

Neither is. Measured: `sites.phone` is populated for **leopardess only**. So repointing alone fixes
one site and leaves robot-hands and gamesdesign serving the dead control; gating alone fixes all
three sites' dead control and leaves leopardess's real number unpublished. The general lesson is
worth keeping: **"the value is unreachable" and "the template assumes a value" are two defects
wearing one symptom**, and fixing the one you noticed first leaves the other live.

### Three judgement calls, recorded because each could have gone wrong quietly

- **The gate wraps the whole `cb-detail-item`, not the value.** Gating the value alone leaves the
  icon and label over nothing — `bugs_closed/111` exactly. I only got this right because the
  `on_missing` LANDMINE says to read what ENCLOSES the field first, and I read it before editing.
- **All three items gated, including the email.** gamesdesign has no email either, so gating only
  the phone would have shipped `<a href="mailto:"></a>` on its next rebuild. Fixing one attribute
  while arming the one beside it would have been absurd, and it was not obvious until I listed the
  per-site values.
- **I did not repoint the phone on my own authority.** Publishing a number is not a config change,
  and `bugs_closed/140` had already traced this one as the owner's own with *"whether six
  businesses should share one number"* recorded as an owner question. Asked; authorised.

### Answering the pre-commit architecture signal, rather than ignoring it

The hook flagged *"migration + platform code in one commit — needs a staged rollout order"* on
`ce3c28da1`. **It is a point fix and the staging concern does not apply**: the only `platform/`
file is a **test**, which ships no behaviour and cannot be out of order with the migration. Had it
been production Go the signal would have been right, and the migration would have needed holding
until the roll — which is exactly the 494 shape from 08-19. Recorded so the next reader can see the
signal was read rather than skipped.

### The measurement discipline that paid, and the one that failed

**Paid:** a Go test against the real renderer with a mutation — un-gate the phone item and it fails
printing `<a href="tel:"></a>`, i.e. it reproduces the live defect verbatim before the fix ships.
**Failed:** every row-level census I ran. `content_data ? 'contact_email'` said six rows were
damaged; the pages said otherwise in one direction (email present) and worse in another (dead
`tel:`). **A `page_components` row is a claim about a page, not a measurement of one** — the
LANDMINE says so, and I still had to be shown twice in two days.

---

## 2026-08-22 — scoping RFC_042's detector, and the census that answered it before it was built

The owner asked for a `bugs_open/` handoff scoping RFC_042 option (c), *"so we can measure it"*.
Filed as **`bugs_open/355_HANDOFF_2026-08-22_eight_of_nine_content_data_writers_cannot_be_observed_losing_keys.md`**.
**MISSTEP — filed as 354 and had to be renumbered to 355, which is the migration-497 collision of
last week repeating in a different sequence.** 353 was the highest across both dirs when the session
started, so 354 was written. It was `git mv`'d to 355 minutes after committing: another session had
committed `bugs_open/354_HANDOFF_2026-08-22_a_workflow_that_ends_at_its_error_terminal_is_recorded_COMPLETED_with_error_NULL.md`
at 10:17, ten minutes before this commit at 10:27. **The lesson is sharper than "allocate immediately
before naming the file", which is what was written down last time and was followed here.** Writing
the file does not claim the number and neither does `git add` — only a commit does, and only if you
win the race. On a tree this many sessions share, the check that would have caught it is to re-run
`ls bugs_open/ bugs_closed/ | grep -oE '^[0-9]+' | sort -n | tail -1` **immediately before the
commit**, not at the start of the work. Renamed rather than kept because the estate already carries
six numbers naming two unrelated cases, and knowingly adding a seventh is worse than inheriting one;
the move commit names BOTH paths (a `git mv` under a pathspec commit otherwise ships a copy) and was
verified with `git ls-tree`.

**The scoping turned into a measurement, and the measurement came back empty.** The archive trigger
on `page_components` has been recording pre-images with `slot_name` since 2026-08-09 — 6,210 rows —
and its `op` column splits them cleanly: `delete` (5,830) is the funnel, because `save_page_sections`
is DELETE+INSERT; `overwrite` (380) is every in-place writer, i.e. exactly the eight that PBP-039's
carry does not protect. Re-render is not a third case: it emits sections that `save_page_sections`
ingests and holds no `UPDATE page_components` of its own (its own comments at :30, :1091, :1149).

Pairing consecutive generations and diffing schema-declared non-LLM keys over the `overwrite`
population: **279 of 380 judgeable, 0 losses.**

**MISSTEP, caught before it was written down anywhere.** The first control I reached for was the LLM
arm of the same query — and it also returned 0. I nearly took two zeros as corroboration. They are
one zero counted twice: both arms share the joins, the pairing and the schema resolution, so a
defect in any of those silences both simultaneously. The control that actually discriminates had to
come from a population whose non-zero answer was established independently: the `op='delete'` run,
which returns **72 losses (static=24, renderer=48), dated 08-09/08-11/08-12 and none since** —
matching what RFC_042 §4.6 had measured by an entirely different method. Only then was the zero
readable. Generalised into 016b §9's existing demand-control entry as a strengthening: *a control
drawn from your own query shares your own blindness.* Not a `WRONG_CALLS` row — nothing false was
ever asserted — but it was one query away from being one.

**Second misstep, same shape, caught by asking why the number was small.** The first join resolved
the schema through `page_component_history.component_id`, and reported **92 of 380 judgeable**. That
FK is `ON DELETE SET NULL`, so every archived row whose page_components row was later deleted by a
regeneration has lost its pointer — **221 of 380, 58%**. A slot-keyed fallback recovers it to 279.
A census keyed on the FK alone returns a clean, plausible, three-times-too-small denominator.

**What keeps 355 open despite the zero** — four blind spots, each quantified in the file:
1. 101 pairs unjudgeable by any route;
2. **`application_name` cannot name the writer** — every app write carries the pgx connection default
   `app - <ip>:<port>`, hand SQL carries `psql`. So even a positive result could not be attributed,
   which is the exact question RFC_042 §4.3 says the detector exists to answer;
3. `pch_op_check` permits only `overwrite`/`delete` — **the trigger never fires on INSERT**, so a row
   born with incomplete content_data is invisible for ever (bound: 119 of 1,850 deployed rows carry
   no content_data — 77 NULL, 42 `{}`);
4. the window is 13 days; the older `save_page_sections_overwrite` archive reaches back to 2026-03-16
   but carries no `slot_name`, so it can only be paired at page granularity — the granularity that
   could not see 238.

**Writer census re-verified at HEAD.** RFC_042's nine stands. Three files newly matched the census
grep since it was written and all three are excluded on reading: `v3_site_actions.go` (writes
`sites.content_data`, a different table, plus `build_status`), `store_generated_component_action.go`
(`build_status` only, :1177), `create_tool_component_regenerate.go` (`rendered_html` only, :316).
Two paths destroy content_data by deleting the row rather than writing it —
`remove_duplicate_page_sections_action.go:297` and `internal/core-manager/admin/tool_admin_handlers.go:184`
— deliberately out of scope, recorded so the next census does not "discover" them.

**The finding-with-no-reader count is now two, not one.** `CONTENT_DATA_REGRESSION` 41 rows
(2026-08-08 → 08-21, 0 resolved) and `STRUCTURAL_KEY_CARRY_MISS` 28 rows (08-11 → 08-17, 0 resolved);
grep confirms neither code appears anywhere but its own write site and prose. So 355's candidate A3 —
ship the consumer in the same commit as the detector — is written as non-negotiable rather than as
advice. A third unread code would be the pattern, not an accident.

**The cheapest thing in the file is not the detector.** Candidate A1: `SET LOCAL application_name =
'action:<name>'` inside the existing transaction at each write site. No schema change, no config key,
no migration; the trigger already captures the column. It closes blind spot 2 permanently and makes
every future census attributable, including ones nobody has thought of. Worth shipping even if the
owner takes RFC_042 option (a).

Also this session: RFC_042 updated with an 08-22 header note and a re-read warning on its own
recommendation; the census + its control written into this lane's RUNBOOK with the two ⚠ traps;
016b §9's demand-control entry strengthened.

## 2026-08-22 (later) — OWNER RULED RFC_042 §6: option (c)

Ruled in chat, in the session that picked the lane up via `bugs_open/238` (which is closed; the
session verified the closure before reporting — 355's §2.3 census independently re-confirms it).
Recorded in RFC_042 §6 (decision block) and 355 (status). The ruling commissions the detector as 355
§4 scopes it, in that order:

- **A1** first — `application_name` self-attribution at the nine write sites, transaction-scoped.
- **A2 + A3 in one commit** — the per-key differ extending `writeContentDataRegressionLog`, plus its
  consumer. The two existing unread codes (`CONTENT_DATA_REGRESSION` 41, `STRUCTURAL_KEY_CARRY_MISS`
  28) are the anti-pattern the consumer half exists to end.
- **A4 NOT commissioned** — refusal waits for a measured population from A2/A3.

**NOT ruled:** the joint-with-RFC_008 half of §6. The owner named option (c) only; nothing in this
ruling decides the `rendered_html` seam, and RFC_008 stays open. Do not read this as licence to build
one seam over both columns.

Next in this session: read the nine write sites at HEAD (tx context, exec layer, what identifiers and
schema are in scope), design A1 against what the code actually does — noting ⚠ the sketch in 355 §4
(`SET LOCAL application_name = $1` via a bind parameter) is not executable as written: SET is a
utility statement and takes no parameters, so the real form is `SELECT set_config('application_name',
$1, true)` or an interpolated literal from a compile-time constant. Council-submit A1 as its own
coherent task, then A2+A3 as a second.

## 2026-08-22 (later still) — two sessions, one ruling; division of labour agreed; 358 filed

**Supersedes the "Next in this session" plan two entries up.** The owner gave the option-(c)
directive to BOTH sessions working this territory today. The owning-lane session was already
executing when this session's ruling record landed: mig `552` committed (`e7567d1fc`, closes the
archive's content-data-only-UPDATE blind spot — a hole 355 §3 had not listed), with
`cmd/content-loss-check` (A2+A3 one binary) in progress and A1 to follow. Coordinated by direct
session message rather than by collision; the owning lane keeps the build.

This session's contributions, all handed over by message:
- The set_config correction (SET takes no bind parameters; `_, _ =` would have eaten the failure).
- The three explorer censuses. The load-bearing one for A1: **only 2 of the 9 write sites run in a
  transaction** (apply_adoption_plan, admin HandleUpdateComponent) — the other 7 are bare-pool
  autocommit, where a transaction-scoped stamp is a silent no-op and a session-level SET is
  forbidden under pgbouncer `pool_mode=transaction`. Options weighed in the message (multi-statement
  simple-protocol implicit tx / short explicit tx / stamp-where-tx-exists), owning lane decides.
- The non-llm-key definition warning: the 355 census's `source not like 'llm%'` counts
  renderer/static as non-llm, and that is the class the 72-loss demand control lives in — a detector
  that excludes renderer/static zeroes its own control and refuses for ever.
- **`bugs_open/358` filed and committed (`a57b26696`)** — the unread-finding-codes CLASS file the
  lane asked for: 16 finding-shaped codes with no automated reader, `resolved` never set on any of
  45,426 rows ever, mig 466 retention deletes unresolved at 30d. The check's docs point at 358 as
  class owner; 358 excludes the two codes content-loss-check consumes. Verified at CONSTANT level,
  not just literal (the one real reader, `page_build_failure_guard.go:131`, binds a const — a
  literal-grep verdicts it unread; trap recorded in 358 §3/§8). 016b §10 row added.
